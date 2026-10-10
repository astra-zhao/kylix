#!/usr/bin/env bash
# e2e.sh — curl the flashcard server through browse / mark / review.
#
# The expected bodies are the parity program's stdout (the same unit the
# server calls). Default backend is Go. Set VOCAB_BACKEND=llvm to repeat
# the same curls against an LLVM-built server. Set VOCAB_BACKEND=both to
# run Go and then LLVM.
#
# Port 8091. KylixAdmin's default is 8090, so the two can run together.
# This does not start an emulator.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
KYLIX="${KYLIX:-$ROOT/kylix}"
if [ ! -x "$KYLIX" ]; then
  echo "e2e: building kylix CLI" >&2
  (cd "$ROOT" && go build -o "$ROOT/kylix" ./cmd/kylix/)
  KYLIX="$ROOT/kylix"
fi

BACKEND="${VOCAB_BACKEND:-go}"
case "$BACKEND" in
  go|llvm|both) ;;
  *) echo "VOCAB_BACKEND must be go, llvm, or both" >&2; exit 2 ;;
esac

WORK="$(mktemp -d /tmp/kylix_vocab_e2e.XXXXXX)"
SERVER_PID=""
cleanup() {
  if [ -n "$SERVER_PID" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -rf "$WORK"
}
trap cleanup EXIT

if curl -sf -o /dev/null --max-time 1 "http://127.0.0.1:8091/api/words"; then
  echo "port 8091 is already in use; stop that process and retry" >&2
  exit 1
fi

echo "[vocab] expected bodies from Go parity" >&2
(cd "$ROOT/apps/vocab" && "$KYLIX" build --backend=go -o "$WORK/parity.go" \
  ../../stdlib/stringutil.klx vocab.klx parity.klx)
(cd "$ROOT" && go build -o "$WORK/parity_go" "$WORK/parity.go")
"$WORK/parity_go" > "$WORK/go.txt"

take() {
  local key="$1"
  local out="$2"
  local line
  line="$(sed -n "s/^${key}=//p" "$WORK/go.txt")"
  if [ -z "$line" ]; then
    echo "missing parity line ${key}" >&2
    exit 1
  fi
  printf '%s' "$line" > "$out"
}

take browse "$WORK/browse.exp"
take mark "$WORK/mark.exp"
take review "$WORK/review.exp"
take again "$WORK/again.exp"
take both "$WORK/both.exp"
take bad "$WORK/bad.exp"

same() {
  local name="$1"
  local exp="$2"
  local got="$3"
  if ! cmp -s "$exp" "$got"; then
    echo "DIFF $name" >&2
    diff -u "$exp" "$got" >&2 || true
    exit 1
  fi
}

run_backend() {
  local backend="$1"
  echo "[vocab] e2e backend=$backend" >&2
  if [ -n "$SERVER_PID" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    SERVER_PID=""
  fi

  if [ "$backend" = "go" ]; then
    (cd "$ROOT/apps/vocab" && "$KYLIX" build --backend=go -o "$WORK/server.go" \
      ../../stdlib/stringutil.klx vocab.klx server.klx)
    (cd "$ROOT" && go build -o "$WORK/vocab_server" "$WORK/server.go")
  else
    (cd "$ROOT/apps/vocab" && "$KYLIX" build --backend=llvm -o "$WORK/vocab_server" \
      ../../stdlib/stringutil.klx vocab.klx server.klx)
  fi

  "$WORK/vocab_server" > "$WORK/server.log" 2>&1 &
  SERVER_PID=$!

  local i
  for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
    if curl -sf -o /dev/null --max-time 1 "http://127.0.0.1:8091/api/words"; then
      break
    fi
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      echo "server exited" >&2
      cat "$WORK/server.log" >&2 || true
      exit 1
    fi
    sleep 0.25
  done
  if ! curl -sf -o /dev/null --max-time 1 "http://127.0.0.1:8091/api/words"; then
    echo "server did not accept /api/words" >&2
    cat "$WORK/server.log" >&2 || true
    exit 1
  fi

  curl -sf --max-time 5 "http://127.0.0.1:8091/api/words" > "$WORK/browse.got"
  same browse "$WORK/browse.exp" "$WORK/browse.got"

  curl -sf --max-time 5 "http://127.0.0.1:8091/api/words?known=" > "$WORK/empty.got"
  same empty "$WORK/browse.exp" "$WORK/empty.got"

  curl -sf --max-time 5 "http://127.0.0.1:8091/api/words?known=3,1,1" > "$WORK/both.got"
  same both "$WORK/both.exp" "$WORK/both.got"

  curl -sf --max-time 5 -X POST "http://127.0.0.1:8091/api/mark" \
    -H 'Content-Type: application/json' \
    --data '{"known":"","id":"1"}' > "$WORK/mark.got"
  same mark "$WORK/mark.exp" "$WORK/mark.got"

  curl -sf --max-time 5 "http://127.0.0.1:8091/api/review?known=1" > "$WORK/review.got"
  same review "$WORK/review.exp" "$WORK/review.got"

  curl -sf --max-time 5 -X POST "http://127.0.0.1:8091/api/mark" \
    -H 'Content-Type: application/json' \
    --data '{"known":"1","id":"1"}' > "$WORK/again.got"
  same again "$WORK/again.exp" "$WORK/again.got"

  local code
  code="$(curl -sS --max-time 5 -o "$WORK/bad.got" -w '%{http_code}' \
    -X POST "http://127.0.0.1:8091/api/mark" \
    -H 'Content-Type: application/json' \
    --data '{"known":"1","id":"9"}')"
  if [ "$code" != "400" ]; then
    echo "mark of unknown id status=$code" >&2
    exit 1
  fi
  same bad "$WORK/bad.exp" "$WORK/bad.got"

  if ! grep -q 'apple' "$WORK/browse.got"; then
    echo "browse body lost apple" >&2
    exit 1
  fi
  if grep -q 'apple' "$WORK/review.got"; then
    echo "review still contains apple" >&2
    exit 1
  fi
  echo "[vocab] e2e PASS ($backend)" >&2
}

if [ "$BACKEND" = "both" ]; then
  run_backend go
  run_backend llvm
elif [ "$BACKEND" = "llvm" ]; then
  run_backend llvm
else
  run_backend go
fi
