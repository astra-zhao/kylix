package llvmgen_test

import (
	"strings"
	"testing"
)

// generateLambdaIR parses src and returns the generated LLVM IR.
func generateLambdaIR(t *testing.T, src string) string {
	t.Helper()
	ir := generateExcIR(t, src) // reuse the shared helper in exc_test.go
	return ir
}

func TestLambda_NoCaptureProcedureEmitsNamedFunction(t *testing.T) {
	ir := generateLambdaIR(t, `program p;
begin
  var greet := procedure(name: String)
  begin
    WriteLn(name);
  end;
  greet('Alice');
end.`)
	// Lambda becomes a named function with (env, params) signature.
	assertExcContains(t, ir, "define void @__lambda_0(ptr %env, ptr %name)")
	// Closure pair stores the function pointer.
	assertExcContains(t, ir, "store ptr @__lambda_0")
	// No captures → env is null.
	assertExcContains(t, ir, "store ptr null, ptr")
}

func TestLambda_ClosureIndirectCall(t *testing.T) {
	ir := generateLambdaIR(t, `program p;
begin
  var greet := procedure(name: String)
  begin
    WriteLn(name);
  end;
  greet('Alice');
end.`)
	// Call through the closure: load {ptr,ptr}, extractvalue both fields, indirect call.
	assertExcContains(t, ir, "load { ptr, ptr }")
	assertExcContains(t, ir, "extractvalue { ptr, ptr }")
	// Indirect call passes env (ptr) as the first argument, with explicit
	// function-type signature.
	assertExcContains(t, ir, "call void (ptr, ptr) %")
}

func TestLambda_FunctionWithReturnType(t *testing.T) {
	ir := generateLambdaIR(t, `program p;
begin
  var sq := function(x: Integer): Integer;
  begin
    result := x * x;
  end;
  WriteLn(sq(7));
end.`)
	// Function lambda returns i64 (not void).
	assertExcContains(t, ir, "define i64 @__lambda_0(ptr %env, i64 %x)")
	assertExcContains(t, ir, "ret i64")
	// Result alloca present for block-bodied function lambda.
	assertExcContains(t, ir, "%result = alloca i64")
}

func TestLambda_CaptureVariableEmitsEnvStruct(t *testing.T) {
	ir := generateLambdaIR(t, `program p;
begin
  var n := 10;
  var add := function(x: Integer): Integer;
  begin
    result := x + n;
  end;
  WriteLn(add(5));
end.`)
	// Env struct literal type appears with the captured i64 field.
	assertExcContains(t, ir, "getelementptr { i64 }")
	// Creation site stores the captured value into the env.
	assertExcContains(t, ir, "store i64")
	// Env is malloc'd at creation (not null).
	assertExcContains(t, ir, "call ptr @malloc")
}

func TestLambda_CaptureMaterializedInBody(t *testing.T) {
	// Inside the lambda body, the captured variable is loaded from env into a
	// local alloca so emitIdentLoad works transparently.
	ir := generateLambdaIR(t, `program p;
begin
  var n := 10;
  var add := function(x: Integer): Integer;
  begin
    result := x + n;
  end;
end.`)
	// The body loads env field then stores into a local named after the capture.
	assertExcContains(t, ir, "getelementptr { i64 }, ptr %env, i32 0, i32 0")
	assertExcContains(t, ir, "%v_n_int = alloca i64")
}

func TestLambda_NotNullPointerStub(t *testing.T) {
	// Regression: previously lambda emitted `inttoptr i64 0 to ptr` stub.
	ir := generateLambdaIR(t, `program p;
begin
  var f := procedure()
  begin
  end;
end.`)
	if strings.Contains(ir, "inttoptr i64 0 to ptr") {
		t.Errorf("lambda still using null pointer stub\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "lambda/closure unsupported") {
		t.Errorf("lambda still marked unsupported\nIR:\n%s", ir)
	}
}
