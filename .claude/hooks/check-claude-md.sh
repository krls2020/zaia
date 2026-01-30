#!/bin/bash
# PostToolUse hook: Remind to update CLAUDE.md when key Go files change
#
# Receives JSON on stdin with tool_input.file_path

input=$(cat)
CHANGED_FILE=$(echo "$input" | grep -o '"file_path":"[^"]*"' | head -1 | sed 's/"file_path":"//;s/"//')

if [ -z "$CHANGED_FILE" ]; then
    exit 0
fi

# Key patterns that should trigger CLAUDE.md review
if echo "$CHANGED_FILE" | grep -qE "(platform/client\.go|commands/root\.go|go\.mod|knowledge/engine\.go|output/envelope\.go|auth/manager\.go|auth/storage\.go|platform/errors\.go)"; then
    echo ""
    echo "══════════════════════════════════════════════════════════"
    echo "  CLAUDE.md CHECK"
    echo ""
    echo "  Klicovy soubor zmenen: $(basename "$CHANGED_FILE")"
    echo ""
    echo "  Over zda CLAUDE.md potrebuje aktualizaci:"
    echo "  - Novy/zmeny interface -> sekce 'Klicove typy'"
    echo "  - Novy command -> sekce 'Prikazy ZAIA CLI'"
    echo "  - Nova zavislost -> sekce 'Zavislosti'"
    echo "  - Zmena auth flow -> sekce 'Architektonicka rozhodnuti'"
    echo "  - Novy error code -> sekce 'Error codes'"
    echo "  - Zmena stavu impl -> sekce 'Aktualni stav implementace'"
    echo "══════════════════════════════════════════════════════════"
    echo ""
fi

exit 0
