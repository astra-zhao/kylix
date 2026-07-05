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
