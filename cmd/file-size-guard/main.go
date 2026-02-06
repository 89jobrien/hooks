package main

import (
	"encoding/json"
	"fmt"
	"hooks/internal/hooks"
	"io"
	"os"
	"strconv"
)

func main() {
	if hooks.IsHookDisabled("file-size-guard") {
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

	maxLines := 500
	if v := os.Getenv("HOOK_MAX_FILE_LINES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxLines = n
		}
	}

	result, code := hooks.FileSizeGuard(input, maxLines)
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	os.Exit(code)
}
