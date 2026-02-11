package hooks

import (
	"regexp"
)

var (
	commitMsgRe       = regexp.MustCompile(`\bgit\s+commit\s+.*-m\s+"([^"]*)"`)
	commitMsgSingleRe = regexp.MustCompile(`\bgit\s+commit\s+.*-m\s+'([^']*)'`)
	conventionalRe    = regexp.MustCompile(`^(feat|fix|chore|docs|refactor|test|ci|perf|style|build|revert)(\(.+\))?!?:\s+.+`)
)

// CommitMsgLint is a preToolUse hook that validates conventional commit messages.
func CommitMsgLint(input HookInput) (HookResult, int) {
	if input.ToolName != "Shell" {
		return Allow(), 0
	}

	cmd := input.Command()
	if cmd == "" {
		return Allow(), 0
	}

	// Only care about git commit commands
	var msg string
	if matches := commitMsgRe.FindStringSubmatch(cmd); len(matches) > 1 {
		msg = matches[1]
	} else if matches := commitMsgSingleRe.FindStringSubmatch(cmd); len(matches) > 1 {
		msg = matches[1]
	} else {
		// Not a git commit with -m, or uses heredoc — pass through
		return Allow(), 0
	}

	if msg == "" {
		return Deny("Blocked: empty commit message"), 2
	}

	if !conventionalRe.MatchString(msg) {
		return Deny("Blocked: commit message doesn't follow conventional commits format. " +
			"Expected: type(scope): description (e.g., 'feat: add auth', 'fix(api): handle timeout')"), 2
	}

	return Allow(), 0
}
