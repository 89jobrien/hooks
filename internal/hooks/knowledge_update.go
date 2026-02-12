package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	openAIAPIURL = "https://api.openai.com/v1/chat/completions"
	maxTokens    = 1000
	timeout      = 30 * time.Second
)

type knowledgeEntities struct {
	FilesCreated      []string `json:"files_created"`
	ComponentsAdded   []string `json:"components_added"`
	DependenciesAdded []string `json:"dependencies_added"`
	PatternsUsed      []string `json:"patterns_used"`
	Technologies      []string `json:"technologies"`
	Concepts          []string `json:"concepts,omitempty"`
}

type knowledgeUpdate struct {
	Timestamp string            `json:"timestamp"`
	SessionID string            `json:"session_id"`
	Cwd       string            `json:"cwd"`
	Entities  knowledgeEntities `json:"entities"`
	Summary   map[string]int    `json:"summary"`
}

// KnowledgeUpdate is a stop hook that extracts knowledge entities from session transcript
// using OpenAI API and saves them to a knowledge graph.
func KnowledgeUpdate(input HookInput, workDir string) (HookResult, int) {
	transcriptPath := input.TranscriptPath()
	if transcriptPath == "" {
		return NoOp(), 0
	}

	// Check if stop_hook_active is set
	if input.StopHookActive() {
		return NoOp(), 0
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return NoOp(), 0 // Fail silently if no API key
	}

	sessionID := input.SessionID()
	if sessionID == "" {
		sessionID = "unknown"
	}

	cwd := input.Cwd()
	if cwd == "" {
		cwd = workDir
	}
	if cwd == "" {
		cwd = "."
	}

	// Read transcript
	transcript, err := os.ReadFile(transcriptPath)
	if err != nil {
		return NoOp(), 0 // Fail silently
	}

	// Call OpenAI to extract entities
	entities, err := extractEntitiesWithLLM(string(transcript), apiKey)
	if err != nil {
		return NoOp(), 0 // Fail silently
	}

	// Skip if nothing interesting found
	if len(entities.FilesCreated) == 0 &&
		len(entities.DependenciesAdded) == 0 &&
		len(entities.PatternsUsed) == 0 &&
		len(entities.ComponentsAdded) == 0 &&
		len(entities.Technologies) == 0 &&
		len(entities.Concepts) == 0 {
		return NoOp(), 0
	}

	// Save knowledge update
	updatePath, err := saveKnowledgeUpdate(entities, cwd, sessionID)
	if err != nil {
		return NoOp(), 0 // Fail silently
	}

	// Build summary message
	var msg strings.Builder
	msg.WriteString("\n")
	msg.WriteString(strings.Repeat("=", 50))
	msg.WriteString("\n[Success] Knowledge entities extracted:\n")

	if len(entities.FilesCreated) > 0 {
		msg.WriteString(fmt.Sprintf("  Files created: %d\n", len(entities.FilesCreated)))
	}

	if len(entities.DependenciesAdded) > 0 {
		deps := entities.DependenciesAdded
		if len(deps) > 5 {
			deps = deps[:5]
		}
		msg.WriteString(fmt.Sprintf("  Dependencies: %s\n", strings.Join(deps, ", ")))
	}

	if len(entities.PatternsUsed) > 0 {
		msg.WriteString(fmt.Sprintf("  Patterns: %s\n", strings.Join(entities.PatternsUsed, ", ")))
	}

	if len(entities.Technologies) > 0 {
		msg.WriteString(fmt.Sprintf("  Technologies: %s\n", strings.Join(entities.Technologies, ", ")))
	}

	if len(entities.ComponentsAdded) > 0 {
		msg.WriteString(fmt.Sprintf("  Components: %d\n", len(entities.ComponentsAdded)))
	}

	msg.WriteString(fmt.Sprintf("\n  Saved to: %s\n", updatePath))
	msg.WriteString(strings.Repeat("=", 50))
	msg.WriteString("\n")

	return NoOpMsg(msg.String()), 0
}

func extractEntitiesWithLLM(transcript string, apiKey string) (knowledgeEntities, error) {
	var entities knowledgeEntities

	// Truncate transcript if too long (keep last 8000 chars for context)
	if len(transcript) > 8000 {
		transcript = "..." + transcript[len(transcript)-8000:]
	}

	prompt := `Analyze the following development session transcript and extract knowledge entities in JSON format.

Return a JSON object with these fields:
- files_created: array of file paths that were created
- components_added: array of functions, classes, or components added (e.g., "UserService", "LoginForm")
- dependencies_added: array of packages/dependencies installed (e.g., "react", "express", "pandas")
- patterns_used: array of design patterns observed (e.g., "TDD", "Repository", "Factory", "Observer", "MVC", "REST")
- technologies: array of technologies/frameworks used (e.g., "React", "FastAPI", "PostgreSQL", "Docker")
- concepts: array of key concepts or topics discussed

Only include entities that are actually mentioned or created in the transcript. Return only valid JSON, no markdown formatting.

Transcript:
` + transcript

	// Get model from env or use default
	model := os.Getenv("HOOK_KNOWLEDGE_UPDATE_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	// Prepare OpenAI API request
	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a knowledge extraction assistant. Extract structured knowledge from development session transcripts. Always return valid JSON only.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  maxTokens,
		"temperature": 0.3,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return entities, err
	}

	// Make API call
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", openAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return entities, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return entities, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return entities, fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return entities, err
	}

	if len(apiResponse.Choices) == 0 {
		return entities, fmt.Errorf("no choices in API response")
	}

	content := apiResponse.Choices[0].Message.Content

	// Parse JSON response
	var response struct {
		FilesCreated      []string `json:"files_created"`
		ComponentsAdded   []string `json:"components_added"`
		DependenciesAdded []string `json:"dependencies_added"`
		PatternsUsed      []string `json:"patterns_used"`
		Technologies      []string `json:"technologies"`
		Concepts          []string `json:"concepts"`
	}

	// Clean up content (remove markdown code blocks if present)
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return entities, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	entities = knowledgeEntities{
		FilesCreated:      response.FilesCreated,
		ComponentsAdded:   response.ComponentsAdded,
		DependenciesAdded: response.DependenciesAdded,
		PatternsUsed:      response.PatternsUsed,
		Technologies:      response.Technologies,
		Concepts:          response.Concepts,
	}

	return entities, nil
}

// repoNameFromCwd returns a directory-safe name for the repo (git top-level base or cwd base).
func repoNameFromCwd(cwd string) string {
	dir := cwd
	if dir == "" || dir == "." {
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}
	if dir == "" || dir == "." {
		return "default"
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	if out, err := cmd.Output(); err == nil {
		if root := strings.TrimSpace(string(out)); root != "" {
			name := filepath.Base(root)
			if name != "" && name != "." && name != ".." {
				return name
			}
		}
	}
	name := filepath.Base(filepath.Clean(dir))
	if name == "" || name == "." || name == ".." {
		return "default"
	}
	return name
}

func saveKnowledgeUpdate(entities knowledgeEntities, cwd, sessionID string) (string, error) {
	knowledgeDir := os.Getenv("HOOK_KNOWLEDGE_DIR")
	if knowledgeDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		knowledgeDir = filepath.Join(homeDir, "logs", "claude", "knowledge-updates")
	}
	repoName := repoNameFromCwd(cwd)
	repoDir := filepath.Join(knowledgeDir, repoName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now()
	updateFile := filepath.Join(repoDir, fmt.Sprintf("knowledge_%s.json", timestamp.Format("20060102_150405")))

	update := knowledgeUpdate{
		Timestamp: timestamp.Format(time.RFC3339),
		SessionID: sessionID,
		Cwd:       cwd,
		Entities:  entities,
		Summary: map[string]int{
			"files_created":      len(entities.FilesCreated),
			"dependencies_added": len(entities.DependenciesAdded),
			"patterns_observed":  len(entities.PatternsUsed),
			"technologies_used":  len(entities.Technologies),
			"components_added":   len(entities.ComponentsAdded),
		},
	}

	jsonData, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(updateFile, jsonData, 0644); err != nil {
		return "", err
	}

	return updateFile, nil
}
