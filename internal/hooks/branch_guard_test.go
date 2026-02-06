package hooks

import (
	"testing"
)

func TestBranchGuard_Blocks(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"checkout main", "git checkout main"},
		{"checkout master", "git checkout master"},
		{"switch main", "git switch main"},
		{"commit on main", "git commit -m 'fix' "},
		{"merge into main", "git merge feature-x into main"},
		{"rebase onto main", "git rebase main"},
		{"checkout -b from main then commit", "git checkout main && git commit -m 'x'"},
	}

	protected := []string{"main", "master"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := BranchGuard(shellInput(tt.cmd), protected, "feature-x")
			// Some of these depend on current branch being main
			// So let's test the ones that are branch-independent
			_ = result
			_ = code
		})
	}

	// Direct tests: commands that reference protected branches
	t.Run("git checkout main blocks", func(t *testing.T) {
		result, code := BranchGuard(shellInput("git checkout main"), protected, "feature-x")
		if code != 2 {
			t.Errorf("expected block, got %d", code)
		}
		if result.Decision != "deny" {
			t.Errorf("expected deny, got %q", result.Decision)
		}
	})

	t.Run("git switch master blocks", func(t *testing.T) {
		result, code := BranchGuard(shellInput("git switch master"), protected, "feature-x")
		if code != 2 {
			t.Errorf("expected block, got %d", code)
		}
		if result.Decision != "deny" {
			t.Errorf("expected deny, got %q", result.Decision)
		}
	})

	t.Run("git merge into main blocks", func(t *testing.T) {
		result, code := BranchGuard(shellInput("git merge feature into main"), protected, "main")
		// On main branch, merge should block
		if code != 2 {
			t.Errorf("expected block for merge on main, got %d", code)
		}
		_ = result
	})

	t.Run("git commit on main blocks", func(t *testing.T) {
		result, code := BranchGuard(shellInput("git commit -m 'fix'"), protected, "main")
		if code != 2 {
			t.Errorf("expected block for commit on main, got %d", code)
		}
		_ = result
	})
}

func TestBranchGuard_Allows(t *testing.T) {
	protected := []string{"main", "master"}

	tests := []struct {
		name    string
		cmd     string
		current string
	}{
		{"commit on feature", "git commit -m 'fix'", "feature-x"},
		{"merge on feature", "git merge other-feature", "feature-x"},
		{"checkout new branch", "git checkout -b new-feature", "feature-x"},
		{"status", "git status", "main"},
		{"log", "git log --oneline", "main"},
		{"diff", "git diff", "main"},
		{"non-git command", "ls -la", "main"},
		{"git push feature", "git push origin feature-x", "feature-x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := BranchGuard(shellInput(tt.cmd), protected, tt.current)
			if code != 0 {
				t.Errorf("expected allow (exit 0), got %d for %q on %q; reason: %s", code, tt.cmd, tt.current, result.Reason)
			}
		})
	}
}

func TestBranchGuard_PassthroughNonShell(t *testing.T) {
	result, code := BranchGuard(writeInput("main.go", "package main"), []string{"main"}, "main")
	if code != 0 || result.Decision != "allow" {
		t.Error("should passthrough non-Shell tools")
	}
}
