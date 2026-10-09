#!/usr/bin/env bash
# host_check.sh — prove the shared mobile core on this machine.
#
#   1. Go backend vs LLVM backend, same parity.klx, byte-identical stdout.
#   2. LLVM --shared .so, dlopen, the C ABI the Android/iOS shells call.
#
# It does not cross-link Android or iOS (no claim that a simulator ran).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "host_check: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

WORK="$(mktemp -d /tmp/kylix_mobile_core.XXXXXX)"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

echo "[core] parity Go vs LLVM" >&2
(cd "$ROOT/apps/shared" && "$KYLIX" build --backend=go -o "$WORK/parity.go" \
  ../../stdlib/stringutil.klx mobilecore.klx parity.klx)
(cd "$ROOT" && go build -o "$WORK/parity_go" "$WORK/parity.go")
"$WORK/parity_go" > "$WORK/go.txt"

(cd "$ROOT/apps/shared" && "$KYLIX" build --backend=llvm -o "$WORK/parity_ll" \
  ../../stdlib/stringutil.klx mobilecore.klx parity.klx)
"$WORK/parity_ll" > "$WORK/ll.txt"

if ! cmp -s "$WORK/go.txt" "$WORK/ll.txt"; then
  echo "PARITY DIFF" >&2
  diff -u "$WORK/go.txt" "$WORK/ll.txt" >&2 || true
  exit 1
fi
echo "[core] parity PASS" >&2

echo "[core] C ABI .so" >&2
(cd "$ROOT/apps/shared" && "$KYLIX" build --backend=llvm --shared -o "$WORK/libkylixlogic.so" \
  ../../stdlib/stringutil.klx mobilecore.klx mobilecore_lib.klx)
cc -o "$WORK/c_host_check" "$ROOT/apps/shared/c_host_check.c" -ldl
"$WORK/c_host_check" "$WORK/libkylixlogic.so"

echo "[core] host check PASS" >&2
