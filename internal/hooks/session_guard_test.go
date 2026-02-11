package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("commit", "--allow-empty", "-m", "init")
	return dir
}

func TestSessionGuard_CleanRepo(t *testing.T) {
	dir := initGitRepo(t)
	result, code := SessionGuard(HookInput{}, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for session hook, got %q", result.Decision)
	}
	if strings.Contains(strings.ToLower(result.Reason), "warning") {
		t.Errorf("expected no warning in clean repo, got %q", result.Reason)
	}
}

func TestSessionGuard_UntrackedFiles(t *testing.T) {
	dir := initGitRepo(t)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hello"), 0644)

	result, code := SessionGuard(HookInput{}, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	msg := strings.ToLower(result.Reason)
	if !strings.Contains(msg, "untracked") && !strings.Contains(msg, "warning") {
		t.Errorf("expected untracked warning, got %q", result.Reason)
	}
}

func TestSessionGuard_StagedChanges(t *testing.T) {
	dir := initGitRepo(t)
	os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("hello"), 0644)
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = dir
	cmd.Run()

	result, code := SessionGuard(HookInput{}, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	msg := strings.ToLower(result.Reason)
	if !strings.Contains(msg, "uncommitted") && !strings.Contains(msg, "staged") && !strings.Contains(msg, "warning") {
		t.Errorf("expected staged warning, got %q", result.Reason)
	}
}

func TestSessionGuard_ModifiedFiles(t *testing.T) {
	dir := initGitRepo(t)
	fpath := filepath.Join(dir, "tracked.txt")
	os.WriteFile(fpath, []byte("v1"), 0644)
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		cmd.Run()
	}
	run("add", "tracked.txt")
	run("commit", "-m", "add tracked")
	os.WriteFile(fpath, []byte("v2"), 0644)

	result, code := SessionGuard(HookInput{}, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	msg := strings.ToLower(result.Reason)
	if !strings.Contains(msg, "modified") && !strings.Contains(msg, "warning") {
		t.Errorf("expected modified warning, got %q", result.Reason)
	}
}

func TestSessionGuard_NotGitRepo(t *testing.T) {
	dir := t.TempDir() // not a git repo
	result, code := SessionGuard(HookInput{}, dir)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if result.Decision != "" {
		t.Errorf("expected empty decision for session hook, got %q", result.Decision)
	}
}

func TestSessionGuard_AlwaysExitsZero(t *testing.T) {
	// Even with dirty state, should never block
	dir := initGitRepo(t)
	os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("x"), 0644)

	_, code := SessionGuard(HookInput{}, dir)
	if code != 0 {
		t.Errorf("session guard must always exit 0, got %d", code)
	}
}
