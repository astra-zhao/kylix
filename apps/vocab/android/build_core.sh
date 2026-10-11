#!/usr/bin/env bash
# build_core.sh — cross-compile the flashcard core for one Android ABI.
#
#   bash apps/vocab/android/build_core.sh arm64
#   bash apps/vocab/android/build_core.sh amd64
#
# `x86_64` is accepted here as a synonym of amd64. `kylix mobile build android`
# itself only accepts amd64 or arm64. Does not compile server.klx.
set -euo pipefail

ABI="${1:-arm64}"
case "$ABI" in
  x86_64) ABI=amd64 ;;
esac

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "build_core: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

exec "$KYLIX" mobile build android "$ABI"
