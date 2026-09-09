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
    12_special_features \
    13_stdlib_phase6 \
    14_body_binding \
    15_jwt \
    16_openapi \
    17_database \
    18_cache \
    19_http \
    20_websocket \
    21_variant; do
    run_single_file_dir "$dir"
done

# 22_web_pages needs two special cases (v0.7.0 P5):
#   - example59 uses the template unit — multi-file build with stdlib/template_engine.klx
#   - example60 starts a real BootRun HTTP server — never exits, so it is
#     compiled, launched in the background, curl-checked on four endpoints,
#     and killed (real end-to-end coverage of redirect + error pages).
run_web_pages_test() {
    echo "Testing 22_web_pages..."
    cd "$ROOT/22_web_pages" 2>/dev/null || return 0

    # ---- example59: template engine (multi-file with stdlib unit) ----
    TOTAL=$((TOTAL + 1))
    if $KYLIX build "$ROOT/../../stdlib/template_engine.klx" example59_template.klx 2>&1 | grep -q "✓ Compiled"; then
        if go run main.go >/dev/null 2>&1; then
            echo "  ✓ example59_template"
            PASS=$((PASS + 1))
        else
            echo "  ✗ example59_template (go run failed)"
            FAIL=$((FAIL + 1))
        fi
    else
        echo "  ✗ example59_template (compile failed)"
        FAIL=$((FAIL + 1))
    fi

    # ---- example60: web framework E2E (launch server + curl + kill) ----
    TOTAL=$((TOTAL + 1))
    local ok=1 tries=0
    if ! $KYLIX build example60_web_framework.klx 2>&1 | grep -q "✓ Compiled" \
            || ! go build -o /tmp/kylix_web_e2e example60_web_framework.go 2>/dev/null; then
        echo "  ✗ example60_web_framework (compile failed)"
        FAIL=$((FAIL + 1))
        return 0
    fi
    /tmp/kylix_web_e2e >/dev/null 2>&1 &
    local pid=$!
    while [ $tries -lt 25 ]; do
        curl -s -o /dev/null http://127.0.0.1:8077/api/new && break
        sleep 0.2
        tries=$((tries + 1))
    done
    [ "$(curl -s http://127.0.0.1:8077/api/new)" = "you made it" ] || ok=0
    curl -si http://127.0.0.1:8077/api/old | head -1 | grep -q "^HTTP/1.1 302" || ok=0
    curl -si http://127.0.0.1:8077/api/old | grep -qi "^Location: /api/new" || ok=0
    curl -s http://127.0.0.1:8077/api/missing | grep -q "Custom 404" || ok=0
    curl -s http://127.0.0.1:8077/api/boom | grep -q "Custom 500" || ok=0
    kill -9 $pid 2>/dev/null
    wait $pid 2>/dev/null
    if [ "$ok" = 1 ]; then
        echo "  ✓ example60_web_framework (E2E)"
        PASS=$((PASS + 1))
    else
        echo "  ✗ example60_web_framework (E2E mismatch)"
        FAIL=$((FAIL + 1))
    fi
}

run_module_test

run_web_pages_test

# 23_regex (v0.7.1 P0b): example61 uses the pure-Kylix regex engine unit —
# multi-file build with stdlib/regex_engine.klx (no Go regexp dependency).
run_regex_test() {
    echo "Testing 23_regex..."
    cd "$ROOT/23_regex" 2>/dev/null || return 0
    TOTAL=$((TOTAL + 1))
    if $KYLIX build "$ROOT/../../stdlib/regex_engine.klx" example61_regex_engine.klx 2>&1 | grep -q "✓ Compiled"; then
        local out
        out=$(go run main.go 2>&1)
        if [ "$(echo "$out" | tail -1)" = "done" ] && [ "$(echo "$out" | wc -l | tr -d ' ')" = "33" ]; then
            echo "  ✓ example61_regex_engine"
            PASS=$((PASS + 1))
        else
            echo "  ✗ example61_regex_engine (run output mismatch)"
            FAIL=$((FAIL + 1))
        fi
    else
        echo "  ✗ example61_regex_engine (compile failed)"
        FAIL=$((FAIL + 1))
    fi
}

run_regex_test

echo ""
echo "================================"
echo "Results: $PASS/$TOTAL passed, $FAIL failed"
echo "================================"

if [ "$FAIL" -ne 0 ]; then
    exit 1
fi
