package hooks

import (
	"regexp"
	"strings"
)

var secretPatterns = []struct {
	pattern *regexp.Regexp
	name    string
}{
	{regexp.MustCompile(`AKIA[0-9A-Z]{16}`), "AWS Access Key"},
	{regexp.MustCompile(`(?i)aws_secret_access_key\s*[=:]\s*"[A-Za-z0-9/+=]{40}"`), "AWS Secret Key"},
	{regexp.MustCompile(`ghp_[A-Za-z0-9]{20,}`), "GitHub Personal Access Token"},
	{regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`), "GitHub Fine-grained Token"},
	{regexp.MustCompile(`xox[bpors]-[A-Za-z0-9\-]{10,}`), "Slack Token"},
	{regexp.MustCompile(`sk_(?:live|test)_[A-Za-z0-9]{20,}`), "Stripe Secret Key"},
	{regexp.MustCompile(`SG\.[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}`), "SendGrid API Key"},
	{regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`), "Private Key"},
	{regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_\-]{10,}`), "JWT Token"},
	// Generic patterns — more prone to false positives, so we check context
	{regexp.MustCompile(`(?i)(?:api_key|apikey|api_secret)\s*[=:]\s*"[A-Za-z0-9\-_]{20,}"`), "API Key assignment"},
	{regexp.MustCompile(`(?i)(?:password|passwd)\s*[=:]\s*"[^"$]{8,}"`), "Hardcoded password"},
	{regexp.MustCompile(`(?i)(?:secret)\s*[=:]\s*"[A-Za-z0-9\-_]{10,}"`), "Hardcoded secret"},
}

// SecretScanner is a postToolUse hook that scans written file contents for secrets.
func SecretScanner(input HookInput) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	contents := input.Contents()
	path := input.Path()
	if contents == "" {
		return Allow(), 0
	}

	// Skip example/template files
	lower := strings.ToLower(path)
	if strings.Contains(lower, "example") || strings.Contains(lower, "template") || strings.Contains(lower, "sample") {
		return Allow(), 0
	}

	// Skip test files for generic patterns (but still catch real tokens)
	isTest := strings.Contains(lower, "_test.") || strings.Contains(lower, "test_") || strings.HasSuffix(lower, ".test.go")

	for _, sp := range secretPatterns {
		if sp.pattern.MatchString(contents) {
			// For generic patterns, skip test files
			if isTest && isGenericPattern(sp.name) {
				continue
			}
			return Deny("Blocked: potential " + sp.name + " detected in " + path), 2
		}
	}

	return Allow(), 0
}

func isGenericPattern(name string) bool {
	return name == "API Key assignment" || name == "Hardcoded password" || name == "Hardcoded secret"
}
