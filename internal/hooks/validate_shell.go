package hooks

import "regexp"

var shellDenyRules = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	// Destructive filesystem
	{regexp.MustCompile(`(?:^|\s|;|&&|\|\|)(?:sudo\s+)?rm\s+.*-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\s+/(\*|\s|$)`), "recursive force delete from root"},
	{regexp.MustCompile(`(?:^|\s|;|&&|\|\|)(?:sudo\s+)?rm\s+.*-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*\s+/(\*|\s|$)`), "recursive force delete from root"},
	{regexp.MustCompile(`(?:^|\s|;|&&|\|\|)(?:sudo\s+)?mkfs`), "disk format command"},
	{regexp.MustCompile(`\bdd\b.*of=/dev/`), "dd write to block device"},
	{regexp.MustCompile(`>\s*/dev/(?:sd|nvme|hd|vd)`), "write redirect to block device"},
	{regexp.MustCompile(`(?:^|\s|;|&&|\|\|)(?:sudo\s+)?chmod\s+.*-[a-zA-Z]*R[a-zA-Z]*\s+777\s+/`), "recursive chmod 777 from root"},
	{regexp.MustCompile(`:\(\)\s*\{.*\|.*&.*\}.*:`), "fork bomb detected"},

	// Git footguns
	{regexp.MustCompile(`\bgit\s+push\s+.*(?:-f\b|--force)`), "force push (use --force-with-lease)"},
	{regexp.MustCompile(`\bgit\s+reset\s+--hard`), "git reset --hard (destructive)"},

	// Remote code execution
	{regexp.MustCompile(`\b(?:curl|wget)\b.*\|\s*(?:bash|sh|zsh|python|python3|perl|ruby)`), "remote script execution via pipe"},
	{regexp.MustCompile(`\benv\b.*\|\s*(?:curl|wget|nc|netcat)`), "environment variable exfiltration"},
}

// ValidateShell is a preToolUse hook that blocks dangerous shell commands.
func ValidateShell(input HookInput) (HookResult, int) {
	if input.ToolName != "Shell" {
		return Allow(), 0
	}

	cmd := input.Command()
	if cmd == "" {
		return Allow(), 0
	}

	for _, rule := range shellDenyRules {
		if rule.pattern.MatchString(cmd) {
			return Deny("Blocked: " + rule.reason), 2
		}
	}

	return Allow(), 0
}
