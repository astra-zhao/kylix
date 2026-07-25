package llvmgen_test

import (
	"strings"
	"testing"
)

// TestBareMemberMethodCall_Statement (v5.6.0) guards the fix for bare
// parameterless method-call statements like `self.CollectImports;`. In Pascal
// these are written without parentheses, so the AST is a bare MemberExpression
// (not a CallExpression wrapping one). Pre-fix, emitStatement fell through to
// emitExpr→emitMember→emitFieldAccess, which looked up the member as a FIELD,
// failed ("field CollectImports not found"), and emitted a no-op `add i64 0, 0`
// — silently dropping the call. In the bootstrap that dropped
// `self.CollectImports` so no imports were ever collected (empty `import ()`).
// Now a bare MemberExpression statement whose member is a method of the
// receiver's class is lowered to a zero-argument method call.
func TestBareMemberMethodCall_Statement(t *testing.T) {
	ir := generateIR(t, `program p;
type
  TCounter = class
    n: Integer;
    procedure Bump;
    begin
      self.n := self.n + 1;
    end;
  end;
var c: TCounter;
begin
  c := TCounter.Create;
  c.Bump;
end.`)
	// `c.Bump;` is a bare MemberExpression statement resolving to method Bump.
	// It must emit a call to the method (TCounter_Bump), NOT a no-op field
	// access ("field Bump not found").
	if !strings.Contains(ir, "call") || !strings.Contains(ir, "TCounter_Bump") {
		t.Errorf("expected a call to TCounter_Bump for the bare `c.Bump;` statement (not a no-op field access)\nIR:\n%s", ir)
	}
	if strings.Contains(ir, "field Bump not found") {
		t.Errorf("bare method-call statement was treated as a field access (\"field Bump not found\" no-op)\nIR:\n%s", ir)
	}
}
