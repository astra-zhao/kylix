package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_encoding.go — LLVM IR implementation for the `encoding` stdlib module.
//
// Pure byte-manipulation codecs (no external C library beyond libc malloc/
// memcpy). Mirrors the Go-backend stdlib/encoding.go surface for the most
// commonly used helpers:
//   - HexEncode(s) -> String       : 2 hex chars per input byte
//   - HexDecode(s) -> String       : 1 byte per 2 hex chars
//   - Base64Encode(s) -> String   : standard base64 alphabet
//   - Base64Decode(s) -> String   : reverse the above
//
// UrlEncode/CsvEncode/JsonLinesEncode are deliberately not implemented here
// (they involve compound types — [][]string / []map — that the LLVM backend
// does not yet lower) and fall through to the default "not implemented" stub.

// emitEncodingCall dispatches a `encoding.Func(args)` / bare `Func(args)`
// call to the codec IR emitter. It emits the `call` instruction at the call
// site and queues the function body (deduped) for module-end emission.
func (g *Generator) emitEncodingCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "HexEncode":
		return g.emitEncodingHexEncodeCall(args)
	case "HexDecode":
		return g.emitEncodingHexDecodeCall(args)
	case "Base64Encode":
		return g.emitEncodingBase64EncodeCall(args)
	case "Base64Decode":
		return g.emitEncodingBase64DecodeCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; encoding.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitEncodingBody dispatches the deferred body emitter (called by
// emitPendingStdlib for each queued encoding function).
func (g *Generator) emitEncodingBody(funcName string) {
	switch funcName {
	case "HexEncode":
		g.emitEncodingHexEncodeBody()
	case "HexDecode":
		g.emitEncodingHexDecodeBody()
	case "Base64Encode":
		g.emitEncodingBase64EncodeBody()
	case "Base64Decode":
		g.emitEncodingBase64DecodeBody()
	}
}

// ---- HexEncode: ptr @__kylix_encoding_HexEncode(ptr %str) ----
//
//	For each input byte, emit two hex chars ("0".."9","a".."f"). Output
//	buffer = 2*len + 1 (null terminator). Uses snprintf("%02x").
func (g *Generator) emitEncodingHexEncodeCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("encoding.HexEncode expects 1 argument, got %d", len(args))
	}
	argReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("encoding", "HexEncode", "HexEncode", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_encoding_HexEncode(ptr %s)", r, argReg))
	return r, "ptr", nil
}

func (g *Generator) emitEncodingHexEncodeBody() {
	g.line("define ptr @__kylix_encoding_HexEncode(ptr %str) {")
	g.line("entry:")
	// len = strlen(str)
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%str)", ln))
	// out = malloc(2*len + 1)
	twoLen := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 1", twoLen, ln))
	outSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", outSize, twoLen))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", out, outSize))
	// null-terminate first (in case len==0)
	g.line(fmt.Sprintf("  store i8 0, ptr %s", out))
	// fmt string "%02x"
	fmtStr := g.addString("%02x")
	fmtPtr := g.ptrTo(fmtStr, 5)
	// loop i = 0..len
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	done := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", done, curI, ln))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, exitLbl, bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	// byte = str[i]
	bytePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%str, i64 %s", bytePtr, curI))
	bv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", bv, bytePtr))
	bvI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", bvI64, bv))
	// outPos = 2*i
	outPos := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 1", outPos, curI))
	// dst = out + outPos
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, out, outPos))
	// snprintf(dst, 3, "%02x", byte)  // 3 = 2 hex chars + null
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 3, ptr %s, i64 %s)", dst, fmtPtr, bvI64))
	// i++
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// ---- HexDecode: ptr @__kylix_encoding_HexDecode(ptr %str) ----
//
//	For each pair of input hex chars, emit one byte. Output = len/2 + 1.
//	Stops early on non-hex input (best-effort, like the Go version's
//	lax error handling).
func (g *Generator) emitEncodingHexDecodeCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("encoding.HexDecode expects 1 argument, got %d", len(args))
	}
	argReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("encoding", "HexDecode", "HexDecode", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_encoding_HexDecode(ptr %s)", r, argReg))
	return r, "ptr", nil
}

func (g *Generator) emitEncodingHexDecodeBody() {
	g.line("define ptr @__kylix_encoding_HexDecode(ptr %str) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%str)", ln))
	// outSize = len/2 + 1
	halfLen := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 1", halfLen, ln))
	outSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", outSize, halfLen))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", out, outSize))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", out))
	// i = 0 (input index), o = 0 (output index)
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	oSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", oSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", oSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	// need at least 2 chars left: i+1 < len
	iPlus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iPlus1, curI))
	hasPair := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", hasPair, iPlus1, ln))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasPair, bodyLbl, exitLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	// hi = hexval(str[i]); lo = hexval(str[i+1])
	hiPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%str, i64 %s", hiPtr, curI))
	hiC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", hiC, hiPtr))
	loPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%str, i64 %s", loPtr, iPlus1))
	loC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", loC, loPtr))
	// call helper to convert each nibble char → 0..15 (returns i64; -1 = invalid)
	hiVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_encoding_hexval(i8 %s)", hiVal, hiC))
	loVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_encoding_hexval(i8 %s)", loVal, loC))
	// byte = (hi<<4) | lo
	hiShift := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 4", hiShift, hiVal))
	byteVal := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", byteVal, hiShift, loVal))
	byteI8 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", byteI8, byteVal))
	// store at out[o]
	curO := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO, oSlot))
	dstPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dstPtr, out, curO))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", byteI8, dstPtr))
	// o++
	nextO := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextO, curO))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextO, oSlot))
	// i += 2
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	// null-terminate at out[o]
	curO2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO2, oSlot))
	termPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, out, curO2))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")

	// hexval helper: char '0'..'9' → 0..9, 'a'..'f'/'A'..'F' → 10..15, else -1
	g.line("define i64 @__kylix_encoding_hexval(i8 %c) {")
	g.line("entry:")
	cI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %%c to i64", cI64))
	// sub '0' (48)
	sub0 := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 48", sub0, cI64))
	isDigit := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 10", isDigit, sub0))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_digit, label %%check_upper", isDigit))
	g.line("ret_digit:")
	g.line(fmt.Sprintf("  ret i64 %s", sub0))
	g.line("check_upper:")
	// sub 'A' (65); if < 6 → val = sub+10
	subA := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 65", subA, cI64))
	isUpper := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 6", isUpper, subA))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_upper, label %%check_lower", isUpper))
	g.line("ret_upper:")
	upperVal := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 10", upperVal, subA))
	g.line(fmt.Sprintf("  ret i64 %s", upperVal))
	g.line("check_lower:")
	// sub 'a' (97); if < 6 → val = sub+10
	subLa := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 97", subLa, cI64))
	isLower := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 6", isLower, subLa))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_lower, label %%ret_neg", isLower))
	g.line("ret_lower:")
	lowerVal := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 10", lowerVal, subLa))
	g.line(fmt.Sprintf("  ret i64 %s", lowerVal))
	g.line("ret_neg:")
	g.line("  ret i64 -1")
	g.line("}")
	g.line("")
}

// ---- Base64Encode: ptr @__kylix_encoding_Base64Encode(ptr %str) ----
//
//	Standard base64: 3 input bytes → 4 output chars. Output buffer size =
//	4*ceil(len/3) + 1.
func (g *Generator) emitEncodingBase64EncodeCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("encoding.Base64Encode expects 1 argument, got %d", len(args))
	}
	argReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("encoding", "Base64Encode", "Base64Encode", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_encoding_Base64Encode(ptr %s)", r, argReg))
	return r, "ptr", nil
}

func (g *Generator) emitEncodingBase64EncodeBody() {
	tblReg := g.addBase64Table()
	g.line("define ptr @__kylix_encoding_Base64Encode(ptr %str) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%str)", ln))
	plus2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", plus2, ln))
	div3 := g.tmp()
	g.line(fmt.Sprintf("  %s = udiv i64 %s, 3", div3, plus2))
	outLen := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 2", outLen, div3))
	outSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", outSize, outLen))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", out, outSize))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", out))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	oSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", oSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", oSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	hasMore := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", hasMore, curI, ln))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasMore, bodyLbl, exitLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	b0Ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%str, i64 %s", b0Ptr, curI))
	b0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", b0, b0Ptr))
	b0v := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", b0v, b0))
	iPlus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iPlus1, curI))
	iPlus2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iPlus2, iPlus1))
	has1 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", has1, iPlus1, ln))
	has2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", has2, iPlus2, ln))
	b1Ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%str, i64 %s", b1Ptr, iPlus1))
	b1Load := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", b1Load, b1Ptr))
	b1v := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", b1v, b1Load))
	b1sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 0", b1sel, has1, b1v))
	b2Ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%str, i64 %s", b2Ptr, iPlus2))
	b2Load := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", b2Load, b2Ptr))
	b2v := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", b2v, b2Load))
	b2sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 0", b2sel, has2, b2v))
	b0sh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 16", b0sh, b0v))
	b1sh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 8", b1sh, b1sel))
	lo12 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", lo12, b1sh, b2sel))
	triple := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", triple, b0sh, lo12))
	i0 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 18", i0, triple))
	i0m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 63", i0m, i0))
	i1 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 12", i1, triple))
	i1m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 63", i1m, i1))
	i2 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 6", i2, triple))
	i2m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 63", i2m, i2))
	i3m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 63", i3m, triple))
	curO := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO, oSlot))
	c0ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i64 %s", c0ptr, tblReg, i0m))
	c0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c0, c0ptr))
	d0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", d0, out, curO))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", c0, d0))
	c1ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i64 %s", c1ptr, tblReg, i1m))
	c1 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c1, c1ptr))
	oPlus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", oPlus1, curO))
	d1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", d1, out, oPlus1))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", c1, d1))
	c2ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i64 %s", c2ptr, tblReg, i2m))
	c2load := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c2load, c2ptr))
	c2sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i8 %s, i8 61", c2sel, has1, c2load))
	oPlus2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", oPlus2, curO))
	d2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", d2, out, oPlus2))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", c2sel, d2))
	c3ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i64 %s", c3ptr, tblReg, i3m))
	c3load := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c3load, c3ptr))
	c3sel := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i8 %s, i8 61", c3sel, has2, c3load))
	oPlus3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", oPlus3, curO))
	d3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", d3, out, oPlus3))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", c3sel, d3))
	oPlus4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 4", oPlus4, curO))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", oPlus4, oSlot))
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	curO3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO3, oSlot))
	termPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, out, curO3))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// addBase64Table emits the standard base64 alphabet as a private global
// constant and returns its register name. Idempotent via base64TableEmitted.
func (g *Generator) addBase64Table() string {
	if g.base64TableEmitted {
		return "@__kylix_b64_table"
	}
	g.base64TableEmitted = true
	alpha := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var escaped string
	for _, c := range []byte(alpha) {
		escaped += fmt.Sprintf("\\%02X", c)
	}
	g.line(fmt.Sprintf("@__kylix_b64_table = private unnamed_addr constant [64 x i8] c\"%s\", align 1", escaped))
	return "@__kylix_b64_table"
}

// ---- Base64Decode: ptr @__kylix_encoding_Base64Decode(ptr %str) ----
func (g *Generator) emitEncodingBase64DecodeCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("encoding.Base64Decode expects 1 argument, got %d", len(args))
	}
	argReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("encoding", "Base64Decode", "Base64Decode", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_encoding_Base64Decode(ptr %s)", r, argReg))
	return r, "ptr", nil
}

func (g *Generator) emitEncodingBase64DecodeBody() {
	g.line("define ptr @__kylix_encoding_Base64Decode(ptr %str) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%str)", ln))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", out, ln))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", out))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	oSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", oSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", oSlot))
	condLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", condLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	iPlus3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", iPlus3, curI))
	hasQuad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", hasQuad, iPlus3, ln))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasQuad, bodyLbl, exitLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	var vs [4]string
	for k := 0; k < 4; k++ {
		off := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", off, curI, k))
		cp := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%str, i64 %s", cp, off))
		cv := g.tmp()
		g.line(fmt.Sprintf("  %s = load i8, ptr %s", cv, cp))
		vr := g.tmp()
		g.line(fmt.Sprintf("  %s = call i64 @__kylix_encoding_b64val(i8 %s)", vr, cv))
		vs[k] = vr
	}
	v0sh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 18", v0sh, vs[0]))
	v1sh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 12", v1sh, vs[1]))
	v2sh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 6", v2sh, vs[2]))
	lo := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", lo, v2sh, vs[3]))
	mid := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", mid, v1sh, lo))
	triple := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", triple, v0sh, mid))
	b0 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 16", b0, triple))
	b0m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", b0m, b0))
	b1 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 8", b1, triple))
	b1m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", b1m, b1))
	b2m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", b2m, triple))
	curO := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO, oSlot))
	for k, val := range []string{b0m, b1m, b2m} {
		b8 := g.tmp()
		g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", b8, val))
		op := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", op, curO, k))
		dp := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dp, out, op))
		g.line(fmt.Sprintf("  store i8 %s, ptr %s", b8, dp))
	}
	oPlus3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", oPlus3, curO))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", oPlus3, oSlot))
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 4", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	curO2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO2, oSlot))
	termPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, out, curO2))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")

	// b64val helper
	g.line("define i64 @__kylix_encoding_b64val(i8 %c) {")
	g.line("entry:")
	cI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %%c to i64", cI64))
	subA := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 65", subA, cI64))
	isUpper := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 26", isUpper, subA))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_upper, label %%check_lower", isUpper))
	g.line("ret_upper:")
	g.line(fmt.Sprintf("  ret i64 %s", subA))
	g.line("check_lower:")
	subLa := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 97", subLa, cI64))
	adjLa := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 26", adjLa, subLa))
	isLower := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 26", isLower, subLa))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_lower, label %%check_digit", isLower))
	g.line("ret_lower:")
	g.line(fmt.Sprintf("  ret i64 %s", adjLa))
	g.line("check_digit:")
	subD := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 48", subD, cI64))
	adjD := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 52", adjD, subD))
	isDigit := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 10", isDigit, subD))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_digit, label %%check_plus", isDigit))
	g.line("ret_digit:")
	g.line(fmt.Sprintf("  ret i64 %s", adjD))
	g.line("check_plus:")
	isPlus := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 43", isPlus, cI64))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_plus, label %%check_slash", isPlus))
	g.line("ret_plus:")
	g.line("  ret i64 62")
	g.line("check_slash:")
	isSlash := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 47", isSlash, cI64))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_slash, label %%ret_zero", isSlash))
	g.line("ret_slash:")
	g.line("  ret i64 63")
	g.line("ret_zero:")
	g.line("  ret i64 0")
	g.line("}")
	g.line("")
}
