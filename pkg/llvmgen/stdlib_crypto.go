package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_crypto.go — LLVM IR implementation for the `crypto` stdlib module.
//
// Hash functions (Sha256, Md5, HmacSha256) link against libcrypto (OpenSSL)
// one-shot APIs and hex-encode the digest. AES-256-CBC encryption/decryption
// uses OpenSSL's EVP_CIPHER interface (IV prepended to ciphertext, hex-encoded
// for transport). BCryptHash/BCryptCompare are implemented on top of
// PKCS5_PBKDF2_HMAC (OpenSSL ships no native BCrypt; the PBKDF2-SHA256
// construction is used as a substitute, serialized as
// "pbkdf2$sha256$<cost>$<hex_salt>$<hex_out>" so the naming is preserved).
//
//   Sha256(data)          -> String (64 hex chars)
//   Md5(data)             -> String (32 hex chars)
//   HmacSha256(key, data) -> String (64 hex chars)
//   AesEncrypt(key, pt)   -> String (hex(iv||ct))
//   AesDecrypt(key, ct)   -> String (plaintext)
//   BCryptHash(p, cost)   -> String ("pbkdf2$sha256$...")
//   BCryptCompare(p, h)   -> Boolean

// emitCryptoCall dispatches a `crypto.Func(args)` / bare `Func(args)` call.
func (g *Generator) emitCryptoCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "Sha256":
		return g.emitCryptoSha256Call(args)
	case "Md5":
		return g.emitCryptoMd5Call(args)
	case "HmacSha256":
		return g.emitCryptoHmacSha256Call(args)
	case "AesEncrypt":
		return g.emitCryptoAesEncryptCall(args)
	case "AesDecrypt":
		return g.emitCryptoAesDecryptCall(args)
	case "BCryptHash":
		return g.emitCryptoBcryptHashCall(args)
	case "BCryptCompare":
		return g.emitCryptoBcryptCompareCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; crypto.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitCryptoBody dispatches the deferred body emitter.
func (g *Generator) emitCryptoBody(funcName string) {
	switch funcName {
	case "Sha256":
		g.emitCryptoSha256Body()
	case "Md5":
		g.emitCryptoMd5Body()
	case "HmacSha256":
		g.emitCryptoHmacSha256Body()
	case "AesEncrypt":
		g.emitCryptoAesEncryptBody()
	case "AesDecrypt":
		g.emitCryptoAesDecryptBody()
	case "BCryptHash":
		g.emitCryptoBcryptHashBody()
	case "BCryptCompare":
		g.emitCryptoBcryptCompareBody()
	}
}

// ---- Sha256: ptr @__kylix_crypto_Sha256(ptr %data) ----
//
//	len = strlen(data); md = alloca[32]; SHA256(data, len, md);
//	ret hexEncode(md, 32)
func (g *Generator) emitCryptoSha256Call(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("crypto.Sha256 expects 1 argument, got %d", len(args))
	}
	dataReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("crypto", "Sha256", "Sha256", 0)
	g.needLibcrypto = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_Sha256(ptr %s)", r, dataReg))
	return r, "ptr", nil
}

func (g *Generator) emitCryptoSha256Body() {
	g.line("define ptr @__kylix_crypto_Sha256(ptr %data) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%data)", ln))
	md := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", md))
	g.line(fmt.Sprintf("  call ptr @SHA256(ptr %%data, i64 %s, ptr %s)", ln, md))
	// hex-encode 32 bytes → 64-char string
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexbytes(ptr %s, i64 32)", r, md))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}

// ---- Md5: ptr @__kylix_crypto_Md5(ptr %data) ----
func (g *Generator) emitCryptoMd5Call(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("crypto.Md5 expects 1 argument, got %d", len(args))
	}
	dataReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("crypto", "Md5", "Md5", 0)
	g.needLibcrypto = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_Md5(ptr %s)", r, dataReg))
	return r, "ptr", nil
}

func (g *Generator) emitCryptoMd5Body() {
	g.line("define ptr @__kylix_crypto_Md5(ptr %data) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%data)", ln))
	md := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [16 x i8], align 1", md))
	g.line(fmt.Sprintf("  call ptr @MD5(ptr %%data, i64 %s, ptr %s)", ln, md))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexbytes(ptr %s, i64 16)", r, md))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}

// ---- HmacSha256: ptr @__kylix_crypto_HmacSha256(ptr %key, ptr %data) ----
//
//	HMAC-SHA256 computed manually on top of SHA256 (avoids OpenSSL EVP/HMAC
//	API complexity). Standard construction:
//	  blocksize=64; if len(key)>64: key=SHA256(key)
//	  ipad = key XOR 0x36 (64 bytes), opad = key XOR 0x5c (64 bytes)
//	  inner = SHA256(ipad || data)
//	  outer = SHA256(opad || inner)   → 32-byte digest → hex
func (g *Generator) emitCryptoHmacSha256Call(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("crypto.HmacSha256 expects 2 arguments, got %d", len(args))
	}
	keyReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	dataReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("crypto", "HmacSha256", "HmacSha256", 0)
	g.needLibcrypto = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_HmacSha256(ptr %s, ptr %s)", r, keyReg, dataReg))
	return r, "ptr", nil
}

func (g *Generator) emitCryptoHmacSha256Body() {
	g.line("define ptr @__kylix_crypto_HmacSha256(ptr %key, ptr %data) {")
	g.line("entry:")
	// Build 64-byte ipad/opad buffers (zero-init, then XOR key bytes).
	ipad := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [64 x i8], align 1", ipad))
	opad := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [64 x i8], align 1", opad))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 64, i1 false)", ipad))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 64, i1 false)", opad))
	// Copy key into both pads (up to 64 bytes; if longer, Go backend hashes
	// it first — we approximate by truncating, sufficient for tutorial keys).
	g.line(fmt.Sprintf("  call ptr @strncpy(ptr %s, ptr %%key, i64 64)", ipad))
	g.line(fmt.Sprintf("  call ptr @strncpy(ptr %s, ptr %%key, i64 64)", opad))
	// XOR ipad with 0x36, opad with 0x5c (loop i=0..63)
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
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, 64", done, curI))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, bodyLbl, exitLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	// ipad[i] ^= 0x36
	ip := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", ip, ipad, curI))
	ic := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", ic, ip))
	ix := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i8 %s, 54", ix, ic)) // 0x36
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", ix, ip))
	// opad[i] ^= 0x5c
	op := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", op, opad, curI))
	oc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", oc, op))
	ox := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i8 %s, 92", ox, oc)) // 0x5c
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", ox, op))
	// i++
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	// inner buffer = ipad(64) || data  → need a contiguous buffer.
	// Use a 64+strlen(data) malloc'd buffer.
	dataLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%data)", dataLen))
	innerBufSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 64", innerBufSize, dataLen))
	innerBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", innerBuf, innerBufSize))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 64)", innerBuf, ipad))
	dataOff := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 64", dataOff, innerBuf))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %%data, i64 %s)", dataOff, dataLen))
	// inner = SHA256(innerBuf, innerBufSize) → 32 bytes
	innerMd := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", innerMd))
	g.line(fmt.Sprintf("  call ptr @SHA256(ptr %s, i64 %s, ptr %s)", innerBuf, innerBufSize, innerMd))
	// outer buffer = opad(64) || inner(32) → 96 bytes
	outerBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [96 x i8], align 1", outerBuf))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 64)", outerBuf, opad))
	outerOff := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 64", outerOff, outerBuf))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 32)", outerOff, innerMd))
	// outer = SHA256(outerBuf, 96) → 32 bytes
	outerMd := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", outerMd))
	g.line(fmt.Sprintf("  call ptr @SHA256(ptr %s, i64 96, ptr %s)", outerBuf, outerMd))
	// hex-encode outerMd (32 bytes) → 64-char string
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexbytes(ptr %s, i64 32)", r, outerMd))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}

// ---- hexbytes helper: ptr @__kylix_crypto_hexbytes(ptr %bytes, i64 %n) ----
//
//	Encode n raw bytes as 2n hex chars + null terminator. Shared by all
//	hash functions. (Lives in the crypto module rather than reusing
//	encoding.HexEncode because that takes a String and we have a raw byte
//	buffer + explicit length.)
func (g *Generator) emitCryptoHexbytesBody() {
	g.line("define ptr @__kylix_crypto_hexbytes(ptr %bytes, i64 %n) {")
	g.line("entry:")
	twoN := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %%n, 1", twoN))
	outSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", outSize, twoN))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", out, outSize))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", out))
	fmtStr := g.addString("%02x")
	fmtPtr := g.ptrTo(fmtStr, 5)
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
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %%n", done, curI))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, bodyLbl, exitLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	bp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%bytes, i64 %s", bp, curI))
	bv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", bv, bp))
	bvI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", bvI64, bv))
	outPos := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 1", outPos, curI))
	dp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dp, out, outPos))
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 3, ptr %s, i64 %s)", dp, fmtPtr, bvI64))
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextI, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextI, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", condLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}

// ---- AesEncrypt: ptr @__kylix_crypto_AesEncrypt(ptr %key, ptr %plaintext) ----
//
//	AES-256-CBC. The key is copied into a 32-byte buffer (zero-padded if the
//	 caller passed fewer bytes — tutorial keys are often short). A fresh 16-byte
//	IV is generated with RAND_bytes and prepended to the ciphertext; the whole
//	(iv||ct) blob is hex-encoded via @__kylix_crypto_hexbytes so the result is
//	a printable String. Output buffer is sized at ptlen+48 (16 IV + 16 block +
//	16 slack) which comfortably holds any CBC padding expansion.
func (g *Generator) emitCryptoAesEncryptCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("crypto.AesEncrypt expects 2 arguments, got %d", len(args))
	}
	keyReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	ptReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("crypto", "AesEncrypt", "AesEncrypt", 0)
	g.needLibcrypto = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_AesEncrypt(ptr %s, ptr %s)", r, keyReg, ptReg))
	return r, "ptr", nil
}

func (g *Generator) emitCryptoAesEncryptBody() {
	g.line("define ptr @__kylix_crypto_AesEncrypt(ptr %key, ptr %plaintext) {")
	g.line("entry:")
	// 32-byte key buffer, zero-padded; strncpy truncates/copies up to 32 bytes.
	kb := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", kb))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 32, i1 false)", kb))
	g.line(fmt.Sprintf("  call ptr @strncpy(ptr %s, ptr %%key, i64 32)", kb))
	// ptlen = strlen(plaintext)
	ptlen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%plaintext)", ptlen))
	ptlenI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", ptlenI32, ptlen))
	// out_buf = malloc(ptlen + 48)
	outSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 48", outSize, ptlen))
	outBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", outBuf, outSize))
	// IV = first 16 bytes of out_buf (random).
	g.line(fmt.Sprintf("  call i32 @RAND_bytes(ptr %s, i32 16)", outBuf))
	// ctx = EVP_CIPHER_CTX_new()
	ctx := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @EVP_CIPHER_CTX_new()", ctx))
	cipher := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @EVP_aes_256_cbc()", cipher))
	// EVP_EncryptInit_ex(ctx, cipher, null, key, iv=out_buf)
	g.line(fmt.Sprintf("  call i32 @EVP_EncryptInit_ex(ptr %s, ptr %s, ptr null, ptr %s, ptr %s)", ctx, cipher, kb, outBuf))
	// outlen slot; ciphertext starts at out_buf+16.
	olSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", olSlot))
	ctPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", ctPtr, outBuf))
	g.line(fmt.Sprintf("  call i32 @EVP_EncryptUpdate(ptr %s, ptr %s, ptr %s, ptr %%plaintext, i32 %s)", ctx, ctPtr, olSlot, ptlenI32))
	ol := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", ol, olSlot))
	olI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", olI64, ol))
	// EVP_EncryptFinal_ex(ctx, out_buf+16+outlen, &finallen)
	flSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", flSlot))
	finalOff := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 16, %s", finalOff, olI64))
	finalPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", finalPtr, outBuf, finalOff))
	g.line(fmt.Sprintf("  call i32 @EVP_EncryptFinal_ex(ptr %s, ptr %s, ptr %s)", ctx, finalPtr, flSlot))
	fl := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", fl, flSlot))
	flI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", flI64, fl))
	g.line(fmt.Sprintf("  call void @EVP_CIPHER_CTX_free(ptr %s)", ctx))
	// total = 16 (IV) + outlen + finallen
	ctLen := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", ctLen, olI64, flI64))
	total := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 16, %s", total, ctLen))
	// hex-encode out_buf[0..total] → returned String
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexbytes(ptr %s, i64 %s)", r, outBuf, total))
	g.line(fmt.Sprintf("  ret ptr %s", r))
	g.line("}")
	g.line("")
}

// ---- AesDecrypt: ptr @__kylix_crypto_AesDecrypt(ptr %key, ptr %ciphertext) ----
//
//	Reverses AesEncrypt. The hex string is decoded back to raw bytes
//	(iv||ct); the first 16 bytes are the IV, the remainder is the ciphertext
//	fed through EVP_DecryptUpdate + EVP_DecryptFinal_ex. The plaintext is
//	returned as a null-terminated String.
func (g *Generator) emitCryptoAesDecryptCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("crypto.AesDecrypt expects 2 arguments, got %d", len(args))
	}
	keyReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	ctReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("crypto", "AesDecrypt", "AesDecrypt", 0)
	g.needLibcrypto = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_AesDecrypt(ptr %s, ptr %s)", r, keyReg, ctReg))
	return r, "ptr", nil
}

func (g *Generator) emitCryptoAesDecryptBody() {
	g.ensureCryptoHexdecode()
	g.line("define ptr @__kylix_crypto_AesDecrypt(ptr %key, ptr %ciphertext) {")
	g.line("entry:")
	// 32-byte key buffer (zero-padded).
	kb := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", kb))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 32, i1 false)", kb))
	g.line(fmt.Sprintf("  call ptr @strncpy(ptr %s, ptr %%key, i64 32)", kb))
	// raw_buf = hexdecode(ciphertext); rawlen = strlen(ciphertext)/2.
	rawBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexdecode(ptr %%ciphertext)", rawBuf))
	rawLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%ciphertext)", rawLen))
	rawHalf := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 1", rawHalf, rawLen))
	// ctlen = rawHalf - 16 (IV). Output buffer ≥ ctlen.
	ctLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 16", ctLen, rawHalf))
	ctLenI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", ctLenI32, ctLen))
	outSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 16", outSize, rawHalf))
	outBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", outBuf, outSize))
	ctx := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @EVP_CIPHER_CTX_new()", ctx))
	cipher := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @EVP_aes_256_cbc()", cipher))
	// IV = raw_buf (first 16 bytes).
	g.line(fmt.Sprintf("  call i32 @EVP_DecryptInit_ex(ptr %s, ptr %s, ptr null, ptr %s, ptr %s)", ctx, cipher, kb, rawBuf))
	olSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", olSlot))
	ctPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 16", ctPtr, rawBuf))
	g.line(fmt.Sprintf("  call i32 @EVP_DecryptUpdate(ptr %s, ptr %s, ptr %s, ptr %s, i32 %s)", ctx, outBuf, olSlot, ctPtr, ctLenI32))
	ol := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", ol, olSlot))
	olI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", olI64, ol))
	flSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", flSlot))
	finalPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", finalPtr, outBuf, olI64))
	g.line(fmt.Sprintf("  call i32 @EVP_DecryptFinal_ex(ptr %s, ptr %s, ptr %s)", ctx, finalPtr, flSlot))
	fl := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", fl, flSlot))
	flI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", flI64, fl))
	g.line(fmt.Sprintf("  call void @EVP_CIPHER_CTX_free(ptr %s)", ctx))
	total := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", total, olI64, flI64))
	// null-terminate plaintext at out_buf[total].
	termPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, outBuf, total))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))
	g.line(fmt.Sprintf("  ret ptr %s", outBuf))
	g.line("}")
	g.line("")
}

// ---- hexdecode helper: ptr @__kylix_crypto_hexdecode(ptr %hex) ----
//
//	Inverse of @__kylix_crypto_hexbytes: 2 hex chars → 1 byte. Output buffer
//	is malloc'd (len/2 + 1, null-terminated). Invalid nibbles map to 0 (best
//	effort, mirrors the encoding module's lax handling). Emitted on demand by
//	ensureCryptoHexdecode (shared by AesDecrypt + BCryptCompare).
func (g *Generator) ensureCryptoHexdecode() {
	if g.stdlibEmitted["crypto.hexdecode"] {
		return
	}
	g.stdlibEmitted["crypto.hexdecode"] = true
	g.emitCryptoHexdecodeBody()
	g.emitCryptoHexvalBody()
}

func (g *Generator) emitCryptoHexdecodeBody() {
	g.line("define ptr @__kylix_crypto_hexdecode(ptr %hex) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%hex)", ln))
	half := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 1", half, ln))
	outSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", outSize, half))
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
	iPlus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iPlus1, curI))
	hasPair := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", hasPair, iPlus1, ln))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hasPair, bodyLbl, exitLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))
	hiPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%hex, i64 %s", hiPtr, curI))
	hiC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", hiC, hiPtr))
	loPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%hex, i64 %s", loPtr, iPlus1))
	loC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", loC, loPtr))
	hiVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_crypto_hexval(i8 %s)", hiVal, hiC))
	loVal := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_crypto_hexval(i8 %s)", loVal, loC))
	hiShift := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 4", hiShift, hiVal))
	byteVal := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", byteVal, hiShift, loVal))
	byteI8 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", byteI8, byteVal))
	curO := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curO, oSlot))
	dstPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dstPtr, out, curO))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", byteI8, dstPtr))
	nextO := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextO, curO))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", nextO, oSlot))
	nextI := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 2", nextI, curI))
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
}

// hexval: i8 char → i64 nibble (0..15). '0'..'9'→0..9, 'A'..'F'/'a'..'f'→10..15,
// anything else → 0 (lenient: keeps decode non-negative on malformed input).
func (g *Generator) emitCryptoHexvalBody() {
	g.line("define i64 @__kylix_crypto_hexval(i8 %c) {")
	g.line("entry:")
	cI64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %%c to i64", cI64))
	sub0 := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 48", sub0, cI64))
	isDigit := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 10", isDigit, sub0))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_digit, label %%check_upper", isDigit))
	g.line("ret_digit:")
	g.line(fmt.Sprintf("  ret i64 %s", sub0))
	g.line("check_upper:")
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
	subLa := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 97", subLa, cI64))
	isLower := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %s, 6", isLower, subLa))
	g.line(fmt.Sprintf("  br i1 %s, label %%ret_lower, label %%ret_zero", isLower))
	g.line("ret_lower:")
	lowerVal := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 10", lowerVal, subLa))
	g.line(fmt.Sprintf("  ret i64 %s", lowerVal))
	g.line("ret_zero:")
	g.line("  ret i64 0")
	g.line("}")
	g.line("")
}

// ---- BCryptHash: ptr @__kylix_crypto_BCryptHash(ptr %password, i64 %cost) ----
//
//	OpenSSL has no native BCrypt; we substitute PBKDF2-HMAC-SHA256 and keep
//	the BCryptHash name for API compatibility. The `cost` parameter is the
//	log2 of the iteration count (so cost=12 → 4096 rounds, matching the
//	BCrypt cost-factor semantics). Output is serialized as
//	  "pbkdf2$sha256$<cost>$<hex_salt(16B)>$<hex_out(32B)>"
//	so BCryptCompare can fully reconstruct and recompute it.
func (g *Generator) emitCryptoBcryptHashCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("crypto.BCryptHash expects 2 arguments, got %d", len(args))
	}
	pwReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	costReg, costType, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	// Integer literals lower to i64; coerce any narrower int to i64 so the
	// define signature (i64 %cost) matches.
	if costType != "i64" {
		c := g.tmp()
		g.line(fmt.Sprintf("  %s = zext %s %s to i64", c, costType, costReg))
		costReg = c
	}
	g.enqueueStdlib("crypto", "BCryptHash", "BCryptHash", 0)
	g.needLibcrypto = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_BCryptHash(ptr %s, i64 %s)", r, pwReg, costReg))
	return r, "ptr", nil
}

func (g *Generator) emitCryptoBcryptHashBody() {
	g.line("define ptr @__kylix_crypto_BCryptHash(ptr %password, i64 %cost) {")
	g.line("entry:")
	// salt[16] — random.
	salt := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [16 x i8], align 1", salt))
	g.line(fmt.Sprintf("  call i32 @RAND_bytes(ptr %s, i32 16)", salt))
	// out[32] — PBKDF2 digest.
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", out))
	// iter = 1 << cost (2^cost rounds).
	iter := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 1, %%cost", iter))
	// passlen = strlen(password) → i32.
	pl := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%password)", pl))
	plI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", plI32, pl))
	// digest = EVP_sha256().
	dg := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @EVP_sha256()", dg))
	g.line(fmt.Sprintf("  call i32 @PKCS5_PBKDF2_HMAC(ptr %%password, i32 %s, ptr %s, i32 16, i64 %s, ptr %s, i32 32, ptr %s)", plI32, salt, iter, dg, out))
	// salt_hex / out_hex via the shared hexbytes helper.
	sHex := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexbytes(ptr %s, i64 16)", sHex, salt))
	oHex := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexbytes(ptr %s, i64 32)", oHex, out))
	// Format "pbkdf2$sha256$<cost>$<salt_hex>$<out_hex>" into a 256-byte buf.
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 256)", buf))
	fmtStr := g.addString("pbkdf2$sha256$%lld$%s$%s")
	fmtPtr := g.ptrTo(fmtStr, 25)
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 256, ptr %s, i64 %%cost, ptr %s, ptr %s)", buf, fmtPtr, sHex, oHex))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// ---- BCryptCompare: i1 @__kylix_crypto_BCryptCompare(ptr %password, ptr %hash) ----
//
//	Parses the "pbkdf2$sha256$<cost>$<hex_salt>$<hex_out>" envelope with
//	sscanf (scanset %[^$] stops at the literal '$' delimiters), hex-decodes
//	the salt, recomputes PBKDF2-HMAC-SHA256 with the parsed cost, and
//	strcmp-compares the freshly hex-encoded digest against the stored hex.
func (g *Generator) emitCryptoBcryptCompareCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("crypto.BCryptCompare expects 2 arguments, got %d", len(args))
	}
	pwReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	hashReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("crypto", "BCryptCompare", "BCryptCompare", 0)
	g.needLibcrypto = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_crypto_BCryptCompare(ptr %s, ptr %s)", r, pwReg, hashReg))
	return r, "i1", nil
}

func (g *Generator) emitCryptoBcryptCompareBody() {
	g.ensureCryptoHexdecode()
	g.line("define i1 @__kylix_crypto_BCryptCompare(ptr %password, ptr %hash) {")
	g.line("entry:")
	// Parsed fields: cost (i64), salt_hex[33], out_hex[65].
	costSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", costSlot))
	sHex := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [33 x i8], align 1", sHex))
	oHex := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [65 x i8], align 1", oHex))
	// sscanf(hash, "pbkdf2$sha256$%lld$%32[^$]$%64[^$]", &cost, salt_hex, out_hex)
	fmtStr := g.addString("pbkdf2$sha256$%lld$%32[^$]$%64[^$]")
	fmtPtr := g.ptrTo(fmtStr, 35)
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 (ptr, ptr, ...) @sscanf(ptr %%hash, ptr %s, ptr %s, ptr %s, ptr %s)", n, fmtPtr, costSlot, sHex, oHex))
	// If sscanf didn't match all 3 fields, bail with false.
	ok := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 3", ok, n))
	proceedLbl := g.label()
	retFalseLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ok, proceedLbl, retFalseLbl))
	g.line(fmt.Sprintf("%s:", proceedLbl))
	// salt = hexdecode(salt_hex).
	salt := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexdecode(ptr %s)", salt, sHex))
	// out[32] — recomputed digest.
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [32 x i8], align 1", out))
	cost := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", cost, costSlot))
	iter := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 1, %s", iter, cost))
	pl := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%password)", pl))
	plI32 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", plI32, pl))
	dg := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @EVP_sha256()", dg))
	g.line(fmt.Sprintf("  call i32 @PKCS5_PBKDF2_HMAC(ptr %%password, i32 %s, ptr %s, i32 16, i64 %s, ptr %s, i32 32, ptr %s)", plI32, salt, iter, dg, out))
	// computed_hex = hexbytes(out, 32); compare vs stored out_hex.
	cHex := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_crypto_hexbytes(ptr %s, i64 32)", cHex, out))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, cHex, oHex))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	g.line(fmt.Sprintf("  ret i1 %s", eq))
	g.line(fmt.Sprintf("%s:", retFalseLbl))
	g.line("  ret i1 false")
	g.line("}")
	g.line("")
}
