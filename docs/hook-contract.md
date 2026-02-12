# Hook contract

Any agent or wrapper that invokes these hook binaries must follow this contract. The same binaries are used by Cursor, Claude (in Cursor), and OpenCode (via the adapter plugin).

## Stdin

Single JSON object, one line (or newline-terminated):

```json
{"tool_name": "Shell", "tool_input": {"command": "ls"}}
```

- **tool_name** (string): Event or tool identifier. For tool hooks: `Shell`, `Write`, `Read`, `Edit`, `MultiEdit`, `Grep`, etc. For lifecycle events: `SessionStart`, `beforeSubmitPrompt`, `Stop`, `PreCompact`, `SessionEnd`.
- **tool_input** (object): Payload for the hook. Shape depends on tool/event.

### tool_input shapes

- **Shell**: `{"command": "<string>"}`
- **Write**: `{"path": "<string>", "contents": "<string>"}`
- **Read**: `{"path": "<string>"}`
- **Edit / MultiEdit**: agent-specific (path and edit descriptors).
- **Grep**: `{"pattern": "<string>", ...}`
- **SessionStart / beforeSubmitPrompt**: `{"session_id": "<string>", "cwd": "<string>", ...}`; beforeSubmitPrompt may include `{"prompt": "<string>", ...}`.
- **Stop**: `{"transcript_path": "<string>", "file_path": "<string>", "stop_hook_active": <bool>, ...}`

## Stdout

Single JSON object, one line:

```json
{"decision": "allow"}
```

or to block:

```json
{"decision": "deny", "reason": "explanation"}
```

- **decision** (string, optional): `allow` | `deny`. Omitted for lifecycle-only hooks (SessionStart, SessionEnd, Stop, PreCompact) that do not gate actions; use empty object or reason-only output.
- **reason** (string, optional): Shown to user when denying.
- **message** (string, optional): Informational.
- **lint_command** (string, optional): Command to run (e.g. for fix suggestions).

## Exit codes

- **0**: Allow (honor JSON `decision` if present).
- **2**: Block the action (same as `decision: "deny"`).
- **Other**: Fail-open; the action proceeds.

On stdin parse errors, hooks must output `{"decision": "allow"}` and exit 0 (fail-open).

## Event names (config → agents)

| Config event         | Cursor hooks.json      | Claude settings.json | OpenCode plugin      |
|----------------------|------------------------|----------------------|----------------------|
| sessionStart         | sessionStart           | SessionStart         | session.* (future)   |
| beforeSubmitPrompt   | beforeSubmitPrompt     | UserPromptSubmit     | (future)             |
| preToolUse           | preToolUse             | PreToolUse           | tool.execute.before  |
| postToolUse          | postToolUse            | PostToolUse          | tool.execute.after   |
| stop                 | stop                   | Stop                 | (future)             |
| preCompact           | preCompact             | PreCompact           | (future)             |
| sessionEnd           | sessionEnd             | SessionEnd           | (future)             |

## Tool name mapping (agents using different names)

Agents that use different tool names (e.g. OpenCode) must map to this contract when calling the same binaries:

| This contract | OpenCode (example) |
|----------------|--------------------|
| Shell          | bash               |
| Write          | write              |
| Read           | read               |
| Edit           | edit               |
| MultiEdit      | (edit or multi)     |

When generating stdin for a hook, use the contract names (Shell, Write, Read, Edit) in `tool_name`.

## Supported agents

- **Cursor**: Uses `.cursor/hooks.json`; all events supported. When using the Codex model (e.g. GPT-5.3 Codex) in Cursor, the same hooks apply.
- **Claude (in Cursor)**: Uses `.claude/settings.json`; same hook binaries and contract.
- **OpenCode**: Via the generated adapter plugin (`.opencode/plugins/cursor-hooks-adapter.js`) and manifest; preToolUse/postToolUse map to `tool.execute.before` / `tool.execute.after`. Session/lifecycle hooks are future work.
- **Other agents**: Can support these hooks by invoking the binaries with the same stdin/stdout/exit contract; add a gen-config backend when the agent has a documented config schema.
