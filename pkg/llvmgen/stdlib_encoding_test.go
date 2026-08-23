package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_encoding tests — verify the IR generation for HexEncode/HexDecode/
// Base64Encode/Base64Decode lowers to libc-backed defines (not stubs) and
// emits the expected libc calls.

func TestEncoding_HexEncodeCallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := HexEncode('AB');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_encoding_HexEncode")
	if strings.Contains(ir, "encoding.HexEncode not implemented") {
		t.Errorf("HexEncode still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestEncoding_HexEncodeBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := HexEncode('AB');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_encoding_HexEncode(ptr %str)")
	assertIRContains(t, ir, "call i64 @strlen")
	assertIRContains(t, ir, "call i32 (ptr, i64, ptr, ...) @snprintf")
}

func TestEncoding_HexDecodeBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := HexDecode('4142');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_encoding_HexDecode")
	assertIRContains(t, ir, "define ptr @__kylix_encoding_HexDecode(ptr %str)")
	// hexval helper is emitted alongside HexDecode
	assertIRContains(t, ir, "define i64 @__kylix_encoding_hexval(i8 %c)")
}

func TestEncoding_Base64EncodeCallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := Base64Encode('hello');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_encoding_Base64Encode")
	if strings.Contains(ir, "encoding.Base64Encode not implemented") {
		t.Errorf("Base64Encode still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestEncoding_Base64EncodeBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := Base64Encode('hello');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_encoding_Base64Encode(ptr %str)")
	// base64 alphabet table constant
	assertIRContains(t, ir, "@__kylix_b64_table = private unnamed_addr constant [64 x i8]")
}

func TestEncoding_Base64DecodeBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := Base64Decode('aGVsbG8=');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_encoding_Base64Decode")
	assertIRContains(t, ir, "define ptr @__kylix_encoding_Base64Decode(ptr %str)")
	// b64val helper is emitted alongside Base64Decode
	assertIRContains(t, ir, "define i64 @__kylix_encoding_b64val(i8 %c)")
}

func TestEncoding_BodyDedup(t *testing.T) {
	// Two HexEncode calls must emit the define exactly once.
	ir := generateIR(t, `program p;
uses encoding;
begin
  var a := HexEncode('x');
  var b := HexEncode('y');
end.`)
	if got := strings.Count(ir, "define ptr @__kylix_encoding_HexEncode"); got != 1 {
		t.Errorf("HexEncode define should appear once, got %d\nIR:\n%s", got, ir)
	}
}

func TestEncoding_NotUsedNoBodies(t *testing.T) {
	// A program that does NOT `uses encoding` should not emit encoding symbols.
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "@__kylix_encoding_") {
		t.Errorf("encoding symbol emitted without `uses encoding`\nIR:\n%s", ir)
	}
}

func TestEncoding_UrlEncodeDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := UrlEncode('a b&c');
end.`)
	// Real call + define (not the not-implemented stub); format string must
	// escape the literal '%' (%%%02X) so encoded bytes are prefixed with '%'.
	assertIRContains(t, ir, "call ptr @__kylix_encoding_UrlEncode")
	assertIRContains(t, ir, "define ptr @__kylix_encoding_UrlEncode(ptr %s)")
	assertIRContains(t, ir, "%%%02X")
}

func TestEncoding_UrlDecodeDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := UrlDecode('a%20b+c');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_encoding_UrlDecode")
	assertIRContains(t, ir, "define ptr @__kylix_encoding_UrlDecode(ptr %s)")
	// Reuses the shared hexval nibble decoder.
	assertIRContains(t, ir, "@__kylix_encoding_hexval")
}

func TestEncoding_UrlDecodeWithoutHexDecodeEmitsHexval(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := UrlDecode('a%20b');
end.`)
	// hexval must be defined even though HexDecode itself is never called.
	assertIRContains(t, ir, "define i64 @__kylix_encoding_hexval(i8 %c)")
}

// ---- Base64URL (v6.8.0) ----

func TestEncoding_Base64URLEncodeCallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := Base64URLEncode('hello');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_encoding_Base64URLEncode")
	if strings.Contains(ir, "encoding.Base64URLEncode not implemented") {
		t.Errorf("Base64URLEncode still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestEncoding_Base64URLEncodeBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := Base64URLEncode('hello');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_encoding_Base64URLEncode(ptr %str)")
	// URL-safe table: '-'(0x2D) and '_'(0x5F) at the tail — NOT '+'/'/'.
	assertIRContains(t, ir, "@__kylix_b64url_table = private unnamed_addr constant [64 x i8]")
	assertIRContains(t, ir, `\2D\5F`)
	if strings.Contains(ir, `\2B\2F`) {
		t.Errorf("b64url table still uses the standard +/ alphabet")
	}
}

func TestEncoding_Base64URLDecodeBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses encoding;
begin
  var s := Base64URLDecode('aGVsbG8');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_encoding_Base64URLDecode")
	assertIRContains(t, ir, "define ptr @__kylix_encoding_Base64URLDecode(ptr %str)")
	// b64urlval helper recognizes '-'(→62) and '_'(→63).
	assertIRContains(t, ir, "define i64 @__kylix_encoding_b64urlval(i8 %c)")
	assertIRContains(t, ir, "ret i64 62")
	assertIRContains(t, ir, "ret i64 63")
}
