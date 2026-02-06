package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodebaseMap_NonSessionEvent(t *testing.T) {
	input := HookInput{
		ToolName: "Write",
		ToolInput: []byte(`{"path": "test.ts", "contents": "const x: any = 5;"}`),
	}
	result, code := CodebaseMap(input, ".", 3, nil)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-session event, got decision=%s code=%d", result.Decision, code)
	}
}

func TestCodebaseMap_StopHookActive(t *testing.T) {
	ti, _ := json.Marshal(map[string]interface{}{
		"stop_hook_active": true,
		"session_id":       "test-session",
		"cwd":              ".",
	})
	input := HookInput{
		ToolName:  "SessionStart",
		ToolInput: ti,
	}
	result, code := CodebaseMap(input, ".", 3, nil)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when stop_hook_active is true, got decision=%s code=%d", result.Decision, code)
	}
}

func TestCodebaseMap_SessionCaching(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)

	ti, _ := json.Marshal(map[string]interface{}{
		"session_id": "test-session-1",
		"cwd":        dir,
	})
	input := HookInput{
		ToolName:  "SessionStart",
		ToolInput: ti,
	}

	// First call should generate tree
	result1, code1 := CodebaseMap(input, dir, 3, []string{"README.md", "src/**"})
	if code1 != 0 || result1.Decision != "allow" {
		t.Fatalf("expected allow on first call, got decision=%s code=%d", result1.Decision, code1)
	}
	if result1.Message == "" {
		t.Error("expected message with tree structure on first call")
	}
	if !strings.Contains(result1.Message, "CODEBASE STRUCTURE") {
		t.Error("expected message to contain 'CODEBASE STRUCTURE'")
	}

	// Second call with same session_id should be cached (no message)
	result2, code2 := CodebaseMap(input, dir, 3, []string{"README.md", "src/**"})
	if code2 != 0 || result2.Decision != "allow" {
		t.Fatalf("expected allow on second call, got decision=%s code=%d", result2.Decision, code2)
	}
	if result2.Message != "" {
		t.Errorf("expected empty message on cached call, got %q", result2.Message)
	}

	// Different session_id should generate tree again
	ti2, _ := json.Marshal(map[string]interface{}{
		"session_id": "test-session-2",
		"cwd":        dir,
	})
	input2 := HookInput{
		ToolName:  "SessionStart",
		ToolInput: ti2,
	}
	result3, code3 := CodebaseMap(input2, dir, 3, []string{"README.md", "src/**"})
	if code3 != 0 || result3.Decision != "allow" {
		t.Fatalf("expected allow on third call with different session, got decision=%s code=%d", result3.Decision, code3)
	}
	if result3.Message == "" {
		t.Error("expected message with tree structure for different session")
	}
}

func TestCodebaseMap_GeneratesTree(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src", "app"), 0755)
	os.MkdirAll(filepath.Join(dir, "tests"), 0755)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "main.py"), []byte("print('hello')"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "app", "util.py"), []byte("def util(): pass"), 0644)
	os.WriteFile(filepath.Join(dir, "tests", "test_main.py"), []byte("def test(): pass"), 0644)

	ti, _ := json.Marshal(map[string]interface{}{
		"session_id": "test-tree",
		"cwd":        dir,
	})
	input := HookInput{
		ToolName:  "SessionStart",
		ToolInput: ti,
	}

	includePatterns := []string{
		"README.md",
		"src/**",
		"tests/**",
		"*.py",
	}

	result, code := CodebaseMap(input, dir, 3, includePatterns)
	if code != 0 || result.Decision != "allow" {
		t.Fatalf("expected allow, got decision=%s code=%d", result.Decision, code)
	}

	if !strings.Contains(result.Message, "CODEBASE STRUCTURE") {
		t.Error("expected message to contain 'CODEBASE STRUCTURE'")
	}
	if !strings.Contains(result.Message, "README.md") {
		t.Error("expected tree to include README.md")
	}
	if !strings.Contains(result.Message, "src") {
		t.Error("expected tree to include src directory")
	}
}

func TestCodebaseMap_RespectsMaxDepth(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "a", "b", "c", "d"), 0755)
	os.WriteFile(filepath.Join(dir, "a", "file.txt"), []byte("test"), 0644)

	ti, _ := json.Marshal(map[string]interface{}{
		"session_id": "test-depth",
		"cwd":        dir,
	})
	input := HookInput{
		ToolName:  "SessionStart",
		ToolInput: ti,
	}

	result, code := CodebaseMap(input, dir, 2, []string{"a/**"})
	if code != 0 || result.Decision != "allow" {
		t.Fatalf("expected allow, got decision=%s code=%d", result.Decision, code)
	}

	// Should not include "d" directory (depth 3) when maxDepth is 2
	if strings.Contains(result.Message, "d") {
		t.Error("expected tree to respect maxDepth and not include 'd' directory")
	}
}

func TestCodebaseMap_EmptyIncludePatterns(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)

	ti, _ := json.Marshal(map[string]interface{}{
		"session_id": "test-default",
		"cwd":        dir,
	})
	input := HookInput{
		ToolName:  "SessionStart",
		ToolInput: ti,
	}

	// Should use default patterns
	result, code := CodebaseMap(input, dir, 3, nil)
	if code != 0 || result.Decision != "allow" {
		t.Fatalf("expected allow, got decision=%s code=%d", result.Decision, code)
	}
	// Default patterns include *.md, so README.md should be included
	if !strings.Contains(result.Message, "README.md") {
		t.Error("expected default patterns to include README.md")
	}
}

func TestCodebaseMap_NonexistentDirectory(t *testing.T) {
	ti, _ := json.Marshal(map[string]interface{}{
		"session_id": "test-nonexistent",
		"cwd":        "/nonexistent/path/12345",
	})
	input := HookInput{
		ToolName:  "SessionStart",
		ToolInput: ti,
	}

	result, code := CodebaseMap(input, "/nonexistent/path/12345", 3, nil)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for nonexistent directory, got decision=%s code=%d", result.Decision, code)
	}
	if result.Message != "" {
		t.Errorf("expected empty message for nonexistent directory, got %q", result.Message)
	}
}

func TestMatchesIncludePattern(t *testing.T) {
	tests := []struct {
		name           string
		entryPath      string
		projectRoot    string
		includePattern string
		wantMatch      bool
	}{
		{
			name:           "exact name match",
			entryPath:      "/project/README.md",
			projectRoot:    "/project",
			includePattern: "README.md",
			wantMatch:      true,
		},
		{
			name:           "glob pattern match",
			entryPath:      "/project/src/main.py",
			projectRoot:    "/project",
			includePattern: "*.py",
			wantMatch:      true,
		},
		{
			name:           "directory pattern with **",
			entryPath:      "/project/tests/test_file.py",
			projectRoot:    "/project",
			includePattern: "tests/**",
			wantMatch:      true,
		},
		{
			name:           "no match",
			entryPath:      "/project/src/main.go",
			projectRoot:    "/project",
			includePattern: "*.py",
			wantMatch:      false,
		},
		{
			name:           "relative path match",
			entryPath:      "/project/docs/guide.md",
			projectRoot:    "/project",
			includePattern: "docs/*.md",
			wantMatch:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesIncludePattern(tt.entryPath, []string{tt.includePattern}, tt.projectRoot)
			if got != tt.wantMatch {
				t.Errorf("matchesIncludePattern(%q, %q, %q) = %v, want %v",
					tt.entryPath, tt.includePattern, tt.projectRoot, got, tt.wantMatch)
			}
		})
	}
}
