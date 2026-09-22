package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_db_pg tests — v0.12.0 P5d. The postgres backend is emitted only for
// programs that call DbOpenPg: the declares alone would make compile.go link
// -lpq (it scans the IR text), and the tutorial CI jobs do not install libpq.

const pgProgram = `program p;
uses db;
begin
  var db := DbOpenPg('postgres://localhost/x');
  DbExec(db, 'CREATE TABLE t (a BIGINT)');
  var n := DbQueryScalar(db, 'SELECT COUNT(*) FROM t');
  var rows := DbQueryRows(db, 'SELECT a FROM t WHERE a = ?', '1');
end.`

func TestDbPg_EmittedWhenUsed(t *testing.T) {
	ir := generateIR(t, pgProgram)
	assertIRContains(t, ir, "declare ptr @PQconnectdb(ptr noundef)")
	assertIRContains(t, ir, "define ptr @__kylix_db_pg_open(ptr %dsn)")
	assertIRContains(t, ir, "define ptr @__kylix_db_pg_rewrite(ptr %sql, i64 %nargs)")
	assertIRContains(t, ir, "define i64 @__kylix_db_pg_exec(ptr %conn, ptr %sql, i64 %argc, ptr %argv, ptr %argtypes)")
	assertIRContains(t, ir, "define ptr @__kylix_db_pg_scalar(ptr %conn, ptr %sql, i64 %argc, ptr %argv, ptr %argtypes)")
	assertIRContains(t, ir, "define { ptr, i64, i64 } @__kylix_db_pg_rows(ptr %conn, ptr %sql, i64 %argc, ptr %argv, ptr %argtypes)")
	assertIRContains(t, ir, "@__kylix_db_is_pg = global i32 0")
	// the runtime dialect dispatch the shared call sites go through
	assertIRContains(t, ir, "load i32, ptr @__kylix_db_is_pg")
}

// A program that only uses sqlite must not carry a single libpq symbol: that is
// what keeps -lpq off the tutorials' link line.
func TestDbPg_AbsentWithoutDbOpenPg(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
  DbExec(db, 'CREATE TABLE t (a INTEGER)');
  var n := DbQueryScalar(db, 'SELECT COUNT(*) FROM t');
end.`)
	for _, unwanted := range []string{
		"@PQconnectdb",
		"@__kylix_db_pg_",
		"@__kylix_db_is_pg",
	} {
		if strings.Contains(ir, unwanted) {
			t.Errorf("sqlite-only program contains %q", unwanted)
		}
	}
}
