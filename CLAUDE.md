# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make all # Build all hook binaries to bin/
make test # Run tests: go test -v -count=1 ./...
make config # Generate .cursor/hooks.json and .claude/settings.json from config.yaml
make clean # Remove bin/

# Run a single test
go test -v -run TestValidateShell_Blocks ./internal/hooks/

# Build a single hook binary
go build -o bin/validate-shell ./cmd/validate-shell

# Format check (enforced in CI)
gofmt -l .
```

## Architecture

Single Go module (`hooks`). One external dependency: `gopkg.in/yaml.v3`.

### Core layout

- **`internal/hooks/`** — Single `hooks` package with all hook logic. Each hook is a pure function in its own file pair: `<hook>.go` + `<hook>_test.go`.
- **`internal/hooks/hookutil.go`** — Shared types and helpers: `HookInput`, `HookResult`, `ReadInput`, `Run`, `RunOrDisabled`, `IsHookDisabled`, result constructors (`Allow`, `Deny`, `NoOp`, `NoOpMsg`).
- **`cmd/<hook-name>/main.go`** — One thin binary per hook. Reads JSON from stdin, calls hook function, writes JSON to stdout, exits.
- **`cmd/gen-config/`** — Reads `config.yaml` and generates `.cursor/hooks.json`, `.claude/settings.json`, and `.cursor/hooks.env`.
- **`cmd/interactive/`** — TUI for toggling hooks on/off in `config.yaml`.
- **`internal/config/`** — YAML config loading/saving, config path discovery.
- **`config.yaml`** — Source of truth for hook configuration. Generated JSON files are outputs, not manually edited.

### Hook contract

- **Stdin**: `{"tool_name": "Shell"|"Write"|..., "tool_input": {...}}`
- **Stdout**: JSON matching the hook event schema
- **Exit 0**: allow. **Exit 2**: block. **Other**: fail-open.
- PreToolUse/PostToolUse hooks use `{"decision": "allow"|"deny", "reason": "..."}`.
- Lifecycle hooks (SessionStart, SessionEnd, Stop, PreCompact) must NOT include `decision` in output — use `NoOp()`/`NoOpMsg()` instead of `Allow()`/`AllowMsg()`.

### Two patterns for cmd/main.go

1. **Simple hooks** (only need `HookInput`): use `hooks.RunOrDisabled("hook-name", hooks.HookFn)`.
2. **Hooks with options** (need env vars, cwd, etc.): hand-roll main — check `IsHookDisabled`, read env, call hook with extra args, marshal result.

`RunOrDisabled` outputs `{"decision": "allow"}` when disabled, so it's only valid for PreToolUse/PostToolUse hooks. Lifecycle hooks must output `{}` when disabled.

### Adding a new hook

1. Create `internal/hooks/my_hook.go` with `func MyHook(input HookInput) (HookResult, int)`
2. Create `internal/hooks/my_hook_test.go`
3. Create `cmd/my-hook/main.go`
4. Add `my-hook` to the `CMDS` list in `Makefile`
5. Add entry to `config.yaml` under the appropriate event

## Key Conventions

- All hook logic stays in the single `internal/hooks` package. Don't split into sub-packages.
- One file per hook (+ test file). Use table-driven tests.
- Hooks fail-open on errors: if stdin parsing fails, output allow and exit 0.
- Regex patterns are pre-compiled at package level.
- `gofmt` is enforced in CI. Conventional Commits enforced by `commit-msg-lint` hook.
- Environment variables: `HOOK_DISABLED` (comma-separated list), `HOOK_AUDIT_DIR`, `HOOK_PROTECTED_BRANCHES`, `HOOK_MAX_FILE_LINES`, etc. Hooks read env in their `cmd/` main, not in core logic.