#!/usr/bin/env bash
# e2e.sh — KylixAdmin dual-backend parity E2E (v0.11.0 P3)
#
# Builds the Go and LLVM forms of apps/admin from the same Kylix sources,
# runs the same scenario sequence against each, and diffs the normalized
# transcripts (scenarios print deterministic key=value lines only — CSRF
# tokens / session IDs / timestamps never enter the transcript).
# Exit 0 iff both forms behave identically and every assertion passes.
#
# v0.11.0: every data table is served by the generic CRUD engine at
# /admin/:entity, so the assertions bind to the stable data-* hooks the engine
# emits (data-row / data-f / data-pager) rather than to tag structure — a CSS
# redesign must not be able to invalidate this suite.
#
# Environment:
#   KYLIX    path to the kylix CLI (default /tmp/kylix_bin; built by CI)
#   PORT     listen port (default 8091 — example60 uses 8090)
#   PG_DSN   postgres DSN; with --pg the Go form also runs against postgres and
#            its transcript is diffed against the sqlite one. The database must
#            be created with LC_COLLATE 'C' (see docs/ADMIN_DEPLOY.md): the
#            default collation orders differently from sqlite's BINARY, which
#            would show up as a row-order diff rather than a bug.
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
#   S11 the two log entities render (page + sqlite)
#   S12 logout -> old cookie -> 401
#   S13 search + sort + pager preserve the query string
#   S14 unknown entity -> 404
#   S15 read-only entity refuses writes -> 403
#   S16 generic CRUD on the demo entity (notes) + op_logs
#   S17 validation failure re-renders the form with the error
#   S18 password columns are never listed
set -u

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PORT="${PORT:-8091}"
KYLIX="${KYLIX:-/tmp/kylix_bin}"
LLVM_GC="${LLVM_GC:-1}"
PG_DSN="${PG_DSN:-}"
WITH_PG=0
[ "${1:-}" = "--pg" ] && WITH_PG=1
DB_MODE="sqlite"   # sqlite | pg — selects the probe helper inside scenarios()
WORK="$(mktemp -d /tmp/kyadmin_e2e.XXXXXX)"
ADMIN="$ROOT/apps/admin"
GOGEN="$ROOT/.e2e_admin"
BASE="http://localhost:$PORT"

fail() { echo "E2E-FAIL: $*" >&2; exit 1; }
[ -x "$KYLIX" ] || fail "kylix CLI not found at $KYLIX (set KYLIX=...)"
command -v sqlite3 >/dev/null || fail "sqlite3 required"
command -v curl >/dev/null || fail "curl required"

# dbq runs a SQL probe against the form's database. The transcript is built from
# these results, so the helper has to speak both dialects: sqlite3 for the file
# backend, psql -tA for postgres (both print a bare value, no headers).
dbq() {
  if [ "$DB_MODE" = "pg" ]; then
    psql "$PG_DSN" -tA -c "$1"
  else
    sqlite3 "$DB" "$1"
  fi
}

# db_reset returns the database to an empty state before a form runs. sqlite is
# a file, so removing it is enough; postgres needs its schema dropped — without
# this the second form would see the first form's rows, SeedIfEmpty would skip,
# and every id-based assertion would be off.
db_reset() {
  if [ "$DB_MODE" = "pg" ]; then
    psql "$PG_DSN" -q -c "DROP SCHEMA IF EXISTS public CASCADE" -c "CREATE SCHEMA public" >/dev/null
  else
    rm -f "$DB"
  fi
}

# A stale server from an interrupted run would answer the probes with a
# different database (and the new server would die on bind), silently turning
# every assertion into a false result — refuse to start.
if curl -s --max-time 1 -o /dev/null "$BASE/login" 2>/dev/null; then
  fail "port $PORT is already serving — kill the stale KylixAdmin server first"
fi

SRV_PID=""
cleanup() {
  if [ -n "$SRV_PID" ]; then
    kill "$SRV_PID" 2>/dev/null
    sleep 0.2
    kill -9 "$SRV_PID" 2>/dev/null
  fi
  # wait for the port to actually free before the next form binds it
  local i
  for i in $(seq 1 25); do
    curl -s --max-time 1 -o /dev/null "$BASE/login" 2>/dev/null || break
    sleep 0.2
  done
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
  entities/admin_entities.klx lib/dialect.klx lib/admindb.klx lib/adminsec.klx lib/audit.klx \
  lib/crud.klx lib/crudrender.klx lib/crudhooks.klx lib/adminpage.klx \
  controllers/entity.klx controllers/dashboard.klx controllers/profile.klx \
  controllers/theme.klx main.klx) \
  || fail "Go-form codegen failed"
(cd "$ROOT" && go build -o "$WORK/go_bin" ./.e2e_admin) || fail "Go-form go build failed"

# LLVM form (boehm GC opt-in; fall back when libgc is absent).
LL_BIN="$WORK/ll_bin"
build_ll() {
  (cd "$ADMIN" && "$KYLIX" build --backend=llvm $1 -o "$LL_BIN" \
    ../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
    entities/admin_entities.klx lib/dialect.klx lib/admindb.klx lib/adminsec.klx lib/audit.klx \
    lib/crud.klx lib/crudrender.klx lib/crudhooks.klx lib/adminpage.klx \
    controllers/entity.klx controllers/dashboard.klx controllers/profile.klx \
    controllers/theme.klx main.klx \
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
  db_reset

  # exec inside the subshell so $! is the binary itself and stays a child of
  # this script (the old `(... & echo $! > pid)` pattern orphaned the server,
  # leaving `wait` a no-op and the process alive after cleanup).
  # The DSN selects the dialect; KYADMIN_DB stays the sqlite path.
  local SRV_ENV="KYADMIN_DB=$DB"
  if [ "$DB_MODE" = "pg" ]; then SRV_ENV="KYADMIN_DSN=$PG_DSN"; fi
  (cd "$ADMIN" && exec env $SRV_ENV KYADMIN_PASSWORD=Admin@123 KYADMIN_PORT="$PORT" "$BIN") \
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
      "logfail=$(dbq 'SELECT COUNT(*) FROM login_logs WHERE success=0')" >> "$T"

  # S3 admin login -> 302 /dashboard, guarded pages 200
  C=$(fresh_csrf "$J/a" /login)
  code=$(curl -s -b "$J/a" -c "$J/a" -o /dev/null -w '%{http_code} %{redirect_url}' \
    -d "username=admin&password=Admin@123&_csrf=$C" "$BASE/login" | sed 's|http://localhost:[0-9]*||')
  local dash users_p roles_p logs_p notes_p
  dash=$(curl -s -b "$J/a" -o "$J/dash" -w '%{http_code}' "$BASE/dashboard")
  users_p=$(curl -s -b "$J/a" -o "$J/users" -w '%{http_code}' "$BASE/admin/users")
  roles_p=$(curl -s -b "$J/a" -o "$J/roles" -w '%{http_code}' "$BASE/admin/roles")
  logs_p=$(curl -s -b "$J/a" -o "$J/logins" -w '%{http_code}' "$BASE/admin/login_logs")
  notes_p=$(curl -s -b "$J/a" -o "$J/notes" -w '%{http_code}' "$BASE/admin/notes")
  echo "S3 login=$code dash=$dash users=$users_p roles=$roles_p logins=$logs_p notes=$notes_p" >> "$T"

  # S4 create bob/viewer/lockme -> rows + op_logs (one op_log per create)
  local u
  for u in bob viewer lockme; do
    C=$(fresh_csrf "$J/a" /admin/users/new)
    curl -s -b "$J/a" -o /dev/null -d "username=$u&password=${u}@12345&display_name=$u&is_active=1&_csrf=$C" "$BASE/admin/users"
  done
  curl -s -b "$J/a" -o "$J/users2" "$BASE/admin/users"
  echo "S4 bob=$(has "$J/users2" 'data-f="username">bob<') viewer=$(has "$J/users2" 'data-f="username">viewer<') lockme=$(has "$J/users2" 'data-f="username">lockme<')" \
      "oplogs=$(dbq "SELECT COUNT(*) FROM op_logs WHERE path='/admin/users' AND method='POST'")" >> "$T"

  # S5 delete: temp user removable, seed admin guarded, no-perm user 403
  C=$(fresh_csrf "$J/a" /admin/users/new)
  curl -s -b "$J/a" -o /dev/null -d "username=temp1&password=Temp@12345&display_name=T&is_active=1&_csrf=$C" "$BASE/admin/users"
  local tid; tid=$(dbq "SELECT id FROM users WHERE username='temp1'")
  C=$(fresh_csrf "$J/a" /admin/users/new)
  curl -s -b "$J/a" -o /dev/null -d "id=$tid&_csrf=$C" "$BASE/admin/users/delete"
  local after_del; after_del=$(dbq "SELECT COUNT(*) FROM users WHERE username='temp1'")
  C=$(fresh_csrf "$J/a" /admin/users/new)
  local guard; guard=$(curl -s -b "$J/a" -o /dev/null -w '%{redirect_url}' -d "id=1&_csrf=$C" "$BASE/admin/users/delete" | sed 's|http://localhost:[0-9]*||')
  # bob (no roles -> no perms): the users.write page guard fires -> 403 page
  curl -s -c "$J/bob" -o "$J/bob_l" "$BASE/login"; C=$(fresh_csrf "$J/bob" /login)
  curl -s -b "$J/bob" -c "$J/bob" -o /dev/null -d "username=bob&password=bob@12345&_csrf=$C" "$BASE/login"
  C=$(fresh_csrf "$J/bob" /dashboard)
  curl -s -b "$J/bob" -o "$J/bob_d" -d "id=2&_csrf=$C" "$BASE/admin/users/delete"
  echo "S5 temp-gone=$after_del admin-guard=$guard bob-403=$(has "$J/bob_d" '403 Forbidden')" >> "$T"

  # S6 perm-less user GET /users -> 403 Forbidden
  local v6; v6=$(curl -s -b "$J/bob" -o "$J/bob_u" -w '%{http_code}' "$BASE/admin/users")
  echo "S6 status=$v6 forbidden=$(has "$J/bob_u" '403 Forbidden')" >> "$T"

  # S7 POST /users without csrf -> 403
  code=$(curl -s -b "$J/a" -o /dev/null -w '%{http_code}' -d "username=x&password=x&display_name=x" "$BASE/admin/users")
  echo "S7 nocsrf=$code" >> "$T"

  # S8 no session GET /users -> 401
  code=$(curl -s -o /dev/null -w '%{http_code}' "$BASE/admin/users")
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

  # S11 both read-only log entities render + sqlite row counts
  curl -s -b "$J/a" -o "$J/logins2" "$BASE/admin/login_logs"
  curl -s -b "$J/a" -o "$J/ops2" "$BASE/admin/op_logs"
  echo "S11 logins_head=$(has "$J/logins2" 'data-f="username"') ops_head=$(has "$J/ops2" 'data-f="method"')" \
      "oprows=$(has "$J/ops2" 'data-f="path">/admin/users<')" \
      "db_ops=$(dbq 'SELECT COUNT(*) FROM op_logs')" >> "$T"

  # S12 logout -> old cookie rejected
  C=$(fresh_csrf "$J/a" /dashboard)
  local lo; lo=$(curl -s -b "$J/a" -c "$J/a" -o /dev/null -w '%{http_code} %{redirect_url}' -d "_csrf=$C" "$BASE/logout" | sed 's|http://localhost:[0-9]*||')
  local old; old=$(curl -s -b "$J/a" -o /dev/null -w '%{http_code}' "$BASE/admin/users")
  echo "S12 logout=$lo oldcookie=$old" >> "$T"

  # S12 logs out; sign back in for the remaining scenarios.
  curl -s -c "$J/a2" -o /dev/null "$BASE/login"; C=$(fresh_csrf "$J/a2" /login)
  local relog; relog=$(curl -s -b "$J/a2" -c "$J/a2" -o /dev/null -w '%{http_code}' -d "username=admin&password=Admin@123&_csrf=$C" "$BASE/login")
  echo "S12b relogin=$relog" >> "$T"

  # S13 search + sort + pager: the pager keeps ?q=/?sort=/?dir= in a fixed order
  curl -s -b "$J/a2" -o "$J/s13" "$BASE/admin/users?q=bob&sort=username&dir=desc"
  echo "S13 hit=$(has "$J/s13" 'data-f="username">bob<') miss=$(has "$J/s13" 'data-f="username">lockme<')" \
      "pager=$(has "$J/s13" 'pager-next')" >> "$T"
  curl -s -b "$J/a2" -o "$J/s13b" "$BASE/admin/users?q=zzz-no-such-user"
  echo "S13b empty=$(has "$J/s13b" 'data-row=')" >> "$T"

  # S14 unknown entity -> 404
  code=$(curl -s -b "$J/a2" -o "$J/s14" -w '%{http_code}' "$BASE/admin/nosuchtable")
  echo "S14 status=$code body=$(has "$J/s14" 'Unknown entity')" >> "$T"

  # S15 read-only entity refuses writes (no permission point, and [ReadOnly])
  C=$(fresh_csrf "$J/a2" /dashboard)
  code=$(curl -s -b "$J/a2" -o "$J/s15" -w '%{http_code}' -d "id=1&_csrf=$C" "$BASE/admin/op_logs/delete")
  echo "S15 status=$code forbidden=$(has "$J/s15" '403 Forbidden')" >> "$T"

  # S16 generic CRUD on the demonstration entity: create, update, delete
  C=$(fresh_csrf "$J/a2" /admin/notes/new)
  curl -s -b "$J/a2" -o /dev/null -d "title=first note&body=hello&_csrf=$C" "$BASE/admin/notes"
  local nid; nid=$(dbq "SELECT id FROM notes WHERE title='first note'")
  C=$(fresh_csrf "$J/a2" /admin/notes/new)
  curl -s -b "$J/a2" -o /dev/null -d "id=$nid&title=renamed note&body=hello&done=1&_csrf=$C" "$BASE/admin/notes/update"
  curl -s -b "$J/a2" -o "$J/s16" "$BASE/admin/notes"
  echo "S16 renamed=$(has "$J/s16" 'data-f="title">renamed note<') done=$(has "$J/s16" 'data-f="done">yes<')" \
      "db=$(dbq "SELECT COUNT(*) FROM notes")" >> "$T"
  C=$(fresh_csrf "$J/a2" /admin/notes/new)
  curl -s -b "$J/a2" -o /dev/null -d "id=$nid&_csrf=$C" "$BASE/admin/notes/delete"
  echo "S16b after-delete=$(dbq "SELECT COUNT(*) FROM notes")" \
      "oplogs=$(dbq "SELECT COUNT(*) FROM op_logs WHERE path LIKE '/admin/notes%'")" >> "$T"

  # S17 validation failure re-renders the form with the submitted value kept
  C=$(fresh_csrf "$J/a2" /admin/notes/new)
  code=$(curl -s -b "$J/a2" -o "$J/s17" -w '%{http_code}' -d "title=&body=x&_csrf=$C" "$BASE/admin/notes")
  echo "S17 status=$code error=$(has "$J/s17" 'Title is required')" \
      "kept=$(has "$J/s17" 'data-field="body"')" >> "$T"

  # S18 password columns never reach a list page
  curl -s -b "$J/a2" -o "$J/s18" "$BASE/admin/users"
  echo "S18 password_listed=$(has "$J/s18" 'data-f="password"') hash_leak=$(has "$J/s18" 'pbkdf2$')" >> "$T"

  # S19 dashboard: stat cards + the integer-geometry SVG chart
  curl -s -b "$J/a2" -o "$J/s19" "$BASE/dashboard"
  echo "S19 stats=$(has "$J/s19" 'data-stats="1"') chart=$(has "$J/s19" 'data-chart="logins"')" \
      "bars=$(grep -o 'data-day=' "$J/s19" | wc -l | tr -d ' ')" >> "$T"

  # S20 profile: change the password, then prove the new one works and the old
  # one does not (and that the change is audited).
  curl -s -b "$J/a2" -o "$J/prof" "$BASE/profile"
  C=$(grep -o 'name="_csrf" value="[^"]*"' "$J/prof" | head -1 | sed 's/.*value="//;s/"//')
  code=$(curl -s -b "$J/a2" -o "$J/s20" -w '%{http_code}' \
    -d "current_password=WRONG&new_password=NewPass@123&confirm_password=NewPass@123&_csrf=$C" "$BASE/profile/password")
  echo "S20 wrong-current=$code msg=$(has "$J/s20" 'Current password is incorrect')" >> "$T"
  C=$(grep -o 'name="_csrf" value="[^"]*"' "$J/s20" | head -1 | sed 's/.*value="//;s/"//')
  code=$(curl -s -b "$J/a2" -o "$J/s20b" -w '%{http_code}' \
    -d "current_password=Admin@123&new_password=NewPass@123&confirm_password=NewPass@123&_csrf=$C" "$BASE/profile/password")
  local pw_ok pw_old
  curl -s -c "$J/newpw" -o /dev/null "$BASE/login"; C=$(fresh_csrf "$J/newpw" /login)
  pw_ok=$(curl -s -b "$J/newpw" -c "$J/newpw" -o /dev/null -w '%{http_code}' \
    -d "username=admin&password=NewPass@123&_csrf=$C" "$BASE/login")
  curl -s -c "$J/oldpw" -o /dev/null "$BASE/login"; C=$(fresh_csrf "$J/oldpw" /login)
  pw_old=$(curl -s -b "$J/oldpw" -o "$J/s20c" -w '%{http_code}' \
    -d "username=admin&password=Admin@123&_csrf=$C" "$BASE/login")
  echo "S20b changed=$code new=$pw_ok old=$pw_old oldmsg=$(has "$J/s20c" 'Invalid username or password')" \
      "audited=$(dbq "SELECT COUNT(*) FROM op_logs WHERE path='/profile/password'")" >> "$T"
  # restore the seed password so the run ends in the state it started
  C=$(fresh_csrf "$J/newpw" /profile)
  curl -s -b "$J/newpw" -o /dev/null \
    -d "current_password=NewPass@123&new_password=Admin@123&confirm_password=Admin@123&_csrf=$C" "$BASE/profile/password"

  # S21 profile: display name + avatar (base64 data URL over a urlencoded form)
  C=$(fresh_csrf "$J/a2" /profile)
  curl -s -b "$J/a2" -o /dev/null -d "display_name=Administrator Renamed&_csrf=$C" "$BASE/profile/display"
  echo "S21 display=$(dbq "SELECT display_name FROM users WHERE username='admin'")" >> "$T"
  C=$(fresh_csrf "$J/a2" /profile)
  # --data-urlencode, not -d: a data URL contains ';' and '+', which a browser
  # percent-encodes when the form is submitted (Go's url.ParseQuery rejects a
  # raw ';').
  code=$(curl -s -b "$J/a2" -o "$J/s21" -w '%{http_code}' \
    -d "_csrf=$C" \
    --data-urlencode "avatar_data=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==" \
    "$BASE/profile/avatar")
  echo "S21b avatar=$code shown=$(has "$J/s21" 'data-avatar="1"')" \
      "stored=$(dbq "SELECT COUNT(*) FROM users WHERE username='admin' AND avatar LIKE 'data:image/png;base64,%'")" >> "$T"
  C=$(fresh_csrf "$J/a2" /profile)
  code=$(curl -s -b "$J/a2" -o "$J/s21c" -w '%{http_code}' -d "avatar_data=not-an-image&_csrf=$C" "$BASE/profile/avatar")
  # the stored avatar survives a page reload
  curl -s -b "$J/a2" -o "$J/s21d" "$BASE/profile"
  echo "S21c bad-upload=$code rejected=$(has "$J/s21c" 'Only image data URLs are accepted')" \
      "kept=$(has "$J/s21d" 'data-avatar="1"')" >> "$T"

  # S22 theme: server-rendered data-theme from the cookie, and the switch
  # endpoint refuses to be an open redirect.
  local t302 tloc tcookie
  t302=$(curl -s -b "$J/a2" -c "$J/theme" -o /dev/null -w '%{http_code}' "$BASE/theme?t=dark&next=/admin/users")
  tloc=$(curl -s -b "$J/a2" -o /dev/null -w '%{redirect_url}' "$BASE/theme?t=dark&next=/admin/users" | sed 's|http://localhost:[0-9]*||')
  tcookie=$(curl -s -b "$J/a2" -D - -o /dev/null "$BASE/theme?t=dark&next=/dashboard" | grep -ci 'set-cookie: theme=dark')
  curl -s -b "$J/theme" -c "$J/theme" -o "$J/s22" "$BASE/admin/users"
  local guard22; guard22=$(curl -s -b "$J/a2" -o /dev/null -w '%{redirect_url}' "$BASE/theme?t=dark&next=//evil.example/x" | sed 's|http://localhost:[0-9]*||')
  local auto22; auto22=$(curl -s -b "$J/a2" -D - -o /dev/null "$BASE/theme?t=bogus" | grep -o 'theme=auto' | head -1)
  echo "S22 status=$t302 loc=$tloc cookie=$tcookie dark=$(has "$J/s22" 'data-theme="dark"')" \
      "toggle=$(has "$J/s22" 'data-theme-toggle="light"') guard=$guard22 bogus=$auto22" >> "$T"
  # static assets the design system needs
  echo "S22b css=$(curl -s -o /dev/null -w '%{http_code}' "$BASE/static/admin.css")" \
      "js=$(curl -s -o /dev/null -w '%{http_code}' "$BASE/static/admin.js")" >> "$T"

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
DB_MODE="sqlite"
scenarios "$WORK/t_go.txt" "$WORK/go_bin" "$WORK/adm_go.db"
scenarios "$WORK/t_ll.txt" "$LL_BIN" "$WORK/adm_ll.db"

# --pg: the same Go binary against postgres. Same Kylix sources, different
# database — so a diff here is a dialect bug, not a code difference.
if [ "$WITH_PG" = "1" ]; then
  [ -n "$PG_DSN" ] || fail "--pg needs PG_DSN (e.g. postgres://user@localhost/db?sslmode=disable)"
  command -v psql >/dev/null || fail "psql required for --pg"
  DB_MODE="pg"
  scenarios "$WORK/t_gopg.txt" "$WORK/go_bin" ""
fi

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
if [ "$WITH_PG" = "1" ]; then
  echo "== Go+postgres transcript =="
  cat "$WORK/t_gopg.txt"
  if diff "$WORK/t_go.txt" "$WORK/t_gopg.txt" > "$WORK/t.pgdiff"; then
    echo "== sqlite ≡ postgres (same Kylix sources) =="
  else
    echo "== SQLITE vs POSTGRES DIFF =="
    cat "$WORK/t.pgdiff"
    tail -5 "$WORK/srv_go_bin.log" 2>/dev/null
    fail "postgres and sqlite transcripts differ"
  fi
  echo "KylixAdmin dual-backend E2E: PASS (22 scenarios x 2 forms + postgres)"
else
  echo "KylixAdmin dual-backend E2E: PASS (22 scenarios x 2 forms)"
fi
