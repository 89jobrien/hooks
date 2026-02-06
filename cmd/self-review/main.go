package main

import "hooks/internal/hooks"

func main() {
	hooks.RunOrDisabled("self-review", hooks.SelfReview)
}
