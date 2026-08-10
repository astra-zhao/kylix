#!/usr/bin/env bash
# compile_time.sh — Compile-time benchmark for the Kylix compiler (v6.0.0 #7).
#
# Measures wall-clock time to compile the bootstrap sources (src/*.klx,
# ~7.5k lines) on both backends:
#   - Go backend, cold cache (fresh .kylix-cache each round)
#   - Go backend, warm cache (incremental cache hits)
#   - LLVM backend, -O0 and -O2 (native binary via llc + clang)
#
# Each scenario runs 3 times; the table reports all 3 plus the median.
# All artifacts are produced in a throwaway temp dir — the repo stays clean.
#
# Usage:
#   bash benchmarks/compile_time.sh            # uses ./kylix (builds if missing)
#   KYLIX=/path/to/kylix bash benchmarks/compile_time.sh
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT=$(pwd)
KYLIX=${KYLIX:-"$ROOT/kylix"}

if [ ! -x "$KYLIX" ]; then
  echo "> building $KYLIX" >&2
  go build -o "$KYLIX" ./cmd/kylix/
fi

LINES=$(wc -l src/*.klx | tail -1 | awk '{print $1}')
FILES=$(ls src/*.klx | wc -l | tr -d ' ')
echo "> benchmark target: src/*.klx  ($FILES files, $LINES lines)"
echo "> kylix: $KYLIX"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cp src/*.klx "$TMP/"
cd "$TMP"
# After cd, the sources are plain `*.klx` here — NOT `src/*.klx` (that glob
# would expand to nothing and `kylix build` would fall into project mode and
# fail fast, giving a bogus ~0ms measurement). v6.1.0 fix.
SRC=(*.klx)

now_ns() {
  if [ "$(uname -s)" = "Darwin" ]; then
    python3 -c 'import time; print(int(time.monotonic()*1e9))'
  else
    date +%s%N
  fi
}

# run_ms: run a command (discarding output) and print its elapsed ms, or FAIL
# if the build itself fails (so a broken scenario is never reported as a fast
# measurement).
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

median() { sort -n | awk '{a[NR]=$1} END {print a[int((NR+1)/2)]}'; }

row() {
  local label=$1; shift
  local t1=$1 t2=$2 t3=$3
  if [ "$t1" = "FAIL" ] || [ "$t2" = "FAIL" ] || [ "$t3" = "FAIL" ]; then
    printf '| %s | FAIL | FAIL | FAIL | **FAIL** |\n' "$label"
    return 0
  fi
  local m
  m=$(printf '%s\n%s\n%s\n' "$t1" "$t2" "$t3" | median)
  printf '| %s | %dms | %dms | %dms | **%dms** |\n' "$label" "$t1" "$t2" "$t3" "$m"
}

echo ""
echo "Scenario: kylix build "${SRC[@]}"  (3 rounds, wall-clock)"
echo ""
echo "| 场景 | 第1次 | 第2次 | 第3次 | 中位数 |"
echo "|---|---|---|---|---|"

# --- Go backend, cold cache (purge .kylix-cache before every round) ---
g1=$(rm -rf .kylix-cache; run_ms "$KYLIX" build "${SRC[@]}")
g2=$(rm -rf .kylix-cache; run_ms "$KYLIX" build "${SRC[@]}")
g3=$(rm -rf .kylix-cache; run_ms "$KYLIX" build "${SRC[@]}")
row "Go 冷编译（无缓存）" "$g1" "$g2" "$g3"

# --- Go backend, warm cache (cache built by the cold runs above) ---
w1=$(run_ms "$KYLIX" build "${SRC[@]}")
w2=$(run_ms "$KYLIX" build "${SRC[@]}")
w3=$(run_ms "$KYLIX" build "${SRC[@]}")
row "Go 热编译（增量缓存）" "$w1" "$w2" "$w3"

# --- LLVM backend, -O0 ---
l1=$(run_ms "$KYLIX" build --backend=llvm "${SRC[@]}")
l2=$(run_ms "$KYLIX" build --backend=llvm "${SRC[@]}")
l3=$(run_ms "$KYLIX" build --backend=llvm "${SRC[@]}")
row "LLVM -O0" "$l1" "$l2" "$l3"

# --- LLVM backend, -O2 ---
o1=$(run_ms "$KYLIX" build --backend=llvm --llvm-opt=2 "${SRC[@]}")
o2=$(run_ms "$KYLIX" build --backend=llvm --llvm-opt=2 "${SRC[@]}")
o3=$(run_ms "$KYLIX" build --backend=llvm --llvm-opt=2 "${SRC[@]}")
row "LLVM -O2" "$o1" "$o2" "$o3"

echo ""
echo "> done. Record results in docs/compile-performance.md (see CHANGELOG)."
