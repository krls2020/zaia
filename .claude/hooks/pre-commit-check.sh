#!/bin/bash
# PreToolUse hook for Bash: Before git commit, check CLAUDE.md is up to date
#
# Receives JSON on stdin with tool_input.command

input=$(cat)
COMMAND=$(echo "$input" | grep -o '"command":"[^"]*"' | head -1 | sed 's/"command":"//;s/"//')

# Only trigger on git commit commands
if ! echo "$COMMAND" | grep -qE 'git commit'; then
    exit 0
fi

cd /Users/macbook/Documents/Zerops-MCP/zaia || exit 0

# Get staged files relative to zaia/
STAGED=$(git diff --cached --name-only 2>/dev/null | grep '^zaia/' | sed 's|^zaia/||')

if [ -z "$STAGED" ]; then
    # Also try without prefix (if cwd is already zaia/)
    STAGED=$(git diff --cached --name-only 2>/dev/null)
fi

if [ -z "$STAGED" ]; then
    exit 0
fi

# Key Go files that indicate structural changes
KEY_PATTERNS="internal/platform/client\.go|internal/commands/root\.go|internal/knowledge/engine\.go|internal/output/envelope\.go|internal/auth/manager\.go|internal/platform/errors\.go|go\.mod"

KEY_FILES_CHANGED=$(echo "$STAGED" | grep -E "$KEY_PATTERNS" | head -5)
CLAUDE_MD_CHANGED=$(echo "$STAGED" | grep -E "CLAUDE\.md$")

if [ -n "$KEY_FILES_CHANGED" ] && [ -z "$CLAUDE_MD_CHANGED" ]; then
    echo ""
    echo "══════════════════════════════════════════════"
    echo "  CLAUDE.md NEBYL ZMENEN"
    echo "══════════════════════════════════════════════"
    echo ""
    echo "  Klicove soubory ve staged area:"
    echo "$KEY_FILES_CHANGED" | sed 's/^/    /'
    echo ""
    echo "  CLAUDE.md by mel byt aktualizovan pokud se zmenily:"
    echo "    - Interfaces, typy, error codes"
    echo "    - Prikazy, flagy, response format"
    echo "    - Zavislosti, architektura"
    echo "    - Stav implementace"
    echo "══════════════════════════════════════════════"
    echo ""
fi

exit 0
