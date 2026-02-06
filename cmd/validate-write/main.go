package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("validate-write", hooks.ValidateWrite)
}
