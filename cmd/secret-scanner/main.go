package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("secret-scanner", hooks.SecretScanner)
}
