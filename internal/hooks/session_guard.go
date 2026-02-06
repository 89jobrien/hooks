package hooks

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SessionGuard is a sessionStart hook that warns about workspace state.
// Always exits 0 (informational only, never blocks).
func SessionGuard(input HookInput, workDir string) (HookResult, int) {
	// Check if inside git repo
	if !isGitRepo(workDir) {
		return HookResult{Decision: "allow", Message: "Not a git repository"}, 0
	}

	var warnings []string

	if n := countGitFiles(workDir, "diff", "--cached", "--name-only"); n > 0 {
		warnings = append(warnings, fmt.Sprintf("warning: %d staged but uncommitted file(s)", n))
	}

	if n := countGitFiles(workDir, "diff", "--name-only"); n > 0 {
		warnings = append(warnings, fmt.Sprintf("warning: %d modified unstaged file(s)", n))
	}

	if n := countGitFiles(workDir, "ls-files", "--others", "--exclude-standard"); n > 0 {
		warnings = append(warnings, fmt.Sprintf("warning: %d untracked file(s)", n))
	}

	if isDetachedHead(workDir) {
		warnings = append(warnings, "warning: detached HEAD state")
	}

	if len(warnings) == 0 {
		return HookResult{Decision: "allow", Message: "workspace clean"}, 0
	}

	return HookResult{Decision: "allow", Message: strings.Join(warnings, "; ")}, 0
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func countGitFiles(dir string, args ...string) int {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	lines := strings.TrimSpace(string(out))
	if lines == "" {
		return 0
	}
	n, _ := strconv.Atoi(fmt.Sprintf("%d", len(strings.Split(lines, "\n"))))
	return n
}

func isDetachedHead(dir string) bool {
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = dir
	return cmd.Run() != nil
}
