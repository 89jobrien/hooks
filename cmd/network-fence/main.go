package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"os"
	"path/filepath"
)

func main() {
	if hooks.IsHookDisabled("network-fence") {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}
	path := os.Getenv("HOOK_ALLOWLISTS_PATH")
	if path == "" {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, ".cursor", "hooks-allowlists.json")
	}

	data, err := os.ReadFile(path)
	var allowedDomains []string
	if err == nil {
		var v struct {
			NetworkFence *struct {
				AllowedDomains []string `json:"allowedDomains"`
			} `json:"networkFence"`
		}
		if json.Unmarshal(data, &v) == nil && v.NetworkFence != nil {
			allowedDomains = v.NetworkFence.AllowedDomains
		}
	}

	input, err := hooks.ReadInput(os.Stdin)
	if err != nil {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}

	result, code := hooks.NetworkFenceWithAllowlist(input, allowedDomains)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
