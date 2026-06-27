package llvmgen_test

import (
	"strings"
	"testing"
)

// ===== Interface fat-pointer codegen tests (Milestone 2 Phase 2) =====

func TestIR_InterfaceDeclEmitsVtableType(t *testing.T) {
	ir := generateIR(t, `program test;
type
  IFoo = interface
    function M(): Integer;
  end;
begin
end.`)
	assertIRContains(t, ir, "%IFoo_vtable = type { ptr }")
	assertIRContains(t, ir, "%IFoo_iface = type { ptr, ptr }")
}

func TestIR_ClassEmitsInterfaceVtable(t *testing.T) {
	ir := generateIR(t, `program test;
type
  IFoo = interface
    function M(): Integer;
  end;
  TFoo = class implements IFoo
    function M(): Integer;
    begin
      result := 1;
    end;
  end;
begin
end.`)
	assertIRContains(t, ir, "@TFoo_IFoo_vtable = constant")
	assertIRContains(t, ir, "ptr @TFoo_M")
}

func TestIR_MemberAccess(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TBox = class
    Width: Integer;
  end;
var
  b: TBox;
begin
  b := TBox.Create;
  WriteLn(b.Width);
end.`)
	assertIRContains(t, ir, "getelementptr inbounds %TBox")
}

func TestIR_DirectMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program test;
type
  TGreeter = class
    function Hello(): Integer;
    begin
      result := 42;
    end;
  end;
var
  g: TGreeter;
begin
  g := TGreeter.Create;
  WriteLn(g.Hello());
end.`)
	// Should load the vtable slot for Hello (via emitVirtualCall).
	assertIRContains(t, ir, "load ptr, ptr")
	assertIRContains(t, ir, "@TGreeter_vtable")
}

func TestIR_InterfaceMethodDispatch(t *testing.T) {
	ir := generateIR(t, `program test;
type
  IFoo = interface
    function M(): Integer;
  end;
  TFoo = class implements IFoo
    function M(): Integer;
    begin
      result := 7;
    end;
  end;
var
  f: IFoo;
  o: TFoo;
begin
  o := TFoo.Create;
  f := o;
  WriteLn(f.M());
end.`)
	// Interface dispatch loads vtable slot 0 via indexed GEP.
	assertIRContains(t, ir, "@TFoo_IFoo_vtable")
	assertIRContains(t, ir, "getelementptr inbounds [1 x ptr]")
}

func TestIR_IsExpression(t *testing.T) {
	ir := generateIR(t, `program test;
type
  IFoo = interface
    function M(): Integer;
  end;
  TFoo = class implements IFoo
    function M(): Integer;
    begin
      result := 1;
    end;
  end;
  TBar = class
    function N(): Integer;
    begin
      result := 2;
    end;
  end;
var
  f: TFoo;
  b: TBar;
  ok1: Boolean;
  ok2: Boolean;
begin
  f := TFoo.Create;
  b := TBar.Create;
  ok1 := f is IFoo;
  ok2 := b is IFoo;
end.`)
	// f is IFoo → 1, b is IFoo → 0.
	assertIRContains(t, ir, "add i1 0, 1 ; TFoo is IFoo")
	assertIRContains(t, ir, "add i1 0, 0 ; TBar is IFoo")
}

func TestIR_AsExpressionBoxesIntoInterface(t *testing.T) {
	ir := generateIR(t, `program test;
type
  IFoo = interface
    function M(): Integer;
  end;
  TFoo = class implements IFoo
    function M(): Integer;
    begin
      result := 1;
    end;
  end;
var
  f: IFoo;
  o: TFoo;
begin
  o := TFoo.Create;
  f := o as IFoo;
end.`)
	// Boxing stores the per-class interface vtable pointer into the iface var.
	if !strings.Contains(ir, "store ptr @TFoo_IFoo_vtable") {
		t.Errorf("expected interface boxing to store @TFoo_IFoo_vtable into iface slot, IR was:\n%s", ir)
	}
}

func TestIR_InterfaceVarAllocatesFatPointerSlots(t *testing.T) {
	ir := generateIR(t, `program test;
type
  IFoo = interface
    function M(): Integer;
  end;
var
  f: IFoo;
begin
end.`)
	assertIRContains(t, ir, "%v_f_iface_vt = alloca ptr")
	assertIRContains(t, ir, "%v_f_iface_data = alloca ptr")
}
