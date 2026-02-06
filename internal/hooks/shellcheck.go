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

const shellcheckTimeout = 10 * time.Second

// ShellCheck is a preToolUse hook that validates shell commands and shell files using shellcheck.
// It runs shellcheck on Shell commands and Write operations for .sh/.bash files.
func ShellCheck(input HookInput, workDir string) (HookResult, int) {
	// Handle Shell tool - check the command
	if input.ToolName == "Shell" {
		return shellCheckCommand(input, workDir)
	}

	// Handle Write tool - check shell files
	if input.ToolName == "Write" {
		return shellCheckFile(input, workDir)
	}

	return Allow(), 0
}

func shellCheckCommand(input HookInput, workDir string) (HookResult, int) {
	cmd := input.Command()
	if cmd == "" {
		return Allow(), 0
	}

	// Skip if not a shell command or if it's just running shellcheck itself
	if !isShellCommand(cmd) || strings.Contains(cmd, "shellcheck") {
		return Allow(), 0
	}

	// Extract shell script from command if it's executing a file
	scriptPath := extractScriptPath(cmd, workDir)
	if scriptPath == "" {
		// Not executing a file, check the command itself
		return checkCommandSyntax(cmd, workDir)
	}

	// Check the script file
	return checkShellFile(scriptPath, workDir)
}

func shellCheckFile(input HookInput, workDir string) (HookResult, int) {
	path := input.Path()
	if path == "" {
		return Allow(), 0
	}

	ext := filepath.Ext(path)
	if ext != ".sh" && ext != ".bash" && !strings.HasSuffix(path, ".sh") && !strings.HasSuffix(path, ".bash") {
		return Allow(), 0
	}

	// Skip vendor directories
	if strings.Contains(path, "/vendor/") || strings.Contains(path, "vendor/") {
		return Allow(), 0
	}

	// Resolve absolute path
	absPath := path
	if !filepath.IsAbs(path) && workDir != "" {
		absPath = filepath.Join(workDir, path)
	}

	return checkShellFile(absPath, workDir)
}

func isShellCommand(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	// Check for common shell patterns
	return strings.HasPrefix(cmd, "sh ") ||
		strings.HasPrefix(cmd, "bash ") ||
		strings.HasPrefix(cmd, "zsh ") ||
		strings.Contains(cmd, ".sh") ||
		strings.Contains(cmd, ".bash") ||
		strings.HasPrefix(cmd, "./") ||
		strings.HasPrefix(cmd, "/")
}

func extractScriptPath(cmd string, workDir string) string {
	// Look for script execution patterns: bash script.sh, sh script.sh, ./script.sh
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	// Check for ./script.sh pattern (first part)
	if strings.HasPrefix(parts[0], "./") {
		script := parts[0]
		if strings.HasSuffix(script, ".sh") || strings.HasSuffix(script, ".bash") {
			// Remove ./ prefix and join with workDir
			script = strings.TrimPrefix(script, "./")
			return filepath.Join(workDir, script)
		}
	}

	// Check if second part looks like a script
	if len(parts) >= 2 {
		script := parts[1]
		if strings.HasSuffix(script, ".sh") || strings.HasSuffix(script, ".bash") {
			if filepath.IsAbs(script) {
				return script
			}
			return filepath.Join(workDir, script)
		}
	}

	return ""
}

func checkCommandSyntax(cmd string, workDir string) (HookResult, int) {
	// For inline commands, we can't easily check them without shellcheck supporting stdin
	// So we'll allow them for now
	return Allow(), 0
}

func checkShellFile(filePath string, workDir string) (HookResult, int) {
	if _, err := os.Stat(filePath); err != nil {
		return Allow(), 0 // File doesn't exist yet, allow
	}

	// Check if shellcheck is available
	if !commandExists("shellcheck") {
		return Allow(), 0 // Fail open if shellcheck not installed
	}

	// Run shellcheck
	ctx, cancel := context.WithTimeout(context.Background(), shellcheckTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "shellcheck",
		"--severity=warning",
		"--enable=all",
		"--exclude=SC1090,SC1091", // Exclude source/include checks
		filePath)

	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return Deny("shellcheck timed out after 10 seconds"), 2
		}
		// Check if command not found
		if strings.Contains(err.Error(), "executable file not found") {
			return Allow(), 0 // Fail open if shellcheck not available
		}
	}

	if exitCode != 0 {
		output := stderr.String()
		if output == "" {
			output = stdout.String()
		}

		var reason strings.Builder
		reason.WriteString(fmt.Sprintf("shellcheck found issues in %s", filepath.Base(filePath)))
		reason.WriteString("\n  Hints:")
		reason.WriteString("\n    - Fix shellcheck warnings before continuing")
		reason.WriteString("\n    - Run 'shellcheck " + filePath + "' to see details")
		reason.WriteString("\n    - Common issues: unquoted variables, missing shebang, unsafe operations")

		if output != "" {
			reason.WriteString("\n\n" + output)
		}

		return Deny(reason.String()), 2
	}

	return Allow(), 0
}
