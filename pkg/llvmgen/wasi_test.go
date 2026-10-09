package llvmgen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"kylix/internal/wasiapi"
	"kylix/pkg/llvmgen"
)

func TestTriple_Wasm32Wasi(t *testing.T) {
	src := "program test; begin end."
	aliases := []string{"wasi/wasm32", "wasm/wasm32", "wasm", "wasm32", "wasi"}
	for _, target := range aliases {
		ir, err := llvmgen.GenerateWithOpts(mustParse(t, src), "test.klx", llvmgen.CompileOpts{Target: target})
		if err != nil {
			t.Fatalf("target %s: %v", target, err)
		}
		if !strings.Contains(ir, `target triple = "wasm32-unknown-wasi"`) {
			t.Errorf("%s: missing wasm32 triple\n%s", target, ir[:min(400, len(ir))])
		}
		if !strings.Contains(ir, `target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"`) {
			t.Errorf("%s: missing wasm32 datalayout", target)
		}
	}
}

func TestWasi_ImportTableAndNoDOM(t *testing.T) {
	src := "program test; begin WriteLn('hi'); end."
	ir, err := llvmgen.GenerateWithOpts(mustParse(t, src), "test.klx", llvmgen.CompileOpts{Target: "wasi/wasm32"})
	if err != nil {
		t.Fatal(err)
	}
	for _, im := range wasiapi.Preview1 {
		needle := `"wasm-import-module"="wasi_snapshot_preview1" "wasm-import-name"="` + im.Name + `"`
		if !strings.Contains(ir, needle) {
			t.Errorf("missing import %s", im.Name)
		}
		if !strings.Contains(ir, "declare ") || !strings.Contains(ir, "@__kylix_wasi_"+im.Name+"(") {
			t.Errorf("missing declare for %s", im.Name)
		}
	}
	if strings.Contains(ir, `wasm-import-module"="env"`) || strings.Contains(ir, "document") {
		t.Error("DOM/env import does not belong on the wasi target")
	}
	for _, sym := range []string{"define ptr @malloc(", "define i32 @puts(", "define i32 @printf(", "define void @_start(", "define void @exit("} {
		if !strings.Contains(ir, sym) {
			t.Errorf("missing %s", sym)
		}
	}
	// Host libc declares must not also be present (declare+define of printf
	// is legal but the wasi path replaces the declare with a define only).
	if strings.Contains(ir, "declare i32 @printf") {
		t.Error("wasi target should define printf, not declare the host libc one")
	}
}

func TestWasi_HostIRUnchanged(t *testing.T) {
	src := "program test; begin WriteLn('hi'); end."
	ir, err := llvmgen.GenerateWithOpts(mustParse(t, src), "test.klx", llvmgen.CompileOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(ir, "wasi_snapshot_preview1") || strings.Contains(ir, "@__kylix_wasi_") || strings.Contains(ir, "wasm32-unknown-wasi") {
		t.Fatal("default target IR picked up the wasi runtime")
	}
	if !strings.Contains(ir, "declare i32 @printf") {
		t.Fatal("default target should still declare host printf")
	}
}

func TestWasi_RejectsBoehm(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "p.klx")
	if err := os.WriteFile(srcPath, []byte("program p; begin end."), 0644); err != nil {
		t.Fatal(err)
	}
	paths, err := llvmgen.FindLLVM()
	if err != nil {
		t.Skip(err)
	}
	_, err = llvmgen.CompileToNativeOpts(srcPath, filepath.Join(dir, "p.wasm"), paths, llvmgen.CompileOpts{
		Target: "wasi/wasm32",
		GC:     "boehm",
	})
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected boehm rejection, got %v", err)
	}
}

func TestWasi_HelloRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("wasmtime"); err != nil {
		t.Skip("wasmtime not installed")
	}
	paths, err := llvmgen.FindLLVM()
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	src := `program wasi_logic;
var
  s: String;
begin
  s := 'sum=' + IntToStr(40 + 2);
  WriteLn(s);
  WriteLn('hello wasi');
  WriteLn(7 * 6);
  WriteLn(1.5);
end.
`
	srcPath := filepath.Join(dir, "main.klx")
	if err := os.WriteFile(srcPath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "logic.wasm")
	if _, err := llvmgen.CompileToNativeOpts(srcPath, out, paths, llvmgen.CompileOpts{Target: "wasi/wasm32"}); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("wasmtime", out)
	got, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wasmtime: %v\n%s", err, got)
	}
	want := "sum=42\nhello wasi\n42\n1.5\n"
	if string(got) != want {
		t.Fatalf("stdout:\n%q\nwant:\n%q", got, want)
	}
	blob, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	text := string(blob)
	for _, name := range []string{"fd_write", "clock_time_get", "random_get", "proc_exit", "wasi_snapshot_preview1"} {
		if !strings.Contains(text, name) {
			t.Errorf("linked wasm missing import %s", name)
		}
	}
}
