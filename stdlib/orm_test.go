package stdlib

import (
	"testing"
	"time"
)

func TestConnectionConfig(t *testing.T) {
	// Test SQLite config
	sqliteConfig := &ConnectionConfig{
		Type:     DBSQLite,
		Database: "test.db",
	}
	connStr := sqliteConfig.ConnectionString()
	if connStr != "test.db" {
		t.Errorf("Expected 'test.db', got '%s'", connStr)
	}

	// Test MySQL config
	mysqlConfig := &ConnectionConfig{
		Type:     DBMySQL,
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "testdb",
	}
	connStr = mysqlConfig.ConnectionString()
	expected := "root:password@tcp(localhost:3306)/testdb?parseTime=true"
	if connStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, connStr)
	}

	// Test PostgreSQL config
	pgConfig := &ConnectionConfig{
		Type:     DBPostgres,
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "testdb",
	}
	connStr = pgConfig.ConnectionString()
	expected = "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"
	if connStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, connStr)
	}
}

func TestQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder("users")

	// Test basic select
	query, args := qb.Select("id", "name").Where("active", "=", true).BuildSelect()
	if query != "SELECT id, name FROM users WHERE active = ?" {
		t.Errorf("Unexpected query: %s", query)
	}
	if len(args) != 1 || args[0] != true {
		t.Errorf("Unexpected args: %v", args)
	}

	// Test multiple conditions
	qb2 := NewQueryBuilder("users")
	query, args = qb2.Where("age", ">", 18).OrWhere("role", "=", "admin").BuildSelect()
	if query != "SELECT * FROM users WHERE age > ? OR role = ?" {
		t.Errorf("Unexpected query: %s", query)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	// Test WHERE IN
	qb3 := NewQueryBuilder("users")
	query, args = qb3.WhereIn("id", []interface{}{1, 2, 3}).BuildSelect()
	if query != "SELECT * FROM users WHERE id IN (?,?,?)" {
		t.Errorf("Unexpected query: %s", query)
	}
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	// Test WHERE BETWEEN
	qb4 := NewQueryBuilder("users")
	query, args = qb4.WhereBetween("age", 18, 65).BuildSelect()
	if query != "SELECT * FROM users WHERE age BETWEEN ? AND ?" {
		t.Errorf("Unexpected query: %s", query)
	}
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	// Test WHERE NULL
	qb5 := NewQueryBuilder("users")
	query, args = qb5.WhereNull("deleted_at").BuildSelect()
	if query != "SELECT * FROM users WHERE deleted_at IS NULL" {
		t.Errorf("Unexpected query: %s", query)
	}

	// Test JOIN
	qb6 := NewQueryBuilder("users")
	query, _ = qb6.Join("orders", "users.id = orders.user_id").BuildSelect()
	if query != "SELECT * FROM users JOIN orders ON users.id = orders.user_id" {
		t.Errorf("Unexpected query: %s", query)
	}

	// Test ORDER BY and LIMIT
	qb7 := NewQueryBuilder("users")
	query, _ = qb7.OrderBy("created_at", "DESC").Limit(10).Offset(20).BuildSelect()
	if query != "SELECT * FROM users ORDER BY created_at DESC LIMIT 10 OFFSET 20" {
		t.Errorf("Unexpected query: %s", query)
	}

	// Test GROUP BY and HAVING
	qb8 := NewQueryBuilder("orders")
	query, args = qb8.Select("user_id", "COUNT(*) as count").
		GroupBy("user_id").
		Having("count", ">", 5).
		BuildSelect()
	expected := "SELECT user_id, COUNT(*) as count FROM orders GROUP BY user_id HAVING count > ?"
	if query != expected {
		t.Errorf("Expected '%s', got '%s'", expected, query)
	}

	// Test COUNT query
	qb9 := NewQueryBuilder("users")
	query, _ = qb9.Where("active", "=", true).BuildCount()
	if query != "SELECT COUNT(*) as count FROM users WHERE active = ?" {
		t.Errorf("Unexpected count query: %s", query)
	}
}

func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		ID        int       `db:"id"`
		Name      string    `db:"name"`
		Email     string    `db:"email"`
		CreatedAt time.Time `db:"created_at"`
		Ignored   string    // No tag, should be ignored
	}

	now := time.Now()
	obj := TestStruct{
		ID:        1,
		Name:      "John",
		Email:     "john@example.com",
		CreatedAt: now,
		Ignored:   "should not appear",
	}

	result := structToMap(obj)

	if result["id"] != 1 {
		t.Errorf("Expected id=1, got %v", result["id"])
	}
	if result["name"] != "John" {
		t.Errorf("Expected name='John', got %v", result["name"])
	}
	if result["email"] != "john@example.com" {
		t.Errorf("Expected email='john@example.com', got %v", result["email"])
	}
	if result["created_at"] != now {
		t.Errorf("Expected created_at=%v, got %v", now, result["created_at"])
	}
	if _, exists := result["Ignored"]; exists {
		t.Error("Field without db tag should not be in map")
	}
}

func TestModel(t *testing.T) {
	model := Model{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if model.ID != 1 {
		t.Errorf("Expected ID=1, got %d", model.ID)
	}
}

// Test ORM with in-memory SQLite database
func TestORMWithSQLite(t *testing.T) {
	// Create in-memory SQLite database
	config := &ConnectionConfig{
		Type:     DBSQLite,
		Database: ":memory:",
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			age INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	orm := NewORM(db)

	// Test Insert
	id, err := orm.Insert("users", map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   25,
	})
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}
	if id != 1 {
		t.Errorf("Expected id=1, got %d", id)
	}

	// Insert more data
	orm.Insert("users", map[string]interface{}{
		"name":  "Bob",
		"email": "bob@example.com",
		"age":   30,
	})

	// Test Find
	user, err := orm.Find("users", 1)
	if err != nil {
		t.Fatalf("Failed to find: %v", err)
	}
	if user["name"] != "Alice" {
		t.Errorf("Expected name='Alice', got %v", user["name"])
	}

	// Test FindAll
	users, err := orm.FindAll("users")
	if err != nil {
		t.Fatalf("Failed to find all: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Test QueryBuilder
	qb := orm.QueryBuilder("users")
	qb.Where("age", ">", 20)
	results, err := orm.Execute(qb)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test Count
	qb2 := orm.QueryBuilder("users")
	count, err := orm.Count(qb2)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count=2, got %d", count)
	}

	// Test Update
	affected, err := orm.Update("users",
		map[string]interface{}{"id": 1},
		map[string]interface{}{"age": 26},
	)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}
	if affected != 1 {
		t.Errorf("Expected 1 row affected, got %d", affected)
	}

	// Test Exists
	exists, err := orm.Exists("users", map[string]interface{}{"email": "alice@example.com"})
	if err != nil {
		t.Fatalf("Failed to check exists: %v", err)
	}
	if !exists {
		t.Error("Expected user to exist")
	}

	// Test Delete
	affected, err = orm.Delete("users", map[string]interface{}{"id": 1})
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}
	if affected != 1 {
		t.Errorf("Expected 1 row affected, got %d", affected)
	}

	// Verify deletion
	users, _ = orm.FindAll("users")
	if len(users) != 1 {
		t.Errorf("Expected 1 user after deletion, got %d", len(users))
	}
}

func TestTransaction(t *testing.T) {
	config := &ConnectionConfig{
		Type:     DBSQLite,
		Database: ":memory:",
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// SQLite :memory: databases are per-connection, pin to 1 connection
	db.SetMaxOpenConns(1)

	_, err = db.Exec("CREATE TABLE test (id INTEGER, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test successful transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO test VALUES (1, 'one')")
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO test VALUES (2, 'two')")
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify data
	rows, err := db.Query("SELECT COUNT(*) FROM test")
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	var count int
	if rows.Next() {
		rows.Scan(&count)
		if count != 2 {
			t.Errorf("Expected 2 rows, got %d", count)
		}
	} else {
		t.Error("No rows returned")
	}
	rows.Close() // Must close before next transaction with MaxOpenConns=1

	// Test rollback
	tx2, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin rollback transaction: %v", err)
	}
	_, err = tx2.Exec("INSERT INTO test VALUES (3, 'three')")
	if err != nil {
		t.Fatalf("Failed to insert in rollback transaction: %v", err)
	}
	err = tx2.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	rows2, err := db.Query("SELECT COUNT(*) FROM test")
	if err != nil {
		t.Fatalf("Failed to query after rollback: %v", err)
	}
	defer rows2.Close()
	var count2 int
	if rows2.Next() {
		rows2.Scan(&count2)
		if count2 != 2 {
			t.Errorf("Expected 2 rows after rollback, got %d", count2)
		}
	} else {
		t.Error("No rows returned after rollback")
	}
}

func TestMigrationManager(t *testing.T) {
	config := &ConnectionConfig{
		Type:     DBSQLite,
		Database: ":memory:",
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	mm := NewMigrationManager(db)

	// Add migrations (SQLite uses AUTOINCREMENT, not AUTO_INCREMENT)
	mm.AddMigration("001", "Create users table",
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
		"DROP TABLE users")

	mm.AddMigration("002", "Create posts table",
		"CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)",
		"DROP TABLE posts")

	// Test migration
	err = mm.Migrate()
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Verify migrations were applied
	applied, err := mm.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}
	if len(applied) != 2 {
		t.Errorf("Expected 2 applied migrations, got %d", len(applied))
	}

	// Test status
	status, err := mm.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	if len(status) != 2 {
		t.Errorf("Expected 2 status entries, got %d", len(status))
	}

	// Test rollback
	err = mm.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	applied, _ = mm.GetAppliedMigrations()
	if len(applied) != 1 {
		t.Errorf("Expected 1 applied migration after rollback, got %d", len(applied))
	}
}
