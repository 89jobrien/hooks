package hooks

import (
	"testing"
)

func TestNoSudo_Blocks(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"sudo rm", "sudo rm -rf /tmp/foo"},
		{"sudo apt", "sudo apt install nginx"},
		{"sudo systemctl", "sudo systemctl restart nginx"},
		{"sudo chown", "sudo chown root:root /etc/config"},
		{"sudo at start", "sudo cat /etc/shadow"},
		{"sudo after &&", "echo hi && sudo rm /tmp/x"},
		{"sudo after ;", "echo hi; sudo rm /tmp/x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := NoSudo(shellInput(tt.cmd))
			if code != 2 {
				t.Errorf("expected block (exit 2), got %d for %q", code, tt.cmd)
			}
			if result.Decision != "deny" {
				t.Errorf("expected deny, got %q", result.Decision)
			}
		})
	}
}

func TestNoSudo_Allows(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"normal command", "ls -la"},
		{"docker without sudo", "docker ps"},
		{"grep for sudo string", `grep -r "sudo" .`},
		{"echo sudo in string", `echo "run sudo to escalate"`},
		{"non-Shell tool", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input HookInput
			if tt.name == "non-Shell tool" {
				input = writeInput("main.go", "package main")
			} else {
				input = shellInput(tt.cmd)
			}
			result, code := NoSudo(input)
			if code != 0 {
				t.Errorf("expected allow (exit 0), got %d for %q; reason: %s", code, tt.cmd, result.Reason)
			}
		})
	}
}
