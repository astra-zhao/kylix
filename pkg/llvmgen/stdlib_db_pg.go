package llvmgen

import (
	"fmt"
	"reflect"

	"kylix/ast"
)

// stdlib_db_pg.go — the postgres backend for the db module (v0.12.0 P5d).
//
// Reached through a separate DbOpenPg entry point rather than a runtime driver
// check inside DbOpen. The sqlite paths inline sqlite3_* calls at the call site
// and the driver string is a runtime value, so a runtime dispatch would force
// every db program to carry both backends — which would also drag `-lpq` into
// the tutorials' link line, and their CI jobs do not install libpq. With a
// separate entry point a program that never calls DbOpenPg emits no libpq
// symbol at all, and compile.go only adds -lpq when it sees the pg prefix.
//
// Handle representation matches sqlite: TDatabase stays an opaque ptr, here a
// PGconn*. The dialect is process-wide (one connection per program), so the
// shared helpers branch on @__kylix_db_is_pg, which DbOpenPg sets and the
// sqlite open paths clear.
//
// Statements arrive with Kylix-facing `?` placeholders and are rewritten to
// libpq's `$n` here — the same rewrite the Go backend performs in its db layer,
// at the same choke point, so the 70-odd call sites stay dialect-agnostic.
//
// Value parity with sqlite: PQftype OIDs are mapped onto the same boxed types
// the sqlite path produces (20/21/23 int8/int2/int4 → int, 700/701 → float,
// 16 → int 0/1, everything else → text). Booleans are stored as INTEGER on both
// dialects for exactly this reason, and NULL goes to nilbox, which renders as
// "" — the same as the Go side.

const dbPgPrefix = "__kylix_db_pg_"

// programUsesPg reports whether the program opens a postgres connection. It is
// a pre-scan because the db call sites are emitted while walking declarations,
// long before main (where DbOpenPg usually appears) is reached — and the
// dispatch those call sites need depends on the answer.
func programUsesPg(prog *ast.Program) bool {
	if prog == nil {
		return false
	}
	return nodeUsesIdent(reflect.ValueOf(prog), "DbOpenPg")
}

// nodeUsesIdent recursively walks an AST node (via reflection) looking for an
// ast.Identifier with the given name — the same technique statementsUseArgs
// uses for the Args builtin.
func nodeUsesIdent(v reflect.Value, name string) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return false
		}
		return nodeUsesIdent(v.Elem(), name)
	case reflect.Struct:
		if id, ok := v.Interface().(ast.Identifier); ok {
			return id.Value == name
		}
		for i := 0; i < v.NumField(); i++ {
			if nodeUsesIdent(v.Field(i), name) {
				return true
			}
		}
		return false
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if nodeUsesIdent(v.Index(i), name) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// emitDbPgGlobals declares the two module globals the postgres backend needs:
// the dialect flag the shared helpers branch on, and the most recent connection
// (DbLastError needs a connection to ask libpq for its error text).
func (g *Generator) emitDbPgGlobals() []string {
	return []string{
		"@__kylix_db_is_pg = global i32 0",
		"@__kylix_db_lastconn = global ptr null",
		"@__kylix_db_pg_lasterr = global ptr null",
	}
}

// emitDbPgDeclares emits the libpq declarations. Called once per module when
// the pg backend is used.
func (g *Generator) emitDbPgDeclares() {
	g.line("; ===== libpq (used by stdlib db, postgres backend v0.12.0) =====")
	g.line("declare ptr @PQconnectdb(ptr noundef)")
	g.line("declare i32 @PQstatus(ptr noundef)")
	g.line("declare ptr @PQerrorMessage(ptr noundef)")
	g.line("declare ptr @PQexecParams(ptr noundef, ptr noundef, i32 noundef, ptr noundef, ptr noundef, ptr noundef, ptr noundef, i32 noundef)")
	g.line("declare i32 @PQresultStatus(ptr noundef)")
	g.line("declare i32 @PQntuples(ptr noundef)")
	g.line("declare i32 @PQnfields(ptr noundef)")
	g.line("declare ptr @PQfname(ptr noundef, i32 noundef)")
	g.line("declare ptr @PQgetvalue(ptr noundef, i32 noundef, i32 noundef)")
	g.line("declare i32 @PQgetisnull(ptr noundef, i32 noundef, i32 noundef)")
	g.line("declare i32 @PQftype(ptr noundef, i32 noundef)")
	g.line("declare ptr @PQcmdTuples(ptr noundef)")
	g.line("declare void @PQclear(ptr noundef)")
	g.line("declare void @PQfinish(ptr noundef)")
}

// ---- call sites ------------------------------------------------------------

// emitDbOpenPgCall emits `DbOpenPg(dsn)`.
func (g *Generator) emitDbOpenPgCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("db.DbOpenPg expects 1 argument (dsn), got %d", len(args))
	}
	dsnReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("db", "DbOpenPg", "DbOpenPg", 1)
	g.needLibpq = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_db_pg_open(ptr %s)", r, dsnReg))
	return r, dbHandleTypeName, nil
}

// emitDbPgPackedCall is the shared shape for the three statement entry points:
// evaluate the arguments, pack them into an argv/argtypes pair (the same
// packing the sqlite bound form uses), and call the module body.
func (g *Generator) emitDbPgPackedCall(funcName string, bodyKey string, retType string, args []ast.Expression) (string, string, error) {
	if len(args) < 2 {
		return "", "", fmt.Errorf("db.%s expects at least 2 arguments (db, sql), got %d", funcName, len(args))
	}
	connReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	sqlReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	bound := args[2:]
	n := len(bound)
	argvSlot := g.tmp()
	typesSlot := g.tmp()
	if n > 0 {
		g.line(fmt.Sprintf("  %s = alloca [%d x i64], align 8", argvSlot, n))
		g.line(fmt.Sprintf("  %s = alloca [%d x i32], align 4", typesSlot, n))
	} else {
		g.line(fmt.Sprintf("  %s = alloca i64, align 8", argvSlot))
		g.line(fmt.Sprintf("  %s = alloca i32, align 4", typesSlot))
	}
	for i, arg := range bound {
		argReg, argType, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		if n > 0 {
			elem := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x i64], ptr %s, i64 0, i64 %d", elem, n, argvSlot, i))
			telem := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x i32], ptr %s, i64 0, i64 %d", telem, n, typesSlot, i))
			if argType == "ptr" {
				asInt := g.tmp()
				g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", asInt, argReg))
				g.line(fmt.Sprintf("  store i64 %s, ptr %s", asInt, elem))
				g.line(fmt.Sprintf("  store i32 0, ptr %s", telem))
			} else {
				g.line(fmt.Sprintf("  store i64 %s, ptr %s", argReg, elem))
				g.line(fmt.Sprintf("  store i32 1, ptr %s", telem))
			}
		}
	}
	g.enqueueStdlib("db", funcName, bodyKey, 0)
	g.needLibpq = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call %s @%s%s(ptr %s, ptr %s, i64 %d, ptr %s, ptr %s)",
		r, retType, dbPgPrefix, bodyKey, connReg, sqlReg, n, argvSlot, typesSlot))
	return r, retType, nil
}

// ---- bodies ----------------------------------------------------------------

// emitDbOpenPgBody — ptr @__kylix_db_pg_open(ptr %dsn).
//
// PQconnectdb returns a connection object even when the connection fails, so
// the status has to be checked; a failed connect returns null, which the Kylix
// side can test (the sqlite path's silent null on an unsupported driver is a
// defect this avoids repeating).
func (g *Generator) emitDbOpenPgBody() {
	g.line(fmt.Sprintf("define ptr @%sopen(ptr %%dsn) {", dbPgPrefix))
	g.line("entry:")
	conn := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @PQconnectdb(ptr %%dsn)", conn))
	st := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQstatus(ptr %s)", st, conn))
	bad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", bad, st))
	failLbl := g.label()
	okLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line(fmt.Sprintf("  call void @PQfinish(ptr %s)", conn))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	g.line("  store i32 1, ptr @__kylix_db_is_pg")
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_db_lastconn", conn))
	g.line(fmt.Sprintf("  ret ptr %s", conn))
	g.line("}")
	g.line("")
}

// emitPgPrepare emits the shared prologue of the three statement bodies:
// rewrite the placeholders, materialise the parameter array, and run
// PQexecParams. It returns the PGresult register.
//
// The parameter array and the text buffers for integer arguments are allocas of
// this frame — sized by argc, which is only known at run time — so they live
// exactly as long as the query. They are deliberately NOT in a helper function:
// returning a pointer into a returned frame would be undefined behaviour.
func (g *Generator) emitPgPrepare(conn, sql, argc, argv, argtypes string) string {
	sql2 := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @%srewrite(ptr %s, i64 %s)", sql2, dbPgPrefix, sql, argc))
	argc32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", argc32, argc))
	// paramValues: argc pointers. paramLengths/paramFormats stay null (all
	// values are NUL-terminated text, the default format).
	vals := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, i64 %s", vals, argc))
	// Scratch space for the integer arguments converted to text: 24 bytes each
	// is enough for a sign plus 20 digits.
	scratchBytes := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 24", scratchBytes, argc))
	scratch := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i8, i64 %s", scratch, scratchBytes))

	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	loopLbl := g.label()
	bodyLbl := g.label()
	intLbl := g.label()
	nextLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	i := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", i, iSlot))
	more := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", more, i, argc))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", more, bodyLbl, doneLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	telem := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i32, ptr %s, i64 %s", telem, argtypes, i))
	tv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tv, telem))
	isInt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 1", isInt, tv))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isInt, intLbl, nextLbl))
	// text argument: the i64 payload is the String pointer again
	g.line(fmt.Sprintf("%s:", intLbl))
	velem := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 %s", velem, argv, i))
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", raw, velem))
	// integers travel as text: snprintf into this frame's scratch
	bufOff := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 24", bufOff, i))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", buf, scratch, bufOff))
	fmtStr := g.addString("%lld")
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 24, ptr %s, i64 %s)", buf, fmtStr, raw))
	pslot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", pslot, vals, i))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", buf, pslot))
	afterLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", afterLbl))
	// text path joins here
	g.line(fmt.Sprintf("%s:", nextLbl))
	velem2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 %s", velem2, argv, i))
	raw2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", raw2, velem2))
	asPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", asPtr, raw2))
	pslot2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", pslot2, vals, i))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", asPtr, pslot2))
	g.line(fmt.Sprintf("  br label %%%s", afterLbl))
	g.line(fmt.Sprintf("%s:", afterLbl))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @PQexecParams(ptr %s, ptr %s, i32 %s, ptr null, ptr %s, ptr null, ptr null, i32 0)",
		res, conn, sql2, argc32, vals))
	return res
}

// emitPgStatusGate emits the result-status check shared by the three statement
// bodies: a statement that fails records libpq's message for DbLastError and
// jumps to failLbl; a successful one clears the slot and falls through. Without
// this a postgres syntax or type error would look exactly like an empty result
// set — the generated code discards the error half of the Go API, so this slot
// is the only channel a Kylix program has.
func (g *Generator) emitPgStatusGate(res, conn string, want int, failLbl string) {
	st := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQresultStatus(ptr %s)", st, res))
	ok := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", ok, st, want))
	okLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ok, okLbl, failLbl))
	// The caller emits failLbl (recording the message, clearing the result and
	// returning a zero value) and then jumps back to okLbl.
	g.line(fmt.Sprintf("%s:", okLbl))
	empty := g.addString("")
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_db_pg_lasterr", empty))
}

// emitPgRecordError emits the failure block a status gate branches to: record
// libpq's message, release the result, and continue to the caller's cleanup.
func (g *Generator) emitPgRecordError(res, conn string) {
	msg := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @PQerrorMessage(ptr %s)", msg, conn))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_db_pg_lasterr", msg))
	g.line(fmt.Sprintf("  call void @PQclear(ptr %s)", res))
}

// emitDbPgExecBody — i64 @__kylix_db_pg_exec(conn, sql, argc, argv, argtypes).
//
// Returns the affected-row count from PQcmdTuples ("" for DDL → 0), which is
// what the Go backend's RowsAffected reports as well.
func (g *Generator) emitDbPgExecBody() {
	g.line(fmt.Sprintf("define i64 @%sexec(ptr %%conn, ptr %%sql, i64 %%argc, ptr %%argv, ptr %%argtypes) {", dbPgPrefix))
	g.line("entry:")
	res := g.emitPgPrepare("%conn", "%sql", "%argc", "%argv", "%argtypes")
	cmdLbl := g.label()
	zeroLbl := g.label()
	doneLbl := g.label()
	// PGRES_COMMAND_OK = 1. A failure lands in zeroLbl after recording the
	// message, so it returns 0 rows and DbLastError explains why.
	g.emitPgStatusGate(res, "%conn", 1, zeroLbl)
	g.line(fmt.Sprintf("  br label %%%s", cmdLbl))
	g.line(fmt.Sprintf("%s:", cmdLbl))
	tuples := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @PQcmdTuples(ptr %s)", tuples, res))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", n, tuples))
	g.line(fmt.Sprintf("  call void @PQclear(ptr %s)", res))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", zeroLbl))
	g.emitPgRecordError(res, "%conn")
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ 0, %%%s ]", out, n, cmdLbl, zeroLbl))
	g.line(fmt.Sprintf("  ret i64 %s", out))
	g.line("}")
	g.line("")
}

// emitDbPgScalarBody — ptr @__kylix_db_pg_scalar(...).
//
// Mirrors the sqlite scalar path: "" when there is no row or the value is NULL.
func (g *Generator) emitDbPgScalarBody() {
	g.line(fmt.Sprintf("define ptr @%sscalar(ptr %%conn, ptr %%sql, i64 %%argc, ptr %%argv, ptr %%argtypes) {", dbPgPrefix))
	g.line("entry:")
	res := g.emitPgPrepare("%conn", "%sql", "%argc", "%argv", "%argtypes")
	rowLbl := g.label()
	emptyLbl := g.label()
	doneLbl := g.label()
	// PGRES_TUPLES_OK = 2; a failed statement returns "" and leaves the message
	// in the error slot for DbLastError.
	g.emitPgStatusGate(res, "%conn", 2, emptyLbl)
	nt := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQntuples(ptr %s)", nt, res))
	hasRow := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i32 %s, 0", hasRow, nt))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasRow, rowLbl, emptyLbl))
	g.line(fmt.Sprintf("%s:", emptyLbl))
	emptyStr := g.addString("")
	g.emitPgRecordError(res, "%conn")
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", rowLbl))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQgetisnull(ptr %s, i32 0, i32 0)", isNull, res))
	notNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", notNull, isNull))
	valLbl := g.label()
	nullLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", notNull, valLbl, nullLbl))
	g.line(fmt.Sprintf("%s:", nullLbl))
	g.line(fmt.Sprintf("  call void @PQclear(ptr %s)", res))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", valLbl))
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @PQgetvalue(ptr %s, i32 0, i32 0)", raw, res))
	dup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", dup, raw))
	g.line(fmt.Sprintf("  call void @PQclear(ptr %s)", res))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = phi ptr [ %s, %%%s ], [ %s, %%%s ], [ %s, %%%s ]", out, emptyStr, emptyLbl, emptyStr, nullLbl, dup, valLbl))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// emitDbPgRowsBody — { ptr, i64, i64 } @__kylix_db_pg_rows(...).
//
// Builds the same slice-of-map-Variant the sqlite path builds, so
// `rows[i]['col']` behaves identically. Values are boxed by PQftype OID:
// int2/int4/int8 → int, float4/float8 → float, bool → int 0/1, anything else →
// text (which is what the Go backend's []byte→string normalisation produces).
func (g *Generator) emitDbPgRowsBody() {
	g.line(fmt.Sprintf("define { ptr, i64, i64 } @%srows(ptr %%conn, ptr %%sql, i64 %%argc, ptr %%argv, ptr %%argtypes) {", dbPgPrefix))
	g.line("entry:")
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca { ptr, i64, i64 }, align 8", resSlot))
	g.line(fmt.Sprintf("  store { ptr, i64, i64 } zeroinitializer, ptr %s", resSlot))
	// Loop counters must be allocas: they are live across the row/column loops.
	rowIdxSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", rowIdxSlot))
	colIdxSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", colIdxSlot))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", rowIdxSlot))
	res := g.emitPgPrepare("%conn", "%sql", "%argc", "%argv", "%argtypes")
	rowLoop := g.label()
	rowBody := g.label()
	rowDone := g.label()
	// A failed statement yields an empty slice (rowDone), with the message in
	// the error slot for DbLastError.
	g.emitPgStatusGate(res, "%conn", 2, rowDone)
	nrows := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQntuples(ptr %s)", nrows, res))
	ncols := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQnfields(ptr %s)", ncols, res))
	g.line(fmt.Sprintf("  br label %%%s", rowLoop))
	g.line(fmt.Sprintf("%s:", rowLoop))
	ri := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", ri, rowIdxSlot))
	rowMore := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, %s", rowMore, ri, nrows))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", rowMore, rowBody, rowDone))

	// ---- row body: a fresh htab for the row
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
	colMore := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, %s", colMore, ci, ncols))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", colMore, colBody, colDone))

	// ---- column body: name + value
	g.line(fmt.Sprintf("%s:", colBody))
	cname := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @PQfname(ptr %s, i32 %s)", cname, res, ci))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQgetisnull(ptr %s, i32 %s, i32 %s)", isNull, res, ri, ci))
	notNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", notNull, isNull))
	valBoxSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", valBoxSlot))
	valLbl := g.label()
	nilLbl := g.label()
	mergeLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", notNull, valLbl, nilLbl))

	// value present: box by OID
	g.line(fmt.Sprintf("%s:", valLbl))
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @PQgetvalue(ptr %s, i32 %s, i32 %s)", raw, res, ri, ci))
	oid := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @PQftype(ptr %s, i32 %s)", oid, res, ci))
	intLbl := g.label()
	floatLbl := g.label()
	boolLbl := g.label()
	textLbl := g.label()
	boxed := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%%s [", oid, textLbl))
	g.line(fmt.Sprintf("    i32 20, label %%%s", intLbl))    // int8
	g.line(fmt.Sprintf("    i32 21, label %%%s", intLbl))    // int2
	g.line(fmt.Sprintf("    i32 23, label %%%s", intLbl))    // int4
	g.line(fmt.Sprintf("    i32 700, label %%%s", floatLbl)) // float4
	g.line(fmt.Sprintf("    i32 701, label %%%s", floatLbl)) // float8
	g.line(fmt.Sprintf("    i32 16, label %%%s", boolLbl))   // bool
	g.line("  ]")
	g.line(fmt.Sprintf("%s:", intLbl))
	iv := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", iv, raw))
	ib := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_int(i64 %s)", ib, iv))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", ib, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", boxed))
	g.line(fmt.Sprintf("%s:", floatLbl))
	fv := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @strtod(ptr %s, ptr null)", fv, raw))
	fb := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_float(double %s)", fb, fv))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", fb, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", boxed))
	// bool → int 0/1: the sqlite driver reports a column declared INTEGER as
	// int64, and both dialects store booleans as INTEGER, so this keeps the two
	// backends' rendering identical.
	g.line(fmt.Sprintf("%s:", boolLbl))
	first := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", first, raw))
	isT := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 116", isT, first)) // 't'
	bv := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 1, i64 0", bv, isT))
	bb := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_int(i64 %s)", bb, bv))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", bb, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", boxed))
	g.line(fmt.Sprintf("%s:", textLbl))
	tdup := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_strdup(ptr %s)", tdup, raw))
	tb := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_str(ptr %s)", tb, tdup))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", tb, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", boxed))
	g.line(fmt.Sprintf("%s:", boxed))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))

	// NULL → nilbox (renders as "" on both backends)
	g.line(fmt.Sprintf("%s:", nilLbl))
	nb := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr @__kylix_variant_nilbox, i32 0, i32 0", nb))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nb, valBoxSlot))
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))

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

	// ---- row done: box the htab as a map-Variant and append it to the slice
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
	g.line(fmt.Sprintf("  %s = %s", newData, g.mallocCall(newBytes)))
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
	riNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i32 %s, 1", riNext, ri))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", riNext, rowIdxSlot))
	g.line(fmt.Sprintf("  br label %%%s", rowLoop))

	g.line(fmt.Sprintf("%s:", rowDone))
	g.emitPgRecordError(res, "%conn")
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load { ptr, i64, i64 }, ptr %s", out, resSlot))
	g.line(fmt.Sprintf("  ret { ptr, i64, i64 } %s", out))
	g.line("}")
	g.line("")
}

// emitDbPgRewriteBody — ptr @__kylix_db_pg_rewrite(ptr %sql, i64 %nargs).
//
// Rewrites the Kylix-facing `?` placeholders into libpq's `$1, $2, ...`. A `?`
// inside a single-quoted literal is left alone, and ” inside a literal is
// treated as an escaped quote — the same rules as the Go backend's rewrite, so
// the two dialects agree on which question marks are parameters.
//
// The buffer is sized from the statement length plus four bytes per argument
// (a placeholder grows from "?" to "$nn" at most).
func (g *Generator) emitDbPgRewriteBody() {
	if g.pgRewriteEmitted {
		return
	}
	g.pgRewriteEmitted = true
	g.line(fmt.Sprintf("define ptr @%srewrite(ptr %%sql, i64 %%nargs) {", dbPgPrefix))
	g.line("entry:")
	length := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%sql)", length))
	extra := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %%nargs, 4", extra))
	cap := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", cap, length, extra))
	cap2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 8", cap2, cap))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, cap2))
	si := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", si))
	di := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", di))
	pi := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", pi))
	lit := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", lit))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", si))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", di))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", pi))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", lit))

	loop := g.label()
	body := g.label()
	quoteLbl := g.label()
	quote2 := g.label()
	quote1 := g.label()
	maybeQ := g.label()
	dowrite := g.label()
	docopy := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", loop))
	i := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", i, si))
	src := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%sql, i64 %s", src, i))
	ch := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", ch, src))
	end := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 0", end, ch))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", end, done, body))
	g.line(fmt.Sprintf("%s:", body))
	isQuote := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 39", isQuote, ch))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isQuote, quoteLbl, maybeQ))
	// '' inside a literal is an escaped quote: copy both characters and stay
	// inside the literal.
	g.line(fmt.Sprintf("%s:", quoteLbl))
	ni := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", ni, i))
	next := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%sql, i64 %s", next, ni))
	nch := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", nch, next))
	isQuote2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 39", isQuote2, nch))
	inLit := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", inLit, lit))
	wasIn := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", wasIn, inLit))
	both := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", both, isQuote2, wasIn))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", both, quote2, quote1))
	g.line(fmt.Sprintf("%s:", quote2))
	d0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", d0, di))
	dst0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst0, buf, d0))
	g.line(fmt.Sprintf("  store i8 39, ptr %s", dst0))
	d0b := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", d0b, d0))
	dst0b := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst0b, buf, d0b))
	g.line(fmt.Sprintf("  store i8 39, ptr %s", dst0b))
	d0c := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", d0c, d0))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", d0c, di))
	si2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", si2, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", si2, si))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", quote1))
	d1 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", d1, di))
	dst1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst1, buf, d1))
	g.line(fmt.Sprintf("  store i8 39, ptr %s", dst1))
	d1b := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", d1b, d1))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", d1b, di))
	flip := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i32 %s, 1", flip, inLit))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", flip, lit))
	si3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", si3, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", si3, si))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	// a '?' outside a literal becomes $n
	g.line(fmt.Sprintf("%s:", maybeQ))
	isParam := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 63", isParam, ch))
	inLit2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", inLit2, lit))
	notIn := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", notIn, inLit2))
	rewrite := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", rewrite, isParam, notIn))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", rewrite, dowrite, docopy))
	g.line(fmt.Sprintf("%s:", dowrite))
	pn := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pn, pi))
	pn1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", pn1, pn))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", pn1, pi))
	dw := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", dw, di))
	dstW := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dstW, buf, dw))
	g.line(fmt.Sprintf("  store i8 36, ptr %s", dstW)) // '$'
	dw1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", dw1, dw))
	digits := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @%sitoa(ptr %s, i64 %s, i64 %s)", digits, dbPgPrefix, buf, dw1, pn1))
	dwEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", dwEnd, dw1, digits))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", dwEnd, di))
	si4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", si4, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", si4, si))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", docopy))
	d2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", d2, di))
	dst2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst2, buf, d2))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", ch, dst2))
	d2b := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", d2b, d2))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", d2b, di))
	si5 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", si5, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", si5, si))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", done))
	dEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", dEnd, di))
	dstEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dstEnd, buf, dEnd))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", dstEnd))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")

	// itoa: writes the decimal form of %n at %buf and returns the digit count
	// (1..3 digits cover any realistic parameter count).
	g.line(fmt.Sprintf("define i64 @%sitoa(ptr %%buf, i64 %%off, i64 %%n) {", dbPgPrefix))
	g.line("entry:")
	base := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%buf, i64 %%off", base))
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = sdiv i64 %%n, 100", h))
	r1 := g.tmp()
	g.line(fmt.Sprintf("  %s = srem i64 %%n, 100", r1))
	t := g.tmp()
	g.line(fmt.Sprintf("  %s = sdiv i64 %s, 10", t, r1))
	u := g.tmp()
	g.line(fmt.Sprintf("  %s = srem i64 %s, 10", u, r1))
	// three-digit case
	three := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %%n, 100", three))
	threeLbl := g.label()
	afterThree := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", three, threeLbl, afterThree))
	g.line(fmt.Sprintf("%s:", threeLbl))
	hch := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 48", hch, h))
	hbyte := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", hbyte, hch))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", hbyte, base))
	g.line(fmt.Sprintf("  br label %%%s", afterThree))
	g.line(fmt.Sprintf("%s:", afterThree))
	// two- and three-digit cases share the tens digit
	base2 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 1, i64 0", base2, three))
	off2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", off2, base, base2))
	two := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %%n, 10", two))
	twoLbl := g.label()
	afterTwo := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", two, twoLbl, afterTwo))
	g.line(fmt.Sprintf("%s:", twoLbl))
	tch := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 48", tch, t))
	tbyte := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", tbyte, tch))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", tbyte, off2))
	g.line(fmt.Sprintf("  br label %%%s", afterTwo))
	g.line(fmt.Sprintf("%s:", afterTwo))
	off3 := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 1, i64 0", off3, two))
	off4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", off4, base2, off3))
	last := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", last, base, off4))
	uch := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 48", uch, u))
	ubyte := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", ubyte, uch))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", ubyte, last))
	// digit count = 1 + two + three
	cnt0 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i1 %s to i64", cnt0, two))
	cnt1 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i1 %s to i64", cnt1, three))
	cnt := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", cnt, cnt0, cnt1))
	cntf := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", cntf, cnt))
	g.line(fmt.Sprintf("  ret i64 %s", cntf))
	g.line("}")
	g.line("")
}
