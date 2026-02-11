package hooks

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
)

var readonlyPatterns = []*regexp.Regexp{
	// Lock files
	regexp.MustCompile(`package-lock\.json$`),
	regexp.MustCompile(`yarn\.lock$`),
	regexp.MustCompile(`pnpm-lock\.yaml$`),
	regexp.MustCompile(`poetry\.lock$`),
	regexp.MustCompile(`Cargo\.lock$`),
	regexp.MustCompile(`uv\.lock$`),
	regexp.MustCompile(`Gemfile\.lock$`),
	regexp.MustCompile(`composer\.lock$`),
	// Generated files
	regexp.MustCompile(`\.min\.js$`),
	regexp.MustCompile(`\.min\.css$`),
	regexp.MustCompile(`\.map$`),
	regexp.MustCompile(`\.d\.ts$`), // TypeScript declarations (usually generated)
	// Vendor/dependency directories
	regexp.MustCompile(`(^|[/\\])node_modules(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])vendor(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])__pycache__(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])\.git(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])dist(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])build(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])\.next(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])\.nuxt(/|\\|$)`),
	// IDE/editor files
	regexp.MustCompile(`(^|[/\\])\.idea(/|\\|$)`),
	regexp.MustCompile(`(^|[/\\])\.vscode[/\\]settings\.json$`), // settings.json specifically
}

var overrideAllowed = []*regexp.Regexp{
	regexp.MustCompile(`\.vscode/launch\.json$`), // Debug configs are ok
	regexp.MustCompile(`\.vscode/tasks\.json$`),  // Task configs are ok
}

// ReadonlyGuard is a preToolUse hook that protects lock files, generated files,
// and vendor directories from modification.
func ReadonlyGuard(input HookInput) (HookResult, int) {
	if input.ToolName != "Write" && input.ToolName != "Edit" && input.ToolName != "MultiEdit" {
		return Allow(), 0
	}

	// Get path - Edit/MultiEdit use "file_path", Write uses "path"
	path := input.Path()
	if path == "" {
		// Try file_path for Edit/MultiEdit
		var m map[string]interface{}
		if err := json.Unmarshal(input.ToolInput, &m); err == nil {
			if v, ok := m["file_path"].(string); ok {
				path = v
			}
		}
		if path == "" {
			return Allow(), 0
		}
	}

	// Normalize path separators
	normalizedPath := filepath.ToSlash(path)

	// Check override patterns first
	for _, pattern := range overrideAllowed {
		if pattern.MatchString(normalizedPath) {
			return Allow(), 0
		}
	}

	// Check readonly patterns
	for _, pattern := range readonlyPatterns {
		if pattern.MatchString(normalizedPath) {
			var reason strings.Builder
			reason.WriteString("Readonly file protection triggered")
			reason.WriteString("\n  File: " + path)
			reason.WriteString("\n  Pattern: " + pattern.String())
			reason.WriteString("\n\nHint: This file is auto-generated or managed by tools.")
			reason.WriteString("\n  - Lock files: Use package manager commands instead")
			reason.WriteString("\n  - Generated files: Modify source files instead")
			reason.WriteString("\n  - Vendor dirs: Don't modify dependencies directly")

			return Deny(reason.String()), 2
		}
	}

	return Allow(), 0
}
