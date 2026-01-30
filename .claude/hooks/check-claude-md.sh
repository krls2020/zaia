#!/bin/bash
# PostToolUse hook: Remind to check CLAUDE.md when key Go files change

input=$(cat)
CHANGED_FILE=$(echo "$input" | grep -o '"file_path":"[^"]*"' | head -1 | sed 's/"file_path":"//;s/"//')

[ -z "$CHANGED_FILE" ] && exit 0

if echo "$CHANGED_FILE" | grep -qE "(platform/client\.go|commands/root\.go|go\.mod|knowledge/engine\.go|output/envelope\.go|auth/manager\.go|auth/storage\.go|platform/errors\.go)"; then
    echo "CLAUDE.md CHECK: Key file changed: $(basename "$CHANGED_FILE"). Check if CLAUDE.md or README.md needs updating."
fi

exit 0
