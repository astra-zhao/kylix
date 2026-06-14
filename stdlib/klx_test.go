package stdlib_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
)

// klxDir returns the absolute path to stdlib/klx/.
func klxDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	// file is .../stdlib/klx_test.go; klx/ is a sibling directory
	return filepath.Join(filepath.Dir(file), "klx")
}

func TestKlxDeclarationsAreParseable(t *testing.T) {
	dir := klxDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("cannot read stdlib/klx dir: %v", err)
	}

	var klxFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".klx") {
			klxFiles = append(klxFiles, filepath.Join(dir, e.Name()))
		}
	}

	if len(klxFiles) == 0 {
		t.Fatal("no .klx files found in stdlib/klx/")
	}

	for _, path := range klxFiles {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("cannot read %s: %v", name, err)
			}
			l := lexer.New(string(src))
			p := parser.New(l)
			prog := p.ParseProgram()

			if errs := p.Errors(); len(errs) > 0 {
				for _, e := range errs {
					t.Errorf("parse error: %s", e)
				}
				return
			}
			if prog == nil {
				t.Error("ParseProgram returned nil")
			}
		})
	}
}
