package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/pkg/llvmgen"
)

// passes_test.go — tests for the process-in-LLVM IR optimization passes
// (v4.5.0 Phase C): DCE removes unreferenced %tN defs; ConstantFold is a
// structural no-op in the MVP.

func TestDCE_RemovesUnreferencedTemp(t *testing.T) {
	// %t0 is defined but never read → should be removed.
	ir := `define i64 @f() {
entry:
  %t0 = add i64 1, 2
  ret i64 42
}`
	out := llvmgen.DefaultPassPipeline().Run(ir)
	if strings.Contains(out, "%t0 = add") {
		t.Errorf("DCE did not remove unreferenced %%t0\nout:\n%s", out)
	}
}

func TestDCE_KeepsReferencedTemp(t *testing.T) {
	// %t0 is read by ret → must survive.
	ir := `define i64 @f() {
entry:
  %t0 = add i64 1, 2
  ret i64 %t0
}`
	out := llvmgen.DefaultPassPipeline().Run(ir)
	if !strings.Contains(out, "%t0 = add") {
		t.Errorf("DCE wrongly removed referenced %%t0\nout:\n%s", out)
	}
}

func TestDCE_NeverRemovesCallOrStore(t *testing.T) {
	// call/store have side effects — must never be deleted even if the
	// result register is unused.
	ir := `define void @f() {
entry:
  %t0 = call i64 @sideEffect()
  store i64 0, ptr %p
  ret void
}
declare i64 @sideEffect()`
	out := llvmgen.DefaultPassPipeline().Run(ir)
	if !strings.Contains(out, "call i64 @sideEffect") {
		t.Errorf("DCE wrongly removed call with side effects\nout:\n%s", out)
	}
	if !strings.Contains(out, "store i64 0") {
		t.Errorf("DCE wrongly removed store\nout:\n%s", out)
	}
}

func TestDCE_WordBoundaryNotPrefixMatch(t *testing.T) {
	// %t1 is referenced (inside %t10's use? no) — ensure "%t1" doesn't match
	// inside "%t10". Here %t1 is genuinely dead, %t10 is referenced.
	ir := `define i64 @f() {
entry:
  %t1 = add i64 0, 0
  %t10 = add i64 1, 2
  ret i64 %t10
}`
	out := llvmgen.DefaultPassPipeline().Run(ir)
	if strings.Contains(out, "%t1 = add") {
		t.Errorf("DCE did not remove dead %%t1 (word-boundary bug?)\nout:\n%s", out)
	}
	if !strings.Contains(out, "%t10 = add") {
		t.Errorf("DCE wrongly removed referenced %%t10\nout:\n%s", out)
	}
}

func TestDCE_NoOpOnCleanIR(t *testing.T) {
	// No dead temps → IR unchanged (no crash, no spurious deletion).
	ir := `define i64 @f() {
entry:
  %t0 = add i64 1, 2
  ret i64 %t0
}`
	out := llvmgen.DefaultPassPipeline().Run(ir)
	if strings.TrimSpace(out) != strings.TrimSpace(ir) {
		t.Errorf("DCE altered already-clean IR\nout:\n%s", out)
	}
}
