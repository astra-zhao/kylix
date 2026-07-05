package llvmgen_test

import (
	"strings"
	"testing"
)

// String comparison operators must lower to strcmp (lexicographic content
// comparison), not icmp (which would compare pointer addresses). See
// expr.go's emitStringCompare / the isStringCmp branch in emitInfix.

func TestStringCompare_EqualityUsesStrcmp(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var a := 'foo';
  var b := 'bar';
  if a = b then
    WriteLn('eq');
end.`)
	assertIRContains(t, ir, "call i32 @strcmp(ptr")
	assertIRContains(t, ir, "icmp eq i32")
	// Exception-runtime boilerplate (__kylix_is_subtype) legitimately contains
	// "icmp eq i64" for an unrelated loop bound check, so only inspect main().
	mainBody := mainFuncBody(t, ir)
	if strings.Contains(mainBody, "icmp eq i64") {
		t.Errorf("string equality must not compare via icmp eq i64 (pointer/garbage compare)\nmain():\n%s", mainBody)
	}
}

func TestStringCompare_NotEqualUsesStrcmp(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var a := 'foo';
  var b := 'bar';
  if a <> b then
    WriteLn('ne');
end.`)
	assertIRContains(t, ir, "call i32 @strcmp(ptr")
	assertIRContains(t, ir, "icmp ne i32")
}

func TestStringCompare_LessThanUsesStrcmp(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var a := 'foo';
  var b := 'bar';
  if a < b then
    WriteLn('lt');
end.`)
	assertIRContains(t, ir, "call i32 @strcmp(ptr")
	assertIRContains(t, ir, "icmp slt i32")
}

func TestStringCompare_LessOrEqualUsesStrcmp(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var a := 'foo';
  var b := 'bar';
  if a <= b then
    WriteLn('le');
end.`)
	assertIRContains(t, ir, "icmp sle i32")
}

func TestStringCompare_GreaterThanUsesStrcmp(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var a := 'foo';
  var b := 'bar';
  if a > b then
    WriteLn('gt');
end.`)
	assertIRContains(t, ir, "icmp sgt i32")
}

func TestStringCompare_GreaterOrEqualUsesStrcmp(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var a := 'foo';
  var b := 'bar';
  if a >= b then
    WriteLn('ge');
end.`)
	assertIRContains(t, ir, "icmp sge i32")
}

func TestStringCompare_LiteralAgainstVariable(t *testing.T) {
	// Regression case from example25_string_ops.klx: comparing a String
	// local against a string literal (both sides ptr) must not fall into
	// the numeric icmp-i64 path.
	ir := generateIR(t, `program p;
begin
  var a := 'foo';
  if a = 'foo' then
    WriteLn('a equals foo');
end.`)
	assertIRContains(t, ir, "call i32 @strcmp(ptr")
	// Exception-runtime boilerplate (__kylix_is_subtype) legitimately contains
	// "icmp eq i64" for an unrelated loop bound check, so only inspect main().
	mainBody := mainFuncBody(t, ir)
	if strings.Contains(mainBody, "icmp eq i64") {
		t.Errorf("literal-vs-variable string equality must use strcmp, not icmp eq i64\nmain():\n%s", mainBody)
	}
}

// mainFuncBody extracts the text of the @main function definition from a
// full IR module, so assertions can avoid false positives from unrelated
// boilerplate (e.g. the exception-runtime subtype-check helper).
func mainFuncBody(t *testing.T, ir string) string {
	t.Helper()
	start := strings.Index(ir, "define i32 @main()")
	if start == -1 {
		t.Fatalf("no @main function found in IR:\n%s", ir)
	}
	end := strings.Index(ir[start:], "\n}\n")
	if end == -1 {
		return ir[start:]
	}
	return ir[start : start+end]
}

func TestNumericCompare_StillUsesIcmpI64(t *testing.T) {
	// Guard against a regression where the isStringCmp check accidentally
	// swallows the plain integer-comparison path.
	ir := generateIR(t, `program p;
begin
  var a := 1;
  var b := 2;
  if a = b then
    WriteLn('eq');
end.`)
	assertIRContains(t, ir, "icmp eq i64")
	if strings.Contains(ir, "call i32 @strcmp") {
		t.Errorf("integer comparison must not call strcmp\nIR:\n%s", ir)
	}
}

func TestFloatCompare_StillUsesFcmp(t *testing.T) {
	// Guard against a regression where the isStringCmp check runs before
	// the float-comparison path and misclassifies double operands.
	ir := generateIR(t, `program p;
begin
  var a := 1.5;
  var b := 2.5;
  if a = b then
    WriteLn('eq');
end.`)
	assertIRContains(t, ir, "fcmp oeq double")
	if strings.Contains(ir, "call i32 @strcmp") {
		t.Errorf("float comparison must not call strcmp\nIR:\n%s", ir)
	}
}

func TestStringSlice_AllocatesNewBufferAndCopies(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var s := 'Hello, World!';
  var sub := s[0:5];
  WriteLn(sub);
end.`)
	assertIRContains(t, ir, "call ptr @malloc")
	assertIRContains(t, ir, "call ptr @memcpy")
	assertIRContains(t, ir, "store i8 0") // null terminator
}

func TestStringSlice_ComputesLengthFromBounds(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var s := 'abcdefgh';
  var sub := s[2:6];
  WriteLn(sub);
end.`)
	// length = high - low
	assertIRContains(t, ir, "sub i64")
	// allocSize = length + 1
	assertIRContains(t, ir, "add i64")
}

func TestStringSlice_OffsetsSrcPointer(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  var s := 'abcdefgh';
  var sub := s[3:7];
  WriteLn(sub);
end.`)
	// src = base + low
	assertIRContains(t, ir, "getelementptr inbounds i8")
}
