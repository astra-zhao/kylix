#!/usr/bin/env bash
# deploy_check.sh — the self-contained-binary gate (v0.12.0 P7)
#
# KylixAdmin bakes its templates and static assets into the binary with
# [Embed('views', 'static')], and the database defaults under $HOME. This script
# proves the claim the deployment guide makes: copy the binary somewhere with
# nothing beside it and it still serves the whole console.
#
#   KYLIX=/tmp/kylix_bin bash apps/admin/deploy_check.sh            # sqlite
#   KYLIX=/tmp/kylix_bin PG_DSN=... bash apps/admin/deploy_check.sh --pg
set -u
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ADMIN="$ROOT/apps/admin"
KYLIX="${KYLIX:-/tmp/kylix_bin}"
PORT="${PORT:-8093}"
PG_DSN="${PG_DSN:-}"
MODE="sqlite"
[ "${1:-}" = "--pg" ] && MODE="pg"
WORK="$(mktemp -d /tmp/kyadmin_deploy.XXXXXX)"
EMPTY="$WORK/empty"

fail() { echo "DEPLOY-FAIL: $*" >&2; exit 1; }
SRV_PID=""
cleanup() {
  [ -n "$SRV_PID" ] && { kill "$SRV_PID" 2>/dev/null; sleep 0.2; kill -9 "$SRV_PID" 2>/dev/null; }
  [ "${KEEP:-0}" = "1" ] || rm -rf "$WORK"
}
trap cleanup EXIT

FILES="../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
  entities/admin_entities.klx lib/dialect.klx lib/migrate.klx lib/admindb.klx \
  lib/adminsec.klx lib/audit.klx lib/crud.klx lib/crudrender.klx lib/crudhooks.klx \
  lib/adminpage.klx controllers/entity.klx controllers/dashboard.klx \
  controllers/profile.klx controllers/theme.klx main.klx"

echo "[deploy] building the LLVM form" >&2
(cd "$ADMIN" && $KYLIX build --backend=llvm --gc=boehm -o "$WORK/kylixadmin" $FILES > "$WORK/build.log" 2>&1) \
  || { tail -5 "$WORK/build.log"; fail "build failed"; }

# The point of the check: nothing but the binary.
mkdir -p "$EMPTY"
cp "$WORK/kylixadmin" "$EMPTY/"
cd "$EMPTY"

if [ "$MODE" = "pg" ]; then
  [ -n "$PG_DSN" ] || fail "--pg needs PG_DSN"
  psql "$PG_DSN" -q -c "DROP SCHEMA IF EXISTS public CASCADE" -c "CREATE SCHEMA public" >/dev/null
  ENV="KYADMIN_DSN=$PG_DSN"
else
  ENV="KYADMIN_DB=$WORK/admin.db"
fi

ls | grep -q '^views$' && fail "the working directory is not empty"
(exec env $ENV KYADMIN_PASSWORD=Admin@123 KYADMIN_PORT="$PORT" ./kylixadmin) > "$WORK/srv.log" 2>&1 &
SRV_PID=$!
for _ in $(seq 1 50); do curl -s -o /dev/null "http://localhost:$PORT/login" && break; sleep 0.3; done

J="$WORK/jar"; rm -f "$J"
tok=$(curl -s -c "$J" "http://localhost:$PORT/login" | grep -o 'name="_csrf" value="[^"]*"' | head -1 | sed 's/.*value="//;s/"//')
[ -n "$tok" ] || { cat "$WORK/srv.log"; fail "login page did not render (templates not embedded?)"; }
curl -s -b "$J" -c "$J" -o /dev/null -d "_csrf=$tok&username=admin&password=Admin@123" "http://localhost:$PORT/login"

users=$(curl -s -b "$J" -o "$WORK/u.html" -w '%{http_code}' "http://localhost:$PORT/admin/users")
rows=$(grep -c 'data-row=' "$WORK/u.html")
css=$(curl -s -o /dev/null -w '%{http_code} %{size_download}' "http://localhost:$PORT/static/admin.css")
js=$(curl -s -o /dev/null -w '%{http_code}' "http://localhost:$PORT/static/admin.js")

echo "[deploy] users=$users rows=$rows css=$css js=$js" >&2
[ "$users" = "200" ] || fail "GET /admin/users = $users"
[ "$rows" -ge 1 ] || fail "no rows rendered"
case "$css" in 200*) ;; *) fail "static css = $css";; esac
[ "$js" = "200" ] || fail "static js = $js"

echo "deploy_check: PASS (${MODE}: self-contained binary in an empty directory)"
