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
- Tasks requiring specialized reasoning models (o3, gpt-5)

## Usage

通过 Bash tool 调用:
```bash
node ~/.claude/skills/codex/scripts/codex.js "<task>" [model] [working_dir]
```

## Timeout Control

- **Built-in**: Script enforces 2-hour timeout by default
- **Override**: Set `CODEX_TIMEOUT` environment variable (in milliseconds, e.g., `CODEX_TIMEOUT=3600000` for 1 hour)
- **Behavior**: On timeout, sends SIGTERM, then SIGKILL after 5s if process doesn't exit
- **Exit code**: Returns 124 on timeout (consistent with GNU timeout)
- **Bash tool**: Always set `timeout: 7200000` parameter for double protection

### Parameters

- `task` (required): Task description, supports `@file` references
- `model` (optional): Model to use (default: gpt-5-codex)
  - `gpt-5-codex`: Default, optimized for code
  - `gpt-5`: Fast general purpose
  - `o3`: Deep reasoning
  - `o4-mini`: Quick tasks
  - `codex-1`: Software engineering
- `working_dir` (optional): Working directory (default: current)

### Return Format

Extracts `agent_message` from Codex JSON stream:
```
Agent response text here...
```

Error format:
```
ERROR: Error message
```

### Invocation Pattern

When calling via Bash tool, always include the timeout parameter:
```
Bash tool parameters:
- command: node ~/.claude/skills/codex/scripts/codex.js "<task>" [model] [working_dir]
- timeout: 7200000
- description: <brief description of the task>
```

### Examples

**Basic code analysis:**
```bash
# Via Bash tool with timeout parameter
node ~/.claude/skills/codex/scripts/codex.js "explain @src/main.ts"
# timeout: 7200000
```

**Refactoring with specific model:**
```bash
node ~/.claude/skills/codex/scripts/codex.js "refactor @src/utils for performance" "gpt-5"
# timeout: 7200000
```

**Multi-file analysis:**
```bash
node ~/.claude/skills/codex/scripts/codex.js "analyze @. and find security issues" "gpt-5-codex" "/path/to/project"
# timeout: 7200000
```

**Quick task:**
```bash
node ~/.claude/skills/codex/scripts/codex.js "add comments to @utils.js" "gpt-5-codex"
# timeout: 7200000
```

## Notes

- Runs with `--dangerously-bypass-approvals-and-sandbox` for automation
- Uses `--skip-git-repo-check` to work in any directory
- Streams progress, returns only final agent message
- Requires Codex CLI installed and authenticated
