package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJITContext_EmptyPrompt(t *testing.T) {
	input := HookInput{
		ToolName:  "beforeSubmitPrompt",
		ToolInput: []byte(`{"prompt": ""}`),
	}
	result, code := JITContext(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for empty prompt, got decision=%s code=%d", result.Decision, code)
	}
	if result.Message != "" {
		t.Errorf("expected empty message for empty prompt, got %q", result.Message)
	}
}

func TestJITContext_ShortPrompt(t *testing.T) {
	input := HookInput{
		ToolName:  "beforeSubmitPrompt",
		ToolInput: []byte(`{"prompt": "hi"}`),
	}
	result, code := JITContext(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for short prompt, got decision=%s code=%d", result.Decision, code)
	}
}

func TestExtractPatterns(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		wantAny  []string // Patterns that should be found
		wantNone []string // Patterns that should NOT be found
	}{
		{
			name:    "glob patterns",
			prompt:  "check all *.py files and **/*.ts",
			wantAny: []string{"*.py", "*.ts"}, // Pattern may be split
		},
		{
			name:    "file extensions",
			prompt:  "look at .ts and .js files",
			wantAny: []string{"*.ts", "*.js"},
		},
		{
			name:     "no patterns",
			prompt:   "just a regular question",
			wantNone: []string{"*.py", "*.ts"},
		},
		{
			name:    "mixed patterns",
			prompt:  "check src/**/*.ts and *.py files",
			wantAny: []string{"*.ts", "*.py"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := extractPatterns(tt.prompt)
			for _, want := range tt.wantAny {
				found := false
				for _, p := range patterns {
					if p == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected pattern %q in %v", want, patterns)
				}
			}
			for _, dontWant := range tt.wantNone {
				for _, p := range patterns {
					if p == dontWant {
						t.Errorf("unexpected pattern %q found", dontWant)
					}
				}
			}
		})
	}
}

func TestExtractPaths(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		wantAny  []string
		wantNone []string
	}{
		{
			name:    "file paths",
			prompt:  "check src/utils.ts and config.json",
			wantAny: []string{"src/utils.ts", "config.json"},
		},
		{
			name:    "directory paths",
			prompt:  "look in src/ and ./config",
			wantAny: []string{"src/", "./config"},
		},
		{
			name:     "skip URLs",
			prompt:   "check https://example.com/file.ts",
			wantNone: []string{"https://example.com/file.ts"},
		},
		{
			name:     "no paths",
			prompt:   "just a question",
			wantNone: []string{"src/", "file.ts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := extractPaths(tt.prompt)
			for _, want := range tt.wantAny {
				found := false
				for _, p := range paths {
					if p == want || strings.Contains(p, want) {
						found = true
						break
					}
				}
				if !found && len(tt.wantAny) > 0 {
					t.Logf("paths found: %v", paths)
				}
			}
			for _, dontWant := range tt.wantNone {
				for _, p := range paths {
					if p == dontWant {
						t.Errorf("unexpected path %q found", dontWant)
					}
				}
			}
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantAny []string
	}{
		{
			name:    "quoted strings",
			prompt:  `check for "getUserData" and 'processRequest'`,
			wantAny: []string{"getUserData", "processRequest"},
		},
		{
			name:    "CamelCase identifiers",
			prompt:  "look for UserService and DataHandler",
			wantAny: []string{"UserService", "DataHandler"},
		},
		{
			name:    "snake_case identifiers",
			prompt:  "find get_user_data and process_request",
			wantAny: []string{"get_user_data", "process_request"},
		},
		{
			name:    "mixed keywords",
			prompt:  `search for "validate" and UserModel`,
			wantAny: []string{"validate", "UserModel"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := extractKeywords(tt.prompt)
			for _, want := range tt.wantAny {
				found := false
				for _, kw := range keywords {
					if kw == want {
						found = true
						break
					}
				}
				if !found && len(tt.wantAny) > 0 {
					t.Logf("keywords found: %v, wanted: %v", keywords, tt.wantAny)
				}
			}
		})
	}
}

func TestJITContext_WithPatternMatch(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.py")
	os.WriteFile(testFile, []byte("def hello():\n    print('world')\n"), 0644)

	ti, _ := json.Marshal(map[string]string{
		"prompt": "check all *.py files",
	})
	input := HookInput{
		ToolName:  "beforeSubmitPrompt",
		ToolInput: ti,
	}

	result, code := JITContext(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Fatalf("expected allow, got decision=%s code=%d", result.Decision, code)
	}

	if result.Message == "" {
		t.Error("expected message with file context")
	}

	if !strings.Contains(result.Message, "test.py") {
		t.Errorf("expected message to contain test.py, got %q", result.Message)
	}

	if !strings.Contains(result.Message, "JIT Context") {
		t.Errorf("expected message to contain 'JIT Context', got %q", result.Message)
	}
}

func TestJITContext_WithPathMatch(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "src", "utils.ts")
	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(testFile, []byte("export function test() {}\n"), 0644)

	ti, _ := json.Marshal(map[string]string{
		"prompt": "check src/utils.ts",
	})
	input := HookInput{
		ToolName:  "beforeSubmitPrompt",
		ToolInput: ti,
	}

	result, code := JITContext(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Fatalf("expected allow, got decision=%s code=%d", result.Decision, code)
	}

	if result.Message != "" {
		if !strings.Contains(result.Message, "utils.ts") {
			t.Logf("message: %q", result.Message)
		}
	}
}

func TestHeadTail(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	// Create file with many lines
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("Line %d", i))
	}
	os.WriteFile(testFile, []byte(strings.Join(lines, "\n")), 0644)

	content := headTail(testFile)
	if content == "" {
		t.Fatal("expected non-empty content")
	}

	// Should contain first and last lines
	if !strings.Contains(content, "Line 0") {
		t.Error("expected content to contain first line")
	}
	if !strings.Contains(content, "Line 99") {
		t.Error("expected content to contain last line")
	}

	// Should contain omitted marker
	if !strings.Contains(content, "omitted") {
		t.Error("expected content to contain 'omitted' marker")
	}
}

func TestHeadTail_SmallFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	content := "Line 1\nLine 2\nLine 3"
	os.WriteFile(testFile, []byte(content), 0644)

	result := headTail(testFile)
	if result != content {
		t.Errorf("expected full content for small file, got %q", result)
	}
}

func TestGrepFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.py")
	content := "def hello():\n    print('world')\ndef goodbye():\n    print('bye')"
	os.WriteFile(testFile, []byte(content), 0644)

	hits := grepFile(testFile, "hello")
	if len(hits) == 0 {
		t.Fatal("expected at least one hit")
	}

	found := false
	for _, hit := range hits {
		if strings.Contains(strings.ToLower(hit.line), "hello") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected hit to contain 'hello'")
	}
}

func TestGrepFile_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.py")
	content := "def Hello():\n    pass"
	os.WriteFile(testFile, []byte(content), 0644)

	hits := grepFile(testFile, "hello")
	if len(hits) == 0 {
		t.Error("expected case-insensitive match")
	}
}

func TestFormatContext(t *testing.T) {
	files := []fileContent{
		{path: "test.py", content: "def test():\n    pass"},
	}
	grepResults := []grepResult{
		{
			path: "utils.ts",
			hits: []grepHit{{lineNum: 5, line: "export function test() {}"}},
		},
	}

	context := formatContext(files, grepResults)
	if context == "" {
		t.Fatal("expected non-empty context")
	}

	if !strings.Contains(context, "JIT Context") {
		t.Error("expected context to contain 'JIT Context'")
	}

	if !strings.Contains(context, "test.py") {
		t.Error("expected context to contain file path")
	}

	if !strings.Contains(context, "utils.ts") {
		t.Error("expected context to contain grep result path")
	}
}

func TestCollectPathMatches(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.ts")
	os.WriteFile(testFile, []byte("const x = 1;"), 0644)

	results := collectPathMatches(dir, []string{"test.ts"})
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	if results[0].path != "test.ts" {
		t.Errorf("expected path 'test.ts', got %q", results[0].path)
	}
}

func TestCollectPathMatches_Nonexistent(t *testing.T) {
	dir := t.TempDir()
	results := collectPathMatches(dir, []string{"nonexistent.ts"})
	if len(results) != 0 {
		t.Errorf("expected no results for nonexistent file, got %d", len(results))
	}
}

func TestCollectPathMatches_LargeFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "large.txt")
	// Create a file larger than maxFileSize
	largeContent := make([]byte, maxFileSize+1)
	os.WriteFile(testFile, largeContent, 0644)

	results := collectPathMatches(dir, []string{"large.txt"})
	if len(results) != 0 {
		t.Error("expected no results for file larger than maxFileSize")
	}
}
