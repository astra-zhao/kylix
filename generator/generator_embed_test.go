package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// generator_embed_test.go — [Embed('dir', …)] baking on the Go backend
// (v0.12.0). The emitted init() registers every file; sysutil.ReadFile and the
// boot static handler consult that registry before the filesystem, so a built
// binary needs no views/ or static/ directory beside it.

func TestEmbedEmittedWithAttribute(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "page.tpl"), []byte("<h1>hi</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, errs := compile("[Embed('" + dir + "')]\nprogram P;\nbegin\n  WriteLn(ReadFile('" +
		filepath.Join(dir, "page.tpl") + "'));\nend.")
	assertNoErrors(t, errs)
	assertContains(t, out, "func init() {")
	assertContains(t, out, "stdlib.RegisterEmbedded(")
	assertContains(t, out, "<h1>hi</h1>")
}

// No attribute → no registry, so a program that does not bake anything is
// unchanged.
func TestEmbedAbsentWithoutAttribute(t *testing.T) {
	out, errs := compile("program P;\nbegin\n  WriteLn('hi');\nend.")
	assertNoErrors(t, errs)
	assertNotContains(t, out, "RegisterEmbedded")
}

// The lookup key is the path as the program asks for it, slash-separated.
func TestEmbedKeyIsSlashSeparated(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "views")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "base.tpl"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, errs := compile("[Embed('" + dir + "')]\nprogram P;\nbegin\n  WriteLn(ReadFile('" +
		filepath.ToSlash(filepath.Join(sub, "base.tpl")) + "'));\nend.")
	assertNoErrors(t, errs)
	key := filepath.ToSlash(filepath.Join(sub, "base.tpl"))
	if !strings.Contains(out, `"`+key+`"`) {
		t.Errorf("expected the registry key %q in the generated code", key)
	}
}
