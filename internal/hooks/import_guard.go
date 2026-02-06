package hooks

import (
	"path/filepath"
	"strings"
)

// ImportGuard is a postToolUse hook that checks for banned imports/patterns.
// bannedPatterns maps file extension -> list of banned strings.
func ImportGuard(input HookInput, bannedPatterns map[string][]string) (HookResult, int) {
	return ImportGuardWithAllowlist(input, bannedPatterns, nil)
}

// ImportGuardWithAllowlist runs import guard; allowedPatterns[ext] lists patterns to allow even if banned.
func ImportGuardWithAllowlist(input HookInput, bannedPatterns map[string][]string, allowedPatterns map[string][]string) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}
	path := input.Path()
	contents := input.Contents()
	if path == "" || contents == "" {
		return Allow(), 0
	}
	ext := filepath.Ext(path)
	banned, ok := bannedPatterns[ext]
	if !ok {
		return Allow(), 0
	}
	if isTestFile(path) {
		return Allow(), 0
	}
	allowed := sliceSet(allowedPatterns[ext])
	for _, pattern := range banned {
		if allowed[pattern] {
			continue
		}
		if strings.Contains(contents, pattern) {
			return Deny("Blocked: banned pattern '" + pattern + "' found in " + filepath.Base(path)), 2
		}
	}
	return Allow(), 0
}

func sliceSet(ss []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range ss {
		m[s] = true
	}
	return m
}
