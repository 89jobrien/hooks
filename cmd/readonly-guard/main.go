package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("readonly-guard", hooks.ReadonlyGuard)
}
