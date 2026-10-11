#!/usr/bin/env bash
# check_artifact.sh — Mach-O archive symbols, then an SDK link smoke.
# Delegates to `kylix mobile artifact ios`.
#
#   bash apps/vocab/ios/check_artifact.sh path/to/libkylixvocab.a simulator
#   bash apps/vocab/ios/check_artifact.sh path/to/libkylixvocab.a device
set -euo pipefail

ARCHIVE="${1:-}"
KIND="${2:-}"
if [ -z "$ARCHIVE" ] || [ -z "$KIND" ]; then
  echo "usage: $0 <libkylixvocab.a> simulator|device" >&2
  exit 2
fi

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "check_artifact: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

exec "$KYLIX" mobile artifact ios "$ARCHIVE" "$KIND"
