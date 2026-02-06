package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("validate-shell", hooks.ValidateShell)
}
