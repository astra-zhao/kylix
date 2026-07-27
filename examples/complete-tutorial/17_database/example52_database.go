package main

import (
	"kylix/stdlib"
	"fmt"
)

func main() {
//line example52_database.klx:6
db := func() *stdlib.Database { _v, _ := stdlib.DbOpenSQLite(":memory:"); return _v }()
//line example52_database.klx:9
func() int64 { _v, _ := stdlib.DbExec(db, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)"); return _v }()
//line example52_database.klx:12
func() int64 { _v, _ := stdlib.DbExec(db, "INSERT INTO users (name, age) VALUES (?, ?)", "alice", 30); return _v }()
//line example52_database.klx:13
func() int64 { _v, _ := stdlib.DbExec(db, "INSERT INTO users (name, age) VALUES (?, ?)", "bob", 25); return _v }()
//line example52_database.klx:16
count := func() string { _v, _ := stdlib.DbQueryScalar(db, "SELECT COUNT(*) FROM users"); return _v }()
//line example52_database.klx:17
fmt.Println(("user count: " + count))
//line example52_database.klx:19
name := func() string { _v, _ := stdlib.DbQueryScalar(db, "SELECT name FROM users WHERE age = 30"); return _v }()
//line example52_database.klx:20
fmt.Println(("age 30 user: " + name))
//line example52_database.klx:22
stdlib.DbClose(db)
//line example52_database.klx:23
fmt.Println("database demo OK")
}
