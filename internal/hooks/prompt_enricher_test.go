package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptEnricher_InjectsConventions(t *testing.T) {
	dir := t.TempDir()
	conventions := "- Always use TDD\n- Use conventional commits\n- Prefer composition over inheritance\n"
	os.MkdirAll(filepath.Join(dir, ".cursor"), 0755)
	os.WriteFile(filepath.Join(dir, ".cursor", "conventions.md"), []byte(conventions), 0644)

	result, code := PromptEnricher(HookInput{}, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(result.Message, "TDD") {
		t.Errorf("expected conventions in message, got %q", result.Message)
	}
}

func TestPromptEnricher_NoFileNoMessage(t *testing.T) {
	dir := t.TempDir()
	result, code := PromptEnricher(HookInput{}, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Message != "" {
		t.Errorf("expected no message when no conventions file, got %q", result.Message)
	}
}

func TestPromptEnricher_AlwaysAllows(t *testing.T) {
	result, code := PromptEnricher(HookInput{}, t.TempDir())
	if code != 0 || result.Decision != "allow" {
		t.Errorf("should always allow")
	}
}
