package llvmgen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"kylix/pkg/llvmgen"
)

// v0.6.10: `var` output params must be passed by pointer and dereferenced
// inside the callee (mirroring the Go backend). Previously the LLVM backend
// ignored IsVar and passed by value, so the callee's writes were lost.
func TestVarParamOutput(t *testing.T) {
	src := `program VarTest;
procedure SetIt(var x: Integer);
begin
  x := 42;
end;
var v: Integer;
begin
  v := 0;
  SetIt(v);
  WriteLn(IntToStr(v));
end.`
	out := compileRunLLVM(t, src)
	if out != "42\n" {
		t.Fatalf("var param output: got %q want %q\n%s", out, "42\n", compileLLVMIr(t, src))
	}
}

// compileLLVMIr parses + generates IR for src (diagnostic helper).
func compileLLVMIr(t *testing.T, src string) string {
	t.Helper()
	ir := generateIR(t, src)
	return ir
}

// compileRunLLVM writes the IR to a temp file, runs llc + clang, executes the
// binary and returns its stdout. Skips if llc/clang unavailable.
func compileRunLLVM(t *testing.T, src string) string {
	t.Helper()
	paths, err := llvmgen.FindLLVM()
	if err != nil {
		t.Skipf("llvm toolchain unavailable: %v", err)
	}
	dir := t.TempDir()
	ll := filepath.Join(dir, "main.ll")
	o := filepath.Join(dir, "main.o")
	bin := filepath.Join(dir, "main")
	if err := os.WriteFile(ll, []byte(generateIR(t, src)), 0644); err != nil {
		t.Fatalf("write IR: %v", err)
	}
	if out, err := exec.Command(paths.LLC, "-filetype=obj", "-O=0", "-disable-verify", "-o", o, ll).CombinedOutput(); err != nil {
		t.Fatalf("llc: %v\n%s", err, out)
	}
	if out, err := exec.Command(paths.Clang, "-o", bin, o).CombinedOutput(); err != nil {
		t.Fatalf("clang: %v\n%s", err, out)
	}
	out, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	return string(out)
}
