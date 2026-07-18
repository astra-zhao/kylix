package llvmgen_test

import (
	"strings"
	"testing"
)

// variant_test.go — tests for the v5.0.0 Variant runtime (boxed-pointer tagged
// union): scalar decl/assign/compare/print, array of Variant element store/
// load, JsonGetArray producing Variant boxes, and Length(arr) routing.

func TestVariant_ScalarDeclBoxes(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 1.0;
end.`)
	// Variant local allocates a ptr slot with the _var suffix.
	assertIRContains(t, ir, "alloca ptr, align 8")
	// Assigning a Real literal boxes it via box_float.
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_float(double")
	// The Variant runtime bodies are emitted.
	assertIRContains(t, ir, "define ptr @__kylix_variant_box_float(double")
}

func TestVariant_ScalarIntBox(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 42;
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_int(i64")
}

func TestVariant_ScalarStrBox(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 'hello';
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_str(ptr")
}

func TestVariant_ScalarCompare(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 1.0;
  if v = 1.0 then WriteLn('match');
end.`)
	// Comparison routes to the runtime comparator (not a raw fcmp on the box ptr).
	assertIRContains(t, ir, "call i32 @__kylix_variant_compare(ptr")
	assertIRContains(t, ir, "icmp eq i32") // = maps to cmp==0
	assertIRContains(t, ir, "define i32 @__kylix_variant_compare(ptr")
}

func TestVariant_WriteLnDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 1.0;
  WriteLn(v);
end.`)
	assertIRContains(t, ir, "call void @__kylix_variant_println(ptr")
	assertIRContains(t, ir, "define void @__kylix_variant_println(ptr")
}

func TestVariant_StaticArrayElementAssign(t *testing.T) {
	ir := generateIR(t, `program p;
var arr: array[0..2] of Variant;
begin
  arr[0] := 10.0;
  arr[1] := 'x';
end.`)
	// Static array of Variant → [N x ptr] of box pointers.
	assertIRContains(t, ir, "alloca [3 x ptr], align 8")
	// Element assignment boxes the RHS.
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_float(double")
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_str(ptr")
}

func TestVariant_StaticArrayElementRead(t *testing.T) {
	ir := generateIR(t, `program p;
var arr: array[0..2] of Variant;
begin
  arr[0] := 10.0;
  if arr[0] = 10.0 then WriteLn('ok');
end.`)
	// arr[0] read + comparison dispatches via variant_compare.
	assertIRContains(t, ir, "call i32 @__kylix_variant_compare(ptr")
}

func TestVariant_LengthArrayRouting(t *testing.T) {
	ir := generateIR(t, `program p;
var arr: array of Variant;
begin
  WriteLn(Length(arr));
end.`)
	// Length(arr) on a dynamic array reads the slice len word (GEP field 1 of
	// the {ptr,i64,i64} struct), NOT strlen the data pointer.
	assertIRContains(t, ir, "getelementptr inbounds { ptr, i64, i64 }, ptr")
	// Should not fall back to calling strlen on the data pointer (the bare
	// `declare i64 @strlen` libc decl is always emitted; only count actual calls).
	if containsCount(ir, "call i64 @strlen") > 0 {
		t.Errorf("Length(arr) should route to emitArrayLength, not strlen\nIR:\n%s", ir)
	}
}

func TestVariant_JsonGetArrayProducesVariants(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":[1,2,3]}');
  var arr := JsonGetArray(m, 'a');
end.`)
	// parse_array now classifies elements into Variant boxes via value_to_variant.
	assertIRContains(t, ir, "call ptr @__kylix_json_value_to_variant(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_json_value_to_variant(ptr")
	// The Variant box helpers are emitted (value_to_variant calls them).
	assertIRContains(t, ir, "define ptr @__kylix_variant_box_float(double")
	assertIRContains(t, ir, "define ptr @__kylix_variant_box_str(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_variant_box_bool(i1")
}

func TestVariant_JsonArrayGetStringUnboxes(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":[1,2,3]}');
  var arr := JsonGetArray(m, 'a');
  WriteLn(JsonArrayGetString(arr, 0));
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_json_JsonArrayGetString(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_json_JsonArrayGetString(ptr")
	// The body now unboxes via variant_as_str.
	assertIRContains(t, ir, "call ptr @__kylix_variant_as_str(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_variant_as_str(ptr")
}

func TestVariant_RuntimeNotEmittedWhenUnused(t *testing.T) {
	// A program with no Variant usage must not emit the Variant runtime.
	ir := generateIR(t, `program p;
begin
  WriteLn('hello');
end.`)
	if containsCount(ir, "@__kylix_variant_") > 0 {
		t.Errorf("Variant runtime should not be emitted when no Variant is used\nIR:\n%s", ir)
	}
}

// containsCount reports the number of (non-overlapping) occurrences of substr.
func containsCount(s, substr string) int {
	n := 0
	for {
		i := strings.Index(s, substr)
		if i < 0 {
			break
		}
		n++
		s = s[i+len(substr):]
	}
	return n
}

// ===== v5.1.0: Variant arithmetic + map[String]Variant + Variant→concrete =====

func TestVariant_ArithIntAdd(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 5;
  v := v + 3;
end.`)
	// Variant '+' dispatches to the runtime add helper (returns a box).
	assertIRContains(t, ir, "call ptr @__kylix_variant_add(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_variant_add(ptr")
}

func TestVariant_ArithStrConcat(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 'a';
  v := v + 'b';
end.`)
	// '+' with a string operand → variant_add; the str-concat branch calls as_str.
	assertIRContains(t, ir, "call ptr @__kylix_variant_add(ptr")
	assertIRContains(t, ir, "call ptr @__kylix_variant_as_str(ptr")
}

func TestVariant_ArithAllOpsEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 1.0;
  v := v + 2.0;
  v := v - 1.0;
  v := v * 2.0;
  v := v / 2.0;
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_variant_add(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_variant_sub(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_variant_mul(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_variant_div(ptr")
}

func TestVariant_AsIntAndAsBoolHelpers(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
begin
  var m := JsonDecodeMap('{"a":1,"b":true}');
  WriteLn(JsonGetInt(m, 'a'));
  if JsonGetBool(m, 'b') then WriteLn('yes');
end.`)
	// JsonGetInt/JsonGetBool now unbox via variant_as_int/variant_as_bool.
	assertIRContains(t, ir, "call i64 @__kylix_variant_as_int(ptr")
	assertIRContains(t, ir, "define i64 @__kylix_variant_as_int(ptr")
	assertIRContains(t, ir, "call i1 @__kylix_variant_as_bool(ptr")
	assertIRContains(t, ir, "define i1 @__kylix_variant_as_bool(ptr")
}

func TestVariant_VariantToConcreteAssign(t *testing.T) {
	// `n := v` (Variant→Integer) unboxes via coerceValue (variant_as_int).
	ir := generateIR(t, `program p;
var
  v: Variant;
  n: Integer;
begin
  v := 42;
  n := v;
end.`)
	assertIRContains(t, ir, "call i64 @__kylix_variant_as_int(ptr")
}

func TestVariant_MapVariantDeclAndIndex(t *testing.T) {
	ir := generateIR(t, `program p;
uses jsonutil;
var m: map[String]Variant;
begin
  m := JsonDecodeMap('{"pi":3.14}');
  if m['pi'] = 3.14 then WriteLn('match');
end.`)
	// map[String]Variant → htab_get_variant (returns a box).
	assertIRContains(t, ir, "call ptr @__kylix_htab_get_variant(ptr")
	assertIRContains(t, ir, "define ptr @__kylix_htab_get_variant(ptr")
	// m['pi'] (Variant) vs 3.14 (double) → variant_compare.
	assertIRContains(t, ir, "call i32 @__kylix_variant_compare(ptr")
	// parse_flat now produces boxes via value_to_variant.
	assertIRContains(t, ir, "call ptr @__kylix_json_value_to_variant(ptr")
}

func TestVariant_MapVariantWriteBoxes(t *testing.T) {
	// Direct `m['k'] := v` on a Variant map boxes the value (no stringify).
	ir := generateIR(t, `program p;
var m: map[String]Variant;
begin
  m['k'] := 7;
end.`)
	// map[String]Variant put → htab_put with a boxed value (box_int).
	assertIRContains(t, ir, "call ptr @__kylix_variant_box_int(i64")
	assertIRContains(t, ir, "call void @__kylix_htab_put(ptr")
}

func TestVariant_NilboxGlobalEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
var v: Variant;
begin
  v := 1;
end.`)
	// The nilbox global (htab_get_variant's miss sentinel) is emitted.
	assertIRContains(t, ir, "@__kylix_variant_nilbox =")
}

