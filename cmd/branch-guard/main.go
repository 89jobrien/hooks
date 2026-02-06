package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if hooks.IsHookDisabled("branch-guard") {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
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

	if !optIn("HOOK_BRANCH_GUARD") {
		fmt.Println(`{"decision": "allow"}`)
		os.Exit(0)
	}

	protected := []string{"main", "master"}
	if v := os.Getenv("HOOK_PROTECTED_BRANCHES"); v != "" {
		protected = strings.Split(v, ",")
	}

	currentBranch := getCurrentBranch()

	result, code := hooks.BranchGuard(input, protected, currentBranch)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}

func optIn(name string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	return v == "1" || v == "true" || v == "yes"
}

func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
