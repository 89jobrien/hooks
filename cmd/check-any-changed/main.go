package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("check-any-changed", hooks.CheckAnyChanged)
}
