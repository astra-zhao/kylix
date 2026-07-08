package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_crypto tests — verify hash functions lower to OpenSSL-backed
// defines and hex-encode the digest.

func TestCrypto_Sha256CallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := Sha256('hello');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_crypto_Sha256")
	assertIRContains(t, ir, "call ptr @SHA256")
	if strings.Contains(ir, "crypto.Sha256 not implemented") {
		t.Errorf("Sha256 still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestCrypto_Sha256BodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := Sha256('hello');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_crypto_Sha256(ptr %data)")
	assertIRContains(t, ir, "call ptr @__kylix_crypto_hexbytes")
}

func TestCrypto_Md5CallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := Md5('abc');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_crypto_Md5")
	assertIRContains(t, ir, "call ptr @MD5")
}

func TestCrypto_HmacSha256CallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := HmacSha256('key', 'data');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_crypto_HmacSha256")
	assertIRContains(t, ir, "define ptr @__kylix_crypto_HmacSha256(ptr %key, ptr %data)")
}

func TestCrypto_HmacSha256UsesSHA256(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := HmacSha256('key', 'data');
end.`)
	// HMAC body calls SHA256 twice (inner + outer)
	if strings.Count(ir, "call ptr @SHA256") < 2 {
		t.Errorf("HmacSha256 should call SHA256 at least twice (inner+outer), got %d\nIR:\n%s", strings.Count(ir, "call ptr @SHA256"), ir)
	}
}

func TestCrypto_HexbytesBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := Sha256('hello');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_crypto_hexbytes(ptr %bytes, i64 %n)")
}

func TestCrypto_BCryptStub(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := BCryptHash('test', 4);
  var ok := BCryptCompare('test', h);
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_crypto_BCryptHash")
	assertIRContains(t, ir, "call i1 @__kylix_crypto_BCryptCompare")
}

// ===== AES-256-CBC real implementation (EVP_CIPHER API) =====

func TestCrypto_AesEncryptReal(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var ct := AesEncrypt('0123456789abcdef0123456789abcdef', 'hello world');
end.`)
	// Call site passes both key + plaintext through (no longer empty-string stub).
	assertIRContains(t, ir, "call ptr @__kylix_crypto_AesEncrypt")
	assertIRContains(t, ir, "define ptr @__kylix_crypto_AesEncrypt(ptr %key, ptr %plaintext)")
	// Real EVP_CIPHER flow.
	assertIRContains(t, ir, "call ptr @EVP_CIPHER_CTX_new")
	assertIRContains(t, ir, "call ptr @EVP_aes_256_cbc")
	assertIRContains(t, ir, "call i32 @EVP_EncryptInit_ex")
	assertIRContains(t, ir, "call i32 @EVP_EncryptUpdate")
	assertIRContains(t, ir, "call i32 @EVP_EncryptFinal_ex")
	assertIRContains(t, ir, "call void @EVP_CIPHER_CTX_free")
	// Random IV generation + hex-encoded output.
	assertIRContains(t, ir, "call i32 @RAND_bytes")
	assertIRContains(t, ir, "call ptr @__kylix_crypto_hexbytes")
	if strings.Contains(ir, "AesEncrypt not implemented") {
		t.Errorf("AesEncrypt still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestCrypto_AesDecryptReal(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var pt := AesDecrypt('0123456789abcdef0123456789abcdef', 'deadbeefcafebabe');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_crypto_AesDecrypt")
	assertIRContains(t, ir, "define ptr @__kylix_crypto_AesDecrypt(ptr %key, ptr %ciphertext)")
	// Real EVP_CIPHER decrypt flow.
	assertIRContains(t, ir, "call i32 @EVP_DecryptInit_ex")
	assertIRContains(t, ir, "call i32 @EVP_DecryptUpdate")
	assertIRContains(t, ir, "call i32 @EVP_DecryptFinal_ex")
	// Hex-decode of the ciphertext envelope (shared helper).
	assertIRContains(t, ir, "call ptr @__kylix_crypto_hexdecode")
	assertIRContains(t, ir, "define ptr @__kylix_crypto_hexdecode(ptr %hex)")
}

func TestCrypto_AesHexdecodeHelperEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var pt := AesDecrypt('key', 'aabbccdd');
end.`)
	// hexdecode + hexval helpers co-emitted on demand.
	assertIRContains(t, ir, "define i64 @__kylix_crypto_hexval(i8 %c)")
}

// ===== BCrypt PBKDF2-SHA256 real implementation =====

func TestCrypto_BCryptHashReal(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := BCryptHash('password', 10);
end.`)
	// Call site passes password + cost (i64) through.
	assertIRContains(t, ir, "call ptr @__kylix_crypto_BCryptHash")
	assertIRContains(t, ir, "define ptr @__kylix_crypto_BCryptHash(ptr %password, i64 %cost)")
	// PBKDF2 + SHA256 digest + salt generation.
	assertIRContains(t, ir, "call i32 @PKCS5_PBKDF2_HMAC")
	assertIRContains(t, ir, "call ptr @EVP_sha256")
	assertIRContains(t, ir, "call i32 @RAND_bytes")
	// Serialized as "pbkdf2$sha256$<cost>$<salt>$<out>" via snprintf.
	assertIRContains(t, ir, "call i32 (ptr, i64, ptr, ...) @snprintf")
	assertIRContains(t, ir, `c"pbkdf2$sha256$%lld$%s$%s`)
	// iter = 1 << cost (shl).
	assertIRContains(t, ir, "shl i64 1,")
}

func TestCrypto_BCryptCompareReal(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := BCryptHash('password', 8);
  var ok := BCryptCompare('password', h);
end.`)
	// Call site passes password + hash through (no longer zero-arg stub).
	assertIRContains(t, ir, "call i1 @__kylix_crypto_BCryptCompare")
	assertIRContains(t, ir, "define i1 @__kylix_crypto_BCryptCompare(ptr %password, ptr %hash)")
	// Parses the envelope with sscanf (scanset %[^$]).
	assertIRContains(t, ir, "call i32 (ptr, ptr, ...) @sscanf")
	assertIRContains(t, ir, `c"pbkdf2$sha256$%lld$%32[^$]$%64[^$]`)
	// Recomputes PBKDF2 and compares hex digests with strcmp.
	assertIRContains(t, ir, "call i32 @PKCS5_PBKDF2_HMAC")
	assertIRContains(t, ir, "call i32 @strcmp")
}

func TestCrypto_OpenSSLDeclarations(t *testing.T) {
	ir := generateIR(t, `program p;
uses crypto;
begin
  var h := Sha256('hello');
end.`)
	assertIRContains(t, ir, "declare ptr @SHA256")
	assertIRContains(t, ir, "declare ptr @MD5")
	assertIRContains(t, ir, "declare ptr @strncpy")
}

func TestCrypto_NotUsedNoSymbols(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	// Only check that no crypto function BODIES are emitted — the libc/libcrypto
	// `declare` lines are always present unconditionally (same as fopen etc.),
	// which is harmless (unused declares don't trigger linking).
	if strings.Contains(ir, "@__kylix_crypto_") {
		t.Errorf("crypto function body emitted without `uses crypto`\nIR:\n%s", ir)
	}
}
