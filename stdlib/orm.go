// orm.go — Database connection, transaction, and shared types.
package stdlib

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseType represents supported database drivers.
type DatabaseType string

const (
	DBMySQL    DatabaseType = "mysql"
	DBPostgres DatabaseType = "postgres"
	DBSQLite   DatabaseType = "sqlite3"
)

// ConnectionConfig holds all parameters needed to open a database connection.
type ConnectionConfig struct {
	Type     DatabaseType
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Options  map[string]string
}

// ConnectionString builds the driver-specific DSN.
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

// Database wraps *sql.DB with connection-pool configuration.
type Database struct {
	db       *sql.DB
	dbType   DatabaseType
	config   *ConnectionConfig
	maxIdle  int
	maxOpen  int
	lifetime time.Duration
}

// NewDatabase opens and pings the database, applying sensible pool defaults.
func NewDatabase(config *ConnectionConfig) (*Database, error) {
	db, err := sql.Open(string(config.Type), config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
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

func (d *Database) SetMaxIdleConns(n int) {
	d.maxIdle = n
	d.db.SetMaxIdleConns(n)
}

func (d *Database) SetMaxOpenConns(n int) {
	d.maxOpen = n
	d.db.SetMaxOpenConns(n)
}

func (d *Database) SetConnMaxLifetime(dur time.Duration) {
	d.lifetime = dur
	d.db.SetConnMaxLifetime(dur)
}

func (d *Database) Close() error      { return d.db.Close() }
func (d *Database) Ping() error       { return d.db.Ping() }
func (d *Database) GetSQLDB() *sql.DB { return d.db }

func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// Transaction wraps *sql.Tx.
type Transaction struct {
	tx *sql.Tx
}

// Begin starts a new transaction.
func (d *Database) Begin() (*Transaction, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

func (t *Transaction) Commit() error   { return t.tx.Commit() }
func (t *Transaction) Rollback() error { return t.tx.Rollback() }

func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// Model is the base struct for ORM models with standard ID and timestamp fields.
type Model struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
