package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_sysutil_test.go — IR assertions for the sysutil stdlib module.
// Uses the shared generateIR/assertIRContains helpers in codegen_test.go.

func TestSysutil_ReadFileCallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses sysutil;
begin
  var s := sysutil.ReadFile('/etc/hostname');
end.`)
	// Call site lowers to the libc-backed define, NOT an unsupported-receiver stub.
	assertIRContains(t, ir, "call ptr @__kylix_sysutil_ReadFile")
	if strings.Contains(ir, "unsupported receiver for ReadFile") {
		t.Errorf("ReadFile still routed to unsupported-receiver stub\nIR:\n%s", ir)
	}
}

func TestSysutil_ReadFileBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses sysutil;
begin
  var s := sysutil.ReadFile('/etc/hostname');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_sysutil_ReadFile(ptr %path)")
	assertIRContains(t, ir, "call ptr @fopen")
	assertIRContains(t, ir, "call i64 @fread")
	assertIRContains(t, ir, "call i32 @fclose")
}

func TestSysutil_WriteFileVoidCall(t *testing.T) {
	ir := generateIR(t, `program p;
uses sysutil;
begin
  sysutil.WriteFile('/tmp/x.txt', 'hello');
end.`)
	assertIRContains(t, ir, "call void @__kylix_sysutil_WriteFile")
	assertIRContains(t, ir, "define void @__kylix_sysutil_WriteFile(ptr %path, ptr %content)")
	assertIRContains(t, ir, "call i32 @fputs")
}

func TestSysutil_FileExistsBoolean(t *testing.T) {
	ir := generateIR(t, `program p;
uses sysutil;
begin
  var ok := sysutil.FileExists('/tmp/x.txt');
end.`)
	assertIRContains(t, ir, "call i1 @__kylix_sysutil_FileExists")
	assertIRContains(t, ir, "define i1 @__kylix_sysutil_FileExists(ptr %path)")
	assertIRContains(t, ir, "call i32 @access")
	assertIRContains(t, ir, "icmp eq i32")
}

func TestSysutil_PathJoinVariadic(t *testing.T) {
	ir := generateIR(t, `program p;
uses sysutil;
begin
  var p := sysutil.PathJoin('/usr', 'local', 'bin');
end.`)
	// PathJoin is monomorphized by arg count (3 args → _3 suffix).
	assertIRContains(t, ir, "call ptr @__kylix_sysutil_PathJoin_3")
	assertIRContains(t, ir, "define ptr @__kylix_sysutil_PathJoin_3(ptr %p0, ptr %p1, ptr %p2)")
	// Body uses strcat to join segments.
	assertIRContains(t, ir, "call ptr @strcat")
}

func TestSysutil_PathBase(t *testing.T) {
	ir := generateIR(t, `program p;
uses sysutil;
begin
  var b := sysutil.PathBase('/path/to/file.txt');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_sysutil_PathBase")
	assertIRContains(t, ir, "define ptr @__kylix_sysutil_PathBase(ptr %path)")
	assertIRContains(t, ir, "call i64 @strlen")
}

func TestSysutil_BodyDedup(t *testing.T) {
	// Two ReadFile calls must emit the define exactly once.
	ir := generateIR(t, `program p;
uses sysutil;
begin
  var a := sysutil.ReadFile('/x');
  var b := sysutil.ReadFile('/y');
end.`)
	if got := strings.Count(ir, "define ptr @__kylix_sysutil_ReadFile"); got != 1 {
		t.Errorf("ReadFile define should appear once, got %d\nIR:\n%s", got, ir)
	}
}

func TestSyslib_LibcDeclsPresent(t *testing.T) {
	ir := generateIR(t, `program p;
uses sysutil;
begin
  var s := sysutil.ReadFile('/x');
end.`)
	assertIRContains(t, ir, "declare ptr @fopen")
	assertIRContains(t, ir, "declare i32 @fclose")
	assertIRContains(t, ir, "declare i64 @fread")
	assertIRContains(t, ir, "declare i32 @access")
	assertIRContains(t, ir, "declare i32 @fputs")
}

func TestSysutil_NotUsedNoBodies(t *testing.T) {
	// A program that does NOT `uses sysutil` should not emit syslib function
	// bodies. (libc decls like fopen are always present — that's fine, unused
	// declares don't cause link errors; only the @__kylix_sysutil_* defines
	// should be gated on `uses`.)
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "@__kylix_sysutil") {
		t.Errorf("sysutil symbol emitted without `uses sysutil`\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "unsupported receiver") {
		t.Errorf("unexpected unsupported-receiver stub\nIR:\n%s", ir)
	}
}

func TestSysutil_MethodCallUnaffectedRegression(t *testing.T) {
	// A normal object method call must still dispatch via emitMethodCall —
	// the stdlib detection must not intercept it.
	ir := generateIR(t, `program p;
type
  TFoo = class
    procedure Speak;
    begin
      WriteLn('foo');
    end;
  end;
var f: TFoo;
begin
  f := TFoo.Create;
  f.Speak;
end.`)
	// Method dispatch still produces the class method symbol, not a sysutil stub.
	assertIRContains(t, ir, "@TFoo_Speak")
	if strings.Contains(ir, "unsupported receiver for Speak") {
		t.Errorf("object method call broken by stdlib detection\nIR:\n%s", ir)
	}
}
