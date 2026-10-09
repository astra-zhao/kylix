package wasi_test

import (
	"os"
	"strings"
	"testing"

	"kylix/internal/wasiapi"
	"kylix/pkg/wasi"
)

func TestEncodeCIVecs_Layout(t *testing.T) {
	got := wasi.EncodeCIVecs([]wasi.CIVec{{Off: 0x11, Len: 4}, {Off: 0x100, Len: 0}})
	want := []byte{
		0x11, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	if len(got) != wasi.CIVecSize*2 {
		t.Fatalf("len %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte %d: got %02x want %02x", i, got[i], want[i])
		}
	}
}

func TestDecodeIndexedStrings_DropsArgv0(t *testing.T) {
	buf := []byte("prog\x00hello\x00")
	const base uint32 = 0x200
	ptrs := []uint32{base, base + 5}
	got := wasi.DecodeIndexedStrings(base, ptrs, buf, true)
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("got %#v", got)
	}
	all := wasi.DecodeIndexedStrings(base, ptrs, buf, false)
	if len(all) != 2 || all[0] != "prog" || all[1] != "hello" {
		t.Fatalf("all %#v", all)
	}
}

func TestDecodeIndexedStrings_SkipsOutOfRange(t *testing.T) {
	buf := []byte("a\x00")
	got := wasi.DecodeIndexedStrings(10, []uint32{0, 10}, buf, false)
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("got %#v", got)
	}
	if wasi.DecodeIndexedStrings(0, nil, buf, false) != nil {
		t.Fatal("expected nil for empty ptrs")
	}
}

func TestSplitEnv(t *testing.T) {
	k, v, ok := wasi.SplitEnv("NAME=Kylix")
	if !ok || k != "NAME" || v != "Kylix" {
		t.Fatalf("got %q %q %v", k, v, ok)
	}
	if _, _, ok := wasi.SplitEnv("=nope"); ok {
		t.Fatal("empty key should fail")
	}
	if _, _, ok := wasi.SplitEnv("NOEQ"); ok {
		t.Fatal("missing equals should fail")
	}
	k, v, ok = wasi.SplitEnv("EMPTY=")
	if !ok || k != "EMPTY" || v != "" {
		t.Fatalf("empty value: %q %q %v", k, v, ok)
	}
}

func TestErrnoName(t *testing.T) {
	if wasi.ErrnoName(wasi.ErrnoSuccess) != "ESUCCESS" {
		t.Fatal(wasi.ErrnoName(0))
	}
	if wasi.ErrnoName(wasi.ErrnoNoent) != "ENOENT" {
		t.Fatal(wasi.ErrnoName(wasi.ErrnoNoent))
	}
	if wasi.ErrnoName(9999) != "EUNKNOWN" {
		t.Fatal("unknown")
	}
}

func TestPreview1Imports_Dispatchable(t *testing.T) {
	src, err := os.ReadFile("preview1_wasip1.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	if len(wasiapi.Preview1) < 8 {
		t.Fatalf("import table too small: %d", len(wasiapi.Preview1))
	}
	seen := map[string]bool{}
	for _, im := range wasiapi.Preview1 {
		if im.Name == "" || seen[im.Name] {
			t.Fatalf("bad or duplicate import %q", im.Name)
		}
		seen[im.Name] = true
		needle := "//go:wasmimport " + wasiapi.Snapshot + " " + im.Name + "\n"
		if !strings.Contains(text, needle) {
			t.Errorf("preview1_wasip1.go missing %q", needle)
		}
		if im.Name == "proc_exit" {
			if !im.Noreturn || len(im.Results) != 0 {
				t.Errorf("proc_exit should be noreturn void")
			}
		} else if len(im.Results) != 1 || im.Results[0] != "i32" {
			t.Errorf("%s should return i32 errno", im.Name)
		}
	}
	if strings.Contains(text, "wasm-import-module") && strings.Contains(text, "dom") {
		t.Error("DOM import does not belong in the WASI table")
	}
}

func TestRandomBytesHost(t *testing.T) {
	b, err := wasi.RandomBytes(8)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 8 {
		t.Fatalf("len %d", len(b))
	}
	if _, err := wasi.RandomBytes(0); err != nil {
		t.Fatal(err)
	}
	empty, err := wasi.RandomBytes(0)
	if err != nil || empty != nil {
		t.Fatalf("empty %#v %v", empty, err)
	}
}
