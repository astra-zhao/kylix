#!/usr/bin/env bash
# check_artifact.sh — ELF shape, dynamic exports, and a JNI link.
# Delegates to `kylix mobile artifact android`.
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
case "$ABI" in
  x86_64) ABI=amd64 ;;
esac

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "check_artifact: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

exec "$KYLIX" mobile artifact android "$SO" "$ABI"
