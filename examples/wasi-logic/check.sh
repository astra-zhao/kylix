#!/usr/bin/env bash
# Shape + behavior gate for the LLVM wasm32-unknown-wasi target.
# Requires: kylix on $KYLIX (or ./kylix), llc/clang with the wasm32 backend, wasmtime.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
OUT="${1:-/tmp/kylix-wasi-logic.wasm}"

"$KYLIX" build --backend=llvm --target wasi/wasm32 -o "$OUT" "$ROOT/examples/wasi-logic/main.klx"

file "$OUT" | grep -q -i 'WebAssembly'
python3 - "$OUT" << 'PY'
import sys
data = open(sys.argv[1], "rb").read()
# Magic.
if data[:4] != b"\x00asm":
    raise SystemExit("not a wasm module")
for name in (b"wasi_snapshot_preview1", b"fd_write", b"clock_time_get", b"random_get", b"proc_exit"):
    if name not in data:
        raise SystemExit(f"missing import {name!r}")
# No DOM / browser import module.
if b"document" in data or b"webgpu" in data:
    raise SystemExit("DOM-looking import in a pure-logic module")
print("imports ok")
PY

got="$(wasmtime "$OUT")"
# *.txt is gitignored, so the expected transcript lives here.
# Command substitution drops the final newline from wasmtime's stdout.
want=$'sum=42\nhello wasi\n42\n1.5'
if [ "$got" != "$want" ]; then
  echo "stdout mismatch"
  echo "got:  $(printf %q "$got")"
  echo "want: $(printf %q "$want")"
  exit 1
fi
echo "wasmtime ok"

# Go wasip1 import table: the smoke binary's imports must include fd_write
# from our //go:wasmimport lines (the Go runtime also imports preview1).
SMOKE_DIR="$ROOT/examples/wasi-preview1"
SMOKE_OUT="${2:-/tmp/kylix-wasi-preview1.wasm}"
( cd "$SMOKE_DIR" && GOOS=wasip1 GOARCH=wasm go build -o "$SMOKE_OUT" . )
python3 - "$SMOKE_OUT" << 'PY'
import sys
data = open(sys.argv[1], "rb").read()
for name in (b"wasi_snapshot_preview1", b"fd_write", b"clock_time_get", b"random_get", b"proc_exit", b"path_open"):
    if name not in data:
        raise SystemExit(f"go wasip1 smoke missing {name!r}")
print("go import table ok")
PY

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
printf 'preview1 file\n' > "$tmpdir/note.txt"
out="$(wasmtime --dir "$tmpdir" --env NAME=Kylix "$SMOKE_OUT")"
printf '%s\n' "$out" | grep -q 'preview1 ok'
printf '%s\n' "$out" | grep -q 'NAME=Kylix'
printf '%s\n' "$out" | grep -q 'preview1 file'
printf '%s\n' "$out" | grep -q 'stored ok'
echo "go wasmtime ok"
