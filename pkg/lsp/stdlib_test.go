package lsp_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"kylix/pkg/lsp"
)

// klxTestDir returns stdlib/klx/ for use in tests.
func klxTestDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	// pkg/lsp/stdlib_test.go → repo root → stdlib/klx
	root := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	d := filepath.Join(root, "stdlib", "klx")
	if _, err := os.Stat(d); err != nil {
		t.Skipf("stdlib/klx not found: %s", d)
	}
	return d
}

// symbolNames returns all names from a Document's symbol table.
func symbolNames(doc *lsp.Document) map[string]bool {
	names := make(map[string]bool)
	if doc.Symbols == nil {
		return names
	}
	for _, sym := range doc.Symbols.AllSymbols {
		names[sym.Name] = true
	}
	return names
}

func TestStdlibKlxFilesExist(t *testing.T) {
	dir := klxTestDir(t)
	for _, name := range []string{"sysutil.klx", "datetime.klx", "regex.klx", "jsonutil.klx"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("missing stdlib file: %s", name)
		}
	}
}

func TestLoadStdlibSymbols_Sysutil(t *testing.T) {
	klxDir := klxTestDir(t)
	// Point KYLIX_HOME so loadStdlibSymbols can find stdlib/klx/
	t.Setenv("KYLIX_HOME", filepath.Dir(filepath.Dir(klxDir)))

	src := "program Test;\nuses sysutil;\nbegin end."
	doc := lsp.NewDocument("file:///test.klx", src)

	syms := symbolNames(doc)

	// sysutil.klx declares these functions
	for _, want := range []string{"GetEnv", "ReadFile", "WriteFile", "PathJoin"} {
		if !syms[want] && !syms["sysutil."+want] {
			t.Errorf("expected symbol %q from sysutil.klx, not found", want)
		}
	}
}

func TestLoadStdlibSymbols_Datetime(t *testing.T) {
	klxDir := klxTestDir(t)
	t.Setenv("KYLIX_HOME", filepath.Dir(filepath.Dir(klxDir)))

	src := "program Test;\nuses datetime;\nbegin end."
	doc := lsp.NewDocument("file:///test.klx", src)

	syms := symbolNames(doc)

	for _, want := range []string{"Now", "Today", "ParseDate"} {
		if !syms[want] && !syms["datetime."+want] {
			t.Errorf("expected symbol %q from datetime.klx, not found", want)
		}
	}
}

func TestLoadStdlibSymbols_NoUses(t *testing.T) {
	klxTestDir(t) // ensure stdlib/klx exists but don't setenv
	src := "program Test;\nbegin end."
	doc := lsp.NewDocument("file:///test.klx", src)
	// Should not panic and symbols may be nil or empty
	if doc.Symbols == nil {
		t.Error("expected non-nil symbol table even with no uses")
	}
}
