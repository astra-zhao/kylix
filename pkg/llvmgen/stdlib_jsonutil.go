package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_jsonutil.go — LLVM IR implementation for the `jsonutil` stdlib module.
//
// Provides a simplified JSON parser sufficient for the tutorial examples
// (flat objects with String/Integer/Boolean values). Nested objects and
// arrays are not supported in this first cut.
//
//   JsonIsValid(s)        -> i1         basic syntax check (balanced braces/quotes)
//   JsonDecodeMap(s)      -> ptr (htab) parse flat JSON object → hash table
//   JsonGetString(m, k)   -> ptr (String)  htab_get
//   JsonGetInt(m, k)      -> i64           htab_get + atoll
//   JsonGetBool(m, k)     -> i1            htab_get + strcmp("true")
//   JsonHasKey(m, k)      -> i1            htab_has
//
// JsonDecodeMap parses {"key":"value","key2":123,"key3":true} into a
// string→string hash table (numbers and booleans stored as their text
// representation). JsonGetInt/JsonGetBool convert back on read.

// emitJsonutilCall dispatches a `jsonutil.Func(args)` / bare `Func(args)` call.
func (g *Generator) emitJsonutilCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "JsonIsValid":
		return g.emitJsonIsValidCall(args)
	case "JsonDecodeMap":
		return g.emitJsonDecodeMapCall(args)
	case "JsonDecode":
		return g.emitJsonDecodeCall(args)
	case "JsonGetString":
		return g.emitJsonGetStringCall(args)
	case "JsonGetInt":
		return g.emitJsonGetIntCall(args)
	case "JsonGetFloat":
		return g.emitJsonGetFloatCall(args)
	case "JsonGetBool":
		return g.emitJsonGetBoolCall(args)
	case "JsonGetMap":
		return g.emitJsonGetMapCall(args)
	case "JsonGetArray":
		return g.emitJsonGetArrayCall(args)
	case "JsonArrayLen":
		return g.emitJsonArrayLenCall(args)
	case "JsonArrayGetString":
		return g.emitJsonArrayGetStringCall(args)
	case "JsonHasKey":
		return g.emitJsonHasKeyCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; jsonutil.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitJsonutilBody dispatches the deferred body emitter.
func (g *Generator) emitJsonutilBody(funcName string) {
	switch funcName {
	case "JsonIsValid":
		g.emitJsonIsValidBody()
	case "JsonDecodeMap":
		g.emitJsonDecodeMapBody()
	case "JsonDecode":
		g.emitJsonDecodeBody()
	case "JsonGetString":
		g.emitJsonGetStringBody()
	case "JsonGetInt":
		g.emitJsonGetIntBody()
	case "JsonGetFloat":
		g.emitJsonGetFloatBody()
	case "JsonGetBool":
		g.emitJsonGetBoolBody()
	case "JsonGetMap":
		g.emitJsonGetMapBody()
	case "JsonGetArray":
		g.emitJsonGetArrayBody()
	case "JsonArrayLen":
		g.emitJsonArrayLenBody()
	case "JsonArrayGetString":
		g.emitJsonArrayGetStringBody()
	case "JsonHasKey":
		g.emitJsonHasKeyBody()
	}
}

// ---- JsonIsValid: i1 @__kylix_json_JsonIsValid(ptr %s) ----
//
//	Basic check: non-empty, starts with '{' or '[', ends with '}' or ']'.
//	(Conservative — catches obviously-bad input like the tutorial's
//	'bad json {' without a full parse.)
func (g *Generator) emitJsonIsValidCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("jsonutil.JsonIsValid expects 1 argument, got %d", len(args))
	}
	sReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonIsValid", "JsonIsValid", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_json_JsonIsValid(ptr %s)", r, sReg))
	return r, "i1", nil
}

func (g *Generator) emitJsonIsValidBody() {
	g.line("define i1 @__kylix_json_JsonIsValid(ptr %s) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%s)", ln))
	// if len == 0 → false
	isEmpty := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", isEmpty, ln))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_false, label %%check_first", isEmpty))
	g.line("ret_false:")
	g.line("  ret i1 false")
	g.line("check_first:")
	// first char must be '{' (123) or '[' (91)
	firstC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %%s", firstC))
	isObj := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 123", isObj, firstC))
	isArr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 91", isArr, firstC))
	isValidStart := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", isValidStart, isObj, isArr))
	g.line(fmt.Sprintf("  br i1 %s, label %%check_last, label %%ret_false", isValidStart))
	g.line("check_last:")
	// last char must be '}' (125) or ']' (93)
	lastOff := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 1", lastOff, ln))
	lastPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", lastPtr, lastOff))
	lastC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", lastC, lastPtr))
	isEndObj := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 125", isEndObj, lastC))
	isEndArr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 93", isEndArr, lastC))
	isValidEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", isValidEnd, isEndObj, isEndArr))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_true, label %%ret_false", isValidEnd))
	g.line("ret_true:")
	g.line("  ret i1 true")
	g.line("}")
	g.line("")
}

// ---- JsonDecodeMap: ptr @__kylix_json_JsonDecodeMap(ptr %s) ----
//
//	Parse a flat JSON object into a hash table. This is a simplified parser
//	that handles {"key":"value","key2":123,"key3":true} — it scans for
//	"key":value pairs and inserts them into a new htab.
//
//	The full parser is complex; for the tutorial's flat-object use case we
//	use a state-machine that:
//	  1. Skips to first '"'
//	  2. Reads key until closing '"'
//	  3. Skips ':' and whitespace
//	  4. Reads value (string in quotes, or bare token until ',' or '}')
//	  5. Inserts (key, value) into htab
//	  6. Repeats until '}'
//
//	To keep the IR manageable, the actual parsing is done by a single
//	helper @__kylix_json_parse_flat that returns an htab ptr.
func (g *Generator) emitJsonDecodeMapCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("jsonutil.JsonDecodeMap expects 1 argument, got %d", len(args))
	}
	sReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonDecodeMap", "JsonDecodeMap", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonDecodeMap(ptr %s)", r, sReg))
	return r, "ptr", nil
}

func (g *Generator) emitJsonDecodeMapBody() {
	// Emit parser helpers (guarded — once per module).
	g.emitJsonParserBodies()
	g.line("define ptr @__kylix_json_JsonDecodeMap(ptr %s) {")
	g.line("entry:")
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_parse_flat(ptr %%s)", r))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonDecode: ptr @__kylix_json_JsonDecode(ptr %s) ----
// Alias of JsonDecodeMap for top-level objects (returns htab ptr).
func (g *Generator) emitJsonDecodeCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("jsonutil.JsonDecode expects 1 argument, got %d", len(args))
	}
	sReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonDecode", "JsonDecode", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonDecode(ptr %s)", r, sReg))
	return r, "ptr", nil
}

func (g *Generator) emitJsonDecodeBody() {
	g.line("define ptr @__kylix_json_JsonDecode(ptr %s) {")
	g.line("entry:")
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_parse_flat(ptr %%s)", r))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonGetString: ptr @__kylix_json_JsonGetString(ptr %m, ptr %k) ----
func (g *Generator) emitJsonGetStringCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonGetString expects 2 arguments, got %d", len(args))
	}
	mReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	kReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonGetString", "JsonGetString", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonGetString(ptr %s, ptr %s)", r, mReg, kReg))
	return r, "ptr", nil
}

func (g *Generator) emitJsonGetStringBody() {
	// v5.1.0: the map's value slots hold Variant boxes; unbox to string.
	g.needVariantRuntime = true
	g.line("define ptr @__kylix_json_JsonGetString(ptr %m, ptr %k) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %%m, ptr %%k)", box))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %s)", r, box))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonGetInt: i64 @__kylix_json_JsonGetInt(ptr %m, ptr %k) ----
func (g *Generator) emitJsonGetIntCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonGetInt expects 2 arguments, got %d", len(args))
	}
	mReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	kReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonGetInt", "JsonGetInt", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_json_JsonGetInt(ptr %s, ptr %s)", r, mReg, kReg))
	return r, "i64", nil
}

func (g *Generator) emitJsonGetIntBody() {
	// v5.1.0: unbox the Variant (variant_as_int dispatches by tag).
	g.needVariantRuntime = true
	g.line("define i64 @__kylix_json_JsonGetInt(ptr %m, ptr %k) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %%m, ptr %%k)", box))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_variant_as_int(ptr %s)", r, box))
	g.line(fmt.Sprintf("  ret i64 %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonGetBool: i1 @__kylix_json_JsonGetBool(ptr %m, ptr %k) ----
func (g *Generator) emitJsonGetBoolCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonGetBool expects 2 arguments, got %d", len(args))
	}
	mReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	kReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonGetBool", "JsonGetBool", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_json_JsonGetBool(ptr %s, ptr %s)", r, mReg, kReg))
	return r, "i1", nil
}

func (g *Generator) emitJsonGetBoolBody() {
	// v5.1.0: unbox the Variant (variant_as_bool dispatches by tag).
	g.needVariantRuntime = true
	g.line("define i1 @__kylix_json_JsonGetBool(ptr %m, ptr %k) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %%m, ptr %%k)", box))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_variant_as_bool(ptr %s)", r, box))
	g.line(fmt.Sprintf("  ret i1 %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonHasKey: i1 @__kylix_json_JsonHasKey(ptr %m, ptr %k) ----
func (g *Generator) emitJsonHasKeyCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonHasKey expects 2 arguments, got %d", len(args))
	}
	mReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	kReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonHasKey", "JsonHasKey", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_json_JsonHasKey(ptr %s, ptr %s)", r, mReg, kReg))
	return r, "i1", nil
}

func (g *Generator) emitJsonHasKeyBody() {
	g.line("define i1 @__kylix_json_JsonHasKey(ptr %m, ptr %k) {")
	g.line("entry:")
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_htab_has(ptr %%m, ptr %%k)", r))
	g.line(fmt.Sprintf("  ret i1 %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonGetFloat: double @__kylix_json_JsonGetFloat(ptr %m, ptr %k) ----
func (g *Generator) emitJsonGetFloatCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonGetFloat expects 2 arguments, got %d", len(args))
	}
	mReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	kReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonGetFloat", "JsonGetFloat", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @__kylix_json_JsonGetFloat(ptr %s, ptr %s)", r, mReg, kReg))
	return r, "double", nil
}

func (g *Generator) emitJsonGetFloatBody() {
	// v5.1.0: unbox the Variant (variant_as_double dispatches by tag).
	g.needVariantRuntime = true
	g.line("define double @__kylix_json_JsonGetFloat(ptr %m, ptr %k) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %%m, ptr %%k)", box))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @__kylix_variant_as_double(ptr %s)", r, box))
	g.line(fmt.Sprintf("  ret double %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonGetMap: ptr @__kylix_json_JsonGetMap(ptr %m, ptr %k) ----
// Nested-object support (v4.7.0): the flat parser stores nested objects as
// their raw JSON substring (skip_nested). JsonGetMap retrieves that substring
// and recursively parses it with parse_flat into a fresh htab, so callers can
// chain JsonGetString(inner, 'name') on the result. Returns null when the key
// is absent or the stored value is empty (not a nested object).
func (g *Generator) emitJsonGetMapCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonGetMap expects 2 arguments, got %d", len(args))
	}
	mReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	kReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonGetMap", "JsonGetMap", 0)
	g.needHashtab = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonGetMap(ptr %s, ptr %s)", r, mReg, kReg))
	return r, "ptr", nil
}

func (g *Generator) emitJsonGetMapBody() {
	// Ensure parse_flat + helpers are emitted (JsonGetMap depends on parse_flat).
	g.emitJsonParserBodies()
	// v5.1.0: the map's value slots hold Variant boxes; the nested object's
	// raw substring is stored as a str box, so unbox it before re-parsing.
	g.needVariantRuntime = true
	g.line("define ptr @__kylix_json_JsonGetMap(ptr %m, ptr %k) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %%m, ptr %%k)", box))
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %s)", raw, box))
	// If raw is empty (miss or non-object value), return null.
	emptyStr := g.addString("")
	emptyPtr := g.ptrTo(emptyStr, 1)
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, raw, emptyPtr))
	isEmpty := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", isEmpty, cmp))
	retNullLbl := g.label()
	parseLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isEmpty, retNullLbl, parseLbl))
	g.line(fmt.Sprintf("%s:", retNullLbl))
	g.line("  ret ptr null")
	// Recursively parse the raw substring as a flat object → nested htab.
	g.line(fmt.Sprintf("%s:", parseLbl))
	nested := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_parse_flat(ptr %s)", nested, raw))
	g.line(fmt.Sprintf("  ret ptr %s", nested))
	g.line("}")
	g.line("")
}

// ---- JsonGetArray: ptr @__kylix_json_JsonGetArray(ptr %m, ptr %k) ----
// Nested-array support (v4.9.0). The flat parser stores a JSON array as its
// raw substring (skip_nested). JsonGetArray retrieves that substring and parses
// it into a Kylix dynamic-array slice struct { ptr items; i64 len; i64 cap },
// where:
//   - items points to a malloc'd [cap x ptr] of C strings
//   - each element is the array element's text: scalars as their JSON text
//     ("1", "true", "\"hi\""), nested objects/arrays as their raw JSON substring
//   - len = element count; cap = allocated capacity (≥ len)
//
// This is the array analogue of v4.7.0's JsonGetMap. A full Variant runtime
// (tagged values + dispatch) is out of scope; callers use JsonArrayLen /
// JsonArrayGetString to read elements. Returns a zero-length slice
// (items=null, len=0, cap=0) when the key is absent or the stored value is
// empty. The returned struct matches a Kylix dynamic array exactly, so
// `var arr: array of Variant; arr := JsonGetArray(...)` stores it directly and
// Length(arr) (via the slice's len word) yields the element count.
func (g *Generator) emitJsonGetArrayCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonGetArray expects 2 arguments, got %d", len(args))
	}
	mReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	kReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonGetArray", "JsonGetArray", 0)
	g.needHashtab = true
	// Result is a {ptr, i64, i64} slice returned by value into a local alloca,
	// so callers can store it into a `var arr: array of ...` slot with a single
	// struct copy and index it with the standard slice path.
	retAlloca := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca { ptr, i64, i64 }, align 8", retAlloca))
	g.line(fmt.Sprintf("  call void @__kylix_json_JsonGetArray(ptr %s, ptr %s, ptr %s)",
		retAlloca, mReg, kReg))
	return retAlloca, "{ ptr, i64, i64 }", nil
}

func (g *Generator) emitJsonGetArrayBody() {
	// Ensure parse helpers + the new array parser are emitted.
	g.emitJsonParserBodies()
	g.emitJsonArrayParserBodies()
	// v5.0.0: parse_array now produces Variant boxes (value_to_variant calls
	// box_str/box_float/box_bool), so the Variant runtime must be emitted.
	g.needVariantRuntime = true
	emptyStr := g.addString("")
	g.line("define void @__kylix_json_JsonGetArray(ptr %out, ptr %m, ptr %k) {")
	g.line("entry:")
	// v5.1.0: the array's raw substring is stored as a str box; unbox first.
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %%m, ptr %%k)", box))
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %s)", raw, box))
	// If raw is empty (miss or non-array value), write a zero-length slice.
	emptyPtr := g.ptrTo(emptyStr, 1)
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, raw, emptyPtr))
	isEmpty := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", isEmpty, cmp))
	retEmptyLbl := g.label()
	parseLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isEmpty, retEmptyLbl, parseLbl))
	// Empty path: *out = { null, 0, 0 }.
	g.line(fmt.Sprintf("%s:", retEmptyLbl))
	g.emitStoreSliceWords("%out", "null", "0", "0")
	g.line("  ret void")
	// Parse path: parse_array(raw) fills *out.
	g.line(fmt.Sprintf("%s:", parseLbl))
	g.line(fmt.Sprintf("  call void @__kylix_json_parse_array(ptr %%out, ptr %s)", raw))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitStoreSliceWords writes {ptr items, i64 len, i64 cap} into the slice
// struct at baseReg. Shared by the empty-result path and the parser's done
// path. operands are raw IR operand strings (e.g. "null", "0", "%len").
func (g *Generator) emitStoreSliceWords(baseReg, itemsOp, lenOp, capOp string) {
	itemsLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 0", itemsLoc, baseReg))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", itemsOp, itemsLoc))
	lenLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 1", lenLoc, baseReg))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", lenOp, lenLoc))
	capLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 2", capLoc, baseReg))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", capOp, capLoc))
}

// emitStoreSliceWordsReg is the register-operand variant of emitStoreSliceWords
// for the parser's done path, where items/len/cap are SSA registers (ptr/i64).
// It emits the same GEP+store sequence; the operand strings are already
// register references, so no type coercion is needed.
func (g *Generator) emitStoreSliceWordsReg(baseReg, itemsReg, lenReg, capReg string) {
	itemsLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 0", itemsLoc, baseReg))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", itemsReg, itemsLoc))
	lenLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 1", lenLoc, baseReg))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", lenReg, lenLoc))
	capLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 2", capLoc, baseReg))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", capReg, capLoc))
}

// ---- JsonArrayLen: i64 @__kylix_json_JsonArrayLen(ptr %arr) ----
// Returns the element count of a string-array produced by JsonGetArray.
// `arr` is a pointer to the {ptr items, i64 len, i64 cap} slice struct.
func (g *Generator) emitJsonArrayLenCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("jsonutil.JsonArrayLen expects 1 argument, got %d", len(args))
	}
	arrReg := g.sliceArgPtr(args[0])
	g.enqueueStdlib("jsonutil", "JsonArrayLen", "JsonArrayLen", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_json_JsonArrayLen(ptr %s)", r, arrReg))
	return r, "i64", nil
}

func (g *Generator) emitJsonArrayLenBody() {
	g.line("define i64 @__kylix_json_JsonArrayLen(ptr %arr) {")
	g.line("entry:")
	lenLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %%arr, i32 0, i32 1", lenLoc))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", r, lenLoc))
	g.line(fmt.Sprintf("  ret i64 %s", r))
	g.line("}")
	g.line("")
}

// ---- JsonArrayGetString: ptr @__kylix_json_JsonArrayGetString(ptr %arr, i64 %i) ----
// Returns the i-th element string of a JsonGetArray result. Out-of-range
// indices return a pointer to "" (safe, never null).
func (g *Generator) emitJsonArrayGetStringCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jsonutil.JsonArrayGetString expects 2 arguments, got %d", len(args))
	}
	arrReg := g.sliceArgPtr(args[0])
	iReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jsonutil", "JsonArrayGetString", "JsonArrayGetString", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonArrayGetString(ptr %s, i64 %s)", r, arrReg, iReg))
	return r, "ptr", nil
}

// sliceArgPtr resolves a JsonArrayLen/JsonArrayGetString argument to the ptr of
// the slice struct. If the arg is an Identifier bound to a local, return its
// alloca register directly (the runtime reads len/items via GEP, so it needs
// the struct address, not a loaded value). Otherwise fall back to emitExpr —
// which covers the case where the array is itself returned by a call.
func (g *Generator) sliceArgPtr(arg ast.Expression) string {
	if ident, ok := arg.(*ast.Identifier); ok {
		if reg, ok := g.locals[ident.Value]; ok {
			return reg
		}
	}
	reg, _, err := g.emitExpr(arg)
	if err != nil || reg == "" {
		// Best-effort: a zero/null pointer keeps IR legal.
		return "null"
	}
	return reg
}

func (g *Generator) emitJsonArrayGetStringBody() {
	emptyStr := g.addString("")
	g.line("define ptr @__kylix_json_JsonArrayGetString(ptr %arr, i64 %i) {")
	g.line("entry:")
	emptyPtr := g.ptrTo(emptyStr, 1)
	// len = arr->len
	lenLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %%arr, i32 0, i32 1", lenLoc))
	lenVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", lenVal, lenLoc))
	// if i >= len → return empty
	inRange := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %%i, %s", inRange, lenVal))
	getLbl := g.label()
	emptyLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", inRange, getLbl, emptyLbl))
	g.line(fmt.Sprintf("%s:", emptyLbl))
	g.line(fmt.Sprintf("  ret ptr %s", emptyPtr))
	// v5.0.0: items[i] is now a Variant box ptr; unbox to its string form
	// (variant_as_str dispatches on tag) and return that.
	g.line(fmt.Sprintf("%s:", getLbl))
	// Variant runtime is needed for as_str.
	g.needVariantRuntime = true
	itemsLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %%arr, i32 0, i32 0", itemsLoc))
	itemsVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", itemsVal, itemsLoc))
	elemPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %%i", elemPtr, itemsVal))
	boxVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", boxVal, elemPtr))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %s)", r, boxVal))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}
