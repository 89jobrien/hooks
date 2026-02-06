package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

var blockedPaths = []string{
	"/etc",
	"/usr",
	"/bin",
	"/sbin",
	"/var",
	"/System",
	"/Library",
	"/Applications",
	"/Windows",
	"/Program Files",
	"/Program Files (x86)",
	"/ProgramData",
}

// PathValidation is a preToolUse hook that blocks writes outside allowed project directories.
// It prevents accidental modifications to system files or other projects.
func PathValidation(input HookInput, workDir string) (HookResult, int) {
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

	cwd := workDir
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	if cwd == "" {
		cwd = "."
	}

	allowed, reason := isPathAllowed(path, cwd)
	if !allowed {
		var msg strings.Builder
		msg.WriteString("Path validation failed: " + reason)
		msg.WriteString("\n  Attempted path: " + path)
		msg.WriteString("\n  Current directory: " + cwd)
		msg.WriteString("\n\nHint: Only write to project directories or ~/")
		msg.WriteString("\n  - Use relative paths within your project")
		msg.WriteString("\n  - System directories are protected")
		msg.WriteString("\n  - Paths outside project require explicit permission")

		return Deny(msg.String()), 2
	}

	return Allow(), 0
}

func isPathAllowed(filePath string, cwd string) (bool, string) {
	// Expand user home directory
	expandedPath := filePath
	if strings.HasPrefix(filePath, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			expandedPath = strings.Replace(filePath, "~", home, 1)
		}
	}

	// Resolve absolute path
	resolved, err := filepath.Abs(expandedPath)
	if err != nil {
		return false, "Invalid path: " + err.Error()
	}

	// Normalize separators for comparison
	resolved = filepath.Clean(resolved)
	resolvedLower := strings.ToLower(resolved)

	// Check blocked system paths (case-insensitive for Windows)
	// But allow /var/folders (macOS temp) and /var/tmp
	for _, blocked := range blockedPaths {
		blockedLower := strings.ToLower(blocked)
		// Skip /var check if it's a temp directory
		if blockedLower == "/var" {
			if strings.HasPrefix(resolvedLower, "/var/folders/") ||
				strings.HasPrefix(resolvedLower, "/var/tmp/") ||
				resolvedLower == "/var/tmp" {
				continue
			}
		}
		if strings.HasPrefix(resolvedLower, blockedLower) ||
			strings.HasPrefix(resolvedLower, strings.ToLower(filepath.VolumeName(resolved))+blockedLower) {
			// Ensure it's actually the directory, not a subdirectory we want to allow
			if len(resolved) > len(blocked) {
				nextChar := resolved[len(blocked)]
				if nextChar != filepath.Separator && nextChar != '/' && nextChar != '\\' {
					// Not actually a subdirectory, might be a prefix match
					continue
				}
			}
			return false, "System path blocked: " + blocked
		}
	}

	// Check always-allowed paths
	allowedPaths := getAllowedPaths()
	for _, allowed := range allowedPaths {
		allowedLower := strings.ToLower(allowed)
		if strings.HasPrefix(resolvedLower, allowedLower) {
			return true, "Allowed path"
		}
	}

	// Check for path traversal attacks (e.g., ../../../etc/passwd)
	if containsPathTraversal(filePath) {
		// Still check if it resolves to something safe
		cwdResolved, err := filepath.Abs(cwd)
		if err == nil {
			relPath, err := filepath.Rel(cwdResolved, resolved)
			if err == nil && !strings.HasPrefix(relPath, "..") {
				// Path resolves to within cwd, allow it
				return true, "Under cwd"
			}
		}
		// Path traversal detected and doesn't resolve to safe location
		return false, "Path traversal detected"
	}

	// Allow paths under current working directory
	cwdResolved, err := filepath.Abs(cwd)
	if err == nil {
		cwdResolved = filepath.Clean(cwdResolved)
		// Normalize both paths for comparison
		cwdNormalized := filepath.ToSlash(cwdResolved)
		resolvedNormalized := filepath.ToSlash(resolved)

		// Use case-insensitive comparison for Windows
		cwdLower := strings.ToLower(cwdNormalized)
		resolvedLower := strings.ToLower(resolvedNormalized)

		if strings.HasPrefix(resolvedLower, cwdLower) {
			// Ensure it's actually a subdirectory, not just a prefix match
			if len(resolvedLower) == len(cwdLower) {
				return true, "Under cwd"
			}
			// Check if next character is path separator
			nextChar := resolvedLower[len(cwdLower)]
			if nextChar == '/' || nextChar == '\\' {
				return true, "Under cwd"
			}
		}
	}

	// Allow paths under home directory
	home, err := os.UserHomeDir()
	if err == nil {
		home = filepath.Clean(home)
		if strings.HasPrefix(resolved, home) {
			// Warn if path is far from cwd
			if cwdResolved != "" && !strings.HasPrefix(cwdResolved, home) {
				// Path is in home but cwd is not - might be accidental
				// But allow it anyway, just log
				return true, "Under home (outside project)"
			}
			return true, "Under home"
		}
	}

	return false, "Path outside allowed directories"
}

func getAllowedPaths() []string {
	paths := []string{
		"/tmp",
		"/private/tmp",
	}

	// Add ~/.claude if it exists or can be created
	home, err := os.UserHomeDir()
	if err == nil {
		claudeDir := filepath.Join(home, ".claude")
		paths = append(paths, claudeDir)
	}

	// Add paths from environment variable
	if envPaths := os.Getenv("HOOK_PATH_VALIDATION_ALLOWED"); envPaths != "" {
		for _, p := range strings.Split(envPaths, ":") {
			p = strings.TrimSpace(p)
			if p != "" {
				paths = append(paths, p)
			}
		}
	}

	return paths
}

func containsPathTraversal(path string) bool {
	return strings.Contains(path, "..") || strings.Contains(path, "../") || strings.Contains(path, "..\\")
}
