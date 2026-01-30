#!/bin/bash
# PostToolUse hook: Auto-run tests after Go file edits
#
# Receives JSON on stdin with tool_input.file_path
# Provides immediate TDD feedback loop.

input=$(cat)
CHANGED_FILE=$(echo "$input" | grep -o '"file_path":"[^"]*"' | head -1 | sed 's/"file_path":"//;s/"//')

if [ -z "$CHANGED_FILE" ]; then
    exit 0
fi

# Only for Go files
if ! echo "$CHANGED_FILE" | grep -qE '\.go$'; then
    exit 0
fi

# Skip test data, embed, and generated files
if echo "$CHANGED_FILE" | grep -qE '(testdata/|embed/|\.pb\.go$)'; then
    exit 0
fi

cd /Users/macbook/Documents/Zerops-MCP/zaia || exit 0

# Determine package from file path
PKG_DIR=$(echo "$CHANGED_FILE" | sed 's|^/Users/macbook/Documents/Zerops-MCP/zaia/||' | xargs dirname)

# Map to Go package path
if [ -d "$PKG_DIR" ]; then
    echo "── go test ./${PKG_DIR} ──"
    RESULT=$(go test "./${PKG_DIR}" -count=1 -short 2>&1 | tail -20)
    echo "$RESULT"
    if echo "$RESULT" | grep -q "^ok"; then
        echo "── PASS ──"
    elif echo "$RESULT" | grep -q "FAIL"; then
        echo "── FAIL ──"
    fi
    # Also vet
    VET=$(go vet "./${PKG_DIR}" 2>&1)
    if [ -n "$VET" ]; then
        echo "── go vet ──"
        echo "$VET" | tail -5
    fi
fi

exit 0
