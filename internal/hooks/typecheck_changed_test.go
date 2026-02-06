package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypecheckChanged_NonWriteTool(t *testing.T) {
	input := HookInput{
		ToolName: "Shell",
		ToolInput: []byte(`{"command": "ls"}`),
	}
	result, code := TypecheckChanged(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-Write tool, got decision=%s code=%d", result.Decision, code)
	}
}

func TestTypecheckChanged_NonTypeScriptFile(t *testing.T) {
	input := writeInput("app.go", "package main")
	result, code := TypecheckChanged(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-TypeScript file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestTypecheckChanged_NoFilePath(t *testing.T) {
	input := writeInput("", "content")
	result, code := TypecheckChanged(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no file path, got decision=%s code=%d", result.Decision, code)
	}
}

func TestTypecheckChanged_NonexistentFile(t *testing.T) {
	input := writeInput("/nonexistent/file.ts", "const x = 1;")
	result, code := TypecheckChanged(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for nonexistent file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestTypecheckChanged_NoTsconfig(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.ts")
	os.WriteFile(testFile, []byte("const x = 1;"), 0644)

	input := writeInput(testFile, "const x = 1;")
	result, code := TypecheckChanged(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when no tsconfig.json, got decision=%s code=%d", result.Decision, code)
	}
}

func TestTypecheckChanged_TSXFile(t *testing.T) {
	dir := t.TempDir()
	tsconfig := filepath.Join(dir, "tsconfig.json")
	os.WriteFile(tsconfig, []byte(`{"compilerOptions": {}}`), 0644)

	testFile := filepath.Join(dir, "test.tsx")
	os.WriteFile(testFile, []byte("const x = 1;"), 0644)

	input := writeInput(testFile, "const x = 1;")
	result, code := TypecheckChanged(input, dir)
	// Should allow if tsc passes or not available
	if code != 0 && result.Decision == "deny" {
		// If it denies, should be because of type errors
		if !strings.Contains(result.Reason, "Type errors") {
			t.Errorf("expected deny reason to mention type errors, got %q", result.Reason)
		}
	}
}

func TestTypecheckChanged_RelativePath(t *testing.T) {
	dir := t.TempDir()
	tsconfig := filepath.Join(dir, "tsconfig.json")
	os.WriteFile(tsconfig, []byte(`{"compilerOptions": {}}`), 0644)

	testFile := filepath.Join(dir, "test.ts")
	os.WriteFile(testFile, []byte("const x = 1;"), 0644)

	relPath, _ := filepath.Rel(dir, testFile)
	input := writeInput(relPath, "const x = 1;")

	result, code := TypecheckChanged(input, dir)
	// Should allow if tsc passes or not available
	if code != 0 && result.Decision == "deny" {
		if !strings.Contains(result.Reason, "Type errors") {
			t.Errorf("expected deny reason to mention type errors, got %q", result.Reason)
		}
	}
}

func TestParseTscOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		wantErrs int
	}{
		{
			name:     "single error",
			output:   "test.ts:5:10 - error TS2322: Type 'string' is not assignable to type 'number'.",
			wantErrs: 1,
		},
		{
			name: "multiple errors",
			output: `test.ts:5:10 - error TS2322: Type 'string' is not assignable to type 'number'.
test.ts:8:3 - error TS2304: Cannot find name 'x'.`,
			wantErrs: 2,
		},
		{
			name:     "no errors",
			output:   "Found 0 errors.",
			wantErrs: 0,
		},
		{
			name:     "mixed output",
			output:   "Some info\n  error TS1234: Something wrong\nMore info",
			wantErrs: 1,
		},
		{
			name:     "colon error format",
			output:   "test.ts:10:5: error: Cannot find module 'xyz'",
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := parseTscOutput(tt.output)
			if len(errors) != tt.wantErrs {
				t.Errorf("parseTscOutput() returned %d errors, want %d", len(errors), tt.wantErrs)
			}
		})
	}
}

func TestParseTscOutput_Empty(t *testing.T) {
	errors := parseTscOutput("")
	if len(errors) != 0 {
		t.Errorf("expected no errors for empty output, got %d", len(errors))
	}
}

func TestParseTscOutput_Whitespace(t *testing.T) {
	output := "   \n  \n  test.ts:1:1 - error TS1234: Test\n  \n"
	errors := parseTscOutput(output)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0], "error TS1234") {
		t.Errorf("expected error to contain 'error TS1234', got %q", errors[0])
	}
}
