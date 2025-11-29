---
name: codex
description: Execute Codex CLI for code analysis, refactoring, and automated code changes. Use when you need to delegate complex code tasks to Codex AI with file references (@syntax) and structured output.
---

# Codex CLI Integration

## Overview

Execute Codex CLI commands and parse structured JSON responses. Supports file references via `@` syntax, multiple models, and sandbox controls.

## When to Use

- Complex code analysis requiring deep understanding
- Large-scale refactoring across multiple files
- Automated code generation with safety controls

## Fallback Policy

Codex is the **primary execution method** for all code edits and tests. Direct execution is only permitted when:

1. Codex is unavailable (service down, network issues)
2. Codex fails **twice consecutively** on the same task

When falling back to direct execution:
- Log `CODEX_FALLBACK` with the reason
- Retry Codex on the next task (don't permanently switch)
- Document the fallback in the final summary

## Usage

**Mandatory**: Run every automated invocation through the Bash tool in the foreground with **HEREDOC syntax** to avoid shell quoting issues, keeping the `timeout` parameter fixed at `7200000` milliseconds (do not change it or use any other entry point).

```bash
codex-wrapper - [working_dir] <<'EOF'
<task content here>
EOF
```

**Why HEREDOC?** Tasks often contain code blocks, nested quotes, shell metacharacters (`$`, `` ` ``, `\`), and multiline text. HEREDOC (Here Document) syntax passes these safely without shell interpretation, eliminating quote-escaping nightmares.

**Foreground only (no background/BashOutput)**: Never set `background: true`, never accept Claude's "Running in the background" mode, and avoid `BashOutput` streaming loops. Keep a single foreground Bash call per Codex task; if work might be long, split it into smaller foreground runs instead of offloading to background execution.

**Simple tasks** (backward compatibility):
For simple single-line tasks without special characters, you can still use direct quoting:
```bash
codex-wrapper "simple task here" [working_dir]
```

**Resume a session with HEREDOC:**
```bash
codex-wrapper resume <session_id> - [working_dir] <<'EOF'
<task content>
EOF
```

**Cross-platform notes:**
- **Bash/Zsh**: Use `<<'EOF'` (single quotes prevent variable expansion)
- **PowerShell 5.1+**: Use `@'` and `'@` (here-string syntax)
  ```powershell
  codex-wrapper - @'
  task content
  '@
  ```

## Environment Variables

- **CODEX_TIMEOUT**: Override timeout in milliseconds (default: 7200000 = 2 hours)
  - Example: `export CODEX_TIMEOUT=3600000` for 1 hour

## Timeout Control

- **Built-in**: Binary enforces 2-hour timeout by default
- **Override**: Set `CODEX_TIMEOUT` environment variable (in milliseconds, e.g., `CODEX_TIMEOUT=3600000` for 1 hour)
- **Behavior**: On timeout, sends SIGTERM, then SIGKILL after 5s if process doesn't exit
- **Exit code**: Returns 124 on timeout (consistent with GNU timeout)
- **Bash tool**: Always set `timeout: 7200000` parameter for double protection

### Parameters

- `task` (required): Task description, supports `@file` references
- `working_dir` (optional): Working directory (default: current)

### Return Format

Extracts `agent_message` from Codex JSON stream and appends session ID:
```
Agent response text here...

---
SESSION_ID: 019a7247-ac9d-71f3-89e2-a823dbd8fd14
```

Error format (stderr):
```
ERROR: Error message
```

Return only the final agent message and session ID—do not paste raw `BashOutput` logs or background-task chatter into the conversation.

### Invocation Pattern

All automated executions must use HEREDOC syntax through the Bash tool in the foreground, with `timeout` fixed at `7200000` (non-negotiable):

```
Bash tool parameters:
- command: codex-wrapper - [working_dir] <<'EOF'
  <task content>
  EOF
- timeout: 7200000
- description: <brief description of the task>
```

Run every call in the foreground—never append `&` to background it—so logs and errors stay visible for timely interruption or diagnosis.

**Important:** Use HEREDOC (`<<'EOF'`) for all but the simplest tasks. This prevents shell interpretation of quotes, variables, and special characters.

### Examples

**Basic code analysis:**
```bash
# Recommended: with HEREDOC (handles any special characters)
codex-wrapper - <<'EOF'
explain @src/main.ts
EOF
# timeout: 7200000

# Alternative: simple direct quoting (if task is simple)
codex-wrapper "explain @src/main.ts"
```

**Refactoring with multiline instructions:**
```bash
codex-wrapper - <<'EOF'
refactor @src/utils for performance:
- Extract duplicate code into helpers
- Use memoization for expensive calculations
- Add inline comments for non-obvious logic
EOF
# timeout: 7200000
```

**Multi-file analysis:**
```bash
codex-wrapper - "/path/to/project" <<'EOF'
analyze @. and find security issues:
1. Check for SQL injection vulnerabilities
2. Identify XSS risks in templates
3. Review authentication/authorization logic
4. Flag hardcoded credentials or secrets
EOF
# timeout: 7200000
```

**Resume previous session:**
```bash
# First session
codex-wrapper - <<'EOF'
add comments to @utils.js explaining the caching logic
EOF
# Output includes: SESSION_ID: 019a7247-ac9d-71f3-89e2-a823dbd8fd14

# Continue the conversation with more context
codex-wrapper resume 019a7247-ac9d-71f3-89e2-a823dbd8fd14 - <<'EOF'
now add TypeScript type hints and handle edge cases where cache is null
EOF
# timeout: 7200000
```

**Task with code snippets and special characters:**
```bash
codex-wrapper - <<'EOF'
Fix the bug in @app.js where the regex /\d+/ doesn't match "123"
The current code is:
  const re = /\d+/;
  if (re.test(input)) { ... }
Add proper escaping and handle $variables correctly.
EOF
```

### Large Task Protocol

- For every large task, first produce a canonical task list that enumerates the Task ID, description, file/directory scope, dependencies, test commands, and the expected Codex Bash invocation.
- Tasks without dependencies should be executed concurrently via multiple foreground Bash calls (you can keep separate terminal windows) and each run must log start/end times plus any shared resource usage.
- Reuse context aggressively (such as @spec.md or prior analysis output), and after concurrent execution finishes, reconcile against the task list to report which items completed and which slipped.

| ID | Description | Scope | Dependencies | Tests | Command |
| --- | --- | --- | --- | --- | --- |
| T1 | Review @spec.md to extract requirements | docs/, @spec.md | None | None | `codex-wrapper - <<'EOF'`<br/>`analyze requirements @spec.md`<br/>`EOF` |
| T2 | Implement the module and add test cases | src/module | T1 | npm test -- --runInBand | `codex-wrapper - <<'EOF'`<br/>`implement and test @src/module`<br/>`EOF` |

## Notes

- **Binary distribution**: Single Go binary, zero dependencies
- **Installation**: Download from GitHub Releases or use install.sh
- **Cross-platform compatible**: Linux (amd64/arm64), macOS (amd64/arm64)
- All automated runs must use the Bash tool with the fixed timeout to provide dual timeout protection and unified logging/exit semantics
for automation (new sessions only)
- Uses `--skip-git-repo-check` to work in any directory
- Streams progress, returns only final agent message
- Every execution returns a session ID for resuming conversations
- Requires Codex CLI installed and authenticated
