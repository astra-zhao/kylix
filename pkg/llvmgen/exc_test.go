package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/llvmgen"
)

// generateExcIR parses src and returns the generated LLVM IR, failing the test
// on parse or codegen errors. Mirrors the helper in codegen_test.go.
func generateExcIR(t *testing.T, src string) string {
	t.Helper()
	p := parser.New(lexer.New(src))
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	ir, err := llvmgen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	return ir
}

func assertExcContains(t *testing.T, ir, substr string) {
	t.Helper()
	if !strings.Contains(ir, substr) {
		t.Errorf("expected IR to contain %q\nActual IR:\n%s", substr, ir)
	}
}

func assertExcNotContains(t *testing.T, ir, substr string) {
	t.Helper()
	if strings.Contains(ir, substr) {
		t.Errorf("expected IR NOT to contain %q\nActual IR:\n%s", substr, ir)
	}
}

// ===== Runtime declarations & globals =====

func TestExc_DeclsAlwaysPresent(t *testing.T) {
	ir := generateExcIR(t, `program p; begin end.`)
	assertExcContains(t, ir, "declare i32 @setjmp(ptr)")
	assertExcContains(t, ir, "declare void @longjmp(ptr, i32)")
	assertExcContains(t, ir, "declare void @exit(i32)")
}

func TestExc_GlobalsAlwaysPresent(t *testing.T) {
	ir := generateExcIR(t, `program p; begin end.`)
	assertExcContains(t, ir, "@__kylix_exc_obj = global ptr null")
	assertExcContains(t, ir, "@__kylix_exc_type = global i32 0")
	assertExcContains(t, ir, "@__kylix_exc_active = global i1 false")
	assertExcContains(t, ir, "@__kylix_jmpbuf = global ptr null")
}

func TestExc_ExceptionClassInjected(t *testing.T) {
	// Even with no user code, the Exception class struct type is emitted.
	ir := generateExcIR(t, `program p; begin end.`)
	assertExcContains(t, ir, "%Exception = type { ptr, ptr }")
}

func TestExc_IsSubtypeRuntimeEmitted(t *testing.T) {
	ir := generateExcIR(t, `program p; begin end.`)
	assertExcContains(t, ir, "define i1 @__kylix_is_subtype")
}

// ===== raise =====

func TestExc_RaiseStoresObjectAndLongjmps(t *testing.T) {
	ir := generateExcIR(t, `program p;
begin
  raise Exception.Create;
end.`)
	assertExcContains(t, ir, "store ptr")
	assertExcContains(t, ir, "ptr @__kylix_exc_obj")
	assertExcContains(t, ir, "store i32")
	assertExcContains(t, ir, "ptr @__kylix_exc_type")
	assertExcContains(t, ir, "store i1 true, ptr @__kylix_exc_active")
	assertExcContains(t, ir, "call void @longjmp")
	// Uncaught path (no handler installed) calls exit.
	assertExcContains(t, ir, "call void @exit(i32 70)")
}

func TestExc_RaiseRecordsTypeID(t *testing.T) {
	// A custom subclass of Exception should get ID 2 and be stored on raise.
	ir := generateExcIR(t, `program p;
type
  TFooError = class(Exception) end;
begin
  raise TFooError.Create;
end.`)
	// TFooError is a subclass → ID 2. raise stores it.
	assertExcContains(t, ir, "store i32 2, ptr @__kylix_exc_type")
	// The subtype table records the edge TFooError(2) → Exception(1).
	assertExcContains(t, ir, "{ i32 2, i32 1 }")
}

// ===== try / except =====

func TestExc_TryInstallsHandler(t *testing.T) {
	ir := generateExcIR(t, `program p;
begin
  try
    WriteLn('body');
  except
    WriteLn('caught');
  end;
end.`)
	assertExcContains(t, ir, "call i32 @setjmp")
	assertExcContains(t, ir, "store ptr")
	assertExcContains(t, ir, "ptr @__kylix_jmpbuf")
}

func TestExc_OnClauseEmitsSubtypeCheck(t *testing.T) {
	ir := generateExcIR(t, `program p;
type
  TFooError = class(Exception) end;
begin
  try
    raise TFooError.Create;
  except
    on E: TFooError do
      WriteLn('foo');
  end;
end.`)
	assertExcContains(t, ir, "call i1 @__kylix_is_subtype")
}

func TestExc_OnClauseBindsVariable(t *testing.T) {
	ir := generateExcIR(t, `program p;
begin
  try
    raise Exception.Create;
  except
    on E: Exception do
      WriteLn('caught');
  end;
end.`)
	// The on-clause binds E via an alloca %v_E and loads the exception object.
	assertExcContains(t, ir, "%v_E = alloca ptr")
	assertExcContains(t, ir, "load ptr, ptr @__kylix_exc_obj")
}

func TestExc_PlainExceptHandlesAll(t *testing.T) {
	ir := generateExcIR(t, `program p;
begin
  try
    raise Exception.Create;
  except
    WriteLn('caught');
  end;
end.`)
	// A plain except block clears exc_active after handling (no on-clause).
	assertExcContains(t, ir, "store i1 false, ptr @__kylix_exc_active")
}

// ===== finally =====

func TestExc_FinallyEmitted(t *testing.T) {
	ir := generateExcIR(t, `program p;
begin
  try
    WriteLn('body');
  finally
    WriteLn('cleanup');
  end;
end.`)
	// Finally body is emitted (appears in normal and reraise paths).
	count := strings.Count(ir, "cleanup")
	if count < 1 {
		t.Errorf("finally body 'cleanup' not emitted; IR:\n%s", ir)
	}
}

// ===== bare raise =====

func TestExc_BareRaiseOutsideHandlerIsGeneric(t *testing.T) {
	// Bare raise with no enclosing except → generic Exception (ID 1).
	ir := generateExcIR(t, `program p;
begin
  raise;
end.`)
	assertExcContains(t, ir, "store i32 1, ptr @__kylix_exc_type")
	assertExcContains(t, ir, "call void @longjmp")
}

func TestExc_BareRaiseInsideExceptRethrows(t *testing.T) {
	// Bare raise inside an except handler re-throws: a second longjmp to the
	// outer handler (the raise path emits longjmp; the try itself emits one
	// on the reraise branch). We assert at least two longjmp calls appear.
	ir := generateExcIR(t, `program p;
begin
  try
    raise Exception.Create;
  except
    raise;
  end;
end.`)
	count := strings.Count(ir, "call void @longjmp")
	if count < 2 {
		t.Errorf("expected >= 2 longjmp calls (raise + reraise), got %d\nIR:\n%s", count, ir)
	}
}

// ===== nesting =====

func TestExc_NestedTryEmitsTwoSetjmps(t *testing.T) {
	ir := generateExcIR(t, `program p;
begin
  try
    try
      raise Exception.Create;
    except
      WriteLn('inner');
    end;
  except
    WriteLn('outer');
  end;
end.`)
	count := strings.Count(ir, "call i32 @setjmp")
	if count < 2 {
		t.Errorf("expected >= 2 setjmp calls for nested try, got %d\nIR:\n%s", count, ir)
	}
}

// ===== regression: no exception code in trivial program =====

func TestExc_NoRaiseNoLongjmp(t *testing.T) {
	// A program with no raise/try must not emit any longjmp call in main.
	ir := generateExcIR(t, `program p; begin WriteLn('hi'); end.`)
	assertExcNotContains(t, ir, "call void @longjmp")
	assertExcNotContains(t, ir, "call i32 @setjmp")
}
