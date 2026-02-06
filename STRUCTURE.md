# Hooks codebase structure

## Layout

- **Module**: single Go module `hooks` (repo root).
- **Hook logic**: one package `hooks` in `internal/hooks`. Each hook is a pure function `func X(input HookInput, ...opts) (HookResult, int)` in its own file pair `*_hook.go` + `*_hook_test.go`.
- **Binaries**: `cmd/<hook-name>/main.go` per hook (22 hooks) plus `cmd/gen-config/` (config generator). Built by Makefile; each binary depends on `cmd/%/main.go` and `internal/hooks/*.go`.
- **Shared**: `internal/hooks/hookutil.go` — `HookInput`, `HookResult`, `ReadInput`, `IsHookDisabled`, `Run`, `RunOrDisabled`.
- **Config**: `config.yaml` → gen-config → `.cursor/hooks.json` and `.claude/settings.json`. Hooks read env (e.g. `HOOK_AUDIT_DIR`, `HOOK_DISABLED`) in main.

```
hooks/
├── go.mod
├── config.yaml
├── Makefile
├── install.sh
├── scripts/
├── internal/
│   └── hooks/
│       ├── hookutil.go          # HookInput, HookResult, Run, RunOrDisabled, ReadInput, IsHookDisabled
│       ├── hookutil_test.go
│       ├── audit.go
│       ├── audit_test.go
│       ├── branch_guard.go
│       ├── branch_guard_test.go
│       └── ...                  # one .go + _test.go per hook
└── cmd/
    ├── audit/main.go
    ├── branch-guard/main.go
    ├── ...              # one dir per hook binary
    ├── gen-config/main.go
    └── gen-config/gen_config_test.go
```

## Contract

- **Stdin**: `{"tool_name": "Shell"|"Write"|..., "tool_input": {...}}`
- **Stdout**: `{"decision": "allow"|"deny", "reason": "...", "message": "...", "lint_command": "..."}`
- **Exit 0**: allow (use JSON output). **Exit 2**: block. Other exit: fail-open (action proceeds).
- **Fail-open**: On stdin/parse error, output `{"decision": "allow"}` and exit 0.

## Conventions

1. **Single package `hooks`** — Hook logic lives in `internal/hooks` as pure functions. No split into multiple packages for hooks.
2. **One file per hook (+ test)** — e.g. `audit.go` / `audit_test.go`. Optional: group related hooks (e.g. time_tracker start+end) in one file later if desired.
3. **One binary per hook** — `cmd/<name>/main.go`. Thin wrapper: parse stdin, read env and build options, call hook, marshal result, exit. Use `hooks.RunOrDisabled(name, hooks.Fn)` when the hook takes only `HookInput`; otherwise hand-roll and pass options from env.
4. **gen-config** — Separate tool (no hook contract). Stays under `cmd/gen-config/`.

## Do / don't

- **Do**: Keep one `hooks` package, one `cmd/<hook>/main.go` per hook, hookutil as the single place for contract and `Run`/`RunOrDisabled`.
- **Do**: Use `RunOrDisabled` for hooks that need only `HookInput` (e.g. validate-shell, no-long-running, validate-write, secret-scanner, lint-on-write).
- **Don't**: Split hook logic into multiple packages or merge all mains into one binary unless there is a clear need.
