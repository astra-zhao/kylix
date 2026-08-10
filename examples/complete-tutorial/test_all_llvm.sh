#!/bin/bash
# Test all Kylix tutorial examples with the LLVM backend (v6.1.0).
#
# Mirrors test_all.sh but compiles each example*.klx with --backend=llvm and
# runs the resulting native binary (exit-code check). The .ll/.o intermediates
# (written next to the source by the LLVM pipeline) are removed after each
# build so the working tree stays clean.
#
# Usage:
#   KYLIX=$PWD/kylix bash examples/complete-tutorial/test_all_llvm.sh
#   LLVM_OPT=2 ...        # pass --llvm-opt=2 (v6.0.0 #8 -O2 verification)
#   OUTDIR=/tmp/o0 ...    # capture each binary's stdout to OUTDIR/<name>.out
#                         # instead of exit-code checking (for parity diffs:
#                         #   OUTDIR=/tmp/o0 bash ... ; OUTDIR=/tmp/o2 LLVM_OPT=2 bash ...
#                         #   diff -r /tmp/o0 /tmp/o2)

KYLIX=${KYLIX:-kylix}
LLVM_OPT=${LLVM_OPT:-}
OUTDIR=${OUTDIR:-}
PASS=0
FAIL=0
TOTAL=0

ROOT=$(cd "$(dirname "$0")" && pwd)
BINDIR=$(mktemp -d)
trap 'rm -rf "$BINDIR"' EXIT

# clean_artifacts removes .ll/.o/.opt.ll generated next to the source by the
# LLVM pipeline, but ONLY untracked ones — committed artifacts (e.g. the
# historical example01_hello.ll checked into the repo) are preserved.
clean_artifacts() {
    local base="$1"
    for ext in .ll .o .opt.ll; do
        local f="${base}${ext}"
        if [ -f "$f" ] && ! git ls-files --error-unmatch "$f" >/dev/null 2>&1; then
            rm -f "$f"
        fi
    done
}

run_single_file_dir() {
    local dir="$1"
    echo "Testing $dir (LLVM)..."
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
        local name="${f%.klx}"
        if $KYLIX build --backend=llvm ${LLVM_OPT:+--llvm-opt=$LLVM_OPT} -o "$BINDIR/$name" "$f" >/dev/null 2>&1; then
            if [ -n "$OUTDIR" ]; then
                mkdir -p "$OUTDIR"
                if "$BINDIR/$name" > "$OUTDIR/$name.out" 2>&1; then
                    echo "  ✓ $f (captured)"
                    PASS=$((PASS + 1))
                else
                    echo "  ✗ $f (run failed)"
                    FAIL=$((FAIL + 1))
                fi
            elif [ -x "$BINDIR/$name" ] && "$BINDIR/$name" >/dev/null 2>&1; then
                echo "  ✓ $f"
                PASS=$((PASS + 1))
            else
                echo "  ✗ $f (run failed)"
                FAIL=$((FAIL + 1))
            fi
        else
            echo "  ✗ $f (compile failed)"
            FAIL=$((FAIL + 1))
        fi
        # build --backend=llvm writes .ll/.o/.opt.ll next to the source — clean
        # up only the ones this run created (untracked).
        clean_artifacts "${f%.klx}"
    done
}

run_module_test() {
    echo "Testing 11_modules (LLVM)..."
    cd "$ROOT/11_modules" 2>/dev/null || return 0
    TOTAL=$((TOTAL + 2))
    if $KYLIX build --backend=llvm ${LLVM_OPT:+--llvm-opt=$LLVM_OPT} -o "$BINDIR/example33_use_module" \
            math_helper.klx example33_use_module.klx >/dev/null 2>&1; then
        if [ -n "$OUTDIR" ]; then
            mkdir -p "$OUTDIR"
            if "$BINDIR/example33_use_module" > "$OUTDIR/example33_use_module.out" 2>&1; then
                echo "  ✓ modules (2 files, captured)"
                PASS=$((PASS + 2))
            else
                echo "  ✗ modules (run failed)"
                FAIL=$((FAIL + 2))
            fi
        elif [ -x "$BINDIR/example33_use_module" ] && "$BINDIR/example33_use_module" >/dev/null 2>&1; then
            echo "  ✓ modules (2 files)"
            PASS=$((PASS + 2))
        else
            echo "  ✗ modules (run failed)"
            FAIL=$((FAIL + 2))
        fi
    else
        echo "  ✗ modules (compile failed)"
        FAIL=$((FAIL + 2))
    fi
    clean_artifacts math_helper
    clean_artifacts example33_use_module
}

echo "Testing Kylix Tutorial Examples (LLVM backend)"
echo "=============================================="
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

run_module_test

echo ""
echo "=============================================="
echo "Results: $PASS/$TOTAL passed, $FAIL failed"
echo "=============================================="

if [ "$FAIL" -ne 0 ]; then
    exit 1
fi
