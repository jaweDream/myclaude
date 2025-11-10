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

```bash
bash scripts/codex.sh "<task>" [model] [working_dir]
```

**Timeout**: Set `timeout: 7200000` (2 hours) in Bash tool for long tasks.

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

### Examples

**Basic code analysis:**
```bash
bash scripts/codex.sh "explain @src/main.ts"
```

**Refactoring with specific model:**
```bash
bash scripts/codex.sh "refactor @src/utils for performance" "gpt-5"
```

**Multi-file analysis:**
```bash
bash scripts/codex.sh "analyze @. and find security issues" "gpt-5-codex" "/path/to/project"
```

**Quick task:**
```bash
bash scripts/codex.sh "add comments to @utils.js" "gpt-5-codex"
```

## Notes

- Runs with `--dangerously-bypass-approvals-and-sandbox` for automation
- Uses `--skip-git-repo-check` to work in any directory
- Streams progress, returns only final agent message
- Requires Codex CLI installed and authenticated
- Use `timeout: 7200000` (2 hours) for complex tasks that may take longer
