package hooks

import (
	"regexp"
	"strings"
)

var (
	gitCheckoutRe   = regexp.MustCompile(`\bgit\s+(?:checkout|switch)\s+(\S+)`)
	gitCheckoutNewRe = regexp.MustCompile(`\bgit\s+(?:checkout|switch)\s+-[bB]\s`)
	gitCommitRe     = regexp.MustCompile(`\bgit\s+commit\b`)
	gitMergeRe      = regexp.MustCompile(`\bgit\s+merge\b`)
	gitRebaseRe     = regexp.MustCompile(`\bgit\s+rebase\b`)
	gitReadOnlyRe   = regexp.MustCompile(`\bgit\s+(?:status|log|diff|show|branch|remote|fetch|stash\s+list|tag\s+-l)\b`)
)

// BranchGuard is a preToolUse hook that prevents operations on protected branches.
func BranchGuard(input HookInput, protected []string, currentBranch string) (HookResult, int) {
	if input.ToolName != "Shell" {
		return Allow(), 0
	}

	cmd := input.Command()
	if cmd == "" || !strings.Contains(cmd, "git") {
		return Allow(), 0
	}

	// Always allow read-only git commands
	if gitReadOnlyRe.MatchString(cmd) && !gitCommitRe.MatchString(cmd) && !gitMergeRe.MatchString(cmd) {
		return Allow(), 0
	}

	// Block checkout/switch to protected branch (but allow -b new branch)
	if gitCheckoutRe.MatchString(cmd) && !gitCheckoutNewRe.MatchString(cmd) {
		matches := gitCheckoutRe.FindStringSubmatch(cmd)
		if len(matches) > 1 {
			target := matches[1]
			for _, p := range protected {
				if target == p {
					return Deny("Blocked: cannot checkout protected branch '" + p + "'. Use a feature branch."), 2
				}
			}
		}
	}

	// Block commit/merge/rebase on protected branch
	isOnProtected := false
	for _, p := range protected {
		if currentBranch == p {
			isOnProtected = true
			break
		}
	}

	if isOnProtected {
		if gitCommitRe.MatchString(cmd) {
			return Deny("Blocked: cannot commit on protected branch '" + currentBranch + "'. Create a feature branch."), 2
		}
		if gitMergeRe.MatchString(cmd) {
			return Deny("Blocked: cannot merge on protected branch '" + currentBranch + "'."), 2
		}
		if gitRebaseRe.MatchString(cmd) {
			return Deny("Blocked: cannot rebase on protected branch '" + currentBranch + "'."), 2
		}
	}

	return Allow(), 0
}
