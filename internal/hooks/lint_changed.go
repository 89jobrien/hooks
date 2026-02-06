package hooks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LintChanged is a postToolUse hook that runs the appropriate linter on changed files.
// It detects biome, eslint, or ruff based on config files and runs them.
func LintChanged(input HookInput, workDir string) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	path := input.Path()
	if path == "" {
		return Allow(), 0
	}

	// Resolve absolute path
	absPath := path
	if !filepath.IsAbs(path) && workDir != "" {
		absPath = filepath.Join(workDir, path)
	}

	if _, err := os.Stat(absPath); err != nil {
		return Allow(), 0
	}

	projectRoot := workDir
	if projectRoot == "" {
		cwd, _ := os.Getwd()
		projectRoot = cwd
	}

	linter := detectLinter(projectRoot)
	if linter == "" {
		return Allow(), 0
	}

	// Run linter with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch linter {
	case "biome":
		cmd = exec.CommandContext(ctx, "biome", "check", absPath)
	case "eslint":
		cmd = exec.CommandContext(ctx, "eslint", absPath)
	case "ruff":
		cmd = exec.CommandContext(ctx, "ruff", "check", "--fix", "--exit-zero", absPath)
	default:
		return Deny("Unknown linter: " + linter), 2
	}

	cmd.Dir = projectRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return Deny(fmt.Sprintf("%s timed out after 30 seconds", linter)), 2
		}
		// Check if command not found
		if strings.Contains(err.Error(), "executable file not found") {
			return Allow(), 0 // Fail open if linter not installed
		}
	}

	if exitCode != 0 {
		fileName := filepath.Base(path)
		var reason strings.Builder
		reason.WriteString(fmt.Sprintf("%s found issues in %s (exit code %d)", linter, fileName, exitCode))
		reason.WriteString("\n  Hints:")
		reason.WriteString(fmt.Sprintf("\n    - Run '%s --fix %s' to auto-fix issues", linter, fileName))
		reason.WriteString("\n    - Review the linter output below for details")

		if stdout.Len() > 0 {
			reason.WriteString("\n\n" + stdout.String())
		}
		if stderr.Len() > 0 {
			reason.WriteString("\n" + stderr.String())
		}

		return Deny(reason.String()), 2
	}

	return Allow(), 0
}

func detectLinter(projectRoot string) string {
	// Check for biome
	if exists(filepath.Join(projectRoot, "biome.json")) ||
		exists(filepath.Join(projectRoot, "biome.jsonc")) {
		if commandExists("biome") {
			return "biome"
		}
	}

	// Check for eslint
	if exists(filepath.Join(projectRoot, ".eslintrc.js")) ||
		exists(filepath.Join(projectRoot, ".eslintrc.json")) ||
		exists(filepath.Join(projectRoot, ".eslintrc.cjs")) ||
		exists(filepath.Join(projectRoot, ".eslintrc.yaml")) ||
		exists(filepath.Join(projectRoot, ".eslintrc.yml")) ||
		exists(filepath.Join(projectRoot, "eslint.config.js")) ||
		exists(filepath.Join(projectRoot, "eslint.config.mjs")) {
		if commandExists("eslint") {
			return "eslint"
		}
	}

	// Check for ruff
	if exists(filepath.Join(projectRoot, "ruff.toml")) ||
		exists(filepath.Join(projectRoot, "pyproject.toml")) {
		if commandExists("ruff") {
			return "ruff"
		}
	}

	return ""
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
