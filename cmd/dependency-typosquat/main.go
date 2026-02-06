package main

import (
	"encoding/json"
	"hooks/internal/hooks"
	"os"
	"path/filepath"
)

func main() {
	var allowedPackages []string
	path := os.Getenv("HOOK_ALLOWLISTS_PATH")
	if path == "" {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, ".cursor", "hooks-allowlists.json")
	}
	if data, err := os.ReadFile(path); err == nil {
		var v struct {
			DT *struct {
				AllowedPackages []string `json:"allowedPackages"`
			} `json:"dependencyTyposquat"`
		}
		if json.Unmarshal(data, &v) == nil && v.DT != nil {
			allowedPackages = v.DT.AllowedPackages
		}
	}
	hooks.RunOrDisabled("dependency-typosquat", func(input hooks.HookInput) (hooks.HookResult, int) {
		return hooks.DependencyTyposquatWithAllowlist(input, allowedPackages)
	})
}
