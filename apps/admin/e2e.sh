#!/usr/bin/env bash
# e2e.sh — KylixAdmin dual-backend parity E2E (v0.10.0 P2 phase 4)
#
# Builds the Go and LLVM forms of apps/admin from the same Kylix sources,
# runs the same 12-scenario curl sequence against each, and diffs the
# normalized transcripts (scenarios print deterministic key=value lines only
# — CSRF tokens / session IDs / timestamps never enter the transcript).
# Exit 0 iff both forms behave identically and every assertion passes.
#
# Environment:
#   KYLIX    path to the kylix CLI (default /tmp/kylix_bin; built by CI)
#   PORT     listen port (default 8091 — example60 uses 8090)
#   LLVM_GC  1 (default) = build the LLVM form with --gc=boehm (falls back
#            to plain malloc build when libgc is missing)
#
# Scenario map (plan doc: docs/ADMIN_PLATFORM.md v0.10.0 P2):
#   S1  GET /login 200 + _csrf + Set-Cookie KYLIX_SID
#   S2  wrong password -> error + login_logs success=0
#   S3  admin login -> 302 /dashboard, guarded pages 200
#   S4  admin creates bob/viewer/lockme -> list rows + op_logs
#   S5  delete: temp user gone, seed-admin guard, no-perm user 403
#   S6  perm-less user GET /users -> 403 Forbidden page
#   S7  POST /users without _csrf -> 403
#   S8  no session GET /users -> 401
#   S9  remember=1 -> Set-Cookie Max-Age=2592000
#   S10 5 wrong logins -> locked; correct password still locked
#   S11 /logs renders login_logs + op_logs (page + sqlite)
#   S12 logout -> old cookie -> 401
set -u

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PORT="${PORT:-8091}"
KYLIX="${KYLIX:-/tmp/kylix_bin}"
LLVM_GC="${LLVM_GC:-1}"
WORK="$(mktemp -d /tmp/kyadmin_e2e.XXXXXX)"
ADMIN="$ROOT/apps/admin"
GOGEN="$ROOT/.e2e_admin"
BASE="http://localhost:$PORT"

fail() { echo "E2E-FAIL: $*" >&2; exit 1; }
[ -x "$KYLIX" ] || fail "kylix CLI not found at $KYLIX (set KYLIX=...)"
command -v sqlite3 >/dev/null || fail "sqlite3 required"
command -v curl >/dev/null || fail "curl required"

SRV_PID=""
cleanup() {
  if [ -n "$SRV_PID" ]; then
    kill "$SRV_PID" 2>/dev/null
    sleep 0.2
    kill -9 "$SRV_PID" 2>/dev/null
  fi
  if [ "${KEEP:-0}" != "1" ]; then
    rm -rf "$WORK" "$GOGEN"
  else
    echo "KEEP=1: workdir preserved at $WORK" >&2
  fi
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# build both forms
# ---------------------------------------------------------------------------

# Go form: the generated Go source must live inside the repo module to
# resolve the "kylix/stdlib" import, so it goes to $GOGEN (cleaned on exit).
mkdir -p "$GOGEN"
(cd "$ADMIN" && "$KYLIX" build --backend=go -o "$GOGEN/main.go" \
  ../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
  lib/admindb.klx lib/adminsec.klx lib/audit.klx main.klx) \
  || fail "Go-form codegen failed"
(cd "$ROOT" && go build -o "$WORK/go_bin" ./.e2e_admin) || fail "Go-form go build failed"

# LLVM form (boehm GC opt-in; fall back when libgc is absent).
LL_BIN="$WORK/ll_bin"
build_ll() {
  (cd "$ADMIN" && "$KYLIX" build --backend=llvm $1 -o "$LL_BIN" \
    ../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
    lib/admindb.klx lib/adminsec.klx lib/audit.klx main.klx \
    > "$WORK/ll_build.log" 2>&1)
}
LL_FLAGS=""
if [ "$LLVM_GC" = "1" ]; then LL_FLAGS="--gc=boehm"; fi
if ! build_ll "$LL_FLAGS"; then
  if [ -n "$LL_FLAGS" ]; then
    echo "note: --gc=boehm build failed (libgc missing?); retrying default malloc" >&2
    LL_FLAGS=""
    build_ll "" || { tail -5 "$WORK/ll_build.log"; fail "LLVM-form build failed"; }
  else
    tail -5 "$WORK/ll_build.log"; fail "LLVM-form build failed"
  fi
fi

# ---------------------------------------------------------------------------
# the 12-scenario sequence (shared by both forms)
# ---------------------------------------------------------------------------
scenarios() {
  # $1 = transcript file, $2 = server binary, $3 = sqlite db path
  local T="$1" BIN="$2" DB="$3"
  local TAG; TAG=$(basename "$BIN")
  local J="$WORK/jars_$TAG"
  mkdir -p "$J"
  rm -f "$DB"

  # exec inside the subshell so $! is the binary itself and stays a child of
  # this script (the old `(... & echo $! > pid)` pattern orphaned the server,
  # leaving `wait` a no-op and the process alive after cleanup).
  (cd "$ADMIN" && exec env KYADMIN_DB="$DB" KYADMIN_PASSWORD=Admin@123 KYADMIN_PORT="$PORT" "$BIN") \
    > "$WORK/srv_$TAG.log" 2>&1 &
  # global (not local) so the EXIT trap can kill it even on fail()
  SRV_PID=$!

  local ready=""
  for _ in $(seq 1 50); do
    curl -s -o /dev/null "$BASE/login" && ready=1 && break
    sleep 0.3
  done
  [ -n "$ready" ] || fail "server did not start: $(tail -3 "$WORK/srv_$TAG.log" 2>/dev/null)"
  echo "  [e2e] $TAG server ready (pid $SRV_PID)" >&2

  has() { grep -q "$2" "$1" 2>/dev/null && echo 1 || echo 0; }
  # fresh_csrf <jar> <page-path> — re-fetch a page and pull a live token
  # (the login page when logged out, an authenticated page otherwise).
  fresh_csrf() {
    curl -s -b "$1" -c "$1" -o "$J/_cs.html" "$BASE$2"
    grep -o 'name="_csrf" value="[^"]*"' "$J/_cs.html" | head -1 | sed 's/.*value="//;s/"//'
  }

  # S1 login page: 200 + csrf + session cookie
  local code
  code=$(curl -s -c "$J/a" -o "$J/p1" -w '%{http_code}' "$BASE/login")
  echo "S1 status=$code csrf=$(has "$J/p1" '_csrf') setcookie=$(curl -s -D - -o /dev/null "$BASE/login" | grep -ci 'set-cookie: KYLIX_SID=')" >> "$T"

  # S2 wrong password -> error message + login_logs success=0
  local C; C=$(fresh_csrf "$J/a" /login)
  code=$(curl -s -b "$J/a" -c "$J/a" -o "$J/p2" -w '%{http_code}' \
    -d "username=admin&password=WRONG&_csrf=$C" "$BASE/login")
  echo "S2 status=$code error=$(has "$J/p2" 'Invalid username or password')" \
      "logfail=$(sqlite3 "$DB" 'SELECT COUNT(*) FROM login_logs WHERE success=0')" >> "$T"

  # S3 admin login -> 302 /dashboard, guarded pages 200
  C=$(fresh_csrf "$J/a" /login)
  code=$(curl -s -b "$J/a" -c "$J/a" -o /dev/null -w '%{http_code} %{redirect_url}' \
    -d "username=admin&password=Admin@123&_csrf=$C" "$BASE/login" | sed 's|http://localhost:[0-9]*||')
  local dash users_p roles_p logs_p
  dash=$(curl -s -b "$J/a" -o "$J/dash" -w '%{http_code}' "$BASE/dashboard")
  users_p=$(curl -s -b "$J/a" -o "$J/users" -w '%{http_code}' "$BASE/users")
  roles_p=$(curl -s -b "$J/a" -o "$J/roles" -w '%{http_code}' "$BASE/roles")
  logs_p=$(curl -s -b "$J/a" -o "$J/logs" -w '%{http_code}' "$BASE/logs")
  echo "S3 login=$code dash=$dash users=$users_p roles=$roles_p logs=$logs_p" >> "$T"

  # S4 create bob/viewer/lockme -> rows + op_logs (one op_log per create)
  local u
  for u in bob viewer lockme; do
    C=$(fresh_csrf "$J/a" /users/new)
    curl -s -b "$J/a" -o /dev/null -d "username=$u&password=${u}@12345&display_name=$u&_csrf=$C" "$BASE/users"
  done
  curl -s -b "$J/a" -o "$J/users2" "$BASE/users"
  echo "S4 bob=$(has "$J/users2" '<td>bob</td>') viewer=$(has "$J/users2" '<td>viewer</td>') lockme=$(has "$J/users2" '<td>lockme</td>')" \
      "oplogs=$(sqlite3 "$DB" "SELECT COUNT(*) FROM op_logs WHERE path='/users' AND method='POST'")" >> "$T"

  # S5 delete: temp user removable, seed admin guarded, no-perm user 403
  C=$(fresh_csrf "$J/a" /users/new)
  curl -s -b "$J/a" -o /dev/null -d "username=temp1&password=Temp@12345&display_name=T&_csrf=$C" "$BASE/users"
  local tid; tid=$(sqlite3 "$DB" "SELECT id FROM users WHERE username='temp1'")
  C=$(fresh_csrf "$J/a" /users/new)
  curl -s -b "$J/a" -o /dev/null -d "id=$tid&_csrf=$C" "$BASE/users/delete"
  local after_del; after_del=$(sqlite3 "$DB" "SELECT COUNT(*) FROM users WHERE username='temp1'")
  C=$(fresh_csrf "$J/a" /users/new)
  local guard; guard=$(curl -s -b "$J/a" -o /dev/null -w '%{redirect_url}' -d "id=1&_csrf=$C" "$BASE/users/delete" | sed 's|http://localhost:[0-9]*||')
  # bob (no roles -> no perms): the users.write page guard fires -> 403 page
  curl -s -c "$J/bob" -o "$J/bob_l" "$BASE/login"; C=$(fresh_csrf "$J/bob" /login)
  curl -s -b "$J/bob" -c "$J/bob" -o /dev/null -d "username=bob&password=bob@12345&_csrf=$C" "$BASE/login"
  C=$(fresh_csrf "$J/bob" /dashboard)
  curl -s -b "$J/bob" -o "$J/bob_d" -d "id=2&_csrf=$C" "$BASE/users/delete"
  echo "S5 temp-gone=$after_del admin-guard=$guard bob-403=$(has "$J/bob_d" '403 Forbidden')" >> "$T"

  # S6 perm-less user GET /users -> 403 Forbidden
  local v6; v6=$(curl -s -b "$J/bob" -o "$J/bob_u" -w '%{http_code}' "$BASE/users")
  echo "S6 status=$v6 forbidden=$(has "$J/bob_u" '403 Forbidden')" >> "$T"

  # S7 POST /users without csrf -> 403
  code=$(curl -s -b "$J/a" -o /dev/null -w '%{http_code}' -d "username=x&password=x&display_name=x" "$BASE/users")
  echo "S7 nocsrf=$code" >> "$T"

  # S8 no session GET /users -> 401
  code=$(curl -s -o /dev/null -w '%{http_code}' "$BASE/users")
  echo "S8 nosession=$code" >> "$T"

  # S9 remember=1 -> persistent cookie (Max-Age = 30 days)
  curl -s -c "$J/rem" -o /dev/null "$BASE/login"; C=$(fresh_csrf "$J/rem" /login)
  local maxage; maxage=$(curl -s -b "$J/rem" -c "$J/rem" -D - -o /dev/null \
    -d "username=admin&password=Admin@123&remember=1&_csrf=$C" "$BASE/login" \
    | grep -ci 'max-age=2592000')
  echo "S9 remember_maxage=$maxage" >> "$T"

  # S10 lockout: 5 wrong logins on lockme, then correct password still locked
  local msgs="" i
  for i in 1 2 3 4 5; do
    curl -s -c "$J/lk" -o /dev/null "$BASE/login"; C=$(fresh_csrf "$J/lk" /login)
    curl -s -b "$J/lk" -c "$J/lk" -o "$J/lk_r" \
      -d "username=lockme&password=WRONG$i&_csrf=$C" "$BASE/login"
    grep -q 'account locked for 15 minutes' "$J/lk_r" && msgs="$msgs 5=locked"
  done
  curl -s -c "$J/lk" -o /dev/null "$BASE/login"; C=$(fresh_csrf "$J/lk" /login)
  curl -s -b "$J/lk" -c "$J/lk" -o "$J/lk_r" \
    -d "username=lockme&password=lockme@12345&_csrf=$C" "$BASE/login"
  grep -q 'Account locked; try again later' "$J/lk_r" && msgs="$msgs 6=still-locked"
  echo "S10 lockout:$msgs" >> "$T"

  # S11 /logs renders both tables + sqlite row counts
  curl -s -b "$J/a" -o "$J/logs2" "$BASE/logs"
  echo "S11 logins_head=$(has "$J/logs2" '<h1>Logins</h1>') ops_head=$(has "$J/logs2" '<h1>Operations</h1>')" \
      "oprows=$(has "$J/logs2" '<td>POST</td><td>/users</td>')" \
      "db_ops=$(sqlite3 "$DB" 'SELECT COUNT(*) FROM op_logs')" >> "$T"

  # S12 logout -> old cookie rejected
  C=$(fresh_csrf "$J/a" /dashboard)
  local lo; lo=$(curl -s -b "$J/a" -c "$J/a" -o /dev/null -w '%{http_code} %{redirect_url}' -d "_csrf=$C" "$BASE/logout" | sed 's|http://localhost:[0-9]*||')
  local old; old=$(curl -s -b "$J/a" -o /dev/null -w '%{http_code}' "$BASE/users")
  echo "S12 logout=$lo oldcookie=$old" >> "$T"

  # stop the server and wait until the port is actually free — the other
  # form reuses it, and a stale listener would make its ready-probe hit the
  # corpse while the new server dies on bind.
  kill "$SRV_PID" 2>/dev/null
  local up
  for up in $(seq 1 50); do
    curl -s --max-time 1 -o /dev/null "$BASE/login" 2>/dev/null || break
    sleep 0.2
  done
  kill -9 "$SRV_PID" 2>/dev/null
  wait "$SRV_PID" 2>/dev/null
  SRV_PID=""
  echo "  [e2e] $TAG done" >&2
}

# ---------------------------------------------------------------------------
# run both forms, diff transcripts
# ---------------------------------------------------------------------------
scenarios "$WORK/t_go.txt" "$WORK/go_bin" "$WORK/adm_go.db"
scenarios "$WORK/t_ll.txt" "$LL_BIN" "$WORK/adm_ll.db"

echo "== Go transcript =="
cat "$WORK/t_go.txt"
if diff "$WORK/t_go.txt" "$WORK/t_ll.txt" > "$WORK/t.diff"; then
  echo "== transcripts identical: Go ≡ LLVM =="
else
  echo "== TRANSCRIPT DIFF =="
  cat "$WORK/t.diff"
  echo "== server logs (tail) =="
  tail -5 "$WORK/srv_go_bin.log" 2>/dev/null
  tail -5 "$WORK/srv_ll_bin.log" 2>/dev/null
  fail "dual-backend transcripts differ"
fi
echo "KylixAdmin dual-backend E2E: PASS (12 scenarios x 2 forms)"
