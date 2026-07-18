// variant.go — Variant runtime for the LLVM backend.
//
// A Variant value is a boxed pointer to a 16-byte { i32 tag, i64 payload }:
//
//	tag    0=nil, 1=int, 2=float, 3=str, 4=bool
//	payload holds the value's bit pattern:
//	  int  → the i64 value
//	  float→ bitcast double→i64
//	  str  → ptrtoint ptr→i64  (the C-string pointer)
//	  bool → zext i1→i64       (0 or 1)
//
// Storage slots (`var v: Variant` alloca, `array of Variant` element slots) are
// `ptr` (they hold a box pointer). The codegen-time value type "variant" (a
// synthetic llvmType string returned by emitExpr for Variant box values) lets
// emitInfix/emitWriteLn route Variant operands without a parallel type table.
//
// Runtime helpers are emitted once per module (guarded by variantRuntimeEmitted,
// triggered by needVariantRuntime), mirroring the hashtab runtime pattern.
package llvmgen

import (
	"fmt"
	"kylix/ast"
	"strings"
)

// Variant tag constants.
const (
	varTagNil  = 0
	varTagInt  = 1
	varTagFloat = 2
	varTagStr  = 3
	varTagBool = 4
)

// variantT is the synthetic llvmType string for a Variant box value (a ptr).
const variantT = "variant"

// isVariantOperand reports whether an llvmType denotes a Variant box value.
func isVariantOperand(t string) bool { return t == variantT }

// isComparisonOp reports whether the operator is a relational comparison.
func isComparisonOp(op string) bool {
	switch op {
	case "=", "<>", "<", "<=", ">", ">=":
		return true
	}
	return false
}

// variantTypeName is the LLVM struct type name used in box mallocs/GEPs.
const variantTypeName = "{ i32, i64 }"

// emitVariantBox boxes a scalar value register into a Variant box ptr, based on
// the value's LLVM type. If the value is already a Variant box ("variant"), it
// is returned unchanged (no re-box). Sets needVariantRuntime.
func (g *Generator) emitVariantBox(v, llvmT string) string {
	g.needVariantRuntime = true
	if llvmT == variantT {
		return v // already a box ptr
	}
	var helper string
	switch llvmT {
	case "i64":
		helper = "@__kylix_variant_box_int"
	case "double":
		helper = "@__kylix_variant_box_float"
	case "ptr": // string
		helper = "@__kylix_variant_box_str"
	case "i1":
		helper = "@__kylix_variant_box_bool"
	default:
		// Unknown — coerce to int and box.
		v2, _ := g.coerceValue(v, llvmT, "i64")
		v = v2
		helper = "@__kylix_variant_box_int"
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr %s(%s %s)", r, helper, boxArgType(llvmT), v))
	return r
}

// boxArgType maps the value llvmT to the LLVM type expected by the box helper.
func boxArgType(llvmT string) string {
	switch llvmT {
	case "i64", "ptr", "double", "i1":
		return llvmT
	default:
		return "i64"
	}
}

// emitVariantAsStr unboxes a Variant box to a C-string ptr (dispatches on tag).
func (g *Generator) emitVariantAsStr(v string) string {
	g.needVariantRuntime = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %s)", r, v))
	return r
}

// emitVariantAsInt unboxes a Variant box to an i64 (dispatches on tag).
func (g *Generator) emitVariantAsInt(v string) string {
	g.needVariantRuntime = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_variant_as_int(ptr %s)", r, v))
	return r
}

// emitVariantAsBool unboxes a Variant box to an i1 (dispatches on tag).
func (g *Generator) emitVariantAsBool(v string) string {
	g.needVariantRuntime = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_variant_as_bool(ptr %s)", r, v))
	return r
}

// emitVariantAsDouble unboxes a Variant box to a double (dispatches on tag).
func (g *Generator) emitVariantAsDouble(v string) string {
	g.needVariantRuntime = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @__kylix_variant_as_double(ptr %s)", r, v))
	return r
}

// isArithOp reports whether the operator is a supported Variant arithmetic op.
func isArithOp(op string) bool {
	switch op {
	case "+", "-", "*", "/":
		return true
	}
	return false
}

// emitVariantArith boxes any non-Variant operand, then calls the runtime
// arithmetic helper for the operator. Returns (boxPtr, "variant").
func (g *Generator) emitVariantArith(op, lv, lt, rv, rt string) (string, string, error) {
	g.needVariantRuntime = true
	if lt != variantT {
		lv = g.emitVariantBox(lv, lt)
	}
	if rt != variantT {
		rv = g.emitVariantBox(rv, rt)
	}
	helper := map[string]string{
		"+": "@__kylix_variant_add",
		"-": "@__kylix_variant_sub",
		"*": "@__kylix_variant_mul",
		"/": "@__kylix_variant_div",
	}[op]
	if helper == "" {
		// div/mod on Variants are unsupported — return the nil-box address
		// (a Variant holding nil) so IR stays legal.
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr @__kylix_variant_nilbox, i32 0, i32 0", r))
		return r, variantT, nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr %s(ptr %s, ptr %s)", r, helper, lv, rv))
	return r, variantT, nil
}

// emitVariantCompare boxes any non-Variant operand, then calls the runtime
// comparator and maps the i32 result (-1/0/1) to the requested comparison.
// Returns (reg, "i1").
func (g *Generator) emitVariantCompare(op, lv, lt, rv, rt string) (string, string, error) {
	g.needVariantRuntime = true
	if lt != variantT {
		lv = g.emitVariantBox(lv, lt)
	}
	if rt != variantT {
		rv = g.emitVariantBox(rv, rt)
	}
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @__kylix_variant_compare(ptr %s, ptr %s)", cmp, lv, rv))
	r := g.tmp()
	switch op {
	case "=":
		g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", r, cmp))
	case "<>":
		g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", r, cmp))
	case "<":
		g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 0", r, cmp))
	case "<=":
		g.line(fmt.Sprintf("  %s = icmp sle i32 %s, 0", r, cmp))
	case ">":
		g.line(fmt.Sprintf("  %s = icmp sgt i32 %s, 0", r, cmp))
	case ">=":
		g.line(fmt.Sprintf("  %s = icmp sge i32 %s, 0", r, cmp))
	default:
		// Unsupported op on Variants (e.g. arithmetic) — emit a safe zero.
		g.line(fmt.Sprintf("  %s = add i1 0, 0 ; variant op %q unsupported", r, op))
		return r, "i1", nil
	}
	return r, "i1", nil
}

// emitVariantPrint prints a Variant box, optionally with a trailing newline.
// Returns ("0", "void").
func (g *Generator) emitVariantPrint(v string, newline bool) (string, string, error) {
	g.needVariantRuntime = true
	helper := "@__kylix_variant_print"
	if newline {
		helper = "@__kylix_variant_println"
	}
	g.line(fmt.Sprintf("  call void %s(ptr %s)", helper, v))
	return "0", "void", nil
}

// emitVariantRuntimeBodies emits all Variant runtime helpers, once per module.
// Idempotent via variantRuntimeEmitted.
func (g *Generator) emitVariantRuntimeBodies() {
	if g.variantRuntimeEmitted {
		return
	}
	g.variantRuntimeEmitted = true
	g.emitVariantNilboxGlobal()
	g.emitVariantBoxInt()
	g.emitVariantBoxFloat()
	g.emitVariantBoxStr()
	g.emitVariantBoxBool()
	g.emitVariantAsDoubleBody()
	g.emitVariantAsStrBody()
	g.emitVariantAsIntBody()
	g.emitVariantAsBoolBody()
	g.emitVariantCompareBody()
	g.emitVariantArithBody("+")
	g.emitVariantArithBody("-")
	g.emitVariantArithBody("*")
	g.emitVariantArithBody("/")
	g.emitVariantPrintBody(false) // print
	g.emitVariantPrintBody(true)  // println
}

// emitVariantNilboxGlobal emits the global nil-box (tag=0, payload=0) used as
// the "missing key" sentinel by htab_get_variant. A Variant holding nil reads
// as the typed zero via as_* (tag 0 → nil branch).
func (g *Generator) emitVariantNilboxGlobal() {
	g.line("@__kylix_variant_nilbox = internal constant { i32, i64 } { i32 0, i64 0 }")
	g.line("")
}

// boxAddr returns a register holding the address of field `idx` of the box at
// %box (the box struct is { i32, i64 }).
func (g *Generator) boxAddr(box string, idx int32) string {
	loc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds %s, ptr %s, i32 0, i32 %d", loc, variantTypeName, box, idx))
	return loc
}

func (g *Generator) emitVariantBoxInt() {
	g.line("define ptr @__kylix_variant_box_int(i64 %v) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", box))
	tagLoc := g.boxAddr(box, 0)
	g.line(fmt.Sprintf("  store i32 %d, ptr %s", varTagInt, tagLoc))
	payloadLoc := g.boxAddr(box, 1)
	g.line(fmt.Sprintf("  store i64 %%v, ptr %s", payloadLoc))
	g.line(fmt.Sprintf("  ret ptr %s", box))
	g.line("}")
	g.line("")
}

func (g *Generator) emitVariantBoxFloat() {
	g.line("define ptr @__kylix_variant_box_float(double %v) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", box))
	tagLoc := g.boxAddr(box, 0)
	g.line(fmt.Sprintf("  store i32 %d, ptr %s", varTagFloat, tagLoc))
	payloadLoc := g.boxAddr(box, 1)
	bits := g.tmp()
	g.line(fmt.Sprintf("  %s = bitcast double %%v to i64", bits))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", bits, payloadLoc))
	g.line(fmt.Sprintf("  ret ptr %s", box))
	g.line("}")
	g.line("")
}

func (g *Generator) emitVariantBoxStr() {
	g.line("define ptr @__kylix_variant_box_str(ptr %v) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", box))
	tagLoc := g.boxAddr(box, 0)
	g.line(fmt.Sprintf("  store i32 %d, ptr %s", varTagStr, tagLoc))
	payloadLoc := g.boxAddr(box, 1)
	bits := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %%v to i64", bits))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", bits, payloadLoc))
	g.line(fmt.Sprintf("  ret ptr %s", box))
	g.line("}")
	g.line("")
}

func (g *Generator) emitVariantBoxBool() {
	g.line("define ptr @__kylix_variant_box_bool(i1 %v) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", box))
	tagLoc := g.boxAddr(box, 0)
	g.line(fmt.Sprintf("  store i32 %d, ptr %s", varTagBool, tagLoc))
	payloadLoc := g.boxAddr(box, 1)
	ext := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i1 %%v to i64", ext))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", ext, payloadLoc))
	g.line(fmt.Sprintf("  ret ptr %s", box))
	g.line("}")
	g.line("")
}

// as_double loads the payload as a double, coercing by tag.
// tag 1 (int): sitofp payload→double. tag 2 (float): bitcast payload→double.
// tag 4 (bool): zext...payload is already i64, sitofp→double. tag 3 (str): strtod.
// tag 0 (nil): 0.0.
func (g *Generator) emitVariantAsDoubleBody() {
	g.line("define double @__kylix_variant_as_double(ptr %v) {")
	g.line("entry:")
	tagLoc := g.boxAddr("%v", 0)
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagLoc))
	payloadLoc := g.boxAddr("%v", 1)
	payload := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payload, payloadLoc))
	// Result alloca (must dominate the switch branches).
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca double, align 8", res))
	g.line(fmt.Sprintf("  store double 0.0, ptr %s", res)) // default nil
	intLbl := g.label()
	floatLbl := g.label()
	strLbl := g.label()
	boolLbl := g.label()
	endLbl := g.label()
	defLbl := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%%s [", tag, defLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagInt, intLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagFloat, floatLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagStr, strLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagBool, boolLbl))
	g.line(fmt.Sprintf("  ]"))
	// int: sitofp payload→double
	g.line(fmt.Sprintf("%s:", intLbl))
	dv := g.tmp()
	g.line(fmt.Sprintf("  %s = sitofp i64 %s to double", dv, payload))
	g.line(fmt.Sprintf("  store double %s, ptr %s", dv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// float: bitcast payload→double
	g.line(fmt.Sprintf("%s:", floatLbl))
	fv := g.tmp()
	g.line(fmt.Sprintf("  %s = bitcast i64 %s to double", fv, payload))
	g.line(fmt.Sprintf("  store double %s, ptr %s", fv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// bool: sitofp payload→double (payload is zext i1 already i64 0/1)
	g.line(fmt.Sprintf("%s:", boolLbl))
	bv := g.tmp()
	g.line(fmt.Sprintf("  %s = sitofp i64 %s to double", bv, payload))
	g.line(fmt.Sprintf("  store double %s, ptr %s", bv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// str: strtod(payload→ptr, null)
	g.line(fmt.Sprintf("%s:", strLbl))
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", sp, payload))
	sv := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @strtod(ptr %s, ptr null)", sv, sp))
	g.line(fmt.Sprintf("  store double %s, ptr %s", sv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", defLbl))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", endLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load double, ptr %s", out, res))
	g.line(fmt.Sprintf("  ret double %s", out))
	g.line("}")
	g.line("")
}

// as_str unboxes to a freshly-malloc'd C string by tag.
// int→snprintf %lld; float→snprintf %.15g; str→inttoptr payload; bool→"true"/"false"; nil→"".
func (g *Generator) emitVariantAsStrBody() {
	emptyStr := g.addString("")
	g.line("define ptr @__kylix_variant_as_str(ptr %v) {")
	g.line("entry:")
	tagLoc := g.boxAddr("%v", 0)
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagLoc))
	payloadLoc := g.boxAddr("%v", 1)
	payload := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payload, payloadLoc))
	// Result slot.
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", res))
	emptyPtr := g.ptrTo(emptyStr, 1)
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", emptyPtr, res)) // default nil→""
	intLbl := g.label()
	floatLbl := g.label()
	strLbl := g.label()
	boolLbl := g.label()
	endLbl := g.label()
	defLbl := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%%s [", tag, defLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagInt, intLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagFloat, floatLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagStr, strLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagBool, boolLbl))
	g.line(fmt.Sprintf("  ]"))
	// int → snprintf("%lld")
	intLbl_fn(intLbl, g, payload, res, endLbl)
	// float → snprintf("%.15g")
	floatLbl_fn(floatLbl, g, payload, res, endLbl)
	// str → inttoptr payload
	g.line(fmt.Sprintf("%s:", strLbl))
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", sp, payload))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", sp, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// bool → select payload!=0 ? "true" : "false"
	g.line(fmt.Sprintf("%s:", boolLbl))
	trueStr := g.addString("true")
	falseStr := g.addString("false")
	truePtr := g.ptrTo(trueStr, len("true")+1)
	falsePtr := g.ptrTo(falseStr, len("false")+1)
	isz := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", isz, payload))
	sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", sel, isz, truePtr, falsePtr))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", sel, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", defLbl))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", endLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", out, res))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// intLbl_fn emits the int branch of as_str: snprintf("%lld", payload) into a
// 24-byte buffer, store its address in resSlot, branch to endLbl.
func intLbl_fn(lbl string, g *Generator, payload, resSlot, endLbl string) {
	g.line(fmt.Sprintf("%s:", lbl))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 32)", buf))
	fmtReg := g.addString("%lld")
	fmtPtr := g.ptrTo(fmtReg, len("%lld")+1)
	_ = g.tmp() // snprintf ret (ignored)
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %s, i64 32, ptr noundef %s, i64 %s)",
		buf, fmtPtr, payload))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", buf, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
}

// floatLbl_fn emits the float branch of as_str: bitcast payload→double,
// snprintf("%.15g", double) into a 32-byte buffer.
func floatLbl_fn(lbl string, g *Generator, payload, resSlot, endLbl string) {
	g.line(fmt.Sprintf("%s:", lbl))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 32)", buf))
	dv := g.tmp()
	g.line(fmt.Sprintf("  %s = bitcast i64 %s to double", dv, payload))
	fmtReg := g.addString("%.15g")
	fmtPtr := g.ptrTo(fmtReg, len("%.15g")+1)
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %s, i64 32, ptr noundef %s, double %s)",
		buf, fmtPtr, dv))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", buf, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
}

// emitVariantCompareBody emits @__kylix_variant_compare(ptr a, ptr b) → i32.
// Returns -1/0/1. Categories:
//   both numeric (tag int/float/bool): coerce both via as_double, compare.
//   both str (tag 3): strcmp the reconstructed string pointers.
//   both nil (tag 0): equal → 0.
//   mismatched categories: order by tag (a.tag - b.tag, clamped) so '=' on
//   mismatched types is non-0 (not equal).
func (g *Generator) emitVariantCompareBody() {
	g.line("define i32 @__kylix_variant_compare(ptr %a, ptr %b) {")
	g.line("entry:")
	tagALoc := g.boxAddr("%a", 0)
	tagA := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tagA, tagALoc))
	tagBLoc := g.boxAddr("%b", 0)
	tagB := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tagB, tagBLoc))
	// Category flags. numeric = tag in {1,2,4}; str = tag==3; nil = tag==0.
	isAStr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", isAStr, tagA, varTagStr))
	isBStr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", isBStr, tagB, varTagStr))
	bothStr := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", bothStr, isAStr, isBStr))
	isANil := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", isANil, tagA, varTagNil))
	isBNil := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", isBNil, tagB, varTagNil))
	bothNil := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", bothNil, isANil, isBNil))
	// bothNumeric = !isAStr && !isBStr && !isANil && !isBNil
	notAStr := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i1 %s, 1", notAStr, isAStr))
	notBStr := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i1 %s, 1", notBStr, isBStr))
	notANil := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i1 %s, 1", notANil, isANil))
	notBNil := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i1 %s, 1", notBNil, isBNil))
	bn1 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", bn1, notAStr, notBStr))
	bn2 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", bn2, bn1, notANil))
	bothNumeric := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", bothNumeric, bn2, notBNil))
	// Result slot (dominates all branches).
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", res))
	numLbl := g.label()
	strLbl := g.label()
	nilLbl := g.label()
	tagLbl := g.label()
	endLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_str", bothNumeric, numLbl))
	g.line("chk_str:")
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_nil", bothStr, strLbl))
	g.line("chk_nil:")
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bothNil, nilLbl, tagLbl))
	// numeric: da=as_double(a), db=as_double(b); sign(da-db).
	g.line(fmt.Sprintf("%s:", numLbl))
	da := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @__kylix_variant_as_double(ptr %%a)", da))
	db := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @__kylix_variant_as_double(ptr %%b)", db))
	nlt := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp olt double %s, %s", nlt, da, db))
	ngt := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp ogt double %s, %s", ngt, da, db))
	neg := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 -1, i32 0", neg, nlt))
	nsgn := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 1, i32 %s", nsgn, ngt, neg))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", nsgn, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// str: sa=as_str(a), sb=as_str(b); strcmp; sign.
	g.line(fmt.Sprintf("%s:", strLbl))
	sa := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %%a)", sa))
	sb := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %%b)", sb))
	scmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", scmp, sa, sb))
	slt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 0", slt, scmp))
	sgt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i32 %s, 0", sgt, scmp))
	seg := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 -1, i32 0", seg, slt))
	ssgn := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 1, i32 %s", ssgn, sgt, seg))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", ssgn, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// both nil → 0.
	g.line(fmt.Sprintf("%s:", nilLbl))
	g.line(fmt.Sprintf("  store i32 0, ptr %s", res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// mismatched → sign(tagA - tagB).
	g.line(fmt.Sprintf("%s:", tagLbl))
	diff := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i32 %s, %s", diff, tagA, tagB))
	dlt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 0", dlt, diff))
	dgt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i32 %s, 0", dgt, diff))
	deg := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 -1, i32 0", deg, dlt))
	dsgn := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i32 1, i32 %s", dsgn, dgt, deg))
	g.line(fmt.Sprintf("  store i32 %s, ptr %s", dsgn, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", endLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", out, res))
	g.line(fmt.Sprintf("  ret i32 %s", out))
	g.line("}")
	g.line("")
}

// emitVariantPrintBody emits @__kylix_variant_print(ln)(ptr v) → void.
func (g *Generator) emitVariantPrintBody(newline bool) {
	name := "@__kylix_variant_print"
	suffix := ""
	if newline {
		name = "@__kylix_variant_println"
	}
	g.line(fmt.Sprintf("define void %s(ptr %%v) {", name))
	g.line("entry:")
	tagLoc := g.boxAddr("%v", 0)
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagLoc))
	payloadLoc := g.boxAddr("%v", 1)
	payload := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payload, payloadLoc))
	intLbl := g.label()
	floatLbl := g.label()
	strLbl := g.label()
	boolLbl := g.label()
	endLbl := g.label()
	defLbl := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%%s [", tag, defLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagInt, intLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagFloat, floatLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagStr, strLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagBool, boolLbl))
	g.line(fmt.Sprintf("  ]"))
	suffix = printSuffix(newline)
	// int → printf("%lld\n", payload)
	g.line(fmt.Sprintf("%s:", intLbl))
	ifmt := g.addString("%lld" + suffix)
	ifmtPtr := g.ptrTo(ifmt, len("%lld"+suffix)+1)
	g.line(fmt.Sprintf("  call i32 (ptr, ...) @printf(ptr noundef %s, i64 %s)", ifmtPtr, payload))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// float → bitcast payload→double; printf("%.15g\n")
	g.line(fmt.Sprintf("%s:", floatLbl))
	fv := g.tmp()
	g.line(fmt.Sprintf("  %s = bitcast i64 %s to double", fv, payload))
	ffmt := g.addString("%.15g" + suffix)
	ffmtPtr := g.ptrTo(ffmt, len("%.15g"+suffix)+1)
	g.line(fmt.Sprintf("  call i32 (ptr, ...) @printf(ptr noundef %s, double %s)", ffmtPtr, fv))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// str → inttoptr payload→ptr; puts (println) or printf %s (print)
	g.line(fmt.Sprintf("%s:", strLbl))
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", sp, payload))
	if newline {
		g.line(fmt.Sprintf("  call i32 @puts(ptr noundef %s)", sp))
	} else {
		sfmt := g.addString("%s")
		sfmtPtr := g.ptrTo(sfmt, len("%s")+1)
		g.line(fmt.Sprintf("  call i32 (ptr, ...) @printf(ptr noundef %s, ptr %s)", sfmtPtr, sp))
	}
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// bool → "true"/"false" (puts adds newline for println; printf %s for print)
	g.line(fmt.Sprintf("%s:", boolLbl))
	trueStr := g.addString("true")
	falseStr := g.addString("false")
	truePtr := g.ptrTo(trueStr, len("true")+1)
	falsePtr := g.ptrTo(falseStr, len("false")+1)
	isz := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", isz, payload))
	sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", sel, isz, truePtr, falsePtr))
	if newline {
		g.line(fmt.Sprintf("  call i32 @puts(ptr noundef %s)", sel))
	} else {
		sfmt := g.addString("%s")
		sfmtPtr := g.ptrTo(sfmt, len("%s")+1)
		g.line(fmt.Sprintf("  call i32 (ptr, ...) @printf(ptr noundef %s, ptr %s)", sfmtPtr, sel))
	}
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", defLbl))
	nilStr := g.addString("nil")
	nilPtr := g.ptrTo(nilStr, len("nil")+1)
	if newline {
		g.line(fmt.Sprintf("  call i32 @puts(ptr noundef %s)", nilPtr))
	} else {
		sfmt := g.addString("%s")
		sfmtPtr := g.ptrTo(sfmt, len("%s")+1)
		g.line(fmt.Sprintf("  call i32 (ptr, ...) @printf(ptr noundef %s, ptr %s)", sfmtPtr, nilPtr))
	}
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", endLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// printSuffix returns the trailing-newline string for format constants.
// Returns a real newline byte (0x0A) — matching emitWriteLn's addString
// usage, so the IR string-constant emitter handles it the same way.
func printSuffix(newline bool) string {
	if newline {
		return "\n"
	}
	return ""
}

// as_int unboxes to i64 by tag: int→payload; float→fptosi; str→atoll; bool→payload; nil→0.
func (g *Generator) emitVariantAsIntBody() {
	g.line("define i64 @__kylix_variant_as_int(ptr %v) {")
	g.line("entry:")
	tagLoc := g.boxAddr("%v", 0)
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagLoc))
	payloadLoc := g.boxAddr("%v", 1)
	payload := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payload, payloadLoc))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", res))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", res)) // default nil
	intLbl := g.label()
	floatLbl := g.label()
	strLbl := g.label()
	boolLbl := g.label()
	endLbl := g.label()
	defLbl := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%%s [", tag, defLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagInt, intLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagFloat, floatLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagStr, strLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagBool, boolLbl))
	g.line(fmt.Sprintf("  ]"))
	g.line(fmt.Sprintf("%s:", intLbl))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", payload, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", floatLbl))
	// float→fptosi: bitcast payload→double then fptosi→i64
	fdbits := g.tmp()
	g.line(fmt.Sprintf("  %s = bitcast i64 %s to double", fdbits, payload))
	fi := g.tmp()
	g.line(fmt.Sprintf("  %s = fptosi double %s to i64", fi, fdbits))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", fi, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", boolLbl))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", payload, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", strLbl))
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", sp, payload))
	sv := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", sv, sp))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", sv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", defLbl))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", endLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", out, res))
	g.line(fmt.Sprintf("  ret i64 %s", out))
	g.line("}")
	g.line("")
}

// as_bool unboxes to i1 by tag: bool→payload!=0; int→!=0; float→!=0.0; str→strcmp("true")==0; nil→0.
func (g *Generator) emitVariantAsBoolBody() {
	trueStr := g.addString("true")
	g.line("define i1 @__kylix_variant_as_bool(ptr %v) {")
	g.line("entry:")
	tagLoc := g.boxAddr("%v", 0)
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagLoc))
	payloadLoc := g.boxAddr("%v", 1)
	payload := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payload, payloadLoc))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i1, align 1", res))
	g.line(fmt.Sprintf("  store i1 0, ptr %s", res)) // default nil
	intLbl := g.label()
	floatLbl := g.label()
	strLbl := g.label()
	boolLbl := g.label()
	endLbl := g.label()
	defLbl := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%%s [", tag, defLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagInt, intLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagFloat, floatLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagStr, strLbl))
	g.line(fmt.Sprintf("    i32 %d, label %%%s", varTagBool, boolLbl))
	g.line(fmt.Sprintf("  ]"))
	g.line(fmt.Sprintf("%s:", intLbl))
	iv := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", iv, payload))
	g.line(fmt.Sprintf("  store i1 %s, ptr %s", iv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", boolLbl))
	bv := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", bv, payload))
	g.line(fmt.Sprintf("  store i1 %s, ptr %s", bv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", floatLbl))
	fdbits := g.tmp()
	g.line(fmt.Sprintf("  %s = bitcast i64 %s to double", fdbits, payload))
	fv := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp one double %s, 0.0", fv, fdbits))
	g.line(fmt.Sprintf("  store i1 %s, ptr %s", fv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", strLbl))
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", sp, payload))
	truePtr := g.ptrTo(trueStr, len("true")+1)
	scmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", scmp, sp, truePtr))
	sv := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", sv, scmp))
	g.line(fmt.Sprintf("  store i1 %s, ptr %s", sv, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", defLbl))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", endLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load i1, ptr %s", out, res))
	g.line(fmt.Sprintf("  ret i1 %s", out))
	g.line("}")
	g.line("")
}

// emitVariantArithBody emits the runtime arithmetic helper for op (+,-,*,/).
// All return a fresh Variant box. Dispatch:
//   + : either str → str concat; both int → int add; else → double add
//   -,*: both int → int op; else → double op
//   /  : always double (real division)
func (g *Generator) emitVariantArithBody(op string) {
	sym := map[string]string{"+": "add", "-": "sub", "*": "mul", "/": "div"}[op]
	g.line(fmt.Sprintf("define ptr @__kylix_variant_%s(ptr %%a, ptr %%b) {", sym))
	g.line("entry:")
	tagALoc := g.boxAddr("%a", 0)
	tagA := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tagA, tagALoc))
	tagBLoc := g.boxAddr("%b", 0)
	tagB := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tagB, tagBLoc))
	payloadALoc := g.boxAddr("%a", 1)
	payloadA := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payloadA, payloadALoc))
	payloadBLoc := g.boxAddr("%b", 1)
	payloadB := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payloadB, payloadBLoc))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", res))
	g.line(fmt.Sprintf("  store ptr null, ptr %s", res)) // default
	// Category flags.
	isAStr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", isAStr, tagA, varTagStr))
	isBStr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", isBStr, tagB, varTagStr))
	eitherStr := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", eitherStr, isAStr, isBStr))
	bothInt := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", bothInt, g.icmpEqI32(tagA, varTagInt), g.icmpEqI32(tagB, varTagInt)))
	strLbl := g.label()
	intLbl := g.label()
	dblLbl := g.label()
	endLbl := g.label()
	chkIntLbl := g.label()
	// Dispatch by op. '+' concatenates on either-string; '-','*' coerce
	// either-string to double; '/' always uses real (double) division.
	switch op {
	case "+":
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eitherStr, strLbl, chkIntLbl))
	case "-", "*":
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eitherStr, dblLbl, chkIntLbl))
	case "/":
		g.line(fmt.Sprintf("  br label %%%s", dblLbl))
	}
	g.line(fmt.Sprintf("%s:", chkIntLbl))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bothInt, intLbl, dblLbl))
	// str concat (+ only): sa=as_str(a), sb=as_str(b), buf=malloc, strcpy+strcat, box_str
	if op == "+" {
		g.line(fmt.Sprintf("%s:", strLbl))
		sa := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %%a)", sa))
		sb := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %%b)", sb))
		buf := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 512)", buf))
		g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %s)", buf, sa))
		g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", buf, sb))
		sbox := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_str(ptr %s)", sbox, buf))
		g.line(fmt.Sprintf("  store ptr %s, ptr %s", sbox, res))
		g.line(fmt.Sprintf("  br label %%%s", endLbl))
	}
	// both int → int op → box_int
	g.line(fmt.Sprintf("%s:", intLbl))
	var ir string
	switch op {
	case "+":
		ir = g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %s", ir, payloadA, payloadB))
	case "-":
		ir = g.tmp()
		g.line(fmt.Sprintf("  %s = sub i64 %s, %s", ir, payloadA, payloadB))
	case "*":
		ir = g.tmp()
		g.line(fmt.Sprintf("  %s = mul i64 %s, %s", ir, payloadA, payloadB))
	case "/":
		// both-int '/' still real division per design (Variant / always double);
		// but bothInt branch is only taken for -,* ; '/' goes to dblLbl.
		ir = g.tmp()
		g.line(fmt.Sprintf("  %s = sdiv i64 %s, %s", ir, payloadA, payloadB))
	}
	ibox := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_int(i64 %s)", ibox, ir))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", ibox, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	// else → double op → box_float
	g.line(fmt.Sprintf("%s:", dblLbl))
	da := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @__kylix_variant_as_double(ptr %%a)", da))
	db := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @__kylix_variant_as_double(ptr %%b)", db))
	var dr string
	switch op {
	case "+":
		dr = g.tmp()
		g.line(fmt.Sprintf("  %s = fadd double %s, %s", dr, da, db))
	case "-":
		dr = g.tmp()
		g.line(fmt.Sprintf("  %s = fsub double %s, %s", dr, da, db))
	case "*":
		dr = g.tmp()
		g.line(fmt.Sprintf("  %s = fmul double %s, %s", dr, da, db))
	case "/":
		dr = g.tmp()
		g.line(fmt.Sprintf("  %s = fdiv double %s, %s", dr, da, db))
	}
	fbox := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_float(double %s)", fbox, dr))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", fbox, res))
	g.line(fmt.Sprintf("  br label %%%s", endLbl))
	g.line(fmt.Sprintf("%s:", endLbl))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", out, res))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// icmpEqI32 returns a register holding (icmp eq i32 a, b).
func (g *Generator) icmpEqI32(a string, b int) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", r, a, b))
	return r
}

// isVariantTypeExpr reports whether an AST type expression denotes Variant.
// The lexer case-folds `Variant`/`variant`/`VARIANT` to the `variant` keyword,
// so a Variant type annotation parses as *ast.VariantType (with empty Cases
// when no discriminated-union cases follow, e.g. `var v: Variant`). The
// discriminated-union declaration `type T = variant A: Int; B: Str end` is a
// separate (Go-backend) feature; when a VariantType appears as a type
// annotation it means the dynamic Variant.
func isVariantTypeExpr(t ast.Expression) bool {
	if t == nil {
		return false
	}
	if _, ok := t.(*ast.VariantType); ok {
		return true
	}
	// Defensive: also accept a plain Identifier named Variant (in case the
	// lexer ever stops treating it as a keyword).
	if ident, ok := t.(*ast.Identifier); ok {
		return strings.EqualFold(ident.Value, "Variant")
	}
	return false
}
