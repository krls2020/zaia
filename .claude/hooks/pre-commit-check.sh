#!/bin/bash
# PreToolUse hook: Before git commit, check CLAUDE.md is staged

input=$(cat)
COMMAND=$(echo "$input" | grep -o '"command":"[^"]*"' | head -1 | sed 's/"command":"//;s/"//')

echo "$COMMAND" | grep -qE 'git commit' || exit 0

MODULE_ROOT=$(cd "$(dirname "$0")/../.." && pwd)
cd "$MODULE_ROOT" || exit 0

STAGED=$(git diff --cached --name-only 2>/dev/null)
[ -z "$STAGED" ] && exit 0

KEY_PATTERNS="internal/platform/client\.go|internal/commands/root\.go|internal/knowledge/engine\.go|internal/output/envelope\.go|internal/auth/manager\.go|internal/platform/errors\.go|go\.mod"

KEY_FILES_CHANGED=$(echo "$STAGED" | grep -E "$KEY_PATTERNS" | head -5)
CLAUDE_MD_CHANGED=$(echo "$STAGED" | grep -E "CLAUDE\.md$")

if [ -n "$KEY_FILES_CHANGED" ] && [ -z "$CLAUDE_MD_CHANGED" ]; then
    echo "CLAUDE.md CHECK: Key files staged but CLAUDE.md not included. Consider: git add CLAUDE.md"
fi

exit 0
