// db.go — Kylix-friendly database convenience layer.
//
// Wraps the lower-level orm.Database with simple open/query/exec helpers
// tuned for Kylix call sites. SQLite is the default tutorial driver (in-memory
// :memory:) so examples run without external services.
package stdlib

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DbOpen opens a database by driver name and DSN.
// driver: "sqlite3" | "mysql" | "postgres"
// Returns a *Database (with sensible pool defaults) or an error.
func DbOpen(driver, dsn string) (*Database, error) {
	cfg := &ConnectionConfig{
		Type:     DatabaseType(driver),
		Database: dsn,
	}
	// For mysql/postgres the DSN is the connection string verbatim;
	// ConnectionString() falls back to cfg.Database for sqlite3 and unknown
	// drivers, so passing the raw DSN works for all three.
	if driver == "mysql" || driver == "postgres" {
		// ConnectionString() builds from host/port/user/pass fields which we
		// don't have here; bypass by stashing the raw DSN in Database and
		// opening directly.
		return openRaw(driver, dsn)
	}
	return NewDatabase(cfg)
}

// openRaw opens a database from a raw DSN without going through ConnectionConfig.
func openRaw(driver, dsn string) (*Database, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	d := &Database{
		db:      db,
		dbType:  DatabaseType(driver),
		config:  &ConnectionConfig{Type: DatabaseType(driver)},
		maxIdle: 10,
		maxOpen: 100,
	}
	d.db.SetMaxIdleConns(d.maxIdle)
	d.db.SetMaxOpenConns(d.maxOpen)
	// v0.12.0: mirror NewDatabase — without a lifetime, a long-running server
	// holds connections a server-side restart has already dropped.
	d.db.SetConnMaxLifetime(time.Hour)
	return d, nil
}

// DbSetMaxOpenConns caps the pool's open connections (0 = unlimited).
func DbSetMaxOpenConns(db *Database, n int64) {
	if db == nil || db.db == nil {
		return
	}
	db.maxOpen = int(n)
	db.db.SetMaxOpenConns(int(n))
}

// DbSetMaxIdleConns sets how many idle connections are kept (0 = none).
func DbSetMaxIdleConns(db *Database, n int64) {
	if db == nil || db.db == nil {
		return
	}
	db.maxIdle = int(n)
	db.db.SetMaxIdleConns(int(n))
}

// DbSetConnMaxLifetime sets the maximum lifetime of a connection, in seconds
// (0 = no limit).
func DbSetConnMaxLifetime(db *Database, seconds int64) {
	if db == nil || db.db == nil {
		return
	}
	db.lifetime = time.Duration(seconds) * time.Second
	db.db.SetConnMaxLifetime(db.lifetime)
}

// DbLastError returns the most recent statement error ("" when the last
// statement succeeded). The generated code discards the error half of the
// stdlib db results, so this is the only way for a Kylix program to notice a
// failed statement — and on postgres a silently failed statement looks exactly
// like an empty result set.
func DbLastError(db *Database) string {
	if db == nil {
		return "database handle is nil"
	}
	return db.lastErr
}

// noteErr records a statement error for DbLastError.
func (d *Database) noteErr(err error) {
	if d == nil {
		return
	}
	if err != nil {
		d.lastErr = err.Error()
		return
	}
	d.lastErr = ""
}

// DbOpenPg opens a postgres connection from a DSN (libpq connection string or
// URL). v0.12.0: it is a separate entry point rather than a driver string so
// the LLVM backend can emit the postgres code paths only for programs that
// actually use them — the tutorials never call it, so they never link libpq.
func DbOpenPg(dsn string) (*Database, error) {
	return DbOpen("postgres", dsn)
}

// DbOpenSQLite opens an SQLite database file (use ":memory:" for in-memory).
func DbOpenSQLite(path string) (*Database, error) {
	return DbOpen("sqlite3", path)
}

// rewritePlaceholders converts the Kylix-facing `?` placeholders to the form
// the active driver expects: lib/pq needs `$1, $2, ...`, everything else takes
// `?` as written.
//
// It runs here, at the single choke point every statement passes through,
// rather than at the 73 call sites in apps/admin. Kylix string literals have no
// escape sequence, so a `?` can legitimately appear inside SQL text; a
// per-call-site rewrite would be one silent mistake away from shifting every
// parameter by one — and that mistake is invisible on sqlite, where the
// rewrite is the identity.
//
// A `?` inside a single-quoted SQL literal is left alone (with ” as the
// escaped quote, per SQL). None of the current statements need that, but the
// scanner handles it so the rule does not depend on how the SQL is written.
func (d *Database) rewritePlaceholders(query string) string {
	if d == nil || d.dbType != DBPostgres || !strings.ContainsRune(query, '?') {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 0
	inLiteral := false
	for i := 0; i < len(query); i++ {
		c := query[i]
		if c == '\'' {
			// '' inside a literal is an escaped quote, not a terminator.
			if inLiteral && i+1 < len(query) && query[i+1] == '\'' {
				b.WriteString("''")
				i++
				continue
			}
			inLiteral = !inLiteral
			b.WriteByte(c)
			continue
		}
		if c == '?' && !inLiteral {
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// normalizeValue maps driver-specific value types onto the shapes the Kylix
// side expects. sqlite returns TEXT as string and INTEGER as int64; lib/pq
// returns most text types as string too, but numeric/bpchar/name/json/uuid
// arrive as []byte — which would render as "[49 48]" instead of "10". The LLVM
// backend boxes everything that is not an int/float/bool as text, so folding
// []byte into string is also what keeps the two backends aligned.
func normalizeValue(v interface{}) interface{} {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

// DbExec executes a statement (INSERT/UPDATE/DELETE/DDL) and returns rows affected.
func DbExec(db *Database, query string, args ...interface{}) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database is nil")
	}
	res, err := db.Exec(db.rewritePlaceholders(query), args...)
	if err != nil {
		db.noteErr(err)
		return 0, err
	}
	n, err := res.RowsAffected()
	db.noteErr(err)
	return n, err
}

// DbQueryRows runs a SELECT and returns all rows as a slice of map[string]interface{}.
// Each map is keyed by column name. Values use the driver's native Go types.
func DbQueryRows(db *Database, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}
	rows, err := db.Query(db.rewritePlaceholders(query), args...)
	if err != nil {
		db.noteErr(err)
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		db.noteErr(err)
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			db.noteErr(err)
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = normalizeValue(values[i])
		}
		result = append(result, row)
	}
	db.noteErr(rows.Err())
	return result, rows.Err()
}

// DbQueryScalar runs a SELECT expected to return a single row/column and
// returns the first column value as a string, or "" if no rows.
func DbQueryScalar(db *Database, query string, args ...interface{}) (string, error) {
	if db == nil {
		return "", fmt.Errorf("database is nil")
	}
	var v interface{}
	if err := db.QueryRow(db.rewritePlaceholders(query), args...).Scan(&v); err != nil {
		if err == sql.ErrNoRows {
			db.noteErr(nil)
			return "", nil
		}
		db.noteErr(err)
		return "", err
	}
	db.noteErr(nil)
	// A NULL column reads as the empty string, not "<nil>" — the LLVM backend
	// returns "" for the same query (and used to crash on it).
	if v == nil {
		return "", nil
	}
	return fmt.Sprintf("%v", normalizeValue(v)), nil
}

// DbClose closes the database connection pool.
func DbClose(db *Database) error {
	if db == nil {
		return nil
	}
	return db.Close()
}
