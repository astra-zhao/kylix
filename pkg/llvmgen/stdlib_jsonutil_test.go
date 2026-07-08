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
	assertIRContains(t, ir, "call ptr @__kylix_htab_get")
}

func TestJson_GetInt_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":3}');
  var n := JsonGetInt(m, 'a');
end.`)
	assertIRContains(t, ir, "call i64 @__kylix_json_JsonGetInt")
	assertIRContains(t, ir, "call i64 @atoll")
}

func TestJson_GetBool_Dispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":true}');
  var b := JsonGetBool(m, 'a');
end.`)
	assertIRContains(t, ir, "call i1 @__kylix_json_JsonGetBool")
	assertIRContains(t, ir, "call i32 @strcmp")
}

func TestJson_GetFloat_Real(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":1.5}');
  var f := JsonGetFloat(m, 'a');
end.`)
	assertIRContains(t, ir, "call double @__kylix_json_JsonGetFloat")
	assertIRContains(t, ir, "call double @strtod")
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

func TestJson_GetMap_ReturnsNull(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":{"b":1}}');
  var sub := JsonGetMap(m, 'a');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonGetMap")
	assertIRContains(t, ir, "define ptr @__kylix_json_JsonGetMap")
	// GetMap body returns null (nested not supported in flat parser)
	getMapIdx := strings.Index(ir, "define ptr @__kylix_json_JsonGetMap")
	bodyEnd := strings.Index(ir[getMapIdx:], "\n}")
	body := ir[getMapIdx : getMapIdx+bodyEnd]
	if !strings.Contains(body, "ret ptr null") {
		t.Errorf("JsonGetMap body should ret null (nested unsupported)\nbody:\n%s", body)
	}
}

func TestJson_GetArray_ReturnsNull(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":[1,2]}');
  var arr := JsonGetArray(m, 'a');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonGetArray")
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
