package llvmgen

import "fmt"

// stdlib_jsonutil_parser.go — flat-object JSON parser IR for the LLVM backend.
//
// Replaces the v4.4.0 JsonDecodeMap stub (which returned an empty hash table)
// with a real state-machine parser. Handles flat JSON objects with scalar
// values (string / number / true / false / null). Nested objects and arrays
// are not recursed into — their raw JSON substring is stored as the value so
// JsonGetString still returns usable text; JsonGetMap/JsonGetArray return null
// (documented limitation; the tutorial's flat-object use case does not need
// them).
//
// Values are stored in the shared htab as RAW STRINGS (e.g. "Kylix", "3",
// "true", "null", "1.5") — this preserves the htab's caller-managed-string
// contract shared with cache/map, avoiding the tagged-value ambiguity that a
// full recursive JsonValue tree would introduce. JsonGetString returns the
// stored string directly; JsonGetInt/JsonGetFloat/JsonGetBool convert on read
// (atoll/strtod/strcmp).
//
// Helper defines (internal — not in the stdlib dispatch table, emitted once
// per module on first JsonDecodeMap use, guarded by jsonParserEmitted):
//
//	skip_ws(s, &pos)              — advance pos past spaces/tabs/newlines
//	read_string(s, &pos) -> ptr   — read a quoted string, handle \" \\ \n \t ...
//	read_bare(s, &pos) -> ptr     — read a bare token (number/literal) until delim
//	skip_nested(s, &pos) -> ptr   — skip {...} or [...] (depth count), raw substring
//	read_value(s, &pos) -> ptr    — dispatch on first char to one of the above
//	parse_flat(s) -> ptr (htab)   — main: expect '{', loop key:value pairs

// emitJsonParserBodies emits all parser helper defines, once per module.
func (g *Generator) emitJsonParserBodies() {
	if g.jsonParserEmitted {
		return
	}
	g.jsonParserEmitted = true
	g.emitJsonSkipWs()
	g.emitJsonReadString()
	g.emitJsonReadBare()
	g.emitJsonSkipNested()
	g.emitJsonReadValue()
	g.emitJsonParseFlat()
}

// ---- skip_ws: void @__kylix_json_skip_ws(ptr %s, ptr %posSlot) ----
func (g *Generator) emitJsonSkipWs() {
	g.line("define void @__kylix_json_skip_ws(ptr %s, ptr %posSlot) {")
	g.line("entry:")
	cond := g.label()
	body := g.label()
	exit := g.label()
	g.line(fmt.Sprintf("  br label %%%s", cond))
	g.line(fmt.Sprintf("%s:", cond))
	pos := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%posSlot", pos))
	cp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp, pos))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cp))
	// isSpace = (c==' ' || c=='\t' || c=='\n' || c=='\r') || c==0(false guard)
	sp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 32", sp, c)) // ' '
	tab := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 9", tab, c)) // '\t'
	nl := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 10", nl, c)) // '\n'
	cr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 13", cr, c)) // '\r'
	a1 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", a1, sp, tab))
	a2 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", a2, a1, nl))
	isSpace := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", isSpace, a2, cr))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isSpace, body, exit))
	g.line(fmt.Sprintf("%s:", body))
	next := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", next, pos))
	g.line(fmt.Sprintf("  store i64 %s, ptr %%posSlot", next))
	g.line(fmt.Sprintf("  br label %%%s", cond))
	g.line(fmt.Sprintf("%s:", exit))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- read_string: ptr @__kylix_json_read_string(ptr %s, ptr %posSlot) ----
// s[pos] == '"'. Reads content until closing '"', handling common escapes.
func (g *Generator) emitJsonReadString() {
	g.line("define ptr @__kylix_json_read_string(ptr %s, ptr %posSlot) {")
	g.line("entry:")
	pos := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%posSlot", pos))
	start := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", start, pos)) // first content char (past opening ")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%s)", ln))
	bufSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", bufSize, ln))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, bufSize))
	outSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", outSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", outSlot))
	inSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", inSlot))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", start, inSlot))
	cond := g.label()
	g.line(fmt.Sprintf("  br label %%%s", cond))
	g.line(fmt.Sprintf("%s:", cond))
	inIdx := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", inIdx, inSlot))
	cp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp, inIdx))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cp))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 0", isNull, c)) // unterminated
	doneLbl := g.label()
	quoteLbl := g.label()
	escLbl := g.label()
	copyLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNull, doneLbl, quoteLbl))
	g.line(fmt.Sprintf("%s:", quoteLbl))
	isQuote := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 34", isQuote, c)) // '"'
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isQuote, doneLbl, escLbl))
	g.line(fmt.Sprintf("%s:", escLbl))
	isBack := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 92", isBack, c)) // '\'
	g.line(fmt.Sprintf("  br i1 %s, label %%escape, label %%%s", isBack, copyLbl))
	// escape: read next char, map, store, advance 2
	g.line("escape:")
	nextIdx := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextIdx, inIdx))
	cp2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp2, nextIdx))
	c2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c2, cp2))
	// map escape via select chain (default: pass c2 through)
	mapped := c2
	mN := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 110", mN, c2)) // 'n'
	mapped = g.emitSelect(mN, 10, mapped)
	mT := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 116", mT, c2)) // 't'
	mapped = g.emitSelect(mT, 9, mapped)
	mR := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 114", mR, c2)) // 'r'
	mapped = g.emitSelect(mR, 13, mapped)
	mQ := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 34", mQ, c2)) // '"'
	mapped = g.emitSelect(mQ, 34, mapped)
	mB := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 92", mB, c2)) // '\'
	mapped = g.emitSelect(mB, 92, mapped)
	mS := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 47", mS, c2)) // '/'
	mapped = g.emitSelect(mS, 47, mapped)
	mBS := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 98", mBS, c2)) // 'b'
	mapped = g.emitSelect(mBS, 8, mapped)
	mF := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 102", mF, c2)) // 'f'
	mapped = g.emitSelect(mF, 12, mapped)
	outIdx := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", outIdx, outSlot))
	bp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", bp, buf, outIdx))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", mapped, bp))
	newOut := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", newOut, outIdx))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newOut, outSlot))
	skipTwo := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", skipTwo, inIdx))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", skipTwo, inSlot))
	g.line(fmt.Sprintf("  br label %%%s", cond))
	// copy: store c, advance 1
	g.line(fmt.Sprintf("%s:", copyLbl))
	outIdx2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", outIdx2, outSlot))
	bp2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", bp2, buf, outIdx2))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", c, bp2))
	newOut2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", newOut2, outIdx2))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newOut2, outSlot))
	newIn := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", newIn, inIdx))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", newIn, inSlot))
	g.line(fmt.Sprintf("  br label %%%s", cond))
	// done: null-terminate, set pos past closing ", return buf
	g.line(fmt.Sprintf("%s:", doneLbl))
	finOut := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", finOut, outSlot))
	bp3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", bp3, buf, finOut))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", bp3))
	finIn := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", finIn, inSlot))
	newPos := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", newPos, finIn)) // past closing "
	g.line(fmt.Sprintf("  store i64 %s, ptr %%posSlot", newPos))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// emitSelect emits a select instruction: cond ? val : prev, returns new reg.
func (g *Generator) emitSelect(condReg string, val int8, prevReg string) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i8 %d, i8 %s", r, condReg, val, prevReg))
	return r
}

// ---- read_bare: ptr @__kylix_json_read_bare(ptr %s, ptr %posSlot) ----
// Reads a bare token until ',', '}', ']', whitespace, or null. malloc+copy.
func (g *Generator) emitJsonReadBare() {
	g.line("define ptr @__kylix_json_read_bare(ptr %s, ptr %posSlot) {")
	g.line("entry:")
	start := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%posSlot", start))
	endSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", endSlot))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", start, endSlot))
	cond := g.label()
	body := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", cond))
	g.line(fmt.Sprintf("%s:", cond))
	idx := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", idx, endSlot))
	cp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp, idx))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cp))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 0", isNull, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_comma", isNull, done))
	g.line("chk_comma:")
	isComma := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 44", isComma, c)) // ','
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_endobj", isComma, done))
	g.line("chk_endobj:")
	isEndObj := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 125", isEndObj, c)) // '}'
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_endarr", isEndObj, done))
	g.line("chk_endarr:")
	isEndArr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 93", isEndArr, c)) // ']'
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_sp", isEndArr, done))
	g.line("chk_sp:")
	isSp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 32", isSp, c)) // ' '
	isTab := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 9", isTab, c))
	isNl := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 10", isNl, c))
	isCr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 13", isCr, c))
	a1 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", a1, isSp, isTab))
	a2 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", a2, a1, isNl))
	isWs := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", isWs, a2, isCr))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isWs, done, body))
	g.line(fmt.Sprintf("%s:", body))
	next := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", next, idx))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", next, endSlot))
	g.line(fmt.Sprintf("  br label %%%s", cond))
	g.line(fmt.Sprintf("%s:", done))
	end := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", end, endSlot))
	length := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", length, end, start))
	allocSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", allocSize, length))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, allocSize))
	src := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", src, start))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", buf, src, length))
	term := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", term, buf, length))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", term))
	g.line(fmt.Sprintf("  store i64 %s, ptr %%posSlot", end)) // don't consume delimiter
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// ---- skip_nested: ptr @__kylix_json_skip_nested(ptr %s, ptr %posSlot) ----
// s[pos] is '{' or '['. Skip to matching close (depth count), return raw substring.
func (g *Generator) emitJsonSkipNested() {
	g.line("define ptr @__kylix_json_skip_nested(ptr %s, ptr %posSlot) {")
	g.line("entry:")
	start := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%posSlot", start))
	startCp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", startCp, start))
	openChar := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", openChar, startCp))
	// close = (open == '{') ? '}' : ']'  (default ']' for '[')
	isObj := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 123", isObj, openChar)) // '{'
	closeChar := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i8 125, i8 93", closeChar, isObj)) // '}' or ']'
	depthSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", depthSlot))
	g.line(fmt.Sprintf("  store i64 1, ptr %s", depthSlot))
	curSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", curSlot))
	afterOpen := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", afterOpen, start))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", afterOpen, curSlot))
	cond := g.label()
	incLbl := g.label()
	decLbl := g.label()
	advLbl := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", cond))
	g.line(fmt.Sprintf("%s:", cond))
	cur := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", cur, curSlot))
	cp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp, cur))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cp))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 0", isNull, c))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_open", isNull, done))
	g.line("chk_open:")
	isOpen := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, %s", isOpen, c, openChar))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_close", isOpen, incLbl))
	g.line("chk_close:")
	isClose := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, %s", isClose, c, closeChar))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isClose, decLbl, advLbl))
	g.line(fmt.Sprintf("%s:", incLbl))
	d := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", d, depthSlot))
	d2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", d2, d))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", d2, depthSlot))
	g.line(fmt.Sprintf("  br label %%%s", advLbl))
	g.line(fmt.Sprintf("%s:", decLbl))
	d3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", d3, depthSlot))
	d4 := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 1", d4, d3))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", d4, depthSlot))
	isZero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", isZero, d4))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isZero, done, advLbl))
	g.line(fmt.Sprintf("%s:", advLbl))
	next := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", next, cur))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", next, curSlot))
	g.line(fmt.Sprintf("  br label %%%s", cond))
	g.line(fmt.Sprintf("%s:", done))
	end := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", end, curSlot))
	// Advance past the closing '}' / ']' so the caller's pos points at the
	// next token (',' or outer close), not back at this nesting's close
	// char. Without this, parse_flat saw '}' right after a nested value and
	// treated it as the end of the OUTER object, dropping any trailing
	// sibling keys (e.g. '{"user":{...},"version":3}' lost "version").
	endAfter := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", endAfter, end))
	length := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", length, end, start))
	allocSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", allocSize, length))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, allocSize))
	src := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", src, start))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", buf, src, length))
	term := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", term, buf, length))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", term))
	g.line(fmt.Sprintf("  store i64 %s, ptr %%posSlot", endAfter))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// ---- read_value: ptr @__kylix_json_read_value(ptr %s, ptr %posSlot) ----
func (g *Generator) emitJsonReadValue() {
	g.line("define ptr @__kylix_json_read_value(ptr %s, ptr %posSlot) {")
	g.line("entry:")
	pos := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%posSlot", pos))
	cp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp, pos))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cp))
	strLbl := g.label()
	nestLbl := g.label()
	bareLbl := g.label()
	isQuote := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 34", isQuote, c)) // '"'
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_obj", isQuote, strLbl))
	g.line("chk_obj:")
	isObj := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 123", isObj, c)) // '{'
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_arr", isObj, nestLbl))
	g.line("chk_arr:")
	isArr := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 91", isArr, c)) // '['
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isArr, nestLbl, bareLbl))
	g.line(fmt.Sprintf("%s:", strLbl))
	sr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_read_string(ptr %%s, ptr %%posSlot)", sr))
	g.line(fmt.Sprintf("  ret ptr %s", sr))
	g.line(fmt.Sprintf("%s:", nestLbl))
	nr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_skip_nested(ptr %%s, ptr %%posSlot)", nr))
	g.line(fmt.Sprintf("  ret ptr %s", nr))
	g.line(fmt.Sprintf("%s:", bareLbl))
	br := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_read_bare(ptr %%s, ptr %%posSlot)", br))
	g.line(fmt.Sprintf("  ret ptr %s", br))
	g.line("}")
	g.line("")
}

// ---- parse_flat: ptr @__kylix_json_parse_flat(ptr %s) -> htab ----
func (g *Generator) emitJsonParseFlat() {
	g.line("define ptr @__kylix_json_parse_flat(ptr %s) {")
	g.line("entry:")
	htab := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_new()", htab))
	posSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", posSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", posSlot))
	g.line(fmt.Sprintf("  call void @__kylix_json_skip_ws(ptr %%s, ptr %s)", posSlot))
	pos := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pos, posSlot))
	cp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp, pos))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cp))
	isOpen := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 123", isOpen, c)) // '{'
	objLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%ret_empty", isOpen, objLbl))
	g.line("ret_empty:")
	g.line(fmt.Sprintf("  ret ptr %s", htab))
	// parse_obj: pos past '{', enter pair loop
	g.line(fmt.Sprintf("%s:", objLbl))
	afterOpen := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", afterOpen, pos))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", afterOpen, posSlot))
	loopLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	g.line(fmt.Sprintf("  call void @__kylix_json_skip_ws(ptr %%s, ptr %s)", posSlot))
	pos2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pos2, posSlot))
	cp2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp2, pos2))
	c2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c2, cp2))
	isClose := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 125", isClose, c2)) // '}'
	doneLbl := g.label()
	keyLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_quote", isClose, doneLbl))
	g.line("chk_quote:")
	isQuote := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 34", isQuote, c2)) // '"'
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%ret_done", isQuote, keyLbl))
	g.line("ret_done:")
	g.line(fmt.Sprintf("  ret ptr %s", htab)) // malformed — return what we have
	// read key
	g.line(fmt.Sprintf("%s:", keyLbl))
	key := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_read_string(ptr %%s, ptr %s)", key, posSlot))
	g.line(fmt.Sprintf("  call void @__kylix_json_skip_ws(ptr %%s, ptr %s)", posSlot))
	pos3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pos3, posSlot))
	cp3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp3, pos3))
	c3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c3, cp3))
	isColon := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 58", isColon, c3)) // ':'
	colonLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%ret_done", isColon, colonLbl))
	g.line(fmt.Sprintf("%s:", colonLbl))
	afterColon := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", afterColon, pos3))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", afterColon, posSlot))
	g.line(fmt.Sprintf("  call void @__kylix_json_skip_ws(ptr %%s, ptr %s)", posSlot))
	val := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_read_value(ptr %%s, ptr %s)", val, posSlot))
	g.line(fmt.Sprintf("  call void @__kylix_htab_put(ptr %s, ptr %s, ptr %s)", htab, key, val))
	g.line(fmt.Sprintf("  call void @__kylix_json_skip_ws(ptr %%s, ptr %s)", posSlot))
	pos4 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pos4, posSlot))
	cp4 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%s, i64 %s", cp4, pos4))
	c4 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c4, cp4))
	isComma := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 44", isComma, c4)) // ','
	commaLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_close2", isComma, commaLbl))
	g.line("chk_close2:")
	isClose2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 125", isClose2, c4)) // '}'
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%ret_done", isClose2, doneLbl))
	g.line(fmt.Sprintf("%s:", commaLbl))
	afterComma := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", afterComma, pos4))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", afterComma, posSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	g.line(fmt.Sprintf("  ret ptr %s", htab))
	g.line("}")
	g.line("")
}
