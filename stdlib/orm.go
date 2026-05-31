package stdlib

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	DBMySQL    DatabaseType = "mysql"
	DBPostgres DatabaseType = "postgres"
	DBSQLite   DatabaseType = "sqlite3"
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Type     DatabaseType
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Options  map[string]string
}

// ConnectionString generates the database connection string
func (c *ConnectionConfig) ConnectionString() string {
	switch c.Type {
	case DBMySQL:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	case DBPostgres:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.Username, c.Password, c.Database)
	case DBSQLite:
		return c.Database
	default:
		return ""
	}
}

// Database represents a database connection wrapper
type Database struct {
	db       *sql.DB
	dbType   DatabaseType
	config   *ConnectionConfig
	maxIdle  int
	maxOpen  int
	lifetime time.Duration
}

// NewDatabase creates a new database connection
func NewDatabase(config *ConnectionConfig) (*Database, error) {
	connStr := config.ConnectionString()
	db, err := sql.Open(string(config.Type), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{
		db:       db,
		dbType:   config.Type,
		config:   config,
		maxIdle:  10,
		maxOpen:  100,
		lifetime: time.Hour,
	}

	d.db.SetMaxIdleConns(d.maxIdle)
	d.db.SetMaxOpenConns(d.maxOpen)
	d.db.SetConnMaxLifetime(d.lifetime)

	return d, nil
}

// SetMaxIdleConns sets the maximum number of idle connections
func (d *Database) SetMaxIdleConns(n int) {
	d.maxIdle = n
	d.db.SetMaxIdleConns(n)
}

// SetMaxOpenConns sets the maximum number of open connections
func (d *Database) SetMaxOpenConns(n int) {
	d.maxOpen = n
	d.db.SetMaxOpenConns(n)
}

// SetConnMaxLifetime sets the maximum lifetime of connections
func (d *Database) SetConnMaxLifetime(dur time.Duration) {
	d.lifetime = dur
	d.db.SetConnMaxLifetime(dur)
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Ping pings the database
func (d *Database) Ping() error {
	return d.db.Ping()
}

// GetSQLDB returns the underlying *sql.DB
func (d *Database) GetSQLDB() *sql.DB {
	return d.db
}

// Query executes a query that returns rows
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Exec executes a query that doesn't return rows
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Transaction represents a database transaction
type Transaction struct {
	tx *sql.Tx
}

// Begin starts a new transaction
func (d *Database) Begin() (*Transaction, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// Query executes a query within the transaction
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// Model represents a database model
type Model struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// QueryBuilder builds SQL queries
type QueryBuilder struct {
	table       string
	conditions  []condition
	orderBy     []string
	limit       int
	offset      int
	joins       []string
	selectCols  []string
	distinct    bool
	groupBy     []string
	having      []condition
}

type condition struct {
	column   string
	operator string
	value    interface{}
	logic    string // AND, OR
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:      table,
		conditions: make([]condition, 0),
		orderBy:    make([]string, 0),
		selectCols: []string{"*"},
	}
}

// Select specifies the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = columns
	return qb
}

// Distinct adds DISTINCT to the query
func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.distinct = true
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(column string, operator string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{
		column:   column,
		operator: operator,
		value:    value,
		logic:    "AND",
	})
	return qb
}

// OrWhere adds an OR WHERE condition
func (qb *QueryBuilder) OrWhere(column string, operator string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{
		column:   column,
		operator: operator,
		value:    value,
		logic:    "OR",
	})
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	qb.conditions = append(qb.conditions, condition{
		column:   column,
		operator: "IN",
		value:    values,
		logic:    "AND",
	})
	return qb
}

// WhereBetween adds a WHERE BETWEEN condition
func (qb *QueryBuilder) WhereBetween(column string, min, max interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{
		column:   column,
		operator: "BETWEEN",
		value:    []interface{}{min, max},
		logic:    "AND",
	})
	return qb
}

// WhereNull adds a WHERE IS NULL condition
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{
		column:   column,
		operator: "IS NULL",
		value:    nil,
		logic:    "AND",
	})
	return qb
}

// WhereNotNull adds a WHERE IS NOT NULL condition
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{
		column:   column,
		operator: "IS NOT NULL",
		value:    nil,
		logic:    "AND",
	})
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("JOIN %s ON %s", table, condition))
	return qb
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

// RightJoin adds a RIGHT JOIN clause
func (qb *QueryBuilder) RightJoin(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition))
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

// GroupBy adds a GROUP BY clause
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// Having adds a HAVING condition
func (qb *QueryBuilder) Having(column string, operator string, value interface{}) *QueryBuilder {
	qb.having = append(qb.having, condition{
		column:   column,
		operator: operator,
		value:    value,
		logic:    "AND",
	})
	return qb
}

// Limit sets the LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Page sets pagination based on page number and page size
func (qb *QueryBuilder) Page(page, pageSize int) *QueryBuilder {
	qb.limit = pageSize
	qb.offset = (page - 1) * pageSize
	return qb
}

// BuildSelect builds a SELECT query
func (qb *QueryBuilder) BuildSelect() (string, []interface{}) {
	var query strings.Builder
	args := make([]interface{}, 0)

	query.WriteString("SELECT ")
	if qb.distinct {
		query.WriteString("DISTINCT ")
	}
	query.WriteString(strings.Join(qb.selectCols, ", "))
	query.WriteString(" FROM ")
	query.WriteString(qb.table)

	for _, join := range qb.joins {
		query.WriteString(" ")
		query.WriteString(join)
	}

	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		for i, cond := range qb.conditions {
			if i > 0 {
				query.WriteString(" " + cond.logic + " ")
			}
			query.WriteString(cond.column)

			switch cond.operator {
			case "IN":
				values := cond.value.([]interface{})
				placeholders := make([]string, len(values))
				for j, v := range values {
					placeholders[j] = "?"
					args = append(args, v)
				}
				query.WriteString(fmt.Sprintf(" IN (%s)", strings.Join(placeholders, ",")))
			case "BETWEEN":
				values := cond.value.([]interface{})
				query.WriteString(" BETWEEN ? AND ?")
				args = append(args, values[0], values[1])
			case "IS NULL", "IS NOT NULL":
				query.WriteString(" " + cond.operator)
			default:
				query.WriteString(fmt.Sprintf(" %s ?", cond.operator))
				args = append(args, cond.value)
			}
		}
	}

	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		for i, cond := range qb.having {
			if i > 0 {
				query.WriteString(" " + cond.logic + " ")
			}
			query.WriteString(cond.column)
			query.WriteString(fmt.Sprintf(" %s ?", cond.operator))
			args = append(args, cond.value)
		}
	}

	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY ")
		query.WriteString(strings.Join(qb.orderBy, ", "))
	}

	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), args
}

// BuildCount builds a COUNT query
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	origCols := qb.selectCols
	qb.selectCols = []string{"COUNT(*) as count"}
	query, args := qb.BuildSelect()
	qb.selectCols = origCols
	return query, args
}

// ORM provides high-level database operations
type ORM struct {
	db *Database
}

// NewORM creates a new ORM instance
func NewORM(db *Database) *ORM {
	return &ORM{db: db}
}

// Insert inserts a record and returns the last insert ID
func (o *ORM) Insert(table string, data map[string]interface{}) (int64, error) {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := o.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// InsertModel inserts a struct as a record
func (o *ORM) InsertModel(table string, model interface{}) (int64, error) {
	data := structToMap(model)
	return o.Insert(table, data)
}

// Update updates records matching the condition
func (o *ORM) Update(table string, condition map[string]interface{}, data map[string]interface{}) (int64, error) {
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

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		table,
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "),
	)

	result, err := o.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete deletes records matching the condition
func (o *ORM) Delete(table string, condition map[string]interface{}) (int64, error) {
	whereClauses := make([]string, 0, len(condition))
	args := make([]interface{}, 0)

	for col, val := range condition {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		table,
		strings.Join(whereClauses, " AND "),
	)

	result, err := o.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Find retrieves a single record by ID
func (o *ORM) Find(table string, id int64) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ? LIMIT 1", table)
	rows, err := o.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanToMap(rows)
}

// FindAll retrieves all records from a table
func (o *ORM) FindAll(table string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s", table)
	return o.QueryAll(query)
}

// Query retrieves a single record
func (o *ORM) Query(query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := o.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanToMap(rows)
}

// QueryAll retrieves all records matching the query
func (o *ORM) QueryAll(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := o.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanToSlice(rows)
}

// QueryBuilder creates a new query builder for the table
func (o *ORM) QueryBuilder(table string) *QueryBuilder {
	return NewQueryBuilder(table)
}

// Execute executes a query using the query builder
func (o *ORM) Execute(qb *QueryBuilder) ([]map[string]interface{}, error) {
	query, args := qb.BuildSelect()
	return o.QueryAll(query, args...)
}

// Count counts records matching the query
func (o *ORM) Count(qb *QueryBuilder) (int64, error) {
	query, args := qb.BuildCount()
	var count int64
	err := o.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// Exists checks if any records match the query
func (o *ORM) Exists(table string, condition map[string]interface{}) (bool, error) {
	whereClauses := make([]string, 0, len(condition))
	args := make([]interface{}, 0)

	for col, val := range condition {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	query := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM %s WHERE %s)",
		table,
		strings.Join(whereClauses, " AND "),
	)

	var exists bool
	err := o.db.QueryRow(query, args...).Scan(&exists)
	return exists, err
}

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          string
	Down        string
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db       *Database
	table    string
	migrations []Migration
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *Database) *MigrationManager {
	return &MigrationManager{
		db:       db,
		table:    "migrations",
		migrations: make([]Migration, 0),
	}
}

// SetTableName sets the migrations table name
func (m *MigrationManager) SetTableName(name string) {
	m.table = name
}

// AddMigration adds a migration
func (m *MigrationManager) AddMigration(version, description, up, down string) {
	m.migrations = append(m.migrations, Migration{
		Version:     version,
		Description: description,
		Up:          up,
		Down:        down,
	})
}

// CreateMigrationsTable creates the migrations table if it doesn't exist
func (m *MigrationManager) CreateMigrationsTable() error {
	autoInc := "AUTO_INCREMENT"
	if m.db.dbType == DBSQLite {
		autoInc = "AUTOINCREMENT"
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY %s,
			version VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, m.table, autoInc)

	_, err := m.db.Exec(query)
	return err
}

// GetAppliedMigrations returns the list of applied migrations
func (m *MigrationManager) GetAppliedMigrations() ([]string, error) {
	query := fmt.Sprintf("SELECT version FROM %s ORDER BY version", m.table)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make([]string, 0)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// Migrate applies pending migrations
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

	for _, migration := range m.migrations {
		if appliedMap[migration.Version] {
			continue
		}

		// Apply migration
		_, err := m.db.Exec(migration.Up)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		// Record migration
		query := fmt.Sprintf(
			"INSERT INTO %s (version, description) VALUES (?, ?)",
			m.table,
		)
		_, err = m.db.Exec(query, migration.Version, migration.Description)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		fmt.Printf("Applied migration: %s - %s\n", migration.Version, migration.Description)
	}

	return nil
}

// Rollback rolls back the last migration
func (m *MigrationManager) Rollback() error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		return errors.New("no migrations to rollback")
	}

	lastVersion := applied[len(applied)-1]

	// Find the migration
	var migration *Migration
	for _, mig := range m.migrations {
		if mig.Version == lastVersion {
			migration = &mig
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found", lastVersion)
	}

	// Rollback migration
	_, err = m.db.Exec(migration.Down)
	if err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", lastVersion, err)
	}

	// Remove migration record
	query := fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.table)
	_, err = m.db.Exec(query, lastVersion)
	if err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", lastVersion, err)
	}

	fmt.Printf("Rolled back migration: %s - %s\n", migration.Version, migration.Description)

	return nil
}

// RollbackTo rolls back migrations to a specific version
func (m *MigrationManager) RollbackTo(version string) error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Find migrations to rollback (from newest to oldest, stopping at version)
	for i := len(applied) - 1; i >= 0; i-- {
		if applied[i] == version {
			break
		}

		// Find the migration
		var migration *Migration
		for _, mig := range m.migrations {
			if mig.Version == applied[i] {
				migration = &mig
				break
			}
		}

		if migration == nil {
			return fmt.Errorf("migration %s not found", applied[i])
		}

		// Rollback migration
		_, err = m.db.Exec(migration.Down)
		if err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", applied[i], err)
		}

		// Remove migration record
		query := fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.table)
		_, err = m.db.Exec(query, applied[i])
		if err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", applied[i], err)
		}

		fmt.Printf("Rolled back migration: %s - %s\n", migration.Version, migration.Description)
	}

	return nil
}

// Status returns the status of migrations
func (m *MigrationManager) Status() ([]map[string]interface{}, error) {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}

	status := make([]map[string]interface{}, 0)
	for _, migration := range m.migrations {
		status = append(status, map[string]interface{}{
			"version":     migration.Version,
			"description": migration.Description,
			"applied":     appliedMap[migration.Version],
		})
	}

	return status, nil
}

// Helper functions

// structToMap converts a struct to a map using db tags
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

		fieldVal := val.Field(i)
		if !fieldVal.CanInterface() {
			continue
		}

		result[tag] = fieldVal.Interface()
	}

	return result
}

// scanToMap scans a single row into a map
func scanToMap(rows *sql.Rows) (map[string]interface{}, error) {
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}

	return result, nil
}

// scanToSlice scans all rows into a slice of maps
func scanToSlice(rows *sql.Rows) ([]map[string]interface{}, error) {
	results := make([]map[string]interface{}, 0)

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, nil
}
