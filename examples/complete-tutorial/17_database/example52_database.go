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
//line example52_database.klx:24
rows := func() []map[string]interface{} { _v, _ := stdlib.DbQueryRows(db, "SELECT name, age FROM users ORDER BY age"); return _v }()
//line example52_database.klx:25
fmt.Println(("row count: " + fmt.Sprintf("%d", int64(len(rows)))))
//line example52_database.klx:26
fmt.Println(rows[0]["name"])
//line example52_database.klx:27
fmt.Println(rows[1]["name"])
//line example52_database.klx:28
fmt.Println(rows[0]["age"])
//line example52_database.klx:30
stdlib.DbClose(db)
//line example52_database.klx:31
fmt.Println("database demo OK")
}
