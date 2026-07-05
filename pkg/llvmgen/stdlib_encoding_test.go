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
