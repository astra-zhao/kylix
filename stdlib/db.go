// db.go — Kylix-friendly database convenience layer.
//
// Wraps the lower-level orm.Database with simple open/query/exec helpers
// tuned for Kylix call sites. SQLite is the default tutorial driver (in-memory
// :memory:) so examples run without external services.
package stdlib

import (
	"database/sql"
	"fmt"
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
	return d, nil
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
		return 0, err
	}
	return res.RowsAffected()
}

// DbQueryRows runs a SELECT and returns all rows as a slice of map[string]interface{}.
// Each map is keyed by column name. Values use the driver's native Go types.
func DbQueryRows(db *Database, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
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
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
	}
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
			return "", nil
		}
		return "", err
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
