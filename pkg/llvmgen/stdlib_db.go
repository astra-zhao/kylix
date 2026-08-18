package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_db.go — LLVM IR implementation for the `db` stdlib module.
//
// TDatabase is an opaque ptr handle wrapping a sqlite3* connection. Links
// against libsqlite3 (Homebrew path on macOS, system path on Linux).
//
//   DbOpenSQLite(path)        -> ptr (TDatabase)    sqlite3_open
//   DbClose(db)               -> void               sqlite3_close
//   DbExec(db, sql, args...)  -> void               prepare+bind+step+finalize
//   DbQueryScalar(db, sql)    -> ptr (String)       prepare+step+column_text+strdup
//
// DbExec's variadic args are handled by inlining the prepare/bind/step
// sequence at each call site (each call generates a bespoke snippet with the
// right number of sqlite3_bind_text calls). This avoids needing a variadic
// ABI — the Kylix-level variadic is flattened into N bind calls at codegen
// time.

const dbHandleTypeName = "TDatabase"

// emitDbCall dispatches a `db.Func(args)` / bare `Func(args)` call.
func (g *Generator) emitDbCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "DbOpenSQLite":
		return g.emitDbOpenSQLiteCall(args)
	case "DbOpen":
		return g.emitDbOpenCall(args)
	case "DbClose":
		return g.emitDbCloseCall(args)
	case "DbExec":
		return g.emitDbExecCall(args)
	case "DbQueryScalar":
		return g.emitDbQueryScalarCall(args)
	case "DbQueryRows":
		return g.emitDbQueryRowsCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; db.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitDbBody dispatches the deferred body emitter (DbOpenSQLite/DbOpen/DbClose
// have module-level bodies; DbExec/DbQueryScalar are inlined at call sites).
func (g *Generator) emitDbBody(funcName string) {
	switch funcName {
	case "DbOpenSQLite":
		g.emitDbOpenSQLiteBody()
	case "DbOpen":
		g.emitDbOpenBody()
	case "DbClose":
		g.emitDbCloseBody()
	case "DbQueryRows":
		g.emitDbQueryRowsBody()
	}
}

// ---- DbOpenSQLite: ptr @__kylix_db_DbOpenSQLite(ptr %path) ----
func (g *Generator) emitDbOpenSQLiteCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("db.DbOpenSQLite expects 1 argument, got %d", len(args))
	}
	pathReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("db", "DbOpenSQLite", "DbOpenSQLite", 0)
	g.needLibsqlite = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_db_DbOpenSQLite(ptr %s)", r, pathReg))
	return r, dbHandleTypeName, nil
}

func (g *Generator) emitDbOpenSQLiteBody() {
	g.line("define ptr @__kylix_db_DbOpenSQLite(ptr %path) {")
	g.line("entry:")
	// db handle slot: sqlite3_open(path, &db) — db is a sqlite3** (ptr to ptr)
	dbSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", dbSlot))
	rc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @sqlite3_open(ptr %%path, ptr %s)", rc, dbSlot))
	// if rc != 0 (SQLITE_OK=0), return null
	bad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", bad, rc))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	dbVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", dbVal, dbSlot))
	g.line(fmt.Sprintf("  ret ptr %s", dbVal))
	g.line("}")
	g.line("")
}

// ---- DbOpen: ptr @__kylix_db_DbOpen(ptr %driver, ptr %dsn) ----
//
// v6.1.0: driver=="sqlite3" → sqlite3_open(dsn) (same as DbOpenSQLite); any
// other driver (mysql/postgres) returns null — those drivers are not linked.
func (g *Generator) emitDbOpenCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("db.DbOpen expects 2 arguments (driver, dsn), got %d", len(args))
	}
	driverReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	dsnReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("db", "DbOpen", "DbOpen", 0)
	g.needLibsqlite = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_db_DbOpen(ptr %s, ptr %s)", r, driverReg, dsnReg))
	return r, dbHandleTypeName, nil
}

func (g *Generator) emitDbOpenBody() {
	g.line("define ptr @__kylix_db_DbOpen(ptr %driver, ptr %dsn) {")
	g.line("entry:")
	driverStr := g.addString("sqlite3")
	isSQLite := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %%driver, ptr %s)", isSQLite, driverStr))
	same := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", same, isSQLite))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", same, okLbl, failLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	dbSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", dbSlot))
	rc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @sqlite3_open(ptr %%dsn, ptr %s)", rc, dbSlot))
	bad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", bad, rc))
	ok2 := g.label()
	fail2 := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bad, fail2, ok2))
	g.line(fmt.Sprintf("%s:", fail2))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", ok2))
	dbVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", dbVal, dbSlot))
	g.line(fmt.Sprintf("  ret ptr %s", dbVal))
	g.line("}")
	g.line("")
}

// ---- DbClose: void @__kylix_db_DbClose(ptr %db) ----
func (g *Generator) emitDbCloseCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("db.DbClose expects 1 argument, got %d", len(args))
	}
	dbReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("db", "DbClose", "DbClose", 0)
	g.needLibsqlite = true
	g.line(fmt.Sprintf("  call void @__kylix_db_DbClose(ptr %s)", dbReg))
	return "0", "void", nil
}

func (g *Generator) emitDbCloseBody() {
	g.line("define void @__kylix_db_DbClose(ptr %db) {")
	g.line("entry:")
	g.line("  call i32 @sqlite3_close(ptr %db)")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- DbExec: inlined at call site ----
//
//	For DbExec(db, sql, arg1, arg2, ...):
//	  sqlite3_prepare_v2(db, sql, -1, &stmt, 0)
//	  for each arg i (1-based):
//	    if arg is String: sqlite3_bind_text(stmt, i, val, -1, -1)  // -1 = SQLITE_TRANSIENT
//	    if arg is Integer: sqlite3_bind_int64(stmt, i, val)
//	  sqlite3_step(stmt)
//	  sqlite3_finalize(stmt)
func (g *Generator) emitDbExecCall(args []ast.Expression) (string, string, error) {
	if len(args) < 2 {
		return "", "", fmt.Errorf("db.DbExec expects at least 2 arguments (db, sql), got %d", len(args))
	}
	dbReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	sqlReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.needLibsqlite = true

	// prepare
	stmtSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", stmtSlot))
	g.line(fmt.Sprintf("  store ptr null, ptr %s", stmtSlot))
	g.line(fmt.Sprintf("  call i32 @sqlite3_prepare_v2(ptr %s, ptr %s, i32 -1, ptr %s, ptr null)", dbReg, sqlReg, stmtSlot))
	stmt := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", stmt, stmtSlot))

	// bind each arg (args[2:])
	for i, arg := range args[2:] {
		argReg, argType, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		idx := i + 1 // sqlite3 bind indices are 1-based
		if argType == "ptr" {
			// bind_text(stmt, idx, val, -1, -1)
			g.line(fmt.Sprintf("  call i32 @sqlite3_bind_text(ptr %s, i32 %d, ptr %s, i32 -1, i64 -1)", stmt, idx, argReg))
		} else {
			// bind_int64(stmt, idx, val)
			g.line(fmt.Sprintf("  call i32 @sqlite3_bind_int64(ptr %s, i32 %d, i64 %s)", stmt, idx, argReg))
		}
	}

	// step (INSERT/CREATE returns SQLITE_DONE=100; errors are ignored)
	g.line(fmt.Sprintf("  call i32 @sqlite3_step(ptr %s)", stmt))
	// finalize
	g.line(fmt.Sprintf("  call i32 @sqlite3_finalize(ptr %s)", stmt))
	// v6.1.0: return rows affected (matches the Go backend's int64 return).
	// sqlite3_changes(db) reports the count from the most recent DML statement.
	rows := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @sqlite3_changes(ptr %s)", rows, dbReg))
	rows64 := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", rows64, rows))
	return rows64, "i64", nil
}

// ---- DbQueryScalar: inlined at call site ----
//
//	For DbQueryScalar(db, sql):
//	  sqlite3_prepare_v2(db, sql, -1, &stmt, 0)
//	  sqlite3_step(stmt)  → expect SQLITE_ROW=100
//	  text = sqlite3_column_text(stmt, 0)  → const unsigned char*
//	  result = strdup(text)  → caller-owned String
//	  sqlite3_finalize(stmt)
//	  ret result  (null if no row)
func (g *Generator) emitDbQueryScalarCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("db.DbQueryScalar expects 2 arguments, got %d", len(args))
	}
	dbReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	sqlReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.needLibsqlite = true
	g.needHashtab = true // DbQueryScalar uses __kylix_htab_strdup

	// Result lives in an alloca (two paths write it: empty vs row).
	resultSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", resultSlot))

	// prepare
	stmtSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", stmtSlot))
	g.line(fmt.Sprintf("  store ptr null, ptr %s", stmtSlot))
	g.line(fmt.Sprintf("  call i32 @sqlite3_prepare_v2(ptr %s, ptr %s, i32 -1, ptr %s, ptr null)", dbReg, sqlReg, stmtSlot))
	stmt := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", stmt, stmtSlot))

	// step
	stepRc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @sqlite3_step(ptr %s)", stepRc, stmt))
	// SQLITE_ROW = 100
	isRow := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 100", isRow, stepRc))
	rowLbl := g.label()
	emptyLbl := g.label()
	mergeLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isRow, rowLbl, emptyLbl))

	// empty path: result = "" (empty string constant)
	g.line(fmt.Sprintf("%s:", emptyLbl))
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", emptyPtr, resultSlot))
	g.line(fmt.Sprintf("  call i32 @sqlite3_finalize(ptr %s)", stmt))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))

	// row path: result = strdup(column_text(stmt, 0))
	g.line(fmt.Sprintf("%s:", rowLbl))
	colText := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @sqlite3_column_text(ptr %s, i32 0)", colText, stmt))
	dup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", dup, colText))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", dup, resultSlot))
	g.line(fmt.Sprintf("  call i32 @sqlite3_finalize(ptr %s)", stmt))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))

	// merge: load result
	g.line(fmt.Sprintf("%s:", mergeLbl))
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", result, resultSlot))
	return result, "ptr", nil
}

// ---- DbQueryRows: { ptr, i64, i64 } @__kylix_db_DbQueryRows(ptr %db, ptr %sql)
//
// Runs a SELECT and returns every row as an `array of Variant` slice struct
// ({ptr,len,cap}). Each element is a map-Variant box (tag=map) whose payload is
// an htab (map[String]Variant): column name → value box (v6.4.0).
//
// Kylix consumer:
//   var rows := DbQueryRows(db, 'SELECT name FROM u');
//   var row  := rows[0];
//   WriteLn(row['name']);
// rows[i] reads a box (IsVariant array index); row['col'] lowers to
// @__kylix_variant_map_get (array.go emitVariantMapIndex). v6.4.0.
func (g *Generator) emitDbQueryRowsCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("db.DbQueryRows expects 2 arguments, got %d", len(args))
	}
	dbReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	sqlReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.needLibsqlite = true
	g.needHashtab = true
	g.needVariantRuntime = true
	g.needMemcpy = true // append copies the slice buffer
	g.enqueueStdlib("db", "DbQueryRows", "DbQueryRows", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call { ptr, i64, i64 } @__kylix_db_DbQueryRows(ptr %s, ptr %s)", r, dbReg, sqlReg))
	return r, "{ ptr, i64, i64 }", nil
}

func (g *Generator) emitDbQueryRowsBody() {
	g.line("define { ptr, i64, i64 } @__kylix_db_DbQueryRows(ptr %db, ptr %sql) {")
	g.line("entry:")
	// prepare
	stmtSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", stmtSlot))
	g.line(fmt.Sprintf("  store ptr null, ptr %s", stmtSlot))
	g.line(fmt.Sprintf("  call i32 @sqlite3_prepare_v2(ptr %%db, ptr %%sql, i32 -1, ptr %s, ptr null)", stmtSlot))
	stmt := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", stmt, stmtSlot))
	nCol := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @sqlite3_column_count(ptr %s)", nCol, stmt))
	// result slice {data,len,cap} accumulator + column-index slot
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca { ptr, i64, i64 }, align 8", resSlot))
	g.line(fmt.Sprintf("  store { ptr, i64, i64 } zeroinitializer, ptr %s", resSlot))
	colIdxSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", colIdxSlot))

	// ---- row loop: sqlite3_step until not SQLITE_ROW(100)
	rowLoop := g.label()
	rowBody := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", rowLoop))
	g.line(fmt.Sprintf("%s:", rowLoop))
	rc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @sqlite3_step(ptr %s)", rc, stmt))
	isRow := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 100", isRow, rc))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isRow, rowBody, done))

	// ---- row body: build one row's htab (map[String]Variant)
	g.line(fmt.Sprintf("%s:", rowBody))
	htab := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_new()", htab))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", colIdxSlot))
	colLoop := g.label()
	colBody := g.label()
	colDone := g.label()
	g.line(fmt.Sprintf("  br label %%%s", colLoop))
	g.line(fmt.Sprintf("%s:", colLoop))
	ci := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", ci, colIdxSlot))
	ci64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", ci64, ci))
	nCol64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", nCol64, nCol))
	colEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", colEnd, ci64, nCol64))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", colEnd, colBody, colDone))

	// ---- col body: box column value by sqlite type, put into the row htab
	g.line(fmt.Sprintf("%s:", colBody))
	cname := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @sqlite3_column_name(ptr %s, i32 %s)", cname, stmt, ci))
	ctype := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @sqlite3_column_type(ptr %s, i32 %s)", ctype, stmt, ci))
	valBoxSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", valBoxSlot))
	intLbl := g.label()
	floatLbl := g.label()
	strLbl := g.label()
	nilLbl := g.label()
	mergeLbl := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%%s [", ctype, nilLbl))
	g.line(fmt.Sprintf("    i32 1, label %%%s", intLbl)) // SQLITE_INTEGER
	g.line(fmt.Sprintf("    i32 2, label %%%s", floatLbl)) // SQLITE_FLOAT
	g.line(fmt.Sprintf("    i32 3, label %%%s", strLbl))  // SQLITE_TEXT
	g.line(fmt.Sprintf("    i32 4, label %%%s", strLbl))  // SQLITE_BLOB → as text
	g.line(fmt.Sprintf("  ]"))
	// int → box_int
	g.line(fmt.Sprintf("%s:", intLbl))
	iv := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @sqlite3_column_int64(ptr %s, i32 %s)", iv, stmt, ci))
	ib := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_int(i64 %s)", ib, iv))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", ib, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	// float → box_float
	g.line(fmt.Sprintf("%s:", floatLbl))
	fv := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @sqlite3_column_double(ptr %s, i32 %s)", fv, stmt, ci))
	fb := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_float(double %s)", fb, fv))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", fb, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	// text/blob → strdup + box_str
	g.line(fmt.Sprintf("%s:", strLbl))
	tv := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @sqlite3_column_text(ptr %s, i32 %s)", tv, stmt, ci))
	tdup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", tdup, tv))
	tb := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_str(ptr %s)", tb, tdup))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", tb, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	// nil / default → nilbox
	g.line(fmt.Sprintf("%s:", nilLbl))
	nb := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr @__kylix_variant_nilbox, i32 0, i32 0", nb))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nb, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	// merge → htab_put(htab, dup(colName), valueBox)
	g.line(fmt.Sprintf("%s:", mergeLbl))
	vbox := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", vbox, valBoxSlot))
	cnameDup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", cnameDup, cname))
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", htab, cnameDup, vbox))
	ciNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 %s, 1", ciNext, ci))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", ciNext, colIdxSlot))
	g.line(fmt.Sprintf("  br label %%%s", colLoop))

	// ---- col done: box the row htab as a map-Variant, append to result slice
	g.line(fmt.Sprintf("%s:", colDone))
	rowBox := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_map(ptr %s)", rowBox, htab))
	cur := g.tmp()
	g.line(fmt.Sprintf("  %s = load { ptr, i64, i64 }, ptr %s", cur, resSlot))
	oldData := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, i64, i64 } %s, 0", oldData, cur))
	oldLen := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, i64, i64 } %s, 1", oldLen, cur))
	newLen := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", newLen, oldLen))
	newBytes := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 8", newBytes, newLen))
	newData := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", newData, newBytes))
	oldBytes := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 8", oldBytes, oldLen))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", newData, oldData, oldBytes))
	elemSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", elemSlot, newData, oldLen))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", rowBox, elemSlot))
	s1 := g.tmp()
	g.line(fmt.Sprintf("  %s = insertvalue { ptr, i64, i64 } undef, ptr %s, 0", s1, newData))
	s2 := g.tmp()
	g.line(fmt.Sprintf("  %s = insertvalue { ptr, i64, i64 } %s, i64 %s, 1", s2, s1, newLen))
	s3 := g.tmp()
	g.line(fmt.Sprintf("  %s = insertvalue { ptr, i64, i64 } %s, i64 %s, 2", s3, s2, newLen))
	g.line(fmt.Sprintf("  store { ptr, i64, i64 } %s, ptr %s", s3, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", rowLoop))

	// ---- done
	g.line(fmt.Sprintf("%s:", done))
	g.line(fmt.Sprintf("  call i32 @sqlite3_finalize(ptr %s)", stmt))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load { ptr, i64, i64 }, ptr %s", out, resSlot))
	g.line(fmt.Sprintf("  ret { ptr, i64, i64 } %s", out))
	g.line("}")
	g.line("")
}
