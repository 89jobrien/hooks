package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("lint-on-write", hooks.LintOnWrite)
}
