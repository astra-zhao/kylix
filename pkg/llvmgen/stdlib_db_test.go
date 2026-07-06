package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_db tests — verify TDatabase lowers to libsqlite3-backed defines
// and inlined prepare/bind/step/finalize sequences.

func TestDb_DbOpenSQLiteCallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_db_DbOpenSQLite")
	assertIRContains(t, ir, "call i32 @sqlite3_open")
	if strings.Contains(ir, "db.DbOpenSQLite not implemented") {
		t.Errorf("DbOpenSQLite still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestDb_DbOpenSQLiteBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_db_DbOpenSQLite(ptr %path)")
}

func TestDb_DbCloseBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
  DbClose(db);
end.`)
	assertIRContains(t, ir, "define void @__kylix_db_DbClose(ptr %db)")
	assertIRContains(t, ir, "call i32 @sqlite3_close")
}

func TestDb_DbExecInlinedPrepareBindStep(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
  DbExec(db, 'CREATE TABLE t (id INTEGER)');
end.`)
	// DbExec is inlined: prepare + step + finalize (no separate define)
	assertIRContains(t, ir, "call i32 @sqlite3_prepare_v2")
	assertIRContains(t, ir, "call i32 @sqlite3_step")
	assertIRContains(t, ir, "call i32 @sqlite3_finalize")
}

func TestDb_DbExecWithStringBind(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
  DbExec(db, 'INSERT INTO t VALUES (?)', 'alice');
end.`)
	// String arg → bind_text
	assertIRContains(t, ir, "call i32 @sqlite3_bind_text")
}

func TestDb_DbExecWithIntegerBind(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
  DbExec(db, 'INSERT INTO t VALUES (?)', 30);
end.`)
	// Integer arg → bind_int64
	assertIRContains(t, ir, "call i32 @sqlite3_bind_int64")
}

func TestDb_DbQueryScalarInlined(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
  var s := DbQueryScalar(db, 'SELECT COUNT(*) FROM t');
end.`)
	assertIRContains(t, ir, "call ptr @sqlite3_column_text")
	// Uses htab_strdup to copy the column text
	assertIRContains(t, ir, "call ptr @__kylix_htab_strdup")
}

func TestDb_SqliteDeclarations(t *testing.T) {
	ir := generateIR(t, `program p;
uses db;
begin
  var db := DbOpenSQLite(':memory:');
end.`)
	assertIRContains(t, ir, "declare i32 @sqlite3_open")
	assertIRContains(t, ir, "declare i32 @sqlite3_prepare_v2")
	assertIRContains(t, ir, "declare i32 @sqlite3_step")
	assertIRContains(t, ir, "declare i32 @sqlite3_finalize")
	assertIRContains(t, ir, "declare ptr @sqlite3_column_text")
}

func TestDb_NotUsedNoSymbols(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "@__kylix_db_") {
		t.Errorf("db symbol emitted without `uses db`\nIR:\n%s", ir)
	}
}
