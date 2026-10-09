#!/usr/bin/env bash
# build_core.sh — cross-compile the shared Kylix core for one Android ABI.
#
#   bash apps/android/build_core.sh arm64     # device / arm64 emulator
#   bash apps/android/build_core.sh amd64     # x86_64 emulator
#
# Requires the Android NDK (ANDROID_NDK_HOME or the usual SDK ndk/ directory).
# The compiler links with the NDK clang; without it this script stops before
# writing a half-built library. Object-only output (no NDK) is the .o command
# in docs/MOBILE_APPS.md.
set -euo pipefail

ABI="${1:-arm64}"
case "$ABI" in
  arm64) TARGET="android/arm64"; DIR="arm64-v8a" ;;
  amd64|x86_64) TARGET="android/amd64"; DIR="x86_64" ;;
  *) echo "usage: $0 arm64|amd64" >&2; exit 2 ;;
esac

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  KYLIX="${KYLIX:-kylix}"
fi
OUT="$ROOT/apps/android/app/src/main/jniLibs/$DIR"
mkdir -p "$OUT"

echo "[android] $TARGET → $OUT/libkylixlogic.so" >&2
(cd "$ROOT/apps/shared" && "$KYLIX" build --backend=llvm --target "$TARGET" --shared \
  -o "$OUT/libkylixlogic.so" \
  ../../stdlib/stringutil.klx mobilecore.klx mobilecore_lib.klx)
echo "[android] wrote $OUT/libkylixlogic.so" >&2
