#!/bin/bash
# PostToolUse hook: Auto-run tests after Go file edits

input=$(cat)
CHANGED_FILE=$(echo "$input" | grep -o '"file_path":"[^"]*"' | head -1 | sed 's/"file_path":"//;s/"//')

[ -z "$CHANGED_FILE" ] && exit 0

# Only for Go files
echo "$CHANGED_FILE" | grep -qE '\.go$' || exit 0

# Skip test data, embed, and generated files
echo "$CHANGED_FILE" | grep -qE '(testdata/|embed/|\.pb\.go$)' && exit 0

MODULE_ROOT=$(cd "$(dirname "$0")/../.." && pwd)
cd "$MODULE_ROOT" || exit 0

# Determine package from file path
PKG_DIR=$(echo "$CHANGED_FILE" | sed "s|^${MODULE_ROOT}/||" | xargs dirname)

if [ -d "$PKG_DIR" ]; then
    echo "-- go test ./${PKG_DIR} --"
    RESULT=$(go test "./${PKG_DIR}" -count=1 -short 2>&1 | tail -20)
    echo "$RESULT"
    if echo "$RESULT" | grep -q "^ok"; then
        echo "-- PASS --"
    elif echo "$RESULT" | grep -q "FAIL"; then
        echo "-- FAIL --"
    fi
    VET=$(go vet "./${PKG_DIR}" 2>&1)
    if [ -n "$VET" ]; then
        echo "-- go vet --"
        echo "$VET" | tail -5
    fi

    # Fast lint on changed package (non-blocking)
    if command -v golangci-lint &>/dev/null; then
        LINT_OUTPUT=$(golangci-lint run "./${PKG_DIR}" --fast 2>&1)
        LINT_EXIT=$?
        if [ $LINT_EXIT -ne 0 ] && [ -n "$LINT_OUTPUT" ]; then
            echo "-- golangci-lint (fast) --"
            echo "$LINT_OUTPUT" | tail -15
            echo "-- LINT WARNINGS --"
        fi
    fi
fi

exit 0
