package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	if hooks.IsHookDisabled("codebase-map") {
		fmt.Println(`{}`)
		os.Exit(0)
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(`{}`)
		os.Exit(0)
	}

	var input hooks.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Println(`{}`)
		os.Exit(0)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(`{}`)
		os.Exit(0)
	}

	maxDepth := 3
	if v := os.Getenv("HOOK_CODEBASE_MAP_MAX_DEPTH"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d > 0 {
			maxDepth = d
		}
	}

	var includePatterns []string
	if v := os.Getenv("HOOK_CODEBASE_MAP_INCLUDE"); v != "" {
		includePatterns = strings.Split(v, ",")
		for i := range includePatterns {
			includePatterns[i] = strings.TrimSpace(includePatterns[i])
		}
	}

	result, code := hooks.CodebaseMap(input, cwd, maxDepth, includePatterns)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
