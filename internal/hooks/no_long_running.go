package hooks

import "regexp"

var longRunningRules = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	// Node ecosystem
	{regexp.MustCompile(`\b(?:npm|yarn|pnpm)\s+run\s+(?:dev|start|serve|watch)\b`), "long-running dev server (use build/test instead)"},
	{regexp.MustCompile(`\b(?:npm|yarn|pnpm)\s+start\b`), "long-running process (start)"},
	{regexp.MustCompile(`\b(?:yarn|pnpm)\s+(?:dev|serve|watch)\b`), "long-running dev server"},
	{regexp.MustCompile(`\bnpx\s+(?:next|vite|nuxt|remix|astro)\s+dev\b`), "long-running dev server"},
	{regexp.MustCompile(`\bnodemon\b`), "nodemon is a long-running watcher"},

	// Python servers
	{regexp.MustCompile(`\bpython[23]?\s+-m\s+http\.server\b`), "python http.server is long-running"},
	{regexp.MustCompile(`\b(?:flask\s+run|uvicorn\s|gunicorn\s)`), "long-running Python server"},

	// Go
	{regexp.MustCompile(`\bgo\s+run\b.*server`), "looks like a long-running Go server"},
	{regexp.MustCompile(`^\s*air\s*$`), "air is a long-running Go hot-reloader"},

	// Rust
	{regexp.MustCompile(`\bcargo\s+watch\b`), "cargo watch is long-running"},

	// Docker foreground (without -d)
	// Handled specially below

	// File watchers
	{regexp.MustCompile(`\b(?:fswatch|inotifywait)\b`), "file watcher is long-running"},

	// tail -f
	{regexp.MustCompile(`\btail\s+.*-[a-zA-Z]*f`), "tail -f is long-running (use tail -n instead)"},

	// Static site generators
	{regexp.MustCompile(`\b(?:hugo|jekyll|gatsby)\s+(?:server|serve)\b`), "long-running static site dev server"},
}

var (
	dockerComposeUpRe = regexp.MustCompile(`\bdocker(?:\s+|-)+compose\s+up\b`)
	detachedFlagRe    = regexp.MustCompile(`\s-d\b`)
)

// NoLongRunning is a preToolUse hook that blocks long-running foreground processes.
func NoLongRunning(input HookInput) (HookResult, int) {
	if input.ToolName != "Shell" {
		return Allow(), 0
	}

	cmd := input.Command()
	if cmd == "" {
		return Allow(), 0
	}

	for _, rule := range longRunningRules {
		if rule.pattern.MatchString(cmd) {
			return Deny("Blocked: " + rule.reason), 2
		}
	}

	// Special case: docker compose up without -d
	if dockerComposeUpRe.MatchString(cmd) && !detachedFlagRe.MatchString(cmd) {
		return Deny("Blocked: docker compose up without -d (add -d for detached)"), 2
	}

	return Allow(), 0
}
