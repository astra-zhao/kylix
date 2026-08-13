package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_jwt.go — LLVM IR implementation of the `jwt` stdlib module.
//
// HS256 (HMAC-SHA256): token = base64url(header) "." base64url(payload) "."
// base64url(HMAC-SHA256(secret, signing)). The header/payload JSON uses the
// same key order as Go's json.Marshal (exp, iat, sub) so tokens interoperate
// with the Go backend. v6.3.0.

// jwtB64URLAlphabet is the base64url alphabet (RFC 4648 §5, no padding).
const jwtB64URLAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

func (g *Generator) emitJwtCall(funcName string, args []ast.Expression) (string, string, error) {
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	switch funcName {
	case "JwtSign":
		return g.emitJwtSignCall(args)
	case "JwtVerify":
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i1 0, 0 ; JwtVerify stub (v6.3.0)", r))
		return r, "i1", nil
	case "JwtSubject":
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; jwt.%s stub", r, funcName))
		return r, "i64", nil
	}
}

func (g *Generator) emitJwtBody(funcName string) {
	switch funcName {
	case "JwtSign":
		g.emitJwtSignBody()
	case "b64url":
		g.emitJwtB64URLBody()
	case "hexdecode":
		g.emitJwtHexDecodeBody()
	}
}

// ---- JwtSign: ptr @__kylix_jwt_JwtSign(ptr %secret, ptr %subject, i64 %expiresIn)
//
// Payload JSON key order matches Go's json.Marshal (exp < iat < sub). iat/exp
// come from time(); exp is always present (== iat when expiresIn<=0) so the
// payload is built with a single snprintf — the Go backend omits exp only when
// expiresIn==0, so tokens are byte-identical in the expiresIn>0 case and
// always verifiable.
func (g *Generator) emitJwtSignCall(args []ast.Expression) (string, string, error) {
	if len(args) < 2 {
		return "", "", fmt.Errorf("jwt.JwtSign expects at least (secret, subject), got %d", len(args))
	}
	secretReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	subjectReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	expiryReg := "0"
	if len(args) >= 3 {
		r, _, err := g.emitExpr(args[2])
		if err != nil {
			return "", "", err
		}
		expiryReg = r
	}
	g.enqueueStdlib("jwt", "JwtSign", "JwtSign", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_JwtSign(ptr %s, ptr %s, i64 %s)", r, secretReg, subjectReg, expiryReg))
	return r, "ptr", nil
}

func (g *Generator) emitJwtSignBody() {
	g.enqueueStdlib("jwt", "b64url", "b64url", 0)
	g.enqueueStdlib("jwt", "hexdecode", "hexdecode", 0)
	g.enqueueStdlib("crypto", "HmacSha256", "HmacSha256", 0)

	g.line("define ptr @__kylix_jwt_JwtSign(ptr %secret, ptr %subject, i64 %expiresIn) {")
	g.line("entry:")
	// headerEnc = b64url("{\"alg\":\"HS256\",\"typ\":\"JWT\"}")
	headerLit := `{"alg":"HS256","typ":"JWT"}`
	headerPtr := g.ptrTo(g.addString(headerLit), len(headerLit)+1)
	headerEnc := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_b64url(ptr %s, i64 %d)", headerEnc, headerPtr, len(headerLit)))

	// now = time(); exp = now + expiresIn
	nowReg := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @time(ptr null)", nowReg))
	expReg := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %%expiresIn", expReg, nowReg))

	// payload = {"exp":<exp>,"iat":<now>,"sub":"<subject>"}
	payloadBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [512 x i8], align 8", payloadBuf))
	payloadPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [512 x i8], ptr %s, i64 0, i64 0", payloadPtr, payloadBuf))
	fmtSpec := `{"exp":%lld,"iat":%lld,"sub":"%s"}`
	fmtPtr := g.ptrTo(g.addString(fmtSpec), len(fmtSpec)+1)
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 512, ptr %s, i64 %s, i64 %s, ptr %%subject)",
		payloadPtr, fmtPtr, expReg, nowReg))
	payloadLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", payloadLen, payloadPtr))
	payloadEnc := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_b64url(ptr %s, i64 %s)", payloadEnc, payloadPtr, payloadLen))

	// signing = headerEnc "." payloadEnc
	signing := g.jwtConcat(headerEnc, payloadEnc)

	// sig = b64url(hexdecode(HmacSha256(secret, signing), 64), 32)
	hmacHex := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_HmacSha256(ptr %%secret, ptr %s)", hmacHex, signing))
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_hexdecode(ptr %s, i64 64)", raw, hmacHex))
	sig := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_b64url(ptr %s, i64 32)", sig, raw))

	// token = signing "." sig
	token := g.jwtConcatDot(signing, sig)
	g.line(fmt.Sprintf("  ret ptr %s", token))
	g.line("}")
	g.line("")
}

// jwtConcat returns a malloc'd buffer containing a + "." + b.
func (g *Generator) jwtConcat(a, b string) string {
	dot := g.ptrTo(g.addString("."), 2)
	return g.jwtConcat3(a, dot, b)
}

func (g *Generator) jwtConcatDot(a, b string) string {
	return g.jwtConcat(a, b)
}

func (g *Generator) jwtConcat3(a, b, c string) string {
	la := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", la, a))
	lb := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", lb, b))
	lc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", lc, c))
	ab := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", ab, la, lb))
	total := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", total, ab, lc))
	plus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", plus1, total))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, plus1))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %s)", buf, a))
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", buf, b))
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", buf, c))
	return buf
}

// ---- b64url: ptr @__kylix_jwt_b64url(ptr %str, i64 %n)
//
// base64url-encodes the first n bytes of %str (RFC 4648 §5, no padding).
func (g *Generator) emitJwtB64URLBody() {
	table := g.addString(jwtB64URLAlphabet)
	g.line("define ptr @__kylix_jwt_b64url(ptr %str, i64 %n) {")
	g.line("entry:")
	rem := g.tmp()
	g.line(fmt.Sprintf("  %s = srem i64 %%n, 3", rem))
	full := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %%n, %s", full, rem))
	// out = malloc(4*((n+2)/3) + 1)
	nPlus2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %%n, 2", nPlus2))
	div3 := g.tmp()
	g.line(fmt.Sprintf("  %s = sdiv i64 %s, 3", div3, nPlus2))
	four := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 4", four, div3))
	one := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", one, four))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", out, one))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	oSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", oSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", oSlot))

	loopLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	lt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", lt, curI, full))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lt, bodyLbl, exitLbl))

	g.line(fmt.Sprintf("%s:", bodyLbl))
	b0 := g.emitLoadByte("%str", curI, 0)
	b1 := g.emitLoadByte("%str", curI, 1)
	b2 := g.emitLoadByte("%str", curI, 2)
	o := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", o, oSlot))
	// 4 chars
	c0 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 2", c0, b0))
	g.emitStoreB64Char(out, o, 0, table, c0)
	sh := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 3", sh, b0))
	sh1 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 4", sh1, sh))
	sh2 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 4", sh2, b1))
	c1 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", c1, sh1, sh2))
	g.emitStoreB64Char(out, o, 1, table, c1)
	sh3 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 15", sh3, b1))
	sh4 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 2", sh4, sh3))
	sh5 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 6", sh5, b2))
	c2 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", c2, sh4, sh5))
	g.emitStoreB64Char(out, o, 2, table, c2)
	sh6 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 63", sh6, b2))
	g.emitStoreB64Char(out, o, 3, table, sh6)
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	nextO := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 4", nextO, o))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextO, oSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))

	// tail: rem == 1 → 2 chars; rem == 2 → 3 chars.
	g.line(fmt.Sprintf("%s:", exitLbl))
	curI2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI2, iSlot))
	o2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", o2, oSlot))
	// rem >= 1?
	remGe1 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 1", remGe1, rem))
	oneLbl := g.label()
	zeroLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", remGe1, oneLbl, zeroLbl))
	g.line(fmt.Sprintf("%s:", oneLbl))
	b0t := g.emitLoadByte("%str", curI2, 0)
	c0t := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 2", c0t, b0t))
	g.emitStoreB64Char(out, o2, 0, table, c0t)
	// rem >= 2?
	remGe2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 2", remGe2, rem))
	twoLbl := g.label()
	oneDoneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", remGe2, twoLbl, oneDoneLbl))
	g.line(fmt.Sprintf("%s:", twoLbl))
	b1t := g.emitLoadByte("%str", curI2, 1)
	sh7 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 3", sh7, b0t))
	sh8 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 4", sh8, sh7))
	sh9 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 4", sh9, b1t))
	c1t := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", c1t, sh8, sh9))
	g.emitStoreB64Char(out, o2, 1, table, c1t)
	sh10 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 15", sh10, b1t))
	sh11 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 2", sh11, sh10))
	g.emitStoreB64Char(out, o2, 2, table, sh11)
	oAfter2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", oAfter2, o2))
	g.line(fmt.Sprintf("  br label %%%s", zeroLbl))
	g.line(fmt.Sprintf("%s:", oneDoneLbl))
	oAfter1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", oAfter1, o2))
	g.line(fmt.Sprintf("  br label %%%s", zeroLbl))
	// zero: write NUL at out[finalO]
	g.line(fmt.Sprintf("%s:", zeroLbl))
	finalO := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i64 [ %s, %%%s ], [ %s, %%%s ], [ %s, %%%s ]",
		finalO, o2, exitLbl, oAfter2, twoLbl, oAfter1, oneDoneLbl))
	g.emitStoreZero(out, finalO)
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// ---- hexdecode: ptr @__kylix_jwt_hexdecode(ptr %hex, i64 %n)
//
// Decodes %n hex characters into a malloc'd buffer of n/2 bytes.
func (g *Generator) emitJwtHexDecodeBody() {
	g.line("define ptr @__kylix_jwt_hexdecode(ptr %hex, i64 %n) {")
	g.line("entry:")
	half := g.tmp()
	g.line(fmt.Sprintf("  %s = sdiv i64 %%n, 2", half))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, half))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	loopLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	lt := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", lt, curI, half))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lt, bodyLbl, exitLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	hi := g.emitHexNibble("%hex", curI, 0)
	lo := g.emitHexNibble("%hex", curI, 1)
	sh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 4", sh, hi))
	orv := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", orv, sh, lo))
	bp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", bp, buf, curI))
	bv := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", bv, orv))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", bv, bp))
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// emitLoadByte loads src[curI+off] as an i64 (off = 0/1/2).
func (g *Generator) emitLoadByte(src, curI string, off int) string {
	base := curI
	if off > 0 {
		base = g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", base, curI, off))
	}
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", p, src, base))
	v := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", v, p))
	z := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", z, v))
	return z
}

// emitStoreB64Char stores table[idx] at out[o+off].
func (g *Generator) emitStoreB64Char(out, o string, off int, table, idx string) {
	pos := o
	if off > 0 {
		pos = g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", pos, o, off))
	}
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", p, out, pos))
	tp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x i8], ptr %s, i64 0, i64 %s", tp, table, idx))
	tv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", tv, tp))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", tv, p))
}

// emitStoreZero stores i8 0 at out[o].
func (g *Generator) emitStoreZero(out, o string) {
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", p, out, o))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", p))
}

// emitHexNibble returns the 4-bit value of hex[2*curI + basePos].
func (g *Generator) emitHexNibble(hexPtr, curI string, basePos int) string {
	twoI := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 1", twoI, curI))
	pos := g.tmp()
	if basePos == 0 {
		pos = twoI
	} else {
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", pos, twoI, basePos))
	}
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", p, hexPtr, pos))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, p))
	ci := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", ci, c))
	dig := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 48", dig, ci))
	isDigit := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ule i64 %s, 9", isDigit, dig))
	hexv := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 87", hexv, ci)) // 'a'(97)-87 = 10
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", res, isDigit, dig, hexv))
	return res
}
