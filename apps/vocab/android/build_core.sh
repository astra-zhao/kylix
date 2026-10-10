#!/usr/bin/env bash
# build_core.sh — cross-compile the flashcard core for one Android ABI.
#
#   bash apps/vocab/android/build_core.sh arm64
#   bash apps/vocab/android/build_core.sh amd64
#
# Requires the Android NDK (ANDROID_NDK_HOME or the usual SDK ndk/ directory).
# Does not compile server.klx: the phone library has no HTTP server.
set -euo pipefail

ABI="${1:-arm64}"
case "$ABI" in
  arm64) TARGET="android/arm64"; DIR="arm64-v8a" ;;
  amd64|x86_64) TARGET="android/amd64"; DIR="x86_64" ;;
  *) echo "usage: $0 arm64|amd64" >&2; exit 2 ;;
esac

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  KYLIX="${KYLIX:-kylix}"
fi
OUT="$ROOT/apps/vocab/android/app/src/main/jniLibs/$DIR"
mkdir -p "$OUT"

echo "[vocab-android] $TARGET → $OUT/libkylixvocab.so" >&2
(cd "$ROOT/apps/vocab" && "$KYLIX" build --backend=llvm --target "$TARGET" --shared \
  -o "$OUT/libkylixvocab.so" \
  ../../stdlib/stringutil.klx vocab.klx vocab_lib.klx)
echo "[vocab-android] wrote $OUT/libkylixvocab.so" >&2
