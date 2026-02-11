package hooks

import (
	"regexp"
	"strings"
)

var installRe = regexp.MustCompile(`\b(?:npm|yarn|pnpm)\s+(?:install|add|i)\s+([^\s-][^\s]*)`)
var pipInstallRe = regexp.MustCompile(`\b(?:pip|pip3)\s+install\s+([^\s-][^\s]*)`)

// Known typosquats: map of typosquat -> real package
var npmTyposquats = map[string]string{
	"lod-ash":       "lodash",
	"lodahs":        "lodash",
	"expres":        "express",
	"expresss":      "express",
	"requets":       "request",
	"reqeust":       "request",
	"electorn":      "electron",
	"electronjs":    "electron",
	"crossenv":      "cross-env",
	"cross_env":     "cross-env",
	"babelcli":      "@babel/cli",
	"babel-cli":     "@babel/cli",
	"coffe-script":  "coffeescript",
	"event-stream2": "event-stream",
	"gruntcli":      "grunt-cli",
	"mongose":       "mongoose",
	"node-fabric":   "fabric",
	"node-opencv":   "opencv",
	"node-opensl":   "openssl",
	"nodefabric":    "fabric",
	"nodesass":      "node-sass",
	"shadowsock":    "shadowsocks",
	"smb":           "samba",
	"discordi.js":   "discord.js",
	"colored":       "colors",
	"colors.js":     "colors",
}

var pipTyposquats = map[string]string{
	"reqeusts":      "requests",
	"requets":       "requests",
	"reequests":     "requests",
	"python-nmap":   "python-nmap",
	"djago":         "django",
	"djnago":        "django",
	"djangoo":       "django",
	"flaask":        "flask",
	"flaskk":        "flask",
	"urlib3":        "urllib3",
	"urrlib3":       "urllib3",
	"numppy":        "numpy",
	"nuumpy":        "numpy",
	"pandass":       "pandas",
	"scapy":         "scipy",
	"beutifulsoup":  "beautifulsoup4",
	"beautifulsoup": "beautifulsoup4",
	"dateutil":      "python-dateutil",
	"colourama":     "colorama",
}

func typosquatCheck(cmd string, allowedPackages map[string]bool) (deny bool, reason string) {
	// Check npm/yarn/pnpm installs
	if matches := installRe.FindAllStringSubmatch(cmd, -1); len(matches) > 0 {
		for _, m := range matches {
			pkg := strings.ToLower(m[1])
			if allowedPackages[pkg] {
				continue
			}
			if real, ok := npmTyposquats[pkg]; ok {
				return true, "Blocked: suspected typosquat package '" + m[1] + "' (did you mean '" + real + "'?)"
			}
		}
	}
	// Check pip installs
	if matches := pipInstallRe.FindAllStringSubmatch(cmd, -1); len(matches) > 0 {
		for _, m := range matches {
			pkg := strings.ToLower(m[1])
			if allowedPackages[pkg] {
				continue
			}
			if real, ok := pipTyposquats[pkg]; ok {
				return true, "Blocked: suspected typosquat package '" + m[1] + "' (did you mean '" + real + "'?)"
			}
		}
	}
	return false, ""
}

// DependencyTyposquat is a preToolUse hook that blocks known typosquat packages.
func DependencyTyposquat(input HookInput) (HookResult, int) {
	return DependencyTyposquatWithAllowlist(input, nil)
}

// DependencyTyposquatWithAllowlist runs typosquat check; packages in allowedPackages are allowed.
func DependencyTyposquatWithAllowlist(input HookInput, allowedPackages []string) (HookResult, int) {
	if input.ToolName != "Shell" {
		return Allow(), 0
	}
	cmd := input.Command()
	if cmd == "" {
		return Allow(), 0
	}
	allowed := make(map[string]bool)
	for _, p := range allowedPackages {
		allowed[strings.ToLower(p)] = true
	}
	if deny, reason := typosquatCheck(cmd, allowed); deny {
		return Deny(reason), 2
	}
	return Allow(), 0
}
