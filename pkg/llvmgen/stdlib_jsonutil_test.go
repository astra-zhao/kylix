package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_jsonutil tests — verify the v4.5.0 flat-object parser lowers to real
// @__kylix_json_parse_* defines (not the v4.4.0 empty-htab stub).

func TestJson_DecodeMap_ParserEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":"b"}');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_parse_flat")
	assertIRContains(t, ir, "define ptr @__kylix_json_parse_flat")
	// parser helpers emitted once
	assertIRContains(t, ir, "define void @__kylix_json_skip_ws")
	assertIRContains(t, ir, "define ptr @__kylix_json_read_string")
	assertIRContains(t, ir, "define ptr @__kylix_json_read_bare")
	assertIRContains(t, ir, "define ptr @__kylix_json_skip_nested")
	assertIRContains(t, ir, "define ptr @__kylix_json_read_value")
	// no longer the empty-htab stub
	if strings.Contains(ir, "TODO: implement flat-object parser") {
		t.Errorf("JsonDecodeMap still has stub TODO\nIR:\n%s", ir)
	}
}

func TestJson_DecodeMap_NotStubEmpty(t *testing.T) {
	// The v4.4.0 stub body was: call htab_new(); ret. The real body calls
	// parse_flat. Ensure the DecodeMap body calls parse_flat, not just htab_new.
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"k":1}');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_json_JsonDecodeMap")
	// The JsonDecodeMap body should contain parse_flat call (not just htab_new + ret)
	decIdx := strings.Index(ir, "define ptr @__kylix_json_JsonDecodeMap")
	if decIdx < 0 {
		t.Fatalf("JsonDecodeMap define not found")
	}
	bodyEnd := strings.Index(ir[decIdx:], "\n}")
	body := ir[decIdx : decIdx+bodyEnd]
	if !strings.Contains(body, "call ptr @__kylix_json_parse_flat") {
		t.Errorf("JsonDecodeMap body does not call parse_flat\nbody:\n%s", body)
	}
}

func TestJson_GetString_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":"b"}');
  var s := JsonGetString(m, 'a');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonGetString")
	// v5.1.0: jsonutil maps hold Variant boxes; JsonGetString unboxes via
	// htab_get_variant + variant_as_str (not htab_get + raw ptr).
	assertIRContains(t, ir, "call ptr @__kylix_htab_get_variant")
	assertIRContains(t, ir, "call ptr @__kylix_variant_as_str(ptr")
}

func TestJson_GetInt_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":3}');
  var n := JsonGetInt(m, 'a');
end.`)
	assertIRContains(t, ir, "call i64 @__kylix_json_JsonGetInt")
	// v5.1.0: JsonGetInt unboxes via variant_as_int (atoll now lives inside
	// the as_int body, not the JsonGetInt body).
	assertIRContains(t, ir, "call i64 @__kylix_variant_as_int(ptr")
}

func TestJson_GetBool_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":true}');
  var b := JsonGetBool(m, 'a');
end.`)
	assertIRContains(t, ir, "call i1 @__kylix_json_JsonGetBool")
	// v5.1.0: JsonGetBool unboxes via variant_as_bool (strcmp now inside as_bool).
	assertIRContains(t, ir, "call i1 @__kylix_variant_as_bool(ptr")
}

func TestJson_GetFloat_Real(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":1.5}');
  var f := JsonGetFloat(m, 'a');
end.`)
	assertIRContains(t, ir, "call double @__kylix_json_JsonGetFloat")
	// v5.1.0: JsonGetFloat unboxes via variant_as_double (strtod inside as_double).
	assertIRContains(t, ir, "call double @__kylix_variant_as_double(ptr")
}

func TestJson_Decode_AliasParseFlat(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecode('{"a":1}');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonDecode")
	assertIRContains(t, ir, "define ptr @__kylix_json_JsonDecode")
}

func TestJson_GetMap_ParsesNestedObject(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":{"b":1}}');
  var sub := JsonGetMap(m, 'a');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonGetMap")
	assertIRContains(t, ir, "define ptr @__kylix_json_JsonGetMap")
	// v4.7.0: GetMap recursively parses the raw substring. The body should
	// call parse_flat (not just `ret ptr null` like the v4.5.0 stub).
	getMapIdx := strings.Index(ir, "define ptr @__kylix_json_JsonGetMap")
	bodyEnd := strings.Index(ir[getMapIdx:], "\n}")
	body := ir[getMapIdx : getMapIdx+bodyEnd]
	if !strings.Contains(body, "call ptr @__kylix_json_parse_flat") {
		t.Errorf("JsonGetMap body should call parse_flat for nested objects\nbody:\n%s", body)
	}
	// Should still return null when the key is absent/empty (strcmp branch).
	if !strings.Contains(body, "call i32 @strcmp") {
		t.Errorf("JsonGetMap body should strcmp the raw value against empty\nbody:\n%s", body)
	}
}

func TestJson_GetArray_ParsesArray(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":[1,2]}');
  var arr := JsonGetArray(m, 'a');
end.`)
	// v4.9.0: JsonGetArray now returns a {ptr,i64,i64} slice via an out-param
	// (define void), not a bare ptr. The call writes into a temp alloca.
	assertIRContains(t, ir, "call void @__kylix_json_JsonGetArray")
	assertIRContains(t, ir, "define void @__kylix_json_JsonGetArray")
	// The array parser is emitted on first use.
	assertIRContains(t, ir, "define void @__kylix_json_parse_array")
}

func TestJson_ArrayLen_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":[1,2]}');
  var arr := JsonGetArray(m, 'a');
  WriteLn(JsonArrayLen(arr));
end.`)
	assertIRContains(t, ir, "call i64 @__kylix_json_JsonArrayLen")
	assertIRContains(t, ir, "define i64 @__kylix_json_JsonArrayLen")
}

func TestJson_ArrayGetString_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":[1,2]}');
  var arr := JsonGetArray(m, 'a');
  WriteLn(JsonArrayGetString(arr, 0));
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonArrayGetString")
	assertIRContains(t, ir, "define ptr @__kylix_json_JsonArrayGetString")
}

func TestJson_HasKey_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":1}');
  if JsonHasKey(m, 'a') then WriteLn('yes');
end.`)
	assertIRContains(t, ir, "call i1 @__kylix_json_JsonHasKey")
	assertIRContains(t, ir, "call i1 @__kylix_htab_has")
}

func TestJson_IsValid_StillReal(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  if JsonIsValid('{}') then WriteLn('ok');
end.`)
	assertIRContains(t, ir, "call i1 @__kylix_json_JsonIsValid")
	assertIRContains(t, ir, "define i1 @__kylix_json_JsonIsValid")
}

func TestJson_ParserEmittedOnce(t *testing.T) {
	// Multiple JsonDecodeMap calls should still emit parse_flat only once.
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m1 := JsonDecodeMap('{"a":1}');
  var m2 := JsonDecodeMap('{"b":2}');
end.`)
	count := strings.Count(ir, "define ptr @__kylix_json_parse_flat")
	if count != 1 {
		t.Errorf("parse_flat define emitted %d times, want 1", count)
	}
}
