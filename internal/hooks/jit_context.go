package hooks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxFileSize      = 100_000 // 100KB
	maxFilesToReturn = 5
	maxGrepHits      = 10
	maxHeadLines     = 30
	maxTailLines     = 20
	maxKeywordSearches = 2
)

// JITContext is a beforeSubmitPrompt hook that analyzes user prompts and pre-loads
// relevant file context using pattern matching, path extraction, and keyword search.
func JITContext(input HookInput, workDir string) (HookResult, int) {
	prompt := input.Prompt()
	if prompt == "" || len(prompt) < 10 {
		return Allow(), 0
	}

	if workDir == "" {
		cwd, _ := os.Getwd()
		workDir = cwd
	}

	// Extract patterns, paths, and keywords from prompt
	patterns := extractPatterns(prompt)
	paths := extractPaths(prompt)
	keywords := extractKeywords(prompt)

	var matchedFiles []fileContent
	var grepResults []grepResult

	// Collect files matching patterns
	matchedFiles = append(matchedFiles, collectPatternMatches(workDir, patterns)...)
	matchedFiles = append(matchedFiles, collectPathMatches(workDir, paths)...)

	// If no files matched, try keyword search
	if len(matchedFiles) == 0 && len(keywords) > 0 {
		grepResults = collectKeywordMatches(workDir, keywords)
	}

	// Format context
	context := formatContext(matchedFiles, grepResults)
	if context == "" {
		return Allow(), 0
	}

	return AllowMsg(context), 0
}

type fileContent struct {
	path    string
	content string
}

type grepResult struct {
	path string
	hits []grepHit
}

type grepHit struct {
	lineNum int
	line    string
}

func extractPatterns(prompt string) []string {
	var patterns []string
	seen := make(map[string]bool)

	// Match explicit glob patterns like *.py, **/*.ts, src/**/*
	globRe := regexp.MustCompile(`[*?[\]{}]+[.\w/]*|[\w./]+[*?[\]{}]+[\w./]*`)
	matches := globRe.FindAllString(prompt, -1)
	for _, m := range matches {
		m = strings.TrimSpace(m)
		if m != "" && !seen[m] {
			patterns = append(patterns, m)
			seen[m] = true
		}
	}

	// Match file extensions mentioned
	extRe := regexp.MustCompile(`\.(py|ts|js|tsx|jsx|md|json|yaml|yml|toml|sh|sql|go|rs|java|rb)\b`)
	extMatches := extRe.FindAllString(prompt, -1)
	for _, ext := range extMatches {
		pattern := "*" + ext
		if !seen[pattern] {
			patterns = append(patterns, pattern)
			seen[pattern] = true
		}
	}

	return patterns
}

func extractPaths(prompt string) []string {
	var paths []string
	seen := make(map[string]bool)

	// Match paths like src/, ./config, /path/to/file.py, src/utils.ts
	pathRe := regexp.MustCompile(`(?:\.?/)?(?:[\w.-]+/)+[\w.-]*(?:\.[\w]+)?|[\w.-]+\.(?:py|ts|js|md|json|yaml|yml|go|rs|java|rb)`)
	matches := pathRe.FindAllString(prompt, -1)
	for _, m := range matches {
		m = strings.TrimSpace(m)
		// Skip URLs and very short matches
		if m != "" && !strings.HasPrefix(m, "http") && len(m) > 2 && !seen[m] {
			paths = append(paths, m)
			seen[m] = true
		}
	}

	return paths
}

func extractKeywords(prompt string) []string {
	var keywords []string
	seen := make(map[string]bool)

	// Look for quoted strings
	quotedRe := regexp.MustCompile(`["']([^"']+)["']`)
	quotedMatches := quotedRe.FindAllStringSubmatch(prompt, -1)
	for _, match := range quotedMatches {
		if len(match) > 1 {
			kw := strings.TrimSpace(match[1])
			if len(kw) >= 3 && !seen[kw] {
				keywords = append(keywords, kw)
				seen[kw] = true
			}
		}
	}

	// Look for function/class names (CamelCase or snake_case)
	identifierRe := regexp.MustCompile(`\b[A-Z][a-zA-Z]+\b|\b[a-z]+_[a-z_]+\b`)
	identMatches := identifierRe.FindAllString(prompt, -1)
	for _, ident := range identMatches {
		if len(ident) >= 3 && !seen[ident] {
			keywords = append(keywords, ident)
			seen[ident] = true
		}
	}

	// Limit to avoid too many searches
	if len(keywords) > maxKeywordSearches {
		keywords = keywords[:maxKeywordSearches]
	}

	return keywords
}

func collectPatternMatches(root string, patterns []string) []fileContent {
	var results []fileContent
	if len(patterns) == 0 {
		return results
	}

	// Convert glob patterns to filepath.Match patterns
	for _, pattern := range patterns {
		if len(results) >= maxFilesToReturn {
			break
		}

		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if len(results) >= maxFilesToReturn {
				return filepath.SkipAll
			}

			if d.IsDir() {
				return nil
			}

			// Check if file matches pattern
			relPath, _ := filepath.Rel(root, path)
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if !matched {
				matched, _ = filepath.Match(pattern, relPath)
			}
			if !matched {
				matched, _ = filepath.Match(pattern, filepath.ToSlash(relPath))
			}

			if matched {
				info, err := d.Info()
				if err != nil || info.Size() >= maxFileSize {
					return nil
				}

				content := headTail(path)
				if content != "" {
					results = append(results, fileContent{
						path:    relPath,
						content: content,
					})
				}
			}

			return nil
		})

		if err != nil {
			continue
		}
	}

	return results
}

func collectPathMatches(root string, paths []string) []fileContent {
	var results []fileContent

	for _, p := range paths {
		if len(results) >= maxFilesToReturn {
			break
		}

		// Try as relative path first
		fullPath := filepath.Join(root, p)
		if _, err := os.Stat(fullPath); err != nil {
			continue
		}

		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() || info.Size() >= maxFileSize {
			continue
		}

		content := headTail(fullPath)
		if content != "" {
			results = append(results, fileContent{
				path:    p,
				content: content,
			})
		}
	}

	return results
}

func collectKeywordMatches(root string, keywords []string) []grepResult {
	var results []grepResult
	if len(keywords) == 0 {
		return results
	}

	// Search in common source files
	searchPatterns := []string{"*.py", "*.ts", "*.js", "*.go", "*.rs", "*.md"}
	var searchFiles []string

	for _, pattern := range searchPatterns {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			if len(searchFiles) >= 50 {
				return filepath.SkipAll
			}

			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				searchFiles = append(searchFiles, path)
			}
			return nil
		})
	}

	for _, keyword := range keywords {
		if len(results) >= maxFilesToReturn {
			break
		}

		for _, filePath := range searchFiles {
			if len(results) >= maxFilesToReturn {
				break
			}

			hits := grepFile(filePath, keyword)
			if len(hits) > 0 {
				relPath, _ := filepath.Rel(root, filePath)
				results = append(results, grepResult{
					path: relPath,
					hits: hits,
				})
			}
		}
	}

	return results
}

func headTail(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	total := len(lines)
	if total <= maxHeadLines+maxTailLines {
		return strings.Join(lines, "\n")
	}

	head := strings.Join(lines[:maxHeadLines], "\n")
	tail := strings.Join(lines[total-maxTailLines:], "\n")
	omitted := total - maxHeadLines - maxTailLines

	return fmt.Sprintf("%s\n\n... [%d lines omitted] ...\n\n%s", head, omitted, tail)
}

func grepFile(filePath, needle string) []grepHit {
	var hits []grepHit
	file, err := os.Open(filePath)
	if err != nil {
		return hits
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	needleLower := strings.ToLower(needle)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), needleLower) {
			hits = append(hits, grepHit{
				lineNum: lineNum,
				line:    strings.TrimRight(line, "\n"),
			})
			if len(hits) >= maxGrepHits {
				break
			}
		}
	}

	return hits
}

func formatContext(files []fileContent, grepResults []grepResult) string {
	var parts []string

	if len(files) > 0 {
		parts = append(parts, "# JIT Context: Relevant Files\n")
		for i, fc := range files {
			if i >= maxFilesToReturn {
				break
			}
			parts = append(parts, fmt.Sprintf("## %s\n```\n%s\n```\n", fc.path, fc.content))
		}
	}

	if len(grepResults) > 0 {
		parts = append(parts, "# JIT Context: Keyword Matches\n")
		for i, gr := range grepResults {
			if i >= maxFilesToReturn {
				break
			}
			parts = append(parts, fmt.Sprintf("## %s", gr.path))
			for _, hit := range gr.hits {
				linePreview := hit.line
				if len(linePreview) > 100 {
					linePreview = linePreview[:100] + "..."
				}
				parts = append(parts, fmt.Sprintf("  L%d: %s", hit.lineNum, linePreview))
			}
			parts = append(parts, "")
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n")
}
