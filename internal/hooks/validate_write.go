package hooks

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	envFileRe     = regexp.MustCompile(`^\.env($|\..+)`)
	envExemptRe   = regexp.MustCompile(`(?i)\.(example|sample|template)$`)
	sshKeyRe      = regexp.MustCompile(`(?:id_rsa|id_ed25519|id_ecdsa|id_dsa|authorized_keys)`)
	certKeyRe     = regexp.MustCompile(`\.(pem|key|p12|pfx|keystore|jks)$`)
	credentialsRe = regexp.MustCompile(`(?i)^credentials\.(json|yaml|yml|xml)$`)
	secretsRe     = regexp.MustCompile(`(?i)^secrets\.(json|yaml|yml|xml)$`)
	serviceAcctRe = regexp.MustCompile(`"type".*service_account`)
	kubeconfigRe  = regexp.MustCompile(`\.kube/config$`)
	tfvarsRe      = regexp.MustCompile(`\.tfvars$`)
)

type writeRule struct {
	check  func(path, basename, contents string) bool
	reason string
}

var writeDenyRules = []writeRule{
	{
		check: func(_, basename, _ string) bool {
			return envFileRe.MatchString(basename) && !envExemptRe.MatchString(basename)
		},
		reason: "write to env file (may contain secrets)",
	},
	{
		check:  func(path, _, _ string) bool { return sshKeyRe.MatchString(path) },
		reason: "write to SSH key file",
	},
	{
		check:  func(_, basename, _ string) bool { return certKeyRe.MatchString(basename) },
		reason: "write to certificate/key file",
	},
	{
		check:  func(_, basename, _ string) bool { return credentialsRe.MatchString(basename) },
		reason: "write to credentials file",
	},
	{
		check:  func(_, basename, _ string) bool { return secretsRe.MatchString(basename) },
		reason: "write to secrets file",
	},
	{
		check:  func(_, basename, _ string) bool { return basename == ".npmrc" },
		reason: "write to .npmrc (may contain auth tokens)",
	},
	{
		check:  func(_, basename, _ string) bool { return basename == ".pypirc" },
		reason: "write to .pypirc (may contain auth tokens)",
	},
	{
		check:  func(path, _, _ string) bool { return kubeconfigRe.MatchString(path) },
		reason: "write to kubeconfig",
	},
	{
		check: func(_, _, contents string) bool {
			return strings.Contains(contents, `"type"`) && strings.Contains(contents, "service_account")
		},
		reason: "file appears to contain a service account key",
	},
	{
		check:  func(_, basename, _ string) bool { return basename == ".htpasswd" },
		reason: "write to .htpasswd",
	},
	{
		check:  func(_, basename, _ string) bool { return tfvarsRe.MatchString(basename) },
		reason: "write to .tfvars file (may contain secrets)",
	},
}

// Suppress unused warning — the compiled regex is used via strings.Contains in the rule above,
// but we keep serviceAcctRe for potential future use.
var _ = serviceAcctRe

// ValidateWrite is a preToolUse hook that blocks writes to sensitive files.
func ValidateWrite(input HookInput) (HookResult, int) {
	if input.ToolName != "Write" {
		return Allow(), 0
	}

	path := input.Path()
	if path == "" {
		return Allow(), 0
	}

	basename := filepath.Base(path)
	contents := input.Contents()

	for _, rule := range writeDenyRules {
		if rule.check(path, basename, contents) {
			return Deny("Blocked: " + rule.reason), 2
		}
	}

	return Allow(), 0
}
