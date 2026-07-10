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
	g.line("define ptr @__kylix_json_JsonGetString(ptr %m, ptr %k) {")
	g.line("entry:")
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %%m, ptr %%k)", r))
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
	g.line("define i64 @__kylix_json_JsonGetInt(ptr %m, ptr %k) {")
	g.line("entry:")
	s := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %%m, ptr %%k)", s))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", r, s))
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
	trueStr := g.addString("true")
	g.line("define i1 @__kylix_json_JsonGetBool(ptr %m, ptr %k) {")
	g.line("entry:")
	s := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %%m, ptr %%k)", s))
	truePtr := g.ptrTo(trueStr, 5)
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, s, truePtr))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", r, cmp))
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
	g.line("define double @__kylix_json_JsonGetFloat(ptr %m, ptr %k) {")
	g.line("entry:")
	s := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %%m, ptr %%k)", s))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call double @strtod(ptr %s, ptr null)", r, s))
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
	g.line("define ptr @__kylix_json_JsonGetMap(ptr %m, ptr %k) {")
	g.line("entry:")
	// raw = htab_get(m, k) — the nested object's raw JSON substring (or "" on miss).
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get(ptr %%m, ptr %%k)", raw))
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
// Nested-array support not implemented (see JsonGetMap). Returns null.
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
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_JsonGetArray(ptr %s, ptr %s)", r, mReg, kReg))
	return r, "ptr", nil
}

func (g *Generator) emitJsonGetArrayBody() {
	g.line("define ptr @__kylix_json_JsonGetArray(ptr %m, ptr %k) {")
	g.line("entry:")
	g.line("  ret ptr null")
	g.line("}")
	g.line("")
}
