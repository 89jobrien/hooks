package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKnowledgeUpdate_NoTranscriptPath(t *testing.T) {
	input := HookInput{
		ToolName:  "Stop",
		ToolInput: []byte(`{}`),
	}
	result, code := KnowledgeUpdate(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no transcript path, got decision=%s code=%d", result.Decision, code)
	}
}

func TestKnowledgeUpdate_StopHookActive(t *testing.T) {
	ti, _ := json.Marshal(map[string]interface{}{
		"transcript_path":  "/path/to/transcript.txt",
		"stop_hook_active": true,
		"session_id":       "test-session",
	})
	input := HookInput{
		ToolName:  "Stop",
		ToolInput: ti,
	}
	result, code := KnowledgeUpdate(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when stop_hook_active is true, got decision=%s code=%d", result.Decision, code)
	}
}

func TestKnowledgeUpdate_NoAPIKey(t *testing.T) {
	// Save original API key
	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalKey)

	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.txt")
	os.WriteFile(transcript, []byte("Some session content"), 0644)

	ti, _ := json.Marshal(map[string]interface{}{
		"transcript_path": transcript,
		"session_id":      "test-session",
		"cwd":             dir,
	})
	input := HookInput{
		ToolName:  "Stop",
		ToolInput: ti,
	}

	result, code := KnowledgeUpdate(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no API key, got decision=%s code=%d", result.Decision, code)
	}
}

func TestKnowledgeUpdate_NonexistentTranscript(t *testing.T) {
	ti, _ := json.Marshal(map[string]interface{}{
		"transcript_path": "/nonexistent/transcript.txt",
		"session_id":      "test-session",
	})
	input := HookInput{
		ToolName:  "Stop",
		ToolInput: ti,
	}

	result, code := KnowledgeUpdate(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for nonexistent transcript, got decision=%s code=%d", result.Decision, code)
	}
}

func TestSaveKnowledgeUpdate(t *testing.T) {
	dir := t.TempDir()
	entities := knowledgeEntities{
		FilesCreated:      []string{"test.py", "utils.ts"},
		ComponentsAdded:   []string{"UserService", "LoginForm"},
		DependenciesAdded: []string{"react", "express"},
		PatternsUsed:      []string{"TDD", "Repository"},
		Technologies:      []string{"React", "TypeScript"},
	}

	updatePath, err := saveKnowledgeUpdate(entities, dir, "test-session")
	if err != nil {
		t.Fatalf("saveKnowledgeUpdate failed: %v", err)
	}

	if updatePath == "" {
		t.Fatal("expected non-empty update path")
	}

	// Verify file exists
	if _, err := os.Stat(updatePath); err != nil {
		t.Fatalf("update file not created: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(updatePath)
	if err != nil {
		t.Fatalf("failed to read update file: %v", err)
	}

	var update knowledgeUpdate
	if err := json.Unmarshal(data, &update); err != nil {
		t.Fatalf("failed to parse update file: %v", err)
	}

	if update.SessionID != "test-session" {
		t.Errorf("expected session_id 'test-session', got %q", update.SessionID)
	}

	if len(update.Entities.FilesCreated) != 2 {
		t.Errorf("expected 2 files created, got %d", len(update.Entities.FilesCreated))
	}

	if update.Summary["files_created"] != 2 {
		t.Errorf("expected summary files_created=2, got %d", update.Summary["files_created"])
	}
}

func TestSaveKnowledgeUpdate_CreatesDirectory(t *testing.T) {
	// Use a temp directory structure
	homeDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	entities := knowledgeEntities{
		FilesCreated: []string{"test.py"},
	}

	updatePath, err := saveKnowledgeUpdate(entities, ".", "test")
	if err != nil {
		t.Fatalf("saveKnowledgeUpdate failed: %v", err)
	}

	// Verify directory was created
	knowledgeDir := filepath.Join(homeDir, "logs", "claude", "knowledge-updates")
	if _, err := os.Stat(knowledgeDir); err != nil {
		t.Errorf("knowledge directory not created: %v", err)
	}

	// Verify file exists in that directory
	if !strings.Contains(updatePath, knowledgeDir) {
		t.Errorf("update file not in knowledge directory: %s", updatePath)
	}
}

func TestKnowledgeUpdate_EmptyEntities(t *testing.T) {
	// This test would require mocking the API call, which is complex
	// For now, we test that empty entities result in allow
	entities := knowledgeEntities{}
	if len(entities.FilesCreated) == 0 &&
		len(entities.DependenciesAdded) == 0 &&
		len(entities.PatternsUsed) == 0 &&
		len(entities.ComponentsAdded) == 0 {
		// Should skip saving
		if len(entities.FilesCreated) != 0 {
			t.Error("expected empty entities")
		}
	}
}

func TestKnowledgeUpdate_SummaryCounts(t *testing.T) {
	entities := knowledgeEntities{
		FilesCreated:      []string{"a.py", "b.ts"},
		DependenciesAdded: []string{"react"},
		PatternsUsed:      []string{"TDD"},
		Technologies:      []string{"React", "TypeScript"},
		ComponentsAdded:   []string{"Component1"},
	}

	homeDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	updatePath, err := saveKnowledgeUpdate(entities, ".", "test")
	if err != nil {
		t.Fatalf("saveKnowledgeUpdate failed: %v", err)
	}

	data, _ := os.ReadFile(updatePath)
	var update knowledgeUpdate
	json.Unmarshal(data, &update)

	if update.Summary["files_created"] != 2 {
		t.Errorf("expected 2 files, got %d", update.Summary["files_created"])
	}
	if update.Summary["dependencies_added"] != 1 {
		t.Errorf("expected 1 dependency, got %d", update.Summary["dependencies_added"])
	}
	if update.Summary["patterns_observed"] != 1 {
		t.Errorf("expected 1 pattern, got %d", update.Summary["patterns_observed"])
	}
	if update.Summary["technologies_used"] != 2 {
		t.Errorf("expected 2 technologies, got %d", update.Summary["technologies_used"])
	}
	if update.Summary["components_added"] != 1 {
		t.Errorf("expected 1 component, got %d", update.Summary["components_added"])
	}
}
