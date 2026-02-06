package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellCheck_NonShellTool(t *testing.T) {
	input := HookInput{
		ToolName: "Write",
		ToolInput: []byte(`{"path": "test.go", "contents": "package main"}`),
	}
	result, code := ShellCheck(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-shell file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestShellCheck_ShellCommand(t *testing.T) {
	input := HookInput{
		ToolName: "Shell",
		ToolInput: []byte(`{"command": "ls -la"}`),
	}
	result, code := ShellCheck(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-shell command, got decision=%s code=%d", result.Decision, code)
	}
}

func TestShellCheck_ShellFile(t *testing.T) {
	dir := t.TempDir()
	shellFile := filepath.Join(dir, "test.sh")
	os.WriteFile(shellFile, []byte("#!/bin/bash\necho $1\n"), 0644)

	input := writeInput(shellFile, "#!/bin/bash\necho $1\n")
	result, code := ShellCheck(input, dir)

	// Should allow if shellcheck passes or not installed
	if code == 2 && result.Decision == "deny" {
		// If it denies, should be because of shellcheck issues
		if !strings.Contains(result.Reason, "shellcheck") {
			t.Errorf("expected reason to mention shellcheck, got %q", result.Reason)
		}
	}
}

func TestShellCheck_SkipsVendor(t *testing.T) {
	dir := t.TempDir()
	vendorFile := filepath.Join(dir, "vendor", "script.sh")
	os.MkdirAll(filepath.Join(dir, "vendor"), 0755)
	os.WriteFile(vendorFile, []byte("#!/bin/bash\necho test\n"), 0644)

	input := writeInput(vendorFile, "#!/bin/bash\necho test\n")
	result, code := ShellCheck(input, dir)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for vendor file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestShellCheck_NonShellFile(t *testing.T) {
	input := writeInput("test.py", "print('hello')")
	result, code := ShellCheck(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-shell file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestIsShellCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected bool
	}{
		{"bash script", "bash script.sh", true},
		{"sh script", "sh script.sh", true},
		{"executable script", "./script.sh", true},
		{"absolute path", "/usr/bin/script.sh", true},
		{"regular command", "ls -la", false},
		{"git command", "git status", false},
		{"zsh script", "zsh script.sh", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isShellCommand(tt.cmd)
			if result != tt.expected {
				t.Errorf("isShellCommand(%q) = %v, want %v", tt.cmd, result, tt.expected)
			}
		})
	}
}

func TestExtractScriptPath(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		workDir  string
		expected string
	}{
		{"bash script", "bash script.sh", "/tmp", "/tmp/script.sh"},
		{"sh script", "sh test.sh", "/tmp", "/tmp/test.sh"},
		{"executable script", "./script.sh", "/tmp", "/tmp/script.sh"},
		{"absolute path", "bash /usr/bin/script.sh", "/tmp", "/usr/bin/script.sh"},
		{"no script", "ls -la", "/tmp", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractScriptPath(tt.cmd, tt.workDir)
			if result != tt.expected {
				t.Errorf("extractScriptPath(%q, %q) = %q, want %q", tt.cmd, tt.workDir, result, tt.expected)
			}
		})
	}
}

func TestShellCheck_ExecutingScript(t *testing.T) {
	dir := t.TempDir()
	scriptFile := filepath.Join(dir, "test.sh")
	os.WriteFile(scriptFile, []byte("#!/bin/bash\necho $1\n"), 0644)

	ti, _ := json.Marshal(map[string]string{
		"command": "bash test.sh",
	})
	input := HookInput{
		ToolName:  "Shell",
		ToolInput: ti,
	}

	result, code := ShellCheck(input, dir)
	// Should check the script file
	if code == 2 && result.Decision == "deny" {
		if !strings.Contains(result.Reason, "shellcheck") {
			t.Errorf("expected reason to mention shellcheck, got %q", result.Reason)
		}
	}
}

func TestShellCheck_NonexistentFile(t *testing.T) {
	input := writeInput("/nonexistent/script.sh", "#!/bin/bash\necho test\n")
	result, code := ShellCheck(input, ".")
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for nonexistent file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestShellCheck_CommandExists(t *testing.T) {
	// Test with a command that should exist
	if !commandExists("sh") && !commandExists("bash") {
		t.Log("sh/bash not found, skipping commandExists test")
	}

	// Test with a command that shouldn't exist
	if commandExists("nonexistent-command-xyz-123") {
		t.Error("expected nonexistent command to return false")
	}
}
