# Hooks

[![ci](https://github.com/89jobrien/hooks/actions/workflows/ci.yml/badge.svg)](https://github.com/89jobrien/hooks/actions/workflows/ci.yml)
[![release-please](https://github.com/89jobrien/hooks/actions/workflows/release-please.yml/badge.svg)](https://github.com/89jobrien/hooks/actions/workflows/release-please.yml)
[![docker-publish](https://github.com/89jobrien/hooks/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/89jobrien/hooks/actions/workflows/docker-publish.yml)

Go hooks for Cursor/Claude and other agents: security, quality, and session lifecycle.

**Supported agents:** Cursor (including when using the Codex model), Claude in Cursor, and OpenCode via the generated adapter plugin. The same hook binaries and contract work across agents. See [docs/hook-contract.md](docs/hook-contract.md) for the stdin/stdout/exit contract and tool-name mapping.

Codebase layout and conventions: [STRUCTURE.md](STRUCTURE.md).

## Commands

```bash
make all # build 26 binaries (hooks + gen-config + interactive) to bin/
make test # run tests (~0.5s)
make config # from repo root: generate .cursor/hooks.json, .claude/settings.json, .opencode/ (manifest + adapter plugin), and (if env in config) .cursor/hooks.env. Requires bin/ built first (make all).
make clean # remove bin/
```

From repo root: `./hooks/bin/interactive` (or from inside hooks: `bin/interactive`) for an interactive menu to enable/disable hooks, then save to regenerate config. `./hooks/bin/interactive scan` scans for project hooks (from cwd upward) and reports global hooks at `~/.cursor/hooks.json`.

## Container

Build:

```bash
docker build -t ghcr.io/89jobrien/hooks:local .
```

Run a hook:

```bash
docker run --rm ghcr.io/89jobrien/hooks:local audit
```

## Hooks (by event)

| Event | Hooks |
|-------|--------|
| sessionStart | session-guard, time-tracker-start |
| beforeSubmitPrompt | prompt-enricher |
| preToolUse | rate-limiter, dry-run-mode, validate-shell, no-long-running, network-fence, dependency-typosquat, validate-write, file-size-guard *(+ branch-guard, commit-msg-lint, no-sudo if opted in)* |
| postToolUse | audit, cost-estimator, secret-scanner, lint-on-write, test-buddy, import-guard, todo-tracker |
| stop | session-diary |
| preCompact | compact-snapshot |
| sessionEnd | time-tracker-end |

## Env (optional)

**Opt-in (default off)** — not in default config. To enable: uncomment the hook in `hooks/config.yaml` under `preToolUse`, run `make -C hooks config`, then set the env to `1`/`true`/`yes`:

| Var | Hook |
|-----|------|
| `HOOK_BRANCH_GUARD` | branch-guard |
| `HOOK_COMMIT_MSG_LINT` | commit-msg-lint |
| `HOOK_NO_SUDO` | no-sudo |

**Other env:**

| Var | Hook(s) | Default |
|-----|---------|---------|
| `HOOK_AUDIT_DIR` | audit, session-diary, compact-snapshot | `~/.cursor/audit` |
| `HOOK_MAX_FILE_LINES` | file-size-guard | 500 |
| `HOOK_PROTECTED_BRANCHES` | branch-guard | main,master |
| `HOOK_RATE_LIMIT` | rate-limiter | 30 |
| `HOOKS_DRY_RUN` | dry-run-mode | 0 → allow; 1 → block shell, log |
| `HOOK_TODO_DIR` | todo-tracker | `~/.cursor/todos` |
| `HOOK_TIME_DIR` | time-tracker-* | `~/.cursor/time` |
| `HOOK_DIARY_DIR` | session-diary | `~/.cursor/diary` |
| `HOOK_SNAPSHOT_DIR` | compact-snapshot | `~/.cursor/snapshots` |
| `HOOK_COST_DIR` | cost-estimator | `~/.cursor/cost` |
| `HOOK_DRY_RUN_DIR` | dry-run-mode | `~/.cursor/dry-run` |
| `HOOK_RATE_DIR` | rate-limiter | `~/.cursor/rate` |

## Add a new hook

1. **Test**: `hooks/internal/hooks/my_hook_test.go` — table-driven, use `shellInput()` / `writeInput()`.
2. **Impl**: `hooks/internal/hooks/my_hook.go` — `func MyHook(input HookInput) (HookResult, int)`.
3. **Binary**: `hooks/cmd/my-hook/main.go` — read stdin, call `hooks.MyHook`, write JSON, `os.Exit(code)`.
4. **Build**: Add `my-hook` to `CMDS` in `Makefile`.
5. **Config**: Add entry to `hooks/config.yaml` under the right event, then run `make -C hooks config`.

## Config

- **Source of truth**: `hooks/config.yaml` (YAML). Edit this; do not edit the JSON by hand.
- **Disable a hook**: set `enabled: false` on that entry (object form). Omitted from generated JSON and not validated as a binary.
- **Interactive mode**: run `./hooks/bin/interactive` from repo root (or `bin/interactive` from inside hooks). Use the menu to toggle hooks on/off (`t <n>`), then `s` to save and run gen-config, which regenerates `.cursor/hooks.json` and `.claude/settings.json`. Use `q` to quit without saving.
- **Generate**: from repo root run `make -C hooks config` (after `make -C hooks all`) for hooks-as-subdir, or `./.hooks/bin/gen-config` for installed `.hooks/` layout. By default writes:
 - `.cursor/hooks.json` (Cursor)
 - `.claude/settings.json` (Claude; enable Third-party skills in Cursor)
 - `.opencode/hooks-manifest.json` and `.opencode/plugins/cursor-hooks-adapter.js` (OpenCode; preToolUse/postToolUse via plugin)
 - `.cursor/hooks.env` (only if `env:` is set in config.yaml; source before Cursor to set per-hook env)
- **Backends**: optional `output.backends: [cursor, claude, opencode]` in config limits which outputs are generated; empty = all. Optional `output.openCodeDir` (default `.opencode`) sets the OpenCode output directory.
- **Validation**: gen-config checks that every hook name in config has a binary under `hooks/bin/`. Run `make all` before `make config`.

## Externalized allowlists (YAML)

Optional top-level `allowlists:` in `config.yaml`. gen-config writes `.cursor/hooks-allowlists.json`. **network-fence** reads `HOOK_ALLOWLISTS_PATH` (default `.cursor/hooks-allowlists.json`) and uses `networkFence.allowedDomains`; if missing, uses built-in list. import-guard and dependency-typosquat still use built-in lists (format TBD).

## Per-hook options (YAML)

In `config.yaml` add an optional top-level `env:` map. Keys are env var names (e.g. `HOOK_MAX_FILE_LINES`, `HOOK_PROTECTED_BRANCHES`, `HOOK_BRANCH_GUARD`). Values are written to `.cursor/hooks.env`. Source that file before starting Cursor (e.g. `source .cursor/hooks.env && cursor .`) so hooks see the vars.

## Install into another repo

From this repo root: `./install.sh /path/to/target/repo`. install.sh builds here, copies only `bin/` and `config.yaml` into `target/.hooks/`, then runs `target/.hooks/bin/gen-config`. To regenerate later from the target repo: `./.hooks/bin/gen-config` from target root. Optional: source the target repo’s `.cursor/hooks.env` before Cursor.

**Setup layouts:**

- **Installed copy** (install.sh): config at `.hooks/config.yaml`, binaries in `.hooks/bin/`. Regen with `./.hooks/bin/gen-config` from repo root.
- **Hooks as subdirectory** (e.g. submodule at `hooks/`): config at `hooks/config.yaml`, binaries in `hooks/bin/`. Regen with `make -C hooks config` from repo root. `sync-config.sh` is for this layout.

## Sync config after editing YAML

- **Hooks as subdir**: run `./hooks/scripts/sync-config.sh` from repo root or `make -C hooks config`.
- **Installed `.hooks/` layout**: run `./.hooks/bin/gen-config` from repo root.

If you use [pre-commit](https://pre-commit.com/), add a hook that runs when config changes: for subdir use `bash hooks/scripts/sync-config.sh`; for `.hooks/` use `./.hooks/bin/gen-config`.

## Summary (audit / cost)

From repo root: `make -C hooks summary`. Prints approximate tool-call count (from audit logs modified in last 24h) and token total from `~/.cursor/cost/cost.log`. Override dirs with `HOOK_AUDIT_DIR` and `HOOK_COST_DIR`.

## CI

CI runs `gofmt` and `go test` on push and PRs. Release automation is handled by release-please on `main`, which opens a PR to bump the version and update `CHANGELOG.md`, then creates a GitHub Release when merged.