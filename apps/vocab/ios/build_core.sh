#!/usr/bin/env bash
# build_core.sh — static library for the flashcard iOS shell.
#
#   bash apps/vocab/ios/build_core.sh simulator
#   bash apps/vocab/ios/build_core.sh device
#
# Linking an iOS binary needs macOS + Xcode (xcrun). This script refuses to
# run anywhere else. server.klx is not on the link line.
set -euo pipefail

KIND="${1:-simulator}"
case "$KIND" in
  simulator|sim) TARGET="ios/simulator-arm64" ;;
  device) TARGET="ios/arm64" ;;
  *) echo "usage: $0 simulator|device" >&2; exit 2 ;;
esac

if [ "$(uname -s)" != "Darwin" ]; then
  echo "iOS archives must be linked on macOS with Xcode command line tools." >&2
  echo "On this host, stop here. The Swift sources are in apps/vocab/ios/." >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  KYLIX="${KYLIX:-kylix}"
fi
OUT="$ROOT/apps/vocab/ios/Sources/CKylixVocab/lib"
mkdir -p "$OUT"

echo "[vocab-ios] $TARGET → $OUT/libkylixvocab.a" >&2
(cd "$ROOT/apps/vocab" && "$KYLIX" build --backend=llvm --target "$TARGET" --shared \
  -o "$OUT/libkylixvocab.a" \
  ../../stdlib/stringutil.klx vocab.klx vocab_lib.klx)
echo "[vocab-ios] wrote $OUT/libkylixvocab.a" >&2
