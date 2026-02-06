package hooks

import (
	"testing"
)

func TestDependencyTyposquat_Blocks(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		// npm typosquats
		{"npm lod-ash", "npm install lod-ash"},
		{"npm expres", "npm install expres"},
		{"npm requets", "npm install requets"},
		{"npm electorn", "npm install electorn"},
		{"npm crossenv", "npm install crossenv"},
		{"npm babelcli", "npm install babelcli"},
		// pip typosquats
		{"pip reqeusts", "pip install reqeusts"},
		{"pip python-nmap", "pip install python-nmap"},
		{"pip djago", "pip install djago"},
		{"pip flaask", "pip install flaask"},
		{"pip urlib3", "pip install urlib3"},
		// go typosquats are less common but test the pattern
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, code := DependencyTyposquat(shellInput(tt.cmd))
			if code != 2 {
				t.Errorf("expected block (exit 2), got %d for %q", code, tt.cmd)
			}
			if result.Decision != "deny" {
				t.Errorf("expected deny, got %q; reason: %s", result.Decision, result.Reason)
			}
		})
	}
}

func TestDependencyTyposquatWithAllowlist_AllowsListed(t *testing.T) {
	// allowedPackages overrides typosquat block
	result, code := DependencyTyposquatWithAllowlist(shellInput("npm install lod-ash"), []string{"lod-ash"})
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when lod-ash in allowlist, got code=%d decision=%q", code, result.Decision)
	}
	result, code = DependencyTyposquatWithAllowlist(shellInput("pip install reqeusts"), []string{"reqeusts"})
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow when reqeusts in allowlist, got code=%d decision=%q", code, result.Decision)
	}
}

func TestDependencyTyposquatWithAllowlist_BlocksWhenNotListed(t *testing.T) {
	result, code := DependencyTyposquatWithAllowlist(shellInput("npm install lod-ash"), []string{"other-pkg"})
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny when lod-ash not in allowlist, got code=%d decision=%q", code, result.Decision)
	}
}

func TestDependencyTyposquat_Allows(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"npm lodash", "npm install lodash"},
		{"npm express", "npm install express"},
		{"npm requests", "npm install requests"},
		{"npm electron", "npm install electron"},
		{"npm cross-env", "npm install cross-env"},
		{"npm @babel/cli", "npm install @babel/cli"},
		{"pip requests", "pip install requests"},
		{"pip django", "pip install django"},
		{"pip flask", "pip install flask"},
		{"pip urllib3", "pip install urllib3"},
		{"go get valid", "go get github.com/gin-gonic/gin"},
		{"non-install command", "npm test"},
		{"non-Shell tool", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input HookInput
			if tt.name == "non-Shell tool" {
				input = writeInput("main.go", "package main")
			} else {
				input = shellInput(tt.cmd)
			}
			result, code := DependencyTyposquat(input)
			if code != 0 {
				t.Errorf("expected allow (exit 0), got %d for %q; reason: %s", code, tt.cmd, result.Reason)
			}
		})
	}
}
