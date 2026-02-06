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

// TypecheckChanged is a postToolUse hook that runs TypeScript type checking on changed files.
// It runs `npx tsc --noEmit` on .ts/.tsx files when tsconfig.json exists.
func TypecheckChanged(input HookInput, workDir string) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	path := input.Path()
	if path == "" {
		return Allow(), 0
	}

	ext := filepath.Ext(path)
	if ext != ".ts" && ext != ".tsx" {
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

	// Check for tsconfig.json
	tsconfig := filepath.Join(projectRoot, "tsconfig.json")
	if _, err := os.Stat(tsconfig); err != nil {
		return Allow(), 0
	}

	// Run tsc with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npx", "tsc", "--noEmit", absPath)
	cmd.Dir = projectRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return Deny("TypeScript type checking timed out after 60 seconds"), 2
		}
		// Check if command not found
		if strings.Contains(err.Error(), "executable file not found") {
			return Allow(), 0 // Fail open if tsc not available
		}
	}

	if exitCode != 0 {
		fileName := filepath.Base(path)
		output := stderr.String()
		if output == "" {
			output = stdout.String()
		}

		errors := parseTscOutput(output)

		var reason strings.Builder
		reason.WriteString(fmt.Sprintf("Type errors found in %s", fileName))
		if len(errors) > 0 {
			reason.WriteString(fmt.Sprintf("\n  Details: %s", errors[0]))
		} else {
			reason.WriteString("\n  Details: See output below")
		}
		reason.WriteString("\n  Hints:")
		reason.WriteString("\n    - Fix type errors before continuing")
		reason.WriteString("\n    - Run 'npm run type-check' to see all errors")

		if output != "" {
			reason.WriteString("\n\n" + output)
		}

		return Deny(reason.String()), 2
	}

	return Allow(), 0
}

func parseTscOutput(output string) []string {
	var errors []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && (strings.Contains(line, "error TS") || strings.Contains(line, ": error")) {
			errors = append(errors, line)
		}
	}

	return errors
}
