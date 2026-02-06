package hooks

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var anyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`:\s*any\b`),
	regexp.MustCompile(`\bas\s+any\b`),
	regexp.MustCompile(`=\s*any\b`),
	regexp.MustCompile(`<any>`),
	regexp.MustCompile(`<any,`),
	regexp.MustCompile(`,\s*any>`),
	regexp.MustCompile(`Array<any>`),
}

var testUtilityPatterns = []*regexp.Regexp{
	regexp.MustCompile(`expect\.any\(`),
	regexp.MustCompile(`\.any\(\)`),
}

// CheckAnyChanged is a postToolUse hook that checks for 'any' types in TypeScript files.
// It forbids 'any' types to enforce better type safety, but allows test utility patterns.
func CheckAnyChanged(input HookInput) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	path := input.Path()
	contents := input.Contents()

	if path == "" || contents == "" {
		return Allow(), 0
	}

	ext := filepath.Ext(path)
	if ext != ".ts" && ext != ".tsx" {
		return Allow(), 0
	}

	lines := strings.Split(contents, "\n")
	var violations []struct {
		lineNum int
		line    string
	}

	for lineNum, line := range lines {
		lineNum++ // 1-indexed for user display
		hasAny := false
		for _, pattern := range anyPatterns {
			if pattern.MatchString(line) {
				hasAny = true
				break
			}
		}

		if !hasAny {
			continue
		}

		// Check if it's a test utility pattern (allowed)
		isTestUtility := false
		for _, pattern := range testUtilityPatterns {
			if pattern.MatchString(line) {
				isTestUtility = true
				break
			}
		}

		if !isTestUtility {
			violations = append(violations, struct {
				lineNum int
				line    string
			}{lineNum, strings.TrimSpace(line)})
		}
	}

	if len(violations) == 0 {
		return Allow(), 0
	}

	fileName := filepath.Base(path)
	var reason strings.Builder
	reason.WriteString(fmt.Sprintf("TypeScript 'any' types detected: Found %d occurrence(s) in %s", len(violations), fileName))
	reason.WriteString("\n  Hints:")
	reason.WriteString("\n    - Replace 'any' with 'unknown' for better type safety")
	reason.WriteString("\n    - Use specific types when possible")
	reason.WriteString("\n    - Consider using generics for flexible types")
	reason.WriteString("\n  Violations:")
	for _, v := range violations {
		reason.WriteString(fmt.Sprintf("\n    Line %d: %s", v.lineNum, v.line))
	}

	return Deny(reason.String()), 2
}
