package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// stdlib_entitymeta.go — the entitymeta module on the LLVM backend (v0.11.0).
//
// Mirrors stdlib/entitymeta.go exactly: codegen emits a registration sequence
// at the top of main() and the pure-Kylix CRUD engine reads it back through the
// four accessors. Both backends must observe identical strings.
//
// Storage is two fixed-size global arrays plus counters — the same shape the
// boot route table uses (@__kylix_boot_routes). Capacity is validated by the
// generator before emission, so the runtime checks below are belt-and-braces;
// they drop overflow rather than corrupting memory.
//
//	@__kylix_entity_tables  [64 x  {ptr, ptr}]          {table, meta}
//	@__kylix_entity_ntables i64
//	@__kylix_entity_fields  [512 x {ptr, ptr}]          {table, entry}
//	@__kylix_entity_nfields i64
//	@__kylix_entity_names   ptr                         (pre-joined CSV)

const (
	entityTableCapacity = 64
	entityFieldCapacity = 512
)

// entityMetaGlobals returns the module-level globals for the registry. Emitted
// only when the program declares [Entity] classes, so programs without them
// produce byte-identical IR to before this feature (bootstrap fixed point).
func entityMetaGlobals() []string {
	return []string{
		fmt.Sprintf("@__kylix_entity_tables = global [%d x { ptr, ptr }] zeroinitializer", entityTableCapacity),
		"@__kylix_entity_ntables = global i64 0",
		fmt.Sprintf("@__kylix_entity_fields = global [%d x { ptr, ptr }] zeroinitializer", entityFieldCapacity),
		"@__kylix_entity_nfields = global i64 0",
		"@__kylix_entity_names = global ptr null",
	}
}

// emitEntityMetaCall dispatches an entitymeta.* call at the call site and
// queues the callee body for module-end emission.
func (g *Generator) emitEntityMetaCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "RegisterEntity", "RegisterEntityField", "SetEntityNames":
		return g.emitEntityMetaRegisterCall(funcName, args)
	case "EntityMetaOf", "EntityFieldAt":
		return g.emitEntityMetaLookupCall(funcName, args, "ptr")
	case "EntityFieldCount":
		return g.emitEntityMetaCountCall(args)
	case "EntityNames":
		if len(args) != 0 {
			return "", "", fmt.Errorf("entitymeta.EntityNames expects no arguments, got %d", len(args))
		}
		g.enqueueStdlib("entitymeta", "EntityNames", "EntityNames", 0)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_entitymeta_EntityNames()", r))
		return r, "ptr", nil
	default:
		return "", "", fmt.Errorf("entitymeta.%s is not implemented on the LLVM backend", funcName)
	}
}

// emitEntityMetaBody dispatches the deferred body emitters.
func (g *Generator) emitEntityMetaBody(funcName string) {
	switch funcName {
	case "RegisterEntity":
		g.emitEntityMetaRegisterTableBody()
	case "RegisterEntityField":
		g.emitEntityMetaRegisterFieldBody()
	case "SetEntityNames":
		g.emitEntityMetaSetNamesBody()
	case "EntityMetaOf":
		g.emitEntityMetaOfBody()
	case "EntityFieldAt":
		g.emitEntityFieldAtBody()
	case "EntityFieldCount":
		g.emitEntityFieldCountBody()
	case "EntityNames":
		g.line("define ptr @__kylix_entitymeta_EntityNames() {")
		g.line("entry:")
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_entity_names", r))
		g.line(fmt.Sprintf("  ret ptr %s", r))
		g.line("}")
		g.line("")
	}
}

// ---- registration ----------------------------------------------------------

// emitEntityMetaRegisterCall emits `call void @__kylix_entitymeta_<name>(...)`
// with the arguments marshalled as string pointers.
func (g *Generator) emitEntityMetaRegisterCall(funcName string, args []ast.Expression) (string, string, error) {
	want := 2
	if funcName == "SetEntityNames" {
		want = 1
	}
	if len(args) != want {
		return "", "", fmt.Errorf("entitymeta.%s expects %d arguments, got %d", funcName, want, len(args))
	}
	regs := make([]string, 0, want)
	for _, a := range args {
		r, _, err := g.emitExpr(a)
		if err != nil {
			return "", "", err
		}
		regs = append(regs, r)
	}
	g.enqueueStdlib("entitymeta", funcName, funcName, want)
	g.line(fmt.Sprintf("  call void @__kylix_entitymeta_%s(%s)", funcName, strings.Join(regs, ", ")))
	return "", "void", nil
}

// emitEntityMetaRegisterTableBody — void @__kylix_entitymeta_RegisterEntity(
// ptr %table, ptr %meta).
func (g *Generator) emitEntityMetaRegisterTableBody() {
	g.line("define void @__kylix_entitymeta_RegisterEntity(ptr %table, ptr %meta) {")
	g.line("entry:")
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_entity_ntables", n))
	full := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %d", full, n, entityTableCapacity))
	storeLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", full, doneLbl, storeLbl))
	g.line(fmt.Sprintf("%s:", storeLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x { ptr, ptr }], ptr @__kylix_entity_tables, i64 0, i64 %s", slot, entityTableCapacity, n))
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 0", f0, slot))
	g.line(fmt.Sprintf("  store ptr %%table, ptr %s", f0))
	f1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 1", f1, slot))
	g.line(fmt.Sprintf("  store ptr %%meta, ptr %s", f1))
	next := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", next, n))
	g.line(fmt.Sprintf("  store i64 %s, ptr @__kylix_entity_ntables", next))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitEntityMetaRegisterFieldBody — void @__kylix_entitymeta_RegisterEntityField(
// ptr %table, ptr %entry). Fields keep global declaration order; the accessors
// filter by table, so per-table order is preserved.
func (g *Generator) emitEntityMetaRegisterFieldBody() {
	// NOTE: the second parameter must not be named %entry — it would collide
	// with the entry block label (the same trap v0.9.0 hit in BootAppendToSlot).
	g.line("define void @__kylix_entitymeta_RegisterEntityField(ptr %table, ptr %spec) {")
	g.line("entry:")
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_entity_nfields", n))
	full := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %d", full, n, entityFieldCapacity))
	storeLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", full, doneLbl, storeLbl))
	g.line(fmt.Sprintf("%s:", storeLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x { ptr, ptr }], ptr @__kylix_entity_fields, i64 0, i64 %s", slot, entityFieldCapacity, n))
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 0", f0, slot))
	g.line(fmt.Sprintf("  store ptr %%table, ptr %s", f0))
	f1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 1", f1, slot))
	g.line(fmt.Sprintf("  store ptr %%spec, ptr %s", f1))
	next := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", next, n))
	g.line(fmt.Sprintf("  store i64 %s, ptr @__kylix_entity_nfields", next))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitEntityMetaSetNamesBody — void @__kylix_entitymeta_SetEntityNames(ptr %csv).
func (g *Generator) emitEntityMetaSetNamesBody() {
	g.line("define void @__kylix_entitymeta_SetEntityNames(ptr %csv) {")
	g.line("entry:")
	g.line("  store ptr %csv, ptr @__kylix_entity_names")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- accessors -------------------------------------------------------------

// emitEntityMetaLookupCall emits a call to a table-keyed lookup accessor and
// queues its body. Used by EntityMetaOf (table → meta) and EntityFieldAt
// (table + index → entry).
func (g *Generator) emitEntityMetaLookupCall(funcName string, args []ast.Expression, retType string) (string, string, error) {
	if funcName == "EntityMetaOf" {
		if len(args) != 1 {
			return "", "", fmt.Errorf("entitymeta.EntityMetaOf expects 1 argument, got %d", len(args))
		}
		tbl, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		g.enqueueStdlib("entitymeta", funcName, funcName, 1)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_entitymeta_EntityMetaOf(ptr %s)", r, tbl))
		return r, retType, nil
	}
	if len(args) != 2 {
		return "", "", fmt.Errorf("entitymeta.EntityFieldAt expects 2 arguments, got %d", len(args))
	}
	tbl, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	idx, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("entitymeta", funcName, funcName, 2)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_entitymeta_EntityFieldAt(ptr %s, i64 %s)", r, tbl, idx))
	return r, retType, nil
}

// emitEntityMetaCountCall emits `call i64 @__kylix_entitymeta_EntityFieldCount`.
func (g *Generator) emitEntityMetaCountCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("entitymeta.EntityFieldCount expects 1 argument, got %d", len(args))
	}
	tbl, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("entitymeta", "EntityFieldCount", "EntityFieldCount", 1)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_entitymeta_EntityFieldCount(ptr %s)", r, tbl))
	return r, "i64", nil
}

// emitEntityMetaOfBody — ptr @__kylix_entitymeta_EntityMetaOf(ptr %table).
// Linear scan of the table array; returns the empty string when not found so
// the Kylix side sees "" exactly like the Go implementation.
func (g *Generator) emitEntityMetaOfBody() {
	g.line("define ptr @__kylix_entitymeta_EntityMetaOf(ptr %table) {")
	g.line("entry:")
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	nSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", nSlot))
	n0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_entity_ntables", n0))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", n0, nSlot))
	loopLbl := g.label()
	bodyLbl := g.label()
	hitLbl := g.label()
	missLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	i := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", i, iSlot))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", n, nSlot))
	cond := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", cond, i, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", cond, bodyLbl, missLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x { ptr, ptr }], ptr @__kylix_entity_tables, i64 0, i64 %s", slot, entityTableCapacity, i))
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 0", f0, slot))
	tblPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tblPtr, f0))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %%table)", cmp, tblPtr))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	nextLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eq, hitLbl, nextLbl))
	g.line(fmt.Sprintf("%s:", nextLbl))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", hitLbl))
	f1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 1", f1, slot))
	metaPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", metaPtr, f1))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", missLbl))
	empty := g.addString("")
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = phi ptr [ %s, %%%s ], [ %s, %%%s ]", res, metaPtr, hitLbl, empty, missLbl))
	g.line(fmt.Sprintf("  ret ptr %s", res))
	g.line("}")
	g.line("")
}

// emitEntityFieldAtBody — ptr @__kylix_entitymeta_EntityFieldAt(ptr %table,
// i64 %index): the index-th field registered for that table, in declaration
// order. Returns "" when the table is unknown or the index is out of range.
func (g *Generator) emitEntityFieldAtBody() {
	g.line("define ptr @__kylix_entitymeta_EntityFieldAt(ptr %table, i64 %index) {")
	g.line("entry:")
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	seenSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", seenSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", seenSlot))
	nSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", nSlot))
	n0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_entity_nfields", n0))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", n0, nSlot))
	loopLbl := g.label()
	bodyLbl := g.label()
	matchLbl := g.label()
	hitLbl := g.label()
	missLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	i := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", i, iSlot))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", n, nSlot))
	cond := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", cond, i, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", cond, bodyLbl, missLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x { ptr, ptr }], ptr @__kylix_entity_fields, i64 0, i64 %s", slot, entityFieldCapacity, i))
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 0", f0, slot))
	tblPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tblPtr, f0))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %%table)", cmp, tblPtr))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	nextLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eq, matchLbl, nextLbl))
	g.line(fmt.Sprintf("%s:", nextLbl))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", matchLbl))
	seen := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", seen, seenSlot))
	isWanted := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %%index", isWanted, seen))
	seenNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", seenNext, seen))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", seenNext, seenSlot))
	afterLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isWanted, hitLbl, afterLbl))
	g.line(fmt.Sprintf("%s:", afterLbl))
	iNext2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext2, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext2, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", hitLbl))
	f1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 1", f1, slot))
	entryPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", entryPtr, f1))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", missLbl))
	empty := g.addString("")
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = phi ptr [ %s, %%%s ], [ %s, %%%s ]", res, entryPtr, hitLbl, empty, missLbl))
	g.line(fmt.Sprintf("  ret ptr %s", res))
	g.line("}")
	g.line("")
}

// emitEntityFieldCountBody — i64 @__kylix_entitymeta_EntityFieldCount(ptr %table).
func (g *Generator) emitEntityFieldCountBody() {
	g.line("define i64 @__kylix_entitymeta_EntityFieldCount(ptr %table) {")
	g.line("entry:")
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	countSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", countSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", countSlot))
	nSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", nSlot))
	n0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_entity_nfields", n0))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", n0, nSlot))
	loopLbl := g.label()
	bodyLbl := g.label()
	nextLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	i := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", i, iSlot))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", n, nSlot))
	cond := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", cond, i, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", cond, bodyLbl, doneLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x { ptr, ptr }], ptr @__kylix_entity_fields, i64 0, i64 %s", slot, entityFieldCapacity, i))
	f0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr }, ptr %s, i32 0, i32 0", f0, slot))
	tblPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", tblPtr, f0))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %%table)", cmp, tblPtr))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	g.line(fmt.Sprintf("  br label %%%s", nextLbl))
	g.line(fmt.Sprintf("%s:", nextLbl))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", c, countSlot))
	cNext := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 1, i64 0", cNext, eq))
	cSum := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", cSum, c, cNext))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", cSum, countSlot))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", res, countSlot))
	g.line(fmt.Sprintf("  ret i64 %s", res))
	g.line("}")
	g.line("")
}
