#!/usr/bin/env bash
# migrate_check.sh — incremental migration check (v0.12.0 P5e)
#
# The E2E always starts from an empty database, so it only ever exercises the
# "create the tables from [Entity] metadata" path. This script covers the other
# half: a database that already exists with a *partial* table must gain the
# columns the metadata declares, via ALTER TABLE ADD COLUMN.
#
#   KYLIX=/tmp/kylix_bin bash apps/admin/migrate_check.sh            # sqlite
#   PG_DSN=postgres://... bash apps/admin/migrate_check.sh --pg      # postgres
set -u
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ADMIN="$ROOT/apps/admin"
KYLIX="${KYLIX:-/tmp/kylix_bin}"
PORT="${PORT:-8092}"
PG_DSN="${PG_DSN:-}"
MODE="sqlite"
[ "${1:-}" = "--pg" ] && MODE="pg"
WORK="$(mktemp -d /tmp/kyadmin_mig.XXXXXX)"

fail() { echo "MIGRATE-FAIL: $*" >&2; exit 1; }
SRV_PID=""
cleanup() {
  [ -n "$SRV_PID" ] && { kill "$SRV_PID" 2>/dev/null; sleep 0.2; kill -9 "$SRV_PID" 2>/dev/null; }
  [ "${KEEP:-0}" = "1" ] || rm -rf "$WORK"
}
trap cleanup EXIT

BUILD_FILES="../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
  entities/admin_entities.klx lib/dialect.klx lib/migrate.klx lib/admindb.klx \
  lib/adminsec.klx lib/audit.klx lib/crud.klx lib/crudrender.klx lib/crudhooks.klx \
  lib/adminpage.klx controllers/entity.klx controllers/dashboard.klx \
  controllers/profile.klx controllers/theme.klx main.klx"

echo "[migrate] building the LLVM form (${MODE})" >&2
(cd "$ADMIN" && $KYLIX build --backend=llvm --gc=boehm -o "$WORK/bin" $BUILD_FILES > "$WORK/build.log" 2>&1) \
  || { tail -5 "$WORK/build.log"; fail "build failed"; }

if [ "$MODE" = "pg" ]; then
  [ -n "$PG_DSN" ] || fail "--pg needs PG_DSN"
  psql "$PG_DSN" -q -c "DROP SCHEMA IF EXISTS public CASCADE" -c "CREATE SCHEMA public" >/dev/null
  psql "$PG_DSN" -q -c "CREATE TABLE users (id BIGSERIAL PRIMARY KEY, username TEXT, password TEXT)" >/dev/null
  ENV="KYADMIN_DSN=$PG_DSN"
  cols() { psql "$PG_DSN" -tA -c "SELECT string_agg(attname, ',' ORDER BY attnum) FROM pg_catalog.pg_attribute WHERE attrelid = to_regclass('users') AND attnum > 0 AND NOT attisdropped"; }
else
  DB="$WORK/partial.db"
  sqlite3 "$DB" "CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT, password TEXT)"
  ENV="KYADMIN_DB=$DB"
  cols() { sqlite3 "$DB" "SELECT group_concat(name) FROM pragma_table_info('users')"; }
fi

echo "[migrate] before: $(cols)" >&2
(cd "$ADMIN" && exec env $ENV KYADMIN_PASSWORD=Admin@123 KYADMIN_PORT="$PORT" "$WORK/bin") > "$WORK/srv.log" 2>&1 &
SRV_PID=$!
for _ in $(seq 1 50); do curl -s -o /dev/null "http://localhost:$PORT/login" && break; sleep 0.3; done

after="$(cols)"
echo "[migrate] after:  $after" >&2
for want in avatar display_name is_active failed_attempts locked_until created_at updated_at; do
  case ",$after," in
    *",$want,"*) ;;
    *) fail "column $want was not added (after=$after)";;
  esac
done
echo "migrate_check: PASS (${MODE}: 3-column users table extended to $(echo "$after" | tr ',' '\n' | wc -l | tr -d ' ') columns)"
