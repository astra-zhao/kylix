#!/usr/bin/env bash
# ci_gate.sh — v0.9.0: performance regression gate for CI.
#
# Runs each compile_time.sh scenario ONCE against the bootstrap sources and
# fails when a scenario blows past its ceiling. The ceilings are deliberately
# LOOSE (~100× local numbers): shared CI runners have ±30% noise, so this
# gate only catches ORDER-OF-MAGNITUDE regressions — lost incremental cache,
# lost DCE/optimization pass, accidental O(n²) in the emitter — not small
# drift. docs/compile-performance.md holds the local reference numbers.
#
# Ceilings (seconds) and what they catch:
#   go_cold  10   bootstrap build with a fresh cache — cache broken/never hit
#   go_warm   5   second build on the same cache — incremental lookup broken
#   llvm_o0  60   native build — mem2reg/DCE pass dropped, IR blow-up
#   llvm_o2  60   optimized build — -O2 channel silently skipped
#
# Usage: bash benchmarks/ci_gate.sh [KYLIX binary]
set -euo pipefail

cd "$(dirname "$0")/.."
KYLIX=${1:-${KYLIX:-$PWD/kylix}}
if [ ! -x "$KYLIX" ]; then
  echo "> building $KYLIX" >&2
  go build -o "$KYLIX" ./cmd/kylix/
fi

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cp src/*.klx "$TMP/"
cd "$TMP"
SRC=(*.klx)

now_ns() {
  if [ "$(uname -s)" = "Darwin" ]; then
    python3 -c 'import time; print(int(time.monotonic()*1e9))'
  else
    date +%s%N
  fi
}

run_ms() {
  local start end
  start=$(now_ns)
  if ! "$@" >/dev/null 2>&1; then
    echo "FAIL"
    return 0
  fi
  end=$(now_ns)
  echo $(( (end - start) / 1000000 ))
}

check() { # check <name> <ceiling_s> <measured_ms>
  local name=$1 cap=$2 ms=$3
  if [ "$ms" = "FAIL" ]; then
    echo "PERF-GATE FAIL  $name: build itself failed" >&2
    exit 1
  fi
  local cap_ms=$((cap * 1000))
  if [ "$ms" -gt "$cap_ms" ]; then
    printf 'PERF-GATE FAIL  %-9s %6dms > ceiling %ds — order-of-magnitude regression\n' "$name" "$ms" "$cap" >&2
    exit 1
  fi
  printf 'PERF-GATE PASS  %-9s %6dms (ceiling %ds)\n' "$name" "$ms" "$cap"
}

echo "> perf gate: $(basename "$KYLIX") over ${#SRC[@]} bootstrap sources"
go_cold=$(rm -rf .kylix-cache; run_ms "$KYLIX" build "${SRC[@]}")
go_warm=$(run_ms "$KYLIX" build "${SRC[@]}")
llvm_o0=$(run_ms "$KYLIX" build --backend=llvm "${SRC[@]}")
llvm_o2=$(run_ms "$KYLIX" build --backend=llvm --llvm-opt=2 "${SRC[@]}")
check go_cold  10 "$go_cold"
check go_warm   5 "$go_warm"
check llvm_o0  60 "$llvm_o0"
check llvm_o2  60 "$llvm_o2"
echo "> perf gate OK"
