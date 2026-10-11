#!/usr/bin/env bash
# build_core.sh — static library for the flashcard iOS shell.
#
#   bash apps/vocab/ios/build_core.sh simulator
#   bash apps/vocab/ios/build_core.sh device
#
# Non-macOS hosts get a readable error (macOS, Xcode, xcodegen) and no .a.
set -euo pipefail

KIND="${1:-simulator}"

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "build_core: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

exec "$KYLIX" mobile build ios "$KIND"
