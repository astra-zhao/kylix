package llvmgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// embed tests — v0.12.0. A file-level [Embed('dir', …)] attribute bakes those
// directories into the binary; the lookup is emitted only when the attribute is
// present, so the tutorials (and the bootstrap fixed point) are untouched.

func TestEmbed_AbsentWithoutAttribute(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  WriteLn(ReadFile('views/base.tpl'));
end.`)
	for _, unwanted := range []string{
		"@__kylix_embed_get",
		"@__kylix_embed_init",
		"@__kylix_embed_names",
	} {
		if strings.Contains(ir, unwanted) {
			t.Errorf("program without [Embed] contains %q", unwanted)
		}
	}
}

// With the attribute the tables, the lookup and the fill call are emitted; the
// ReadFile path consults them before touching the filesystem.
func TestEmbed_EmittedWithAttribute(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "page.tpl"), []byte("<h1>hi</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	ir := generateIR(t, `[Embed('`+dir+`')]
program p;
uses sysutil;
begin
  WriteLn(ReadFile('`+filepath.Join(dir, "page.tpl")+`'));
end.`)
	assertIRContains(t, ir, "@__kylix_embed_count = global i64 1")
	assertIRContains(t, ir, "@__kylix_embed_names = global [1 x ptr] zeroinitializer")
	assertIRContains(t, ir, "define ptr @__kylix_embed_get(ptr %name)")
	assertIRContains(t, ir, "define void @__kylix_embed_init()")
	assertIRContains(t, ir, "call void @__kylix_embed_init()")
	// the ReadFile path asks the table first, then falls back to fopen
	assertIRContains(t, ir, "call ptr @__kylix_embed_get(ptr")
	assertIRContains(t, ir, "call ptr @fopen(ptr")
}
