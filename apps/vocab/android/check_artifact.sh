#!/usr/bin/env bash
# check_artifact.sh — ELF shape, dynamic exports, and a JNI link of the
# flashcard .so. Does not boot an emulator.
#
#   bash apps/vocab/android/check_artifact.sh path/to/libkylixvocab.so arm64
#   bash apps/vocab/android/check_artifact.sh path/to/libkylixvocab.so amd64
set -euo pipefail

SO="${1:-}"
ABI="${2:-}"
if [ -z "$SO" ] || [ -z "$ABI" ]; then
  echo "usage: $0 <libkylixvocab.so> arm64|amd64" >&2
  exit 2
fi
if [ ! -f "$SO" ]; then
  echo "missing $SO" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
LIST="$ROOT/apps/vocab/vocab_exports.list"

case "$ABI" in
  arm64) WANT="ELF 64-bit LSB shared object, ARM aarch64"; CLANG_PREFIX="aarch64-linux-android30" ;;
  amd64|x86_64) WANT="ELF 64-bit LSB shared object, x86-64"; CLANG_PREFIX="x86_64-linux-android30" ;;
  *) echo "usage: $0 <so> arm64|amd64" >&2; exit 2 ;;
esac

desc=$(file "$SO")
echo "$desc"
if ! printf '%s\n' "$desc" | grep -q "$WANT"; then
  echo "expected file(1) to contain: $WANT" >&2
  exit 1
fi

if command -v llvm-nm >/dev/null 2>&1; then
  NM=llvm-nm
elif command -v llvm-nm-18 >/dev/null 2>&1; then
  NM=llvm-nm-18
else
  NM=nm
fi
blob=$("$NM" -D --defined-only "$SO")
echo "$blob" | grep -E 'vc_|kylix_free' || true

while IFS= read -r sym; do
  case "$sym" in
    ''|\#*) continue ;;
  esac
  if ! printf '%s\n' "$blob" | grep -E -q "(^|[^A-Za-z0-9])_?${sym}(\$|[^A-Za-z0-9_])"; then
    echo "missing dynamic symbol ${sym}" >&2
    exit 1
  fi
done < "$LIST"

NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-}}"
if [ -z "$NDK" ]; then
  echo "ANDROID_NDK_HOME is unset; skipped the JNI link (symbol table already matched)" >&2
  echo "android artifact OK ($ABI)"
  exit 0
fi
CLANG="$(echo "$NDK"/toolchains/llvm/prebuilt/*/bin/${CLANG_PREFIX}-clang)"
if [ ! -x "$CLANG" ]; then
  echo "NDK clang not found under $NDK" >&2
  exit 1
fi

WORK="$(mktemp -d /tmp/kylix_vocab_jni.XXXXXX)"
trap 'rm -rf "$WORK"' EXIT
"$CLANG" -shared -fPIC -Wl,--no-undefined \
  -o "$WORK/libvocabjni.so" \
  "$ROOT/apps/vocab/android/app/src/main/cpp/kylix_jni.c" \
  -L"$(dirname "$SO")" -lkylixvocab
echo "jni link OK ($ABI)"
echo "android artifact OK ($ABI)"
