package llvmgen

import (
	"fmt"
	"kylix/ast"
	"strings"
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
	case "JsonEncode", "JsonEncodePretty":
		return g.emitJsonEncodeCall(args)
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
	case "JsonEncode":
		g.emitJsonEncodeBody()
	case "JsonEncodePretty":
		g.emitJsonEncodeBody() // scalar Variants: pretty == compact (no nesting to indent)
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
	// v0.5.1: the map's value slots hold Variant boxes; unbox to string.
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
	// v0.5.1: unbox the Variant (variant_as_int dispatches by tag).
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
	// v0.5.1: unbox the Variant (variant_as_bool dispatches by tag).
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
	// v0.5.1: unbox the Variant (variant_as_double dispatches by tag).
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
// Nested-object support (v0.4.7): the flat parser stores nested objects as
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
	// v0.5.1: the map's value slots hold Variant boxes; the nested object's
	// raw substring is stored as a str box, so unbox it before re-parsing.
	// v0.6.8: value_to_variant now boxes nested objects as a map-Variant box,
	// so return its htab directly when the box tag is map.
	g.needVariantRuntime = true
	g.line("define ptr @__kylix_json_JsonGetMap(ptr %m, ptr %k) {")
	g.line("entry:")
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %%m, ptr %%k)", box))
	tagLoc := g.boxAddr(box, 0)
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagLoc))
	isMap := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, %d", isMap, tag, varTagMap))
	mapLbl := g.label()
	oldLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isMap, mapLbl, oldLbl))
	// v0.6.8 fast path: the value is already a nested map box → return its htab.
	g.line(fmt.Sprintf("%s:", mapLbl))
	payloadLoc := g.boxAddr(box, 1)
	payload := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", payload, payloadLoc))
	htab := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", htab, payload))
	g.line(fmt.Sprintf("  ret ptr %s", htab))
	// Legacy path: raw substring in a str box → re-parse with parse_flat.
	g.line(fmt.Sprintf("%s:", oldLbl))
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
// Nested-array support (v0.4.9). The flat parser stores a JSON array as its
// raw substring (skip_nested). JsonGetArray retrieves that substring and parses
// it into a Kylix dynamic-array slice struct { ptr items; i64 len; i64 cap },
// where:
//   - items points to a malloc'd [cap x ptr] of C strings
//   - each element is the array element's text: scalars as their JSON text
//     ("1", "true", "\"hi\""), nested objects/arrays as their raw JSON substring
//   - len = element count; cap = allocated capacity (≥ len)
//
// This is the array analogue of v0.4.7's JsonGetMap. A full Variant runtime
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
	// Result is a {ptr, i64, i64} slice written into a local alloca, then loaded
	// back as a value so callers can store it into a `var arr: array of ...`
	// slot with a single struct copy. (v0.6.1: the alloca register itself is a
	// ptr, not a slice value — returning it with the "{ ptr, i64, i64 }" type
	// made `arr := JsonGetArray(...)` emit `store {ptr,i64,i64} %t, ptr %arr`
	// with a ptr-typed source, which llc rejects.)
	retAlloca := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca { ptr, i64, i64 }, align 8", retAlloca))
	g.line(fmt.Sprintf("  call void @__kylix_json_JsonGetArray(ptr %s, ptr %s, ptr %s)",
		retAlloca, mReg, kReg))
	retVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load { ptr, i64, i64 }, ptr %s", retVal, retAlloca))
	return retVal, "{ ptr, i64, i64 }", nil
}

func (g *Generator) emitJsonGetArrayBody() {
	// Ensure parse helpers + the new array parser are emitted.
	g.emitJsonParserBodies()
	g.emitJsonArrayParserBodies()
	// v0.5.0: parse_array now produces Variant boxes (value_to_variant calls
	// box_str/box_float/box_bool), so the Variant runtime must be emitted.
	g.needVariantRuntime = true
	emptyStr := g.addString("")
	g.line("define void @__kylix_json_JsonGetArray(ptr %out, ptr %m, ptr %k) {")
	g.line("entry:")
	// v0.5.1: the array's raw substring is stored as a str box; unbox first.
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
	// v0.5.0: items[i] is now a Variant box ptr; unbox to its string form
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

// ── JsonEncode / JsonEncodePretty: Variant → JSON string ──────────────────
//
// v0.5.9: LLVM implementation of jsonutil.JsonEncode. Encodes a Variant box
// ({i32 tag, i64 payload}) as a JSON string matching Go's json.Marshal for the
// scalar types Variant can hold:
//
//	tag 0 nil   → null
//	tag 1 int   → %lld decimal
//	tag 2 float → Go-style shortest float (see __kylix_json_float_str)
//	tag 3 str   → "..." with Go json escaping (\" \\ \u00XX < > &)
//	tag 4 bool  → true / false
//
// JsonEncodePretty produces the same scalar output (a scalar Variant has no
// nesting to indent). The emitted helpers call malloc/strlen/snprintf/strtod
// from libc plus the Variant runtime, so needVariantRuntime is set.
func (g *Generator) emitJsonEncodeCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("jsonutil.JsonEncode expects 1 argument, got %d", len(args))
	}
	vReg, vType, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.needVariantRuntime = true
	vBox := g.emitVariantBox(vReg, vType)
	g.enqueueStdlib("jsonutil", "JsonEncode", "JsonEncode", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonEncode(ptr %s)", r, vBox))
	return r, "ptr", nil
}

func (g *Generator) emitJsonEncodeBody() {
	g.needVariantRuntime = true
	nullStr := g.addString("null")
	trueStr := g.addString("true")
	falseStr := g.addString("false")
	intFmt := g.addString("%lld")

	g.line("define ptr @__kylix_json_JsonEncode(ptr %v) {")
	g.line("entry:")
	// ptrTo must run after the define — the GEPs it emits belong inside the fn.
	nullPtr := g.ptrTo(nullStr, 5)
	truePtr := g.ptrTo(trueStr, 5)
	falsePtr := g.ptrTo(falseStr, 6)
	intFmtPtr := g.ptrTo(intFmt, 5)
	tagP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr %%v, i32 0, i32 0", tagP))
	tag := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", tag, tagP))
	intLbl := g.label()
	floatLbl := g.label()
	strLbl := g.label()
	boolLbl := g.label()
	g.line(fmt.Sprintf("  switch i32 %s, label %%nil [ i32 1, label %%%s i32 2, label %%%s i32 3, label %%%s i32 4, label %%%s ]",
		tag, intLbl, floatLbl, strLbl, boolLbl))
	g.line("nil:")
	g.line(fmt.Sprintf("  ret ptr %s", nullPtr))
	// int → %lld
	g.line(fmt.Sprintf("%s:", intLbl))
	p1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr %%v, i32 0, i32 1", p1))
	ival := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", ival, p1))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 32)", buf))
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 32, ptr %s, i64 %s)", buf, intFmtPtr, ival))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	// float → shortest Go representation
	g.line(fmt.Sprintf("%s:", floatLbl))
	p2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr %%v, i32 0, i32 1", p2))
	fbits := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", fbits, p2))
	dval := g.tmp()
	g.line(fmt.Sprintf("  %s = bitcast i64 %s to double", dval, fbits))
	fstr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_float_str(double %s)", fstr, dval))
	g.line(fmt.Sprintf("  ret ptr %s", fstr))
	// str → escaped, quoted
	g.line(fmt.Sprintf("%s:", strLbl))
	p3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr %%v, i32 0, i32 1", p3))
	sptr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", sptr, p3))
	esc := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_escape_str(ptr %s)", esc, sptr))
	g.line(fmt.Sprintf("  ret ptr %s", esc))
	// bool → true / false
	g.line(fmt.Sprintf("%s:", boolLbl))
	p4 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr %%v, i32 0, i32 1", p4))
	bval := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", bval, p4))
	isTrue := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", isTrue, bval))
	sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", sel, isTrue, truePtr, falsePtr))
	g.line(fmt.Sprintf("  ret ptr %s", sel))
	g.line("}")
	g.line("")
	g.emitJsonEscapeStr()
	g.emitJsonFloatStr()
}

// emitJsonEscapeStr emits @__kylix_json_escape_str — a quoted, Go-json-escaped
// copy of %s. Escapes match encoding/json's encoder:
//
//	" → \", \ → \\, < → \u003c, > → \u003e, & → \u0026,
//	bytes < 0x20 → \u00xx (two lowercase hex digits).
//
// Output buffer is sized 6*len+3 (worst case: every byte → 6 chars + quotes +
// NUL). Returns a heap-allocated NUL-terminated string. All escape branches
// jump to a single %next block that feeds the loop header's phi nodes.
func (g *Generator) emitJsonEscapeStr() {
	g.line("define ptr @__kylix_json_escape_str(ptr %s) {")
	g.line("entry:")
	lenV := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%s)", lenV))
	sz := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 6", sz, lenV))
	sz2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", sz2, sz))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, sz2))
	g.line(fmt.Sprintf("  store i8 34, ptr %s", buf)) // opening '"'
	b1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", b1, buf))
	loopLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	bodyLbl := g.label()
	finishLbl := g.label()
	ctrlLbl := g.label()
	quoteLbl := g.label()
	bsLbl := g.label()
	ltLbl := g.label()
	gtLbl := g.label()
	ampLbl := g.label()
	plainLbl := g.label()
	nextLbl := g.label()

	g.line(fmt.Sprintf("%s:", loopLbl))
	iPhi := g.tmp()
	wpPhi := g.tmp()
	// Reserve backedge register names before the phi so the loop header can
	// reference them; the %next block defines them.
	iNextName := g.tmp()
	wpNextName := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%entry ], [ %s, %%%s ]", iPhi, iNextName, nextLbl))
	g.line(fmt.Sprintf("  %s = phi ptr [ %s, %%entry ], [ %s, %%%s ]", wpPhi, b1, wpNextName, nextLbl))
	done := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp uge i64 %s, %s", done, iPhi, lenV))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, finishLbl, bodyLbl))

	g.line(fmt.Sprintf("%s:", bodyLbl))
	cp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%s, i64 %s", cp, iPhi))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cp))
	ctrl := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i8 %s, 32", ctrl, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_quote", ctrl, ctrlLbl))
	g.line("chk_quote:")
	isq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 34", isq, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_bs", isq, quoteLbl))
	g.line("chk_bs:")
	isb := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 92", isb, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_lt", isb, bsLbl))
	g.line("chk_lt:")
	islt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 60", islt, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_gt", islt, ltLbl))
	g.line("chk_gt:")
	isgt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 62", isgt, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_amp", isgt, gtLbl))
	g.line("chk_amp:")
	isamp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 38", isamp, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isamp, ampLbl, plainLbl))

	// emitEsc writes bytes[] at the current wp and jumps to next, returning
	// the (i+1, wp+n) values for the next-block phi.
	var incI []string
	var incW []string
	emitEsc := func(lbl string, bytes []byte) {
		g.line(fmt.Sprintf("%s:", lbl))
		for k, b := range bytes {
			loc := wpPhi
			if k > 0 {
				loc = g.tmp()
				g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %d", loc, wpPhi, int64(k)))
			}
			g.line(fmt.Sprintf("  store i8 %d, ptr %s", b, loc))
		}
		wpNew := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %d", wpNew, wpPhi, int64(len(bytes))))
		iNew := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNew, iPhi))
		g.line(fmt.Sprintf("  br label %%%s", nextLbl))
		incI = append(incI, fmt.Sprintf("[ %s, %%%s ]", iNew, lbl))
		incW = append(incW, fmt.Sprintf("[ %s, %%%s ]", wpNew, lbl))
	}

	// plain: original byte
	g.line(fmt.Sprintf("%s:", plainLbl))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", c, wpPhi))
	wpPl := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", wpPl, wpPhi))
	iPl := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iPl, iPhi))
	g.line(fmt.Sprintf("  br label %%%s", nextLbl))
	incI = append(incI, fmt.Sprintf("[ %s, %%%s ]", iPl, plainLbl))
	incW = append(incW, fmt.Sprintf("[ %s, %%%s ]", wpPl, plainLbl))

	// 2-byte escapes: \" and \\
	emitEsc(quoteLbl, []byte{92, 34}) // \"
	emitEsc(bsLbl, []byte{92, 92})    // \\
	// 6-byte HTML escapes
	emitEsc(ltLbl, []byte{92, 117, 48, 48, 51, 99})  // \u003c
	emitEsc(gtLbl, []byte{92, 117, 48, 48, 51, 101}) // \u003e
	emitEsc(ampLbl, []byte{92, 117, 48, 48, 50, 54}) // \u0026

	// control byte < 0x20 → \u00xx with lowercase hex
	g.line(fmt.Sprintf("%s:", ctrlLbl))
	g.line(fmt.Sprintf("  store i8 92, ptr %s", wpPhi)) // '\\'
	cc1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", cc1, wpPhi))
	g.line(fmt.Sprintf("  store i8 117, ptr %s", cc1)) // 'u'
	cc2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 2", cc2, wpPhi))
	g.line(fmt.Sprintf("  store i8 48, ptr %s", cc2)) // '0'
	cc3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 3", cc3, wpPhi))
	g.line(fmt.Sprintf("  store i8 48, ptr %s", cc3)) // '0'
	hi := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i8 %s, 4", hi, c))
	lo := g.tmp()
	g.line(fmt.Sprintf("  %s = and i8 %s, 15", lo, c))
	hexCh := func(v, off string) string {
		lt10 := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp ult i8 %s, 10", lt10, v))
		va := g.tmp()
		g.line(fmt.Sprintf("  %s = add i8 %s, 48", va, v)) // '0'
		vf := g.tmp()
		g.line(fmt.Sprintf("  %s = add i8 %s, 87", vf, v)) // 'a'-10
		ch := g.tmp()
		g.line(fmt.Sprintf("  %s = select i1 %s, i8 %s, i8 %s", ch, lt10, va, vf))
		loc := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", loc, wpPhi, off))
		g.line(fmt.Sprintf("  store i8 %s, ptr %s", ch, loc))
		return loc
	}
	_ = hexCh(hi, "4")
	_ = hexCh(lo, "5")
	wpCtrl := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 6", wpCtrl, wpPhi))
	iCtrl := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iCtrl, iPhi))
	g.line(fmt.Sprintf("  br label %%%s", nextLbl))
	incI = append(incI, fmt.Sprintf("[ %s, %%%s ]", iCtrl, ctrlLbl))
	incW = append(incW, fmt.Sprintf("[ %s, %%%s ]", wpCtrl, ctrlLbl))

	// next: merge all branches → loop backedge
	g.line(fmt.Sprintf("%s:", nextLbl))
	g.line(fmt.Sprintf("  %s = phi i64 %s", iNextName, strings.Join(incI, ", ")))
	g.line(fmt.Sprintf("  %s = phi ptr %s", wpNextName, strings.Join(incW, ", ")))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))

	// finish: closing quote + NUL
	g.line(fmt.Sprintf("%s:", finishLbl))
	g.line(fmt.Sprintf("  store i8 34, ptr %s", wpPhi))
	endP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", endP, wpPhi))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", endP))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// emitJsonFloatStr emits @__kylix_json_float_str(double) — a heap-allocated
// string matching Go's encoding/json float64 output (json.Marshal's
// floatEncoder). Algorithm mirrors the Go source:
//
//	fmt = 'f'; if abs != 0 && (abs < 1e-6 || abs >= 1e21) { fmt = 'e' }
//	strconv.AppendFloat(b, f, fmt, -1, 64)   // shortest round-trip
//	if fmt == 'e' { clean "e-0X" → "e-X" }
//
// Shortest round-trip digits come from %.Ng probing (strtod == f). When Go
// wants 'f' but %.Ng produced an exponent form (large/small magnitudes), the
// digits are expanded to fixed-point by @__kylix_json_float_ffmt. NaN/±Inf → ""
// (Go's json.Marshal errors; Kylix JsonEncode returns a plain String).
func (g *Generator) emitJsonFloatStr() {
	emptyStr := g.addString("")
	zeroStr := g.addString("0")
	gFmt := g.addString("%.*g")
	e0Str := g.addString("e-0")

	g.line("define ptr @__kylix_json_float_str(double %d) {")
	g.line("entry:")
	// ptrTo after define — GEPs belong inside the fn body.
	emptyPtr := g.ptrTo(emptyStr, 1)
	zeroPtr := g.ptrTo(zeroStr, 2)
	gFmtPtr := g.ptrTo(gFmt, 5)
	e0Ptr := g.ptrTo(e0Str, 4)

	isnan := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp uno double %%d, %%d", isnan))
	errLbl := g.label()
	notnanLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isnan, errLbl, notnanLbl))
	g.line(fmt.Sprintf("%s:", notnanLbl))
	abs := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @fabs(double %%d)", abs))
	inf := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp oeq double %s, 0x7FF0000000000000", inf, abs))
	chkfmtLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", inf, errLbl, chkfmtLbl))
	g.line(fmt.Sprintf("%s:", chkfmtLbl))
	z := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp oeq double %s, 0.0", z, abs))
	zeroLbl := g.label()
	fmtLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", z, zeroLbl, fmtLbl))
	// ±0.0 → "0" (Go's json.Marshal emits "0", not "-0")
	g.line(fmt.Sprintf("%s:", zeroLbl))
	g.line(fmt.Sprintf("  ret ptr %s", zeroPtr))
	g.line(fmt.Sprintf("%s:", fmtLbl))
	lt := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp olt double %s, 0.000001", lt, abs))
	ge := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp oge double %s, 1.000000e+21", ge, abs))
	lgor := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", lgor, lt, ge))
	nz := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i1 %s, true", nz, z))
	useE := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", useE, nz, lgor))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 64)", buf))

	// probe shortest %.Ng precision 1..17 (round-trip)
	initLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", initLbl))
	loopLbl := g.label()
	nextLbl := g.label()
	g.line(fmt.Sprintf("%s:", initLbl))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	precName := g.tmp() // reserved backedge
	g.line(fmt.Sprintf("%s:", loopLbl))
	prec := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 1, %%%s ], [ %s, %%%s ]", prec, initLbl, precName, nextLbl))
	prec32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", prec32, prec))
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 64, ptr %s, i32 %s, double %%d)", buf, gFmtPtr, prec32))
	d2 := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @strtod(ptr %s, ptr null)", d2, buf))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = fcmp oeq double %s, %%d", eq, d2))
	dispatchLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eq, dispatchLbl, nextLbl))
	g.line(fmt.Sprintf("%s:", nextLbl))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", precName, prec))
	lim := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ugt i64 %s, 17", lim, precName))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lim, dispatchLbl, loopLbl))

	// dispatch: hasE (exponent form) vs fixed form.
	//   fmt=e  → e_clean (strip e-0X)
	//   fmt=f && hasE → __kylix_json_float_ffmt(buf) (expand to fixed-point)
	//   fmt=f && !hasE → buf as-is
	g.line(fmt.Sprintf("%s:", dispatchLbl))
	hasE := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 101)", hasE, buf)) // 'e'
	hasEBool := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", hasEBool, hasE))
	eCleanLbl := g.label()
	noeLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", useE, eCleanLbl, noeLbl))
	g.line(fmt.Sprintf("%s:", noeLbl))
	ffmtLbl := g.label()
	ffmtDoneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasEBool, ffmtLbl, ffmtDoneLbl))
	g.line(fmt.Sprintf("%s:", ffmtLbl))
	ffmtRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_float_ffmt(ptr %s)", ffmtRes, buf))
	g.line(fmt.Sprintf("  ret ptr %s", ffmtRes))
	g.line(fmt.Sprintf("%s:", ffmtDoneLbl))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	// e-0X → e-X
	g.line(fmt.Sprintf("%s:", eCleanLbl))
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", sp, buf, e0Ptr))
	snull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", snull, sp))
	g.line(fmt.Sprintf("  br i1 %s, label %%e_clean_done, label %%e_del", snull))
	g.line("e_del:")
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 2", dst, sp))
	srcP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 3", srcP, sp))
	x := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", x, srcP))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", x, dst))
	endP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 3", endP, sp))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", endP))
	g.line("  br label %e_clean_done")
	g.line("e_clean_done:")
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	// err: NaN/Inf → ""
	g.line(fmt.Sprintf("%s:", errLbl))
	g.line(fmt.Sprintf("  ret ptr %s", emptyPtr))
	g.line("}")
	g.line("")
	g.emitJsonFloatFfmt()
}

// emitJsonFloatFfmt emits @__kylix_json_float_ffmt(ptr) — expands a %.Ng
// exponent form ("1.2345678901234567e+19", "1e+20", "5e-06") to the fixed-point
// string Go's encoding/json produces for fmt='f' with shortest digits:
//
//	pointPos = exp + 1
//	pointPos >= len(digits) → digits + zeros        (1e+20 → 100000000000000000000)
//	0 < pointPos < len       → digits[:p] . digits[p:]  (1.23e+2 → 123)
//	pointPos <= 0            → 0. zeros digits
//
// Output buffer is 64 bytes (shortest digits ≤ 17, exponent ≤ 308).
func (g *Generator) emitJsonFloatFfmt() {
	g.line("define ptr @__kylix_json_float_ffmt(ptr %in) {")
	g.line("entry:")
	ep := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %%in, i32 101)", ep))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, ep))
	noeLbl := g.label()
	haveELbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, noeLbl, haveELbl))
	g.line(fmt.Sprintf("%s:", noeLbl))
	g.line("  ret ptr %in")
	g.line(fmt.Sprintf("%s:", haveELbl))
	expP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", expP, ep))
	exp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", exp, expP))
	firstC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %%in", firstC))
	isNeg := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 45", isNeg, firstC))
	signLen := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 1, i64 0", signLen, isNeg))
	startPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %%in, i64 %s", startPtr, signLen))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 64)", out))
	negLbl := g.label()
	noNegLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNeg, negLbl, noNegLbl))
	g.line(fmt.Sprintf("%s:", negLbl))
	g.line(fmt.Sprintf("  store i8 45, ptr %s", out))
	g.line(fmt.Sprintf("  br label %%%s", noNegLbl))
	// copy digits (skip '.') from startPtr to out+signLen
	g.line(fmt.Sprintf("%s:", noNegLbl))
	wpInit := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", wpInit, out, signLen))
	initL := g.label()
	copyLoopL := g.label()
	copyNextL := g.label()
	copyDoneL := g.label()
	g.line(fmt.Sprintf("  br label %%%s", initL))
	g.line(fmt.Sprintf("%s:", initL))
	g.line(fmt.Sprintf("  br label %%%s", copyLoopL))
	g.line(fmt.Sprintf("%s:", copyLoopL))
	pos := g.tmp()
	wp := g.tmp()
	nd := g.tmp()
	posNext := g.tmp() // reserved backedge
	wpNext := g.tmp()
	ndNext := g.tmp()
	g.line(fmt.Sprintf("  %s = phi ptr [ %s, %%%s ], [ %s, %%%s ]", pos, startPtr, initL, posNext, copyNextL))
	g.line(fmt.Sprintf("  %s = phi ptr [ %s, %%%s ], [ %s, %%%s ]", wp, wpInit, initL, wpNext, copyNextL))
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%%s ], [ %s, %%%s ]", nd, initL, ndNext, copyNextL))
	atEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, %s", atEnd, pos, ep))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%copy_body", atEnd, copyDoneL))
	g.line("copy_body:")
	cc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", cc, pos))
	isDot := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 46", isDot, cc))
	storeLbl := g.label()
	skipLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isDot, skipLbl, storeLbl))
	g.line(fmt.Sprintf("%s:", storeLbl))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", cc, wp))
	wpS := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", wpS, wp))
	ndS := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", ndS, nd))
	g.line(fmt.Sprintf("  br label %%%s", copyNextL))
	g.line(fmt.Sprintf("%s:", skipLbl))
	g.line(fmt.Sprintf("  br label %%%s", copyNextL))
	g.line(fmt.Sprintf("%s:", copyNextL))
	g.line(fmt.Sprintf("  %s = phi ptr [ %s, %%%s ], [ %s, %%%s ]", wpNext, wpS, storeLbl, wp, skipLbl))
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ %s, %%%s ]", ndNext, ndS, storeLbl, nd, skipLbl))
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", posNext, pos))
	g.line(fmt.Sprintf("  br label %%%s", copyLoopL))
	// done: digits at out[sign..sign+nd), wp = out+sign+nd
	g.line(fmt.Sprintf("%s:", copyDoneL))
	pointPos := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", pointPos, exp))
	geCmp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", geCmp, pointPos, nd))
	g.line(fmt.Sprintf("  br i1 %s, label %%case_a, label %%chk_b", geCmp))
	g.line("chk_b:")
	gtCmp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i64 %s, 0", gtCmp, pointPos))
	g.line(fmt.Sprintf("  br i1 %s, label %%case_b, label %%case_c", gtCmp))

	// case_a: pointPos >= nd → digits + (pointPos-nd) zeros
	g.line("case_a:")
	zerosA := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", zerosA, pointPos, nd))
	aInitL := g.label()
	aLoopL := g.label()
	aNextL := g.label()
	aDoneL := g.label()
	g.line(fmt.Sprintf("  br label %%%s", aInitL))
	g.line(fmt.Sprintf("%s:", aInitL))
	g.line(fmt.Sprintf("  br label %%%s", aLoopL))
	g.line(fmt.Sprintf("%s:", aLoopL))
	aJ := g.tmp()
	aJName := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%%s ], [ %s, %%%s ]", aJ, aInitL, aJName, aNextL))
	aDone := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp uge i64 %s, %s", aDone, aJ, zerosA))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%a_body", aDone, aDoneL))
	g.line("a_body:")
	aP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", aP, wp, aJ))
	g.line(fmt.Sprintf("  store i8 48, ptr %s", aP))
	g.line(fmt.Sprintf("  br label %%%s", aNextL))
	g.line(fmt.Sprintf("%s:", aNextL))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", aJName, aJ))
	g.line(fmt.Sprintf("  br label %%%s", aLoopL))
	g.line(fmt.Sprintf("%s:", aDoneL))
	nulA := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", nulA, wp, zerosA))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", nulA))
	g.line(fmt.Sprintf("  ret ptr %s", out))

	// case_b: 0 < pointPos < nd → insert '.' at out[sign+pointPos]
	g.line("case_b:")
	offB := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", offB, signLen, pointPos))
	srcB := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", srcB, out, offB))
	dstB := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", dstB, srcB))
	lenB := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", lenB, nd, pointPos))
	lenB2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", lenB2, lenB))
	g.line(fmt.Sprintf("  call ptr @memmove(ptr %s, ptr %s, i64 %s)", dstB, srcB, lenB2))
	g.line(fmt.Sprintf("  store i8 46, ptr %s", srcB))
	g.line(fmt.Sprintf("  ret ptr %s", out))

	// case_c: pointPos <= 0 → "0." + zeros + digits
	g.line("case_c:")
	zerosC := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 0, %s", zerosC, pointPos))
	outSign := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", outSign, out, signLen))
	g.line(fmt.Sprintf("  store i8 48, ptr %s", outSign))
	outSign2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 1", outSign2, outSign))
	g.line(fmt.Sprintf("  store i8 46, ptr %s", outSign2))
	offC := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", offC, signLen))
	offC2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", offC2, offC, zerosC))
	dstC := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", dstC, out, offC2))
	lenC := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", lenC, nd))
	g.line(fmt.Sprintf("  call ptr @memmove(ptr %s, ptr %s, i64 %s)", dstC, outSign, lenC))
	cInitL := g.label()
	cLoopL := g.label()
	cNextL := g.label()
	cDoneL := g.label()
	g.line(fmt.Sprintf("  br label %%%s", cInitL))
	g.line(fmt.Sprintf("%s:", cInitL))
	g.line(fmt.Sprintf("  br label %%%s", cLoopL))
	g.line(fmt.Sprintf("%s:", cLoopL))
	cJ := g.tmp()
	cJName := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ 0, %%%s ], [ %s, %%%s ]", cJ, cInitL, cJName, cNextL))
	cDone := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp uge i64 %s, %s", cDone, cJ, zerosC))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%c_body", cDone, cDoneL))
	g.line("c_body:")
	cOff := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", cOff, signLen))
	cOff2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", cOff2, cOff, cJ))
	cP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr i8, ptr %s, i64 %s", cP, out, cOff2))
	g.line(fmt.Sprintf("  store i8 48, ptr %s", cP))
	g.line(fmt.Sprintf("  br label %%%s", cNextL))
	g.line(fmt.Sprintf("%s:", cNextL))
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", cJName, cJ))
	g.line(fmt.Sprintf("  br label %%%s", cLoopL))
	g.line(fmt.Sprintf("%s:", cDoneL))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}
