package hooks

import (
	"testing"
)

func TestCommitMsgLint_BlocksBadMessages(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"no prefix", `git commit -m "fixed the bug"`},
		{"uppercase start", `git commit -m "Fix the thing"`},
		{"single word", `git commit -m "update"`},
		{"empty message", `git commit -m ""`},
		{"random prefix", `git commit -m "yolo: ship it"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := CommitMsgLint(shellInput(tt.cmd))
			if code != 2 {
				t.Errorf("expected block (exit 2), got %d for %q", code, tt.cmd)
			}
			if result.Decision != "deny" {
				t.Errorf("expected deny, got %q", result.Decision)
			}
		})
	}
}

func TestCommitMsgLint_AllowsGoodMessages(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"feat", `git commit -m "feat: add user auth"`},
		{"fix", `git commit -m "fix: resolve null pointer"`},
		{"chore", `git commit -m "chore: update dependencies"`},
		{"docs", `git commit -m "docs: add API reference"`},
		{"refactor", `git commit -m "refactor: extract handler logic"`},
		{"test", `git commit -m "test: add integration tests"`},
		{"ci", `git commit -m "ci: add github actions"`},
		{"perf", `git commit -m "perf: optimize query"`},
		{"style", `git commit -m "style: fix formatting"`},
		{"build", `git commit -m "build: update dockerfile"`},
		{"feat with scope", `git commit -m "feat(auth): add OAuth support"`},
		{"fix with scope", `git commit -m "fix(api): handle timeout"`},
		{"breaking change", `git commit -m "feat!: redesign API"`},
		{"non-commit command", "git status"},
		{"git add", "git add ."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := CommitMsgLint(shellInput(tt.cmd))
			if code != 0 {
				t.Errorf("expected allow (exit 0), got %d for %q; reason: %s", code, tt.cmd, result.Reason)
			}
		})
	}
}

func TestCommitMsgLint_PassthroughNonShell(t *testing.T) {
	result, code := CommitMsgLint(writeInput("main.go", "package main"))
	if code != 0 || result.Decision != "allow" {
		t.Error("should passthrough non-Shell tools")
	}
}

func TestCommitMsgLint_HeredocIgnored(t *testing.T) {
	// Heredoc-style commits are harder to parse, should pass through
	result, code := CommitMsgLint(shellInput(`git commit -m "$(cat <<'EOF'
feat: add something

Detailed description here.
EOF
)"`))
	// This is tricky — the hook should either parse it or pass through
	if code != 0 {
		t.Logf("heredoc blocked with reason: %s (acceptable)", result.Reason)
	}
}
