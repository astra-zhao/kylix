package stdlib

import (
	"testing"
)

func TestDb_OpenSQLiteMemory(t *testing.T) {
	db, err := DbOpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("DbOpenSQLite failed: %v", err)
	}
	defer DbClose(db)
}

func TestDb_ExecAndQuery(t *testing.T) {
	db, err := DbOpenSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer DbClose(db)

	// Create table
	affected, err := DbExec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
	if affected != 0 {
		t.Errorf("CREATE TABLE rows affected = %d, want 0", affected)
	}

	// Insert
	affected, err = DbExec(db, "INSERT INTO users (name, age) VALUES (?, ?)", "alice", 30)
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}
	if affected != 1 {
		t.Errorf("INSERT rows affected = %d, want 1", affected)
	}

	// Insert second row
	DbExec(db, "INSERT INTO users (name, age) VALUES (?, ?)", "bob", 25)

	// Query all rows
	rows, err := DbQueryRows(db, "SELECT name, age FROM users ORDER BY age")
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["name"] != "bob" {
		t.Errorf("first row name = %v, want bob", rows[0]["name"])
	}
}

func TestDb_QueryScalar(t *testing.T) {
	db, _ := DbOpenSQLite(":memory:")
	defer DbClose(db)

	DbExec(db, "CREATE TABLE t (n INTEGER)")
	DbExec(db, "INSERT INTO t VALUES (42)")

	val, err := DbQueryScalar(db, "SELECT n FROM t")
	if err != nil {
		t.Fatalf("DbQueryScalar failed: %v", err)
	}
	if val != "42" {
		t.Errorf("scalar = %q, want 42", val)
	}
}

func TestDb_QueryScalarNoRows(t *testing.T) {
	db, _ := DbOpenSQLite(":memory:")
	defer DbClose(db)

	DbExec(db, "CREATE TABLE t (n INTEGER)")
	val, err := DbQueryScalar(db, "SELECT n FROM t")
	if err != nil {
		t.Fatalf("DbQueryScalar failed: %v", err)
	}
	if val != "" {
		t.Errorf("scalar = %q, want empty for no rows", val)
	}
}

func TestDb_NilGuards(t *testing.T) {
	if _, err := DbExec(nil, "SELECT 1"); err == nil {
		t.Error("DbExec(nil) should error")
	}
	if _, err := DbQueryRows(nil, "SELECT 1"); err == nil {
		t.Error("DbQueryRows(nil) should error")
	}
	if _, err := DbQueryScalar(nil, "SELECT 1"); err == nil {
		t.Error("DbQueryScalar(nil) should error")
	}
	if err := DbClose(nil); err != nil {
		t.Errorf("DbClose(nil) should be nil, got %v", err)
	}
}
