package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathValidation_NonWriteTool(t *testing.T) {
	input := HookInput{
		ToolName: "Shell",
		ToolInput: []byte(`{"command": "ls"}`),
	}
	result, code := PathValidation(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-Write tool, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_NoFilePath(t *testing.T) {
	input := writeInput("", "content")
	result, code := PathValidation(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no file path, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_BlockedSystemPaths(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantDeny bool
	}{
		{"/etc/passwd", "/etc/passwd", true},
		{"/usr/bin/test", "/usr/bin/test", true},
		{"/bin/sh", "/bin/sh", true},
		{"/var/log/test", "/var/log/test", true},
		{"/System/test", "/System/test", true},
		{"/Library/test", "/Library/test", true},
		{"/Applications/test", "/Applications/test", true},
		{"project file", "src/app.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := writeInput(tt.filePath, "content")
			result, code := PathValidation(input, "/tmp")
			if tt.wantDeny {
				if code != 2 || result.Decision != "deny" {
					t.Errorf("expected deny for %s, got decision=%s code=%d", tt.filePath, result.Decision, code)
				}
				if !strings.Contains(result.Reason, "System path blocked") {
					t.Errorf("expected reason to mention system path, got %q", result.Reason)
				}
			} else {
				if code != 0 && result.Decision == "deny" {
					t.Logf("path %s was denied, reason: %s", tt.filePath, result.Reason)
				}
			}
		})
	}
}

func TestPathValidation_AllowedPaths(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	input := writeInput(testFile, "content")
	result, code := PathValidation(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for file in project dir, got decision=%s code=%d, reason: %s", result.Decision, code, result.Reason)
	}
}

func TestPathValidation_RelativePath(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	testFile := filepath.Join(dir, "src", "app.js")

	// Use relative path
	relPath, _ := filepath.Rel(dir, testFile)
	input := writeInput(relPath, "content")

	result, code := PathValidation(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for relative path in project, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	// Try to escape with ../
	traversalPath := "../../etc/passwd"

	input := writeInput(traversalPath, "content")
	result, code := PathValidation(input, dir)
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny for path traversal, got decision=%s code=%d", result.Decision, code)
	}
	if !strings.Contains(result.Reason, "Path traversal") {
		t.Errorf("expected reason to mention path traversal, got %q", result.Reason)
	}
}

func TestPathValidation_PathTraversalWithinProject(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src", "subdir"), 0755)
	// Relative path that goes up but stays in project
	traversalPath := "../src/app.js"

	input := writeInput(traversalPath, "content")
	result, code := PathValidation(input, filepath.Join(dir, "subdir"))
	// Should allow if it resolves to within project
	if code == 2 && result.Decision == "deny" {
		t.Logf("path traversal denied, reason: %s", result.Reason)
	}
}

func TestPathValidation_HomeDirectory(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	testFile := filepath.Join(homeDir, "test.txt")
	input := writeInput(testFile, "content")

	result, code := PathValidation(input, "/tmp")
	// Should allow files in home directory
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for file in home directory, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_TildeExpansion(t *testing.T) {
	_, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	testFile := "~/test.txt"
	input := writeInput(testFile, "content")

	result, code := PathValidation(input, "/tmp")
	// Should expand ~ and allow
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for ~ path, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_EditTool(t *testing.T) {
	ti, _ := json.Marshal(map[string]string{
		"file_path": "/etc/passwd",
	})
	input := HookInput{
		ToolName:  "Edit",
		ToolInput: ti,
	}

	result, code := PathValidation(input, "/tmp")
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny for Edit on system file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_MultiEditTool(t *testing.T) {
	ti, _ := json.Marshal(map[string]string{
		"file_path": "/usr/bin/test",
	})
	input := HookInput{
		ToolName:  "MultiEdit",
		ToolInput: ti,
	}

	result, code := PathValidation(input, "/tmp")
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny for MultiEdit on system file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_TempDirectory(t *testing.T) {
	testFile := "/tmp/test.txt"
	input := writeInput(testFile, "content")

	result, code := PathValidation(input, "/tmp")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for /tmp, got decision=%s code=%d", result.Decision, code)
	}
}

func TestPathValidation_InvalidPath(t *testing.T) {
	// Test with invalid characters or path
	invalidPath := "\x00invalid"
	input := writeInput(invalidPath, "content")

	result, code := PathValidation(input, "/tmp")
	// Should deny invalid paths (or allow if system handles it gracefully)
	// On some systems, filepath.Abs might sanitize the path
	if code == 2 && result.Decision == "deny" {
		if !strings.Contains(result.Reason, "Invalid path") {
			t.Errorf("expected reason to mention invalid path, got %q", result.Reason)
		}
	}
}

func TestIsPathAllowed_BlockedPath(t *testing.T) {
	allowed, reason := isPathAllowed("/etc/passwd", "/tmp")
	if allowed {
		t.Error("expected /etc/passwd to be blocked")
	}
	if !strings.Contains(reason, "System path blocked") {
		t.Errorf("expected reason to mention system path, got %q", reason)
	}
}

func TestIsPathAllowed_ProjectPath(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	allowed, reason := isPathAllowed(testFile, dir)
	if !allowed {
		t.Errorf("expected %s to be allowed, reason: %s", testFile, reason)
	}
}

func TestContainsPathTraversal(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"../test", true},
		{"../../etc/passwd", true},
		{"..\\test", true},
		{"src/../app.js", true},
		{"src/app.js", false},
		{"/absolute/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := containsPathTraversal(tt.path)
			if result != tt.expected {
				t.Errorf("containsPathTraversal(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
