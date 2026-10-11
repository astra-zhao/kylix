#!/usr/bin/env bash
# e2e.sh — curl the flashcard server through browse / mark / review.
#
# Delegates to `kylix mobile check --e2e-only`. VOCAB_BACKEND is go (default),
# llvm, or both. The server is not left running. Port 8091.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "e2e: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

BACKEND="${VOCAB_BACKEND:-go}"
exec "$KYLIX" mobile check --e2e-only --backend="$BACKEND"
