package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
	"path/filepath"
)

var defaultBanned = map[string][]string{
	".go": {"os/exec", "reflect", "fmt.Println"},
	".py": {"os.system", "eval(", "exec("},
	".js": {"eval("},
}

func main() {
	if hooks.IsHookDisabled("import-guard") {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}
	var allowedPatterns map[string][]string
	path := os.Getenv("HOOK_ALLOWLISTS_PATH")
	if path == "" {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, ".cursor", "hooks-allowlists.json")
	}
	if data, err := os.ReadFile(path); err == nil {
		var v struct {
			IG *struct {
				AllowedPatterns map[string][]string `json:"allowedPatterns"`
			} `json:"importGuard"`
		}
		if json.Unmarshal(data, &v) == nil && v.IG != nil {
			allowedPatterns = v.IG.AllowedPatterns
		}
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}
	var input hooks.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}
	result, code := hooks.ImportGuardWithAllowlist(input, defaultBanned, allowedPatterns)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
