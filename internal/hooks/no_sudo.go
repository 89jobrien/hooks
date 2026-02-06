package hooks

import "regexp"

var sudoRe = regexp.MustCompile(`(?:^|;|&&|\|\|)\s*sudo\s`)

// NoSudo is a preToolUse hook that blocks sudo usage in shell commands.
func NoSudo(input HookInput) (HookResult, int) {
	if input.ToolName != "Shell" {
		return Allow(), 0
	}

	cmd := input.Command()
	if cmd == "" {
		return Allow(), 0
	}

	if sudoRe.MatchString(cmd) {
		return Deny("Blocked: sudo is not allowed. Run commands without elevated privileges."), 2
	}

	return Allow(), 0
}
