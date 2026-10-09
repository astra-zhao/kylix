#!/usr/bin/env bash
# build_core.sh — static library for the iOS shell.
#
#   bash apps/ios/build_core.sh simulator    # arm64 simulator (Apple silicon Mac)
#   bash apps/ios/build_core.sh device       # arm64 device
#
# Linking an iOS binary needs macOS + Xcode (xcrun). This script refuses to
# run anywhere else, with the exact error the compiler returns.
set -euo pipefail

KIND="${1:-simulator}"
case "$KIND" in
  simulator|sim) TARGET="ios/simulator-arm64" ;;
  device) TARGET="ios/arm64" ;;
  *) echo "usage: $0 simulator|device" >&2; exit 2 ;;
esac

if [ "$(uname -s)" != "Darwin" ]; then
  echo "iOS archives must be linked on macOS with Xcode command line tools." >&2
  echo "On this host, stop here. The Swift sources are in apps/ios/." >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  KYLIX="${KYLIX:-kylix}"
fi
OUT="$ROOT/apps/ios/Sources/CKylixCore/lib"
mkdir -p "$OUT"

echo "[ios] $TARGET → $OUT/libkylixcore.a" >&2
(cd "$ROOT/apps/shared" && "$KYLIX" build --backend=llvm --target "$TARGET" --shared \
  -o "$OUT/libkylixcore.a" \
  ../../stdlib/stringutil.klx mobilecore.klx mobilecore_lib.klx)
echo "[ios] wrote $OUT/libkylixcore.a" >&2
