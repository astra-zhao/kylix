// db.go — Kylix-friendly database convenience layer.
//
// Wraps the lower-level orm.Database with simple open/query/exec helpers
// tuned for Kylix call sites. SQLite is the default tutorial driver (in-memory
// :memory:) so examples run without external services.
package stdlib

import (
	"database/sql"
	"fmt"
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

// DbOpenSQLite opens an SQLite database file (use ":memory:" for in-memory).
func DbOpenSQLite(path string) (*Database, error) {
	return DbOpen("sqlite3", path)
}

// DbExec executes a statement (INSERT/UPDATE/DELETE/DDL) and returns rows affected.
func DbExec(db *Database, query string, args ...interface{}) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database is nil")
	}
	res, err := db.Exec(query, args...)
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
	rows, err := db.Query(query, args...)
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
			row[col] = values[i]
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
	if err := db.QueryRow(query, args...).Scan(&v); err != nil {
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
	return fmt.Sprintf("%v", v), nil
}

// DbClose closes the database connection pool.
func DbClose(db *Database) error {
	if db == nil {
		return nil
	}
	return db.Close()
}
