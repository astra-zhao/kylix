#!/bin/bash
# Test all Kylix tutorial examples

KYLIX=${KYLIX:-kylix}
PASS=0
FAIL=0
TOTAL=0

ROOT=$(cd "$(dirname "$0")" && pwd)

run_single_file_dir() {
    local dir="$1"
    echo "Testing $dir..."
    cd "$ROOT/$dir" 2>/dev/null || return 0

    shopt -s nullglob
    local files=(example*.klx)
    shopt -u nullglob
    if [ ${#files[@]} -eq 0 ]; then
        echo "  (no example*.klx)"
        return 0
    fi

    for f in "${files[@]}"; do
        TOTAL=$((TOTAL + 1))

        if $KYLIX build "$f" 2>&1 | grep -q "✓ Compiled"; then
            GOFILE="${f%.klx}.go"
            if [ -f "$GOFILE" ] && go run "$GOFILE" >/dev/null 2>&1; then
                echo "  ✓ $f"
                PASS=$((PASS + 1))
            else
                echo "  ✗ $f (go run failed)"
                FAIL=$((FAIL + 1))
            fi
        else
            echo "  ✗ $f (compile failed)"
            FAIL=$((FAIL + 1))
        fi
    done
}

run_module_test() {
    echo "Testing 11_modules..."
    cd "$ROOT/11_modules" 2>/dev/null || return 0
    TOTAL=$((TOTAL + 2))
    if $KYLIX build math_helper.klx example33_use_module.klx 2>&1 | grep -q "✓ Compiled"; then
        if [ -f "main.go" ] && go run main.go >/dev/null 2>&1; then
            echo "  ✓ modules (2 files)"
            PASS=$((PASS + 2))
        else
            echo "  ✗ modules (go run failed)"
            FAIL=$((FAIL + 2))
        fi
    else
        echo "  ✗ modules (compile failed)"
        FAIL=$((FAIL + 2))
    fi
}

echo "Testing Kylix Tutorial Examples"
echo "================================"
echo ""

for dir in \
    01_basics \
    02_control_flow \
    03_functions \
    04_oop \
    05_generics \
    06_advanced_types \
    07_stdlib_core \
    08_stdlib_utils \
    10_exceptions \
    12_special_features; do
    run_single_file_dir "$dir"
done

run_module_test

echo ""
echo "================================"
echo "Results: $PASS/$TOTAL passed, $FAIL failed"
echo "================================"

if [ "$FAIL" -ne 0 ]; then
    exit 1
fi
