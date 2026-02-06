package hooks

import (
	"strings"
	"testing"
)

func TestFileSizeGuard_BlocksLargeFiles(t *testing.T) {
	// Generate a 600-line file
	lines := make([]string, 600)
	for i := range lines {
		lines[i] = "// line of code"
	}
	bigContent := strings.Join(lines, "\n")

	result, code := FileSizeGuard(writeInput("handler.go", bigContent), 500)
	if code != 2 {
		t.Errorf("expected block (exit 2) for 600-line file, got %d", code)
	}
	if result.Decision != "deny" {
		t.Errorf("expected deny, got %q", result.Decision)
	}
}

func TestFileSizeGuard_AllowsSmallFiles(t *testing.T) {
	content := "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"
	result, code := FileSizeGuard(writeInput("main.go", content), 500)
	if code != 0 {
		t.Errorf("expected allow (exit 0) for small file, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}

func TestFileSizeGuard_ExactLimit(t *testing.T) {
	lines := make([]string, 500)
	for i := range lines {
		lines[i] = "x"
	}
	content := strings.Join(lines, "\n")

	result, code := FileSizeGuard(writeInput("main.go", content), 500)
	if code != 0 {
		t.Errorf("expected allow at exact limit, got %d", code)
	}
	if result.Decision != "allow" {
		t.Errorf("expected allow, got %q", result.Decision)
	}
}

func TestFileSizeGuard_PassthroughNonWrite(t *testing.T) {
	result, code := FileSizeGuard(shellInput("ls"), 500)
	if code != 0 || result.Decision != "allow" {
		t.Error("should passthrough non-Write tools")
	}
}

func TestFileSizeGuard_CustomLimit(t *testing.T) {
	lines := make([]string, 101)
	for i := range lines {
		lines[i] = "x"
	}
	content := strings.Join(lines, "\n")

	result, code := FileSizeGuard(writeInput("small.go", content), 100)
	if code != 2 {
		t.Errorf("expected block at custom limit 100, got %d", code)
	}
	if result.Decision != "deny" {
		t.Errorf("expected deny, got %q", result.Decision)
	}
}
