#!/usr/bin/env bash
# check_artifact.sh — ELF shape + dynamic exports for one Android .so.
#
#   bash apps/android/check_artifact.sh path/to/libkylixlogic.so arm64
#   bash apps/android/check_artifact.sh path/to/libkylixlogic.so amd64
#
# Same idea as the Windows CI smoke (`file` must say PE32+): the shared
# object must be the right ELF machine, and the dynamic symbol table must
# list every name in apps/shared/mobile_exports.list. This does not load the
# library (the Android linker is not the host one) and does not boot an
# emulator.
set -euo pipefail

SO="${1:-}"
ABI="${2:-}"
if [ -z "$SO" ] || [ -z "$ABI" ]; then
  echo "usage: $0 <libkylixlogic.so> arm64|amd64" >&2
  exit 2
fi
if [ ! -f "$SO" ]; then
  echo "missing $SO" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
LIST="$ROOT/apps/shared/mobile_exports.list"

case "$ABI" in
  arm64) WANT="ELF 64-bit LSB shared object, ARM aarch64" ;;
  amd64|x86_64) WANT="ELF 64-bit LSB shared object, x86-64" ;;
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
echo "$blob" | grep -E 'mc_|kylix_free' || true

while IFS= read -r sym; do
  case "$sym" in
    ''|\#*) continue ;;
  esac
  if ! printf '%s\n' "$blob" | grep -E -q "(^|[^A-Za-z0-9])_?${sym}(\$|[^A-Za-z0-9_])"; then
    echo "missing dynamic symbol ${sym}" >&2
    exit 1
  fi
done < "$LIST"

echo "android artifact OK ($ABI)"
