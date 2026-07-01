package llvmgen_test

import (
	"strings"
	"testing"
)

func TestInherited_BareCallParentMethod(t *testing.T) {
	// `inherited;` in an override calls the parent's same-named method.
	ir := generateExcIR(t, `program p;
type
  TBase = class
    procedure Greet();
    begin
      WriteLn('base');
    end;
  end;
  TChild = class(TBase)
    procedure Greet();
    begin
      inherited;
    end;
  end;
begin
end.`)
	// TChild.Greet must direct-call @TBase_Greet with %self.
	assertExcContains(t, ir, "call void @TBase_Greet(ptr %self)")
}

func TestInherited_QualifiedCallWithArgs(t *testing.T) {
	// `inherited MethodName(args)` calls a specific parent method.
	ir := generateExcIR(t, `program p;
type
  TBase = class
    procedure Hello(name: String);
    begin
      WriteLn(name);
    end;
  end;
  TChild = class(TBase)
    procedure Hello(name: String);
    begin
      inherited Hello(name);
    end;
  end;
begin
end.`)
	assertExcContains(t, ir, "call void @TBase_Hello(ptr %self, ptr")
}

func TestInherited_FunctionReturnValue(t *testing.T) {
	// `inherited Func(args);` as a statement calls the parent function and
	// stores the return into %result (so the overriding function returns it).
	ir := generateExcIR(t, `program p;
type
  TBase = class
    function Double(x: Integer): Integer;
    begin
      result := x * 2;
    end;
  end;
  TChild = class(TBase)
    function Double(x: Integer): Integer;
    begin
      inherited Double(x);
    end;
  end;
begin
end.`)
	assertExcContains(t, ir, "call i64 @TBase_Double(ptr %self, i64")
	// Result stored into %result for the override to return.
	assertExcContains(t, ir, "store i64")
}

func TestInherited_DeepInheritanceChain(t *testing.T) {
	// inherited skips to the grandparent if the parent doesn't define the method.
	ir := generateExcIR(t, `program p;
type
  TGrand = class
    procedure Root();
    begin
    end;
  end;
  TMid = class(TGrand)
  end;
  TLeaf = class(TMid)
    procedure Root();
    begin
      inherited;
    end;
  end;
begin
end.`)
	// TLeaf.Root must resolve through TMid to TGrand.
	assertExcContains(t, ir, "call void @TGrand_Root(ptr %self)")
}

func TestInherited_NoParentMethodEmitsComment(t *testing.T) {
	// If the method isn't found in the parent chain, emit a comment (no crash).
	ir := generateExcIR(t, `program p;
type
  TBase = class
  end;
  TChild = class(TBase)
    procedure Foo();
    begin
      inherited;
    end;
  end;
begin
end.`)
	if !strings.Contains(ir, "inherited: method Foo not found") {
		t.Errorf("expected 'not found' comment for missing parent method\nIR:\n%s", ir)
	}
}
