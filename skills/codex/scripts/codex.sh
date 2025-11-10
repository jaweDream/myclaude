#!/bin/bash
set -euo pipefail

TASK="${1:-}"
MODEL="${2:-gpt-5-codex}"
WORKDIR="${3:-.}"

if [ -z "$TASK" ]; then
  echo "ERROR: Task required" >&2
  exit 1
fi

codex e -m "$MODEL" --dangerously-bypass-approvals-and-sandbox --skip-git-repo-check -C "$WORKDIR" --json "$TASK" 2>/dev/null | \
  jq -r 'select(.type == "item.completed" and .item.type == "agent_message") | .item.text' | \
  tail -n 1
