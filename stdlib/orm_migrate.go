// orm_migrate.go — ORM high-level operations, migration manager, and scan helpers.
package stdlib

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ORM provides high-level CRUD operations on top of Database.
type ORM struct {
	db *Database
}

// NewORM creates an ORM instance.
func NewORM(db *Database) *ORM {
	return &ORM{db: db}
}

// Insert inserts a row and returns the last inserted ID.
func (o *ORM) Insert(table string, data map[string]interface{}) (int64, error) {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	result, err := o.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// InsertModel inserts a struct using its `db` field tags.
func (o *ORM) InsertModel(table string, model interface{}) (int64, error) {
	return o.Insert(table, structToMap(model))
}

// Update updates rows matching condition with the given data.
func (o *ORM) Update(table string, condition, data map[string]interface{}) (int64, error) {
	setClauses := make([]string, 0, len(data))
	whereClauses := make([]string, 0, len(condition))
	args := make([]interface{}, 0)

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}
	for col, val := range condition {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		table, strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))

	result, err := o.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete deletes rows matching condition.
func (o *ORM) Delete(table string, condition map[string]interface{}) (int64, error) {
	whereClauses := make([]string, 0, len(condition))
	args := make([]interface{}, 0)

	for col, val := range condition {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		table, strings.Join(whereClauses, " AND "))

	result, err := o.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Find retrieves a single row by primary key.
func (o *ORM) Find(table string, id int64) (map[string]interface{}, error) {
	rows, err := o.db.Query(fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1", table), id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanToMap(rows)
}

// FindAll retrieves all rows from a table.
func (o *ORM) FindAll(table string) ([]map[string]interface{}, error) {
	return o.QueryAll(fmt.Sprintf("SELECT * FROM %s", table))
}

// Query retrieves a single row via a raw query.
func (o *ORM) Query(query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := o.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanToMap(rows)
}

// QueryAll retrieves all rows via a raw query.
func (o *ORM) QueryAll(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := o.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanToSlice(rows)
}

// QueryBuilder returns a new query builder for the table.
func (o *ORM) QueryBuilder(table string) *QueryBuilder {
	return NewQueryBuilder(table)
}

// Execute runs a QueryBuilder SELECT and returns all matching rows.
func (o *ORM) Execute(qb *QueryBuilder) ([]map[string]interface{}, error) {
	query, args := qb.BuildSelect()
	return o.QueryAll(query, args...)
}

// Count returns the number of rows matching a QueryBuilder.
func (o *ORM) Count(qb *QueryBuilder) (int64, error) {
	query, args := qb.BuildCount()
	var count int64
	return count, o.db.QueryRow(query, args...).Scan(&count)
}

// Exists reports whether any row in table matches condition.
func (o *ORM) Exists(table string, condition map[string]interface{}) (bool, error) {
	whereClauses := make([]string, 0, len(condition))
	args := make([]interface{}, 0)

	for col, val := range condition {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)",
		table, strings.Join(whereClauses, " AND "))

	var exists bool
	return exists, o.db.QueryRow(query, args...).Scan(&exists)
}

// ── Migration manager ─────────────────────────────────────────────────────────

// Migration holds a versioned schema change with up/down SQL.
type Migration struct {
	Version     string
	Description string
	Up          string
	Down        string
}

// MigrationManager tracks and applies database migrations.
type MigrationManager struct {
	db         *Database
	table      string
	migrations []Migration
}

// NewMigrationManager creates a migration manager using the given database.
func NewMigrationManager(db *Database) *MigrationManager {
	return &MigrationManager{db: db, table: "migrations", migrations: make([]Migration, 0)}
}

func (m *MigrationManager) SetTableName(name string) { m.table = name }

func (m *MigrationManager) AddMigration(version, description, up, down string) {
	m.migrations = append(m.migrations, Migration{version, description, up, down})
}

// CreateMigrationsTable creates the migrations tracking table if absent.
func (m *MigrationManager) CreateMigrationsTable() error {
	autoInc := "AUTO_INCREMENT"
	if m.db.dbType == DBSQLite {
		autoInc = "AUTOINCREMENT"
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY %s,
		version VARCHAR(255) NOT NULL UNIQUE,
		description TEXT,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`, m.table, autoInc)
	_, err := m.db.Exec(query)
	return err
}

// GetAppliedMigrations returns the versions that have already been applied.
func (m *MigrationManager) GetAppliedMigrations() ([]string, error) {
	rows, err := m.db.Query(fmt.Sprintf("SELECT version FROM %s ORDER BY version", m.table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make([]string, 0)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// Migrate applies all pending migrations in order.
func (m *MigrationManager) Migrate() error {
	if err := m.CreateMigrationsTable(); err != nil {
		return err
	}
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}
	appliedMap := make(map[string]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}
	for _, mig := range m.migrations {
		if appliedMap[mig.Version] {
			continue
		}
		if _, err := m.db.Exec(mig.Up); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", mig.Version, err)
		}
		if _, err := m.db.Exec(
			fmt.Sprintf("INSERT INTO %s (version, description) VALUES (?, ?)", m.table),
			mig.Version, mig.Description,
		); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", mig.Version, err)
		}
		fmt.Printf("Applied migration: %s - %s\n", mig.Version, mig.Description)
	}
	return nil
}

// Rollback rolls back the most recently applied migration.
func (m *MigrationManager) Rollback() error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		return errors.New("no migrations to rollback")
	}
	return m.rollbackVersion(applied[len(applied)-1])
}

// RollbackTo rolls back all migrations newer than version (exclusive).
func (m *MigrationManager) RollbackTo(version string) error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}
	for i := len(applied) - 1; i >= 0; i-- {
		if applied[i] == version {
			break
		}
		if err := m.rollbackVersion(applied[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *MigrationManager) rollbackVersion(version string) error {
	var mig *Migration
	for i := range m.migrations {
		if m.migrations[i].Version == version {
			mig = &m.migrations[i]
			break
		}
	}
	if mig == nil {
		return fmt.Errorf("migration %s not found", version)
	}
	if _, err := m.db.Exec(mig.Down); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", version, err)
	}
	if _, err := m.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.table), version); err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", version, err)
	}
	fmt.Printf("Rolled back migration: %s - %s\n", mig.Version, mig.Description)
	return nil
}

// Status returns the applied/pending state of all registered migrations.
func (m *MigrationManager) Status() ([]map[string]interface{}, error) {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}
	appliedMap := make(map[string]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}
	status := make([]map[string]interface{}, 0, len(m.migrations))
	for _, mig := range m.migrations {
		status = append(status, map[string]interface{}{
			"version":     mig.Version,
			"description": mig.Description,
			"applied":     appliedMap[mig.Version],
		})
	}
	return status, nil
}

// ── Scan helpers ──────────────────────────────────────────────────────────────

// structToMap converts a struct to a column→value map using `db` field tags.
func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return result
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		if fv := val.Field(i); fv.CanInterface() {
			result[tag] = fv.Interface()
		}
	}
	return result
}

// scanToMap scans the first row of rows into a map[column]value.
func scanToMap(rows *sql.Rows) (map[string]interface{}, error) {
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, len(columns))
	ptrs := make([]interface{}, len(columns))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		if b, ok := values[i].([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = values[i]
		}
	}
	return result, nil
}

// scanToSlice scans all rows into a []map[column]value.
func scanToSlice(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}
		results = append(results, row)
	}
	return results, nil
}
