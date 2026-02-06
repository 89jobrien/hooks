package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("no-long-running", hooks.NoLongRunning)
}
