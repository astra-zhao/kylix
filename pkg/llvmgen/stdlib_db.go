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
	case "DbClose":
		return g.emitDbCloseCall(args)
	case "DbExec":
		return g.emitDbExecCall(args)
	case "DbQueryScalar":
		return g.emitDbQueryScalarCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; db.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitDbBody dispatches the deferred body emitter (DbOpenSQLite/DbClose have
// module-level bodies; DbExec/DbQueryScalar are inlined at call sites).
func (g *Generator) emitDbBody(funcName string) {
	switch funcName {
	case "DbOpenSQLite":
		g.emitDbOpenSQLiteBody()
	case "DbClose":
		g.emitDbCloseBody()
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

	// step (ignore return — INSERT/CREATE returns SQLITE_DONE=100)
	g.line(fmt.Sprintf("  call i32 @sqlite3_step(ptr %s)", stmt))
	// finalize
	g.line(fmt.Sprintf("  call i32 @sqlite3_finalize(ptr %s)", stmt))
	return "0", "void", nil
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
