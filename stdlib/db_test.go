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

// A NULL column reads as the empty string on both backends. It used to be
// "<nil>" here and a segmentation fault on the LLVM side (strdup(NULL) after
// sqlite3_column_text returned NULL), which took the admin profile page down.
func TestDb_QueryScalarNullColumn(t *testing.T) {
	db, _ := DbOpenSQLite(":memory:")
	defer DbClose(db)

	DbExec(db, "CREATE TABLE t (s TEXT)")
	DbExec(db, "INSERT INTO t VALUES (NULL)")

	val, err := DbQueryScalar(db, "SELECT s FROM t")
	if err != nil {
		t.Fatalf("DbQueryScalar failed: %v", err)
	}
	if val != "" {
		t.Errorf("scalar = %q, want empty for a NULL column", val)
	}
}

// v0.12.0 P5b: the placeholder rewrite runs at the single choke point every
// statement passes through. sqlite takes `?` as written; lib/pq needs $n.
func TestDb_RewritePlaceholders(t *testing.T) {
	sqlite := &Database{dbType: DBSQLite}
	pg := &Database{dbType: DBPostgres}

	cases := []struct {
		name string
		db   *Database
		in   string
		want string
	}{
		{"sqlite is the identity", sqlite, "SELECT * FROM t WHERE a = ? AND b = ?", "SELECT * FROM t WHERE a = ? AND b = ?"},
		{"postgres numbers them", pg, "SELECT * FROM t WHERE a = ? AND b = ?", "SELECT * FROM t WHERE a = $1 AND b = $2"},
		{"no placeholders", pg, "SELECT COUNT(*) FROM t", "SELECT COUNT(*) FROM t"},
		{"limit and offset", pg, "SELECT c FROM t WHERE (length(?) = 0 OR c LIKE ?) LIMIT ? OFFSET ?",
			"SELECT c FROM t WHERE (length($1) = 0 OR c LIKE $2) LIMIT $3 OFFSET $4"},
		// A ? inside a string literal must not consume a parameter number.
		{"literal question mark", pg, "SELECT * FROM t WHERE note = 'why?' AND id = ?",
			"SELECT * FROM t WHERE note = 'why?' AND id = $1"},
		{"escaped quote inside literal", pg, "SELECT * FROM t WHERE note = 'it''s ?' AND id = ?",
			"SELECT * FROM t WHERE note = 'it''s ?' AND id = $1"},
	}
	for _, c := range cases {
		if got := c.db.rewritePlaceholders(c.in); got != c.want {
			t.Errorf("%s: rewritePlaceholders = %q, want %q", c.name, got, c.want)
		}
	}
}

// The postgres driver hands back []byte for numeric/bpchar/name/json/uuid; the
// Kylix side renders values with %v, which would print "[49 48]" rather than
// "10". The LLVM backend boxes everything non-numeric as text, so folding
// []byte into string is also what keeps the backends aligned.
func TestDb_NormalizeValue(t *testing.T) {
	if got := normalizeValue([]byte("10")); got != "10" {
		t.Errorf("normalizeValue([]byte) = %#v, want the string \"10\"", got)
	}
	if got := normalizeValue(int64(7)); got != int64(7) {
		t.Errorf("normalizeValue(int64) = %#v, want it unchanged", got)
	}
	if got := normalizeValue(nil); got != nil {
		t.Errorf("normalizeValue(nil) = %#v, want nil", got)
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
