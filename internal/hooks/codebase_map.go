package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	sessionCache   = make(map[string]bool)
	sessionCacheMu sync.Mutex
)

// CodebaseMap is a sessionStart/beforeSubmitPrompt hook that generates a tree structure of the codebase.
// It caches per session_id to only run once per session.
func CodebaseMap(input HookInput, workDir string, maxDepth int, includePatterns []string) (HookResult, int) {
	// Check if stop_hook_active is set (from Stop event)
	var m map[string]interface{}
	if err := json.Unmarshal(input.ToolInput, &m); err == nil {
		if v, ok := m["stop_hook_active"].(bool); ok && v {
			return Allow(), 0
		}
	}

	sessionID := input.SessionID()
	if sessionID == "" {
		sessionID = "default"
	}

	// Check cache
	sessionCacheMu.Lock()
	if sessionCache[sessionID] {
		sessionCacheMu.Unlock()
		return Allow(), 0
	}
	sessionCacheMu.Unlock()

	cwd := input.Cwd()
	if cwd == "" {
		cwd = workDir
	}
	if cwd == "" {
		cwd = "."
	}

	projectRoot, err := filepath.Abs(cwd)
	if err != nil {
		return Allow(), 0
	}

	if _, err := os.Stat(projectRoot); err != nil {
		return Allow(), 0
	}

	if maxDepth == 0 {
		maxDepth = 3
	}

	if len(includePatterns) == 0 {
		includePatterns = []string{
			"nathan",
			"nathan/**",
			"tests",
			"tests/**",
			"docs",
			"docs/**",
			"scripts",
			"scripts/**",
			"*.py",
			"*.md",
			"*.toml",
			"*.yaml",
			"*.yml",
			"*.json",
			"*.sh",
			"pyproject.toml",
			"README.md",
			"docker-compose*.yml",
		}
	}

	tree := generateTree(projectRoot, maxDepth, includePatterns, projectRoot, 0, "")

	if tree == "" {
		return Allow(), 0
	}

	var msg strings.Builder
	msg.WriteString("\n")
	msg.WriteString(strings.Repeat("=", 60))
	msg.WriteString("\nCODEBASE STRUCTURE\n")
	msg.WriteString(strings.Repeat("=", 60))
	msg.WriteString("\n")
	msg.WriteString(tree)
	msg.WriteString("\n")
	msg.WriteString(strings.Repeat("=", 60))
	msg.WriteString("\n")

	// Mark session as cached
	sessionCacheMu.Lock()
	sessionCache[sessionID] = true
	sessionCacheMu.Unlock()

	return AllowMsg(msg.String()), 0
}

func generateTree(root string, maxDepth int, includePatterns []string, projectRoot string, currentDepth int, prefix string) string {
	if currentDepth >= maxDepth {
		return ""
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return ""
	}

	// Filter entries by include patterns
	var filtered []os.DirEntry
	for _, entry := range entries {
		entryPath := filepath.Join(root, entry.Name())
		if matchesIncludePattern(entryPath, includePatterns, projectRoot) {
			filtered = append(filtered, entry)
		}
	}

	if len(filtered) == 0 {
		return ""
	}

	// Sort: directories first, then by name (case-insensitive)
	sort.Slice(filtered, func(i, j int) bool {
		iIsDir := filtered[i].IsDir()
		jIsDir := filtered[j].IsDir()
		if iIsDir != jIsDir {
			return iIsDir
		}
		return strings.ToLower(filtered[i].Name()) < strings.ToLower(filtered[j].Name())
	})

	var lines []string
	for i, entry := range filtered {
		isLast := i == len(filtered)-1
		connector := "└── "
		if !isLast {
			connector = "├── "
		}
		lines = append(lines, prefix+connector+entry.Name())

		if entry.IsDir() {
			extension := "    "
			if !isLast {
				extension = "│   "
			}
			entryPath := filepath.Join(root, entry.Name())
			subtree := generateTree(entryPath, maxDepth, includePatterns, projectRoot, currentDepth+1, prefix+extension)
			if subtree != "" {
				lines = append(lines, subtree)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func matchesIncludePattern(entryPath string, includePatterns []string, projectRoot string) bool {
	relPath, err := filepath.Rel(projectRoot, entryPath)
	if err != nil {
		return false
	}

	relStr := filepath.ToSlash(relPath) // Normalize to forward slashes
	entryName := filepath.Base(entryPath)

	for _, pattern := range includePatterns {
		patternNormalized := filepath.ToSlash(pattern)

		// Exact name match
		if entryName == pattern {
			return true
		}

		// Match against relative path directly
		if match, _ := filepath.Match(patternNormalized, relStr); match {
			return true
		}

		// Match against name with glob
		if match, _ := filepath.Match(patternNormalized, entryName); match {
			return true
		}

		// Match if entry is under a directory pattern (e.g., "nathan/**" matches "nathan/file.py")
		if strings.HasSuffix(patternNormalized, "/**") {
			dirPattern := strings.TrimSuffix(patternNormalized, "/**")
			if relStr == dirPattern || strings.HasPrefix(relStr, dirPattern+"/") {
				return true
			}
		} else if strings.Contains(patternNormalized, "/") || strings.Contains(patternNormalized, "*") {
			// Pattern with path separators - check if entry path matches
			if match, _ := filepath.Match(patternNormalized, relStr); match {
				return true
			}
			// Try with **/ prefix
			if match, _ := filepath.Match("**/"+patternNormalized, relStr); match {
				return true
			}
		}
	}

	return false
}
