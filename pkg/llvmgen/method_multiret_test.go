package llvmgen_test

// method_multiret_test.go — class methods returning (T1, T2) tuples
// (v0.7.2). The vtable slot returns the %__ret_<Class>_<Method> aggregate;
// the method body packs it via insertvalue (`result := (a, b)`) and call
// sites destructure it via extractvalue — both the tuple form
// `(q, r) := obj.M()` and the var-decl form `var q, r := obj.M()`.

import (
	"strings"
	"testing"
)

const methodMultiRetSrc = `program TestMethodMultiRet;
type
  TCalc = class
  public
    function DivMod(a, b: Integer): (Integer, Integer);
    begin
      result := (a div b, a mod b);
    end;
  end;
begin
  var c := TCalc.Create;
  var q, r := c.DivMod(17, 5);
  WriteLn(q);
  WriteLn(r);
end.`

func TestMethodMultiRet_AggregateType(t *testing.T) {
	ir := generateIR(t, methodMultiRetSrc)
	// The aggregate must be declared in the pre-scan, before any define.
	declIdx := strings.Index(ir, "%__ret_TCalc_DivMod = type { i64, i64 }")
	defIdx := strings.Index(ir, "define %__ret_TCalc_DivMod @TCalc_DivMod(")
	if declIdx < 0 {
		t.Fatalf("missing %%__ret_TCalc_DivMod type declaration")
	}
	if defIdx < 0 {
		t.Fatalf("missing multi-return method define")
	}
	if declIdx > defIdx {
		t.Fatalf("type declaration (offset %d) must precede the define (offset %d)", declIdx, defIdx)
	}
}

func TestMethodMultiRet_BodyPacksTuple(t *testing.T) {
	ir := generateIR(t, methodMultiRetSrc)
	// result := (a div b, a mod b) → insertvalue chain into the aggregate.
	assertIRContains(t, ir, "insertvalue %__ret_TCalc_DivMod undef")
	assertIRContains(t, ir, "insertvalue %__ret_TCalc_DivMod %t")
	// The epilogue rets the aggregate loaded from %result.
	assertIRContains(t, ir, "ret %__ret_TCalc_DivMod")
}

func TestMethodMultiRet_VarDeclDestructure(t *testing.T) {
	ir := generateIR(t, methodMultiRetSrc)
	// var q, r := c.DivMod(17, 5) → extractvalue into two typed slots.
	assertIRContains(t, ir, "extractvalue %__ret_TCalc_DivMod %")
	assertIRContains(t, ir, "= alloca i64, align 8")
}

func TestMethodMultiRet_TupleDestructureForm(t *testing.T) {
	ir := generateIR(t, `program TestTupleForm;
type
  TCalc = class
  public
    function DivMod(a, b: Integer): (Integer, Integer);
    begin
      result := (a div b, a mod b);
    end;
  end;
begin
  var c := TCalc.Create;
  var q: Integer;
  var r: Integer;
  (q, r) := c.DivMod(17, 5);
  WriteLn(q);
end.`)
	assertIRContains(t, ir, "extractvalue %__ret_TCalc_DivMod %")
	assertIRContains(t, ir, "store i64 %t")
}

func TestMethodMultiRet_VtableSlotSignature(t *testing.T) {
	// The indirect call through the vtable must use the aggregate return type.
	ir := generateIR(t, methodMultiRetSrc)
	assertIRContains(t, ir, "call %__ret_TCalc_DivMod (ptr, i64, i64) %")
}
