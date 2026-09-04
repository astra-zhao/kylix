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
// with the Go backend. v0.6.3.

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
		return g.emitJwtVerifyCall(args)
	case "JwtSubject":
		// JwtSubject(claims) → claims['sub'] as String.
		return g.emitJwtClaimsAccess(args, "sub", true)
	case "JwtGetString":
		// JwtGetString(claims, key) → claims[key] as String.
		return g.emitJwtClaimsAccessKey(args, true)
	case "JwtGetInt":
		// JwtGetInt(claims, key) → claims[key] as Integer.
		return g.emitJwtClaimsAccessKey(args, false)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; jwt.%s stub", r, funcName))
		return r, "i64", nil
	}
}

// emitJwtClaimsAccess reads a fixed-key claim from a decoded claims Variant
// map (v0.6.6). wantStr → String, else Integer.
func (g *Generator) emitJwtClaimsAccess(args []ast.Expression, key string, wantStr bool) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("jwt claims access expects 1 argument, got %d", len(args))
	}
	g.needVariantRuntime = true
	claimsReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	keyPtr := g.ptrTo(g.addString(key), len(key)+1)
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_map_get(ptr %s, ptr %s)", box, claimsReg, keyPtr))
	if wantStr {
		return g.emitVariantAsStr(box), "ptr", nil
	}
	return g.emitVariantAsInt(box), "i64", nil
}

// emitJwtClaimsAccessKey reads a dynamic-key claim from a decoded claims map.
func (g *Generator) emitJwtClaimsAccessKey(args []ast.Expression, wantStr bool) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jwt claims access expects 2 arguments, got %d", len(args))
	}
	g.needVariantRuntime = true
	claimsReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	keyReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	box := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_map_get(ptr %s, ptr %s)", box, claimsReg, keyReg))
	if wantStr {
		return g.emitVariantAsStr(box), "ptr", nil
	}
	return g.emitVariantAsInt(box), "i64", nil
}

func (g *Generator) emitJwtBody(funcName string) {
	switch funcName {
	case "JwtSign":
		g.emitJwtSignBody()
	case "JwtVerify":
		g.emitJwtVerifyBody()
	case "b64url":
		g.emitJwtB64URLBody()
	case "b64urldecode":
		g.emitJwtB64URLDecodeBody()
	case "b64urlval":
		g.emitJwtB64URLValBody()
	case "hexdecode":
		g.emitJwtHexDecodeBody()
	}
}

// ---- JwtVerify: ptr @__kylix_jwt_JwtVerify(ptr %secret, ptr %token) → Variant
//
// Verifies the HS256 signature only (v0.6.3): splits the token into
// header.payload.sig, recomputes base64url(HMAC-SHA256(secret, signing)) and
// compares. Returns a non-nil Variant on success, the nil Variant otherwise.
// Payload/claims parsing is a documented limitation.
func (g *Generator) emitJwtVerifyCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("jwt.JwtVerify expects 2 arguments, got %d", len(args))
	}
	secretReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	tokenReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("jwt", "JwtVerify", "JwtVerify", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_JwtVerify(ptr %s, ptr %s)", r, secretReg, tokenReg))
	return r, "variant", nil
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
	// v0.6.6: extraClaims — a Variant map of additional payload claims
	// (omitted / nil → null → no extra claims).
	extraReg := "null"
	if len(args) >= 4 {
		r, extraT, err := g.emitExpr(args[3])
		if err != nil {
			return "", "", err
		}
		// v0.6.6: a map[String]Variant argument is boxed into a Variant map
		// (JwtSign's extraClaims parameter is Variant). A plain Variant arg
		// (e.g. another JwtVerify result) passes through as a box.
		if ident, ok := args[3].(*ast.Identifier); ok && g.mapVars[ident.Value] {
			g.needVariantRuntime = true
			boxed := g.tmp()
			g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_map(ptr %s)", boxed, r))
			r = boxed
		} else if extraT == variantT {
			// already a Variant box — pass through.
		}
		extraReg = r
	}
	g.enqueueStdlib("jwt", "JwtSign", "JwtSign", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_JwtSign(ptr %s, ptr %s, i64 %s, ptr %s)", r, secretReg, subjectReg, expiryReg, extraReg))
	return r, "ptr", nil
}

func (g *Generator) emitJwtSignBody() {
	g.enqueueStdlib("jwt", "b64url", "b64url", 0)
	g.enqueueStdlib("jwt", "hexdecode", "hexdecode", 0)
	g.enqueueStdlib("crypto", "HmacSha256", "HmacSha256", 0)

	g.line("define ptr @__kylix_jwt_JwtSign(ptr %secret, ptr %subject, i64 %expiresIn, ptr %extra) {")
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

	// payload = dynamic build: "{" + extra claims + ",\"sub\":\"...\",\"iat\":N"
	// [+ ",\"exp\":M" when expiresIn > 0] + "}". Extra-claim values are
	// serialized as JSON strings (as_str) — ints stay numeric on read-back
	// via atoll. v0.6.6.
	g.needHashtab = true
	g.needVariantRuntime = true
	payloadBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [2048 x i8], align 8", payloadBuf))
	payloadPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [2048 x i8], ptr %s, i64 0, i64 0", payloadPtr, payloadBuf))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", payloadPtr))
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString("{"), 2))
	// firstSlot tracks whether we've emitted the first extra claim (only the
	// first one gets no leading comma — otherwise the payload would start
	// `{,"role":...` which parse_flat rejects). NB: the alloca MUST live in
	// entry — afterClaimsLbl loads it on the extra==null path too, and an
	// alloca in the walk-claims block does not dominate that use (llc
	// "Instruction does not dominate all uses").
	firstSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i1, align 1", firstSlot))
	g.line(fmt.Sprintf("  store i1 true, ptr %s", firstSlot))
	// Walk extraClaims (a Variant map box, or null).
	extraNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %%extra, null", extraNull))
	afterClaimsLbl := g.label()
	walkClaimsLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", extraNull, afterClaimsLbl, walkClaimsLbl))
	g.line(walkClaimsLbl + ":")
	extraPayloadPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%extra, i64 8", extraPayloadPtr))
	extraPayload := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", extraPayload, extraPayloadPtr))
	htab := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 %s to ptr", htab, extraPayload))
	keysAgg := g.tmp()
	g.line(fmt.Sprintf("  %s = call { ptr, i64 } @__kylix_htab_keys(ptr %s)", keysAgg, htab))
	items := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, i64 } %s, 0", items, keysAgg))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, i64 } %s, 1", n, keysAgg))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	claimLoop := g.label()
	claimBody := g.label()
	claimDone := g.label()
	g.line(fmt.Sprintf("  br label %%%s", claimLoop))
	g.line(fmt.Sprintf("%s:", claimLoop))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	claimEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", claimEnd, curI, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", claimEnd, claimDone, claimBody))
	g.line(fmt.Sprintf("%s:", claimBody))
	// Separator: first claim → '"', later claims → ',"'.
	isFirst := g.tmp()
	g.line(fmt.Sprintf("  %s = load i1, ptr %s", isFirst, firstSlot))
	firstSepLbl := g.label()
	laterSepLbl := g.label()
	afterSepLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isFirst, firstSepLbl, laterSepLbl))
	g.line(fmt.Sprintf("%s:", firstSepLbl))
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString("\""), 2))
	g.line(fmt.Sprintf("  store i1 false, ptr %s", firstSlot))
	g.line(fmt.Sprintf("  br label %%%s", afterSepLbl))
	g.line(fmt.Sprintf("%s:", laterSepLbl))
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString(",\""), 3))
	g.line(fmt.Sprintf("  br label %%%s", afterSepLbl))
	g.line(fmt.Sprintf("%s:", afterSepLbl))
	keyPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds ptr, ptr %s, i64 %s", keyPtr, items, curI))
	key := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", key, keyPtr))
	valBox := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %s, ptr %s)", valBox, htab, key))
	g.jwtStrcat(payloadPtr, key)
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString("\":\""), 4))
	valStr := g.emitVariantAsStr(valBox)
	g.jwtStrcat(payloadPtr, valStr)
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString("\""), 2))
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", claimLoop))
	g.line(fmt.Sprintf("%s:", claimDone))
	g.line(fmt.Sprintf("  br label %%%s", afterClaimsLbl))
	g.line(afterClaimsLbl + ":")
	// "sub" needs a leading comma only if extra claims were already emitted.
	// firstSlot starts true and is flipped to false by the first claim, so
	// firstSlot==true → nothing emitted → no comma; firstSlot==false → comma.
	hasClaims := g.tmp()
	g.line(fmt.Sprintf("  %s = load i1, ptr %s", hasClaims, firstSlot))
	subCommaLbl := g.label()
	subNoCommaLbl := g.label()
	subAfterLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasClaims, subNoCommaLbl, subCommaLbl))
	g.line(fmt.Sprintf("%s:", subCommaLbl))
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString(",\"sub\":\""), 9))
	g.line(fmt.Sprintf("  br label %%%s", subAfterLbl))
	g.line(fmt.Sprintf("%s:", subNoCommaLbl))
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString("\"sub\":\""), 8))
	g.line(fmt.Sprintf("  br label %%%s", subAfterLbl))
	g.line(fmt.Sprintf("%s:", subAfterLbl))
	g.jwtStrcat(payloadPtr, "%subject")
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString("\",\"iat\":"), 9))
	g.jwtStrcat(payloadPtr, g.jwtIntToStr(nowReg))
	// exp only when expiresIn > 0 (Go parity).
	expPos := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i64 %%expiresIn, 0", expPos))
	appendExpLbl := g.label()
	skipExpLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", expPos, appendExpLbl, skipExpLbl))
	g.line(appendExpLbl + ":")
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString(",\"exp\":"), 8))
	g.jwtStrcat(payloadPtr, g.jwtIntToStr(expReg))
	g.line(fmt.Sprintf("  br label %%%s", skipExpLbl))
	g.line(skipExpLbl + ":")
	g.jwtStrcat(payloadPtr, g.ptrTo(g.addString("}"), 2))
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

// jwtStrcat emits `strcat(dst, src)` (appends src to dst).
func (g *Generator) jwtStrcat(dst, src string) {
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", dst, src))
}

// jwtIntToStr converts an i64 register to a NUL-terminated decimal string
// (snprintf into a 24-byte stack buffer).
func (g *Generator) jwtIntToStr(v string) string {
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [24 x i8], align 1", buf))
	bufPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [24 x i8], ptr %s, i64 0, i64 0", bufPtr, buf))
	fmtStr := g.addString("%lld")
	fmtPtr := g.ptrTo(fmtStr, 5)
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 24, ptr %s, i64 %s)", bufPtr, fmtPtr, v))
	return bufPtr
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
	// v0.6.6: with rem==1 (1 byte left) two chars must be written — c1 =
	// (b0 & 3) << 4. Previously only c0 was stored, leaving the second
	// output char as uninitialized heap garbage, so tokens whose payload
	// length was 1 mod 3 carried a corrupted (non-standard) payload segment.
	shb := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 3", shb, b0t))
	shc := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 4", shc, shb))
	g.emitStoreB64Char(out, o2, 1, table, shc)
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

// emitJwtVerifyBody emits the signature-verification body described above.
func (g *Generator) emitJwtVerifyBody() {
	g.enqueueStdlib("jwt", "b64url", "b64url", 0)
	g.enqueueStdlib("jwt", "hexdecode", "hexdecode", 0)
	g.enqueueStdlib("crypto", "HmacSha256", "HmacSha256", 0)
	g.needVariantRuntime = true // box_str / nilbox helpers

	g.line("define ptr @__kylix_jwt_JwtVerify(ptr %secret, ptr %token) {")
	g.line("entry:")
	// d1 = strchr(token, '.')
	d1 := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %%token, i32 46)", d1))
	d1Null := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", d1Null, d1))
	fail1Lbl := g.label()
	ok1Lbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", d1Null, fail1Lbl, ok1Lbl))
	g.line(fail1Lbl + ":")
	g.line("  ret ptr @__kylix_variant_nilbox")
	g.line(ok1Lbl + ":")
	// d2 = strchr(d1+1, '.')
	d1p1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", d1p1, d1))
	d2 := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 46)", d2, d1p1))
	d2Null := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", d2Null, d2))
	fail2Lbl := g.label()
	ok2Lbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", d2Null, fail2Lbl, ok2Lbl))
	g.line(fail2Lbl + ":")
	g.line("  ret ptr @__kylix_variant_nilbox")
	g.line(ok2Lbl + ":")

	// signing = token[0..d2) — the header.payload part (both '.' separators
	// are inside, d2 is the second '.'). Copy it + NUL; simpler and safer than
	// reconstructing header + "." + payload.
	d2Addr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", d2Addr, d2))
	tokAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %%token to i64", tokAddr))
	sl := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", sl, d2Addr, tokAddr))
	slen1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", slen1, sl))
	signing := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", signing, slen1))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %%token, i64 %s)", signing, sl))
	nulPos := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", nulPos, signing, sl))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", nulPos))

	// sig = d2+1 (rest of the token)
	sig := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", sig, d2))

	// expect = b64url(hexdecode(HmacSha256(secret, signing), 64), 32)
	hmacHex := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_HmacSha256(ptr %%secret, ptr %s)", hmacHex, signing))
	raw := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_hexdecode(ptr %s, i64 64)", raw, hmacHex))
	expect := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_b64url(ptr %s, i64 32)", expect, raw))

	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, sig, expect))
	ok := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", ok, cmp))
	okLbl := g.label()
	badLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ok, okLbl, badLbl))
	g.line(okLbl + ":")
	// payload = token[(d1+1) .. d2) — copy + NUL terminate.
	d1p1Addr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", d1p1Addr, d1p1))
	payloadLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", payloadLen, d2Addr, d1p1Addr))
	payloadPlus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", payloadPlus1, payloadLen))
	payloadBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", payloadBuf, payloadPlus1))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", payloadBuf, d1p1, payloadLen))
	payloadNul := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", payloadNul, payloadBuf, payloadLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", payloadNul))

	// decoded = b64urldecode(payload) → the claims JSON.
	g.enqueueStdlib("jwt", "b64urldecode", "b64urldecode", 0)
	decoded := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_jwt_b64url_decode(ptr %s, i64 %s)", decoded, payloadBuf, payloadLen))

	// claims = parse_flat(decoded) → htab whose values are Variant boxes.
	g.needHashtab = true
	g.enqueueStdlib("jsonutil", "JsonDecodeMap", "JsonDecodeMap", 0)
	claims := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_json_parse_flat(ptr %s)", claims, decoded))

	// exp check (Go parity): if claims["exp"] exists and exp < now → expired.
	expKey := g.ptrTo(g.addString("exp"), 4)
	expBox := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_htab_get_variant(ptr %s, ptr %s)", expBox, claims, expKey))
	nilbox := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { i32, i64 }, ptr @__kylix_variant_nilbox, i32 0, i32 0", nilbox))
	isNil := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, %s", isNil, expBox, nilbox))
	retClaimsLbl := g.label()
	chkExpLbl := g.label()
	expiredLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isNil, retClaimsLbl, chkExpLbl))
	g.line(chkExpLbl + ":")
	// exp is a number; use variant_as_int (handles int/float/str boxes — exp
	// decoded as a float box would as_double fine, but jsonutil boxes numbers
	// as their parsed representation; as_int's atoll/bitcast covers all).
	expI := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_variant_as_int(ptr %s)", expI, expBox))
	nowT := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @time(ptr null)", nowT))
	isExpired := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", isExpired, expI, nowT))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isExpired, expiredLbl, retClaimsLbl))
	g.line(expiredLbl + ":")
	g.line("  ret ptr @__kylix_variant_nilbox")
	g.line(retClaimsLbl + ":")
	// success: return the claims map as a Variant (box_map).
	retBox := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_box_map(ptr %s)", retBox, claims))
	g.line(fmt.Sprintf("  ret ptr %s", retBox))
	g.line(badLbl + ":")
	g.line("  ret ptr @__kylix_variant_nilbox")
	g.line("}")
	g.line("")
}

// ---- b64urldecode: ptr @__kylix_jwt_b64url_decode(ptr %str, i64 %n) ----
//
// base64url-decodes the first n bytes (RFC 4648 §5, no padding) into a
// malloc'd, NUL-terminated buffer. Uses the shared @__kylix_jwt_b64url_val
// nibble/value decoder. v0.6.6 (JWT payload → claims JSON).
func (g *Generator) emitJwtB64URLDecodeBody() {
	g.enqueueStdlib("jwt", "b64urlval", "b64urlval", 0)
	g.line("define ptr @__kylix_jwt_b64url_decode(ptr %str, i64 %n) {")
	g.line("entry:")
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %%n)", out)) // output <= n bytes
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	oSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", oSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", oSlot))
	loopLbl := g.label()
	quadLbl := g.label()
	tailLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	iPlus3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 3", iPlus3, curI))
	hasQuad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %%n", hasQuad, iPlus3))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasQuad, quadLbl, tailLbl))
	// quad → 3 bytes
	g.line(fmt.Sprintf("%s:", quadLbl))
	v0 := g.emitJwtB64ValLoad("%str", curI, 0)
	v1 := g.emitJwtB64ValLoad("%str", curI, 1)
	v2 := g.emitJwtB64ValLoad("%str", curI, 2)
	v3 := g.emitJwtB64ValLoad("%str", curI, 3)
	t0 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 18", t0, v0))
	t1 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 12", t1, v1))
	t2 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 6", t2, v2))
	lo12 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", lo12, t2, v3))
	lo18 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", lo18, t1, lo12))
	triple := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", triple, t0, lo18))
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
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 4", iNext, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	// tail: remaining 2 or 3 chars → 1 or 2 bytes
	g.line(fmt.Sprintf("%s:", tailLbl))
	curI2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI2, iSlot))
	rem := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %%n, %s", rem, curI2))
	remGe2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 2", remGe2, rem))
	rem2Lbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", remGe2, rem2Lbl, doneLbl))
	g.line(fmt.Sprintf("%s:", rem2Lbl))
	tv0 := g.emitJwtB64ValLoad("%str", curI2, 0)
	tv1 := g.emitJwtB64ValLoad("%str", curI2, 1)
	// b0 = (v0<<18 | v1<<12) >> 16
	ta0 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 18", ta0, tv0))
	ta1 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 12", ta1, tv1))
	ta2 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", ta2, ta0, ta1))
	tb0 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 16", tb0, ta2))
	tb0m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", tb0m, tb0))
	curO2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO2, oSlot))
	dp2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dp2, out, curO2))
	dp2v := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", dp2v, tb0m))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", dp2v, dp2))
	oPlus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", oPlus1, curO2))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", oPlus1, oSlot))
	remGe3 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 3", remGe3, rem))
	rem3Lbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", remGe3, rem3Lbl, doneLbl))
	g.line(fmt.Sprintf("%s:", rem3Lbl))
	tv2 := g.emitJwtB64ValLoad("%str", curI2, 2)
	// b1 = (v0<<18 | v1<<12 | v2<<6) >> 8
	ta3 := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 6", ta3, tv2))
	ta4 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", ta4, ta2, ta3))
	tb1 := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 8", tb1, ta4))
	tb1m := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", tb1m, tb1))
	curO3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO3, oSlot))
	dp3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dp3, out, curO3))
	dp3v := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", dp3v, tb1m))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", dp3v, dp3))
	oPlus2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", oPlus2, curO3))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", oPlus2, oSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	finalO := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", finalO, oSlot))
	g.emitStoreZero(out, finalO)
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// emitJwtB64ValLoad returns the 6-bit value of b64url char at src[curI+off].
func (g *Generator) emitJwtB64ValLoad(src, curI string, off int) string {
	base := curI
	if off > 0 {
		base = g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %d", base, curI, off))
	}
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", p, src, base))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, p))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_jwt_b64url_val(i8 %s)", r, c))
	return r
}

// emitJwtB64URLValBody — i64 @__kylix_jwt_b64url_val(i8 %c). Maps a base64url
// char to 0..63, -1 for invalid. URL alphabet: '-' → 62, '_' → 63.
func (g *Generator) emitJwtB64URLValBody() {
	g.line("define i64 @__kylix_jwt_b64url_val(i8 %c) {")
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
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_digit, label %%check_dash", isDigit))
	g.line("ret_digit:")
	g.line(fmt.Sprintf("  ret i64 %s", adjD))
	g.line("check_dash:")
	isDash := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 45", isDash, cI64)) // '-'
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_dash, label %%check_under", isDash))
	g.line("ret_dash:")
	g.line("  ret i64 62")
	g.line("check_under:")
	isUnder := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 95", isUnder, cI64)) // '_'
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_under, label %%ret_neg", isUnder))
	g.line("ret_under:")
	g.line("  ret i64 63")
	g.line("ret_neg:")
	g.line("  ret i64 -1")
	g.line("}")
	g.line("")
}
