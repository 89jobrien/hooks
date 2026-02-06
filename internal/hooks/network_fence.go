package hooks

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	// Matches curl/wget with a URL
	netCmdRe = regexp.MustCompile(`\b(curl|wget)\s+.*?(https?://[^\s"']+)`)
	// Direct URL extraction
	urlExtractRe = regexp.MustCompile(`https?://[^\s"']+`)
)

// Allowlisted domains for network access
var allowedDomains = []string{
	"localhost",
	"127.0.0.1",
	"::1",
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
	"registry.npmjs.org",
	"npmjs.com",
	"pypi.org",
	"files.pythonhosted.org",
	"pkg.go.dev",
	"proxy.golang.org",
	"sum.golang.org",
	"hub.docker.com",
	"registry.hub.docker.com",
	"docker.io",
	"ghcr.io",
	"crates.io",
	"rubygems.org",
	"repo.maven.apache.org",
	"dl.google.com",
	"storage.googleapis.com",
	"releases.hashicorp.com",
}

// NetworkFence is a preToolUse hook that blocks curl/wget to non-allowlisted domains.
func NetworkFence(input HookInput) (HookResult, int) {
	return NetworkFenceWithAllowlist(input, nil)
}

// NetworkFenceWithAllowlist uses custom allowedDomains; if nil or empty, uses built-in list.
func NetworkFenceWithAllowlist(input HookInput, customDomains []string) (HookResult, int) {
	if input.ToolName != "Shell" {
		return Allow(), 0
	}

	cmd := input.Command()
	if cmd == "" {
		return Allow(), 0
	}

	if !strings.Contains(cmd, "curl") && !strings.Contains(cmd, "wget") {
		return Allow(), 0
	}

	list := customDomains
	if len(list) == 0 {
		list = allowedDomains
	}

	urls := urlExtractRe.FindAllString(cmd, -1)
	for _, rawURL := range urls {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		host := parsed.Hostname()
		if !isDomainAllowedWith(host, list) {
			return Deny("Blocked: network request to non-allowlisted host: " + host), 2
		}
	}

	return Allow(), 0
}

func isDomainAllowed(host string) bool {
	return isDomainAllowedWith(host, allowedDomains)
}

func isDomainAllowedWith(host string, list []string) bool {
	for _, allowed := range list {
		if host == allowed {
			return true
		}
		if strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}
