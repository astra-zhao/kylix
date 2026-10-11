#!/usr/bin/env bash
# host_check.sh — prove the flashcard core on this machine.
#
# Delegates to `kylix mobile check --host-only` so the script and the
# subcommand cannot drift. Success line: [vocab] host check PASS
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "host_check: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

exec "$KYLIX" mobile check --host-only
