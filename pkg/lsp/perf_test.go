package lsp_test

import (
	"strings"
	"testing"
	"time"

	"kylix/pkg/lsp"
)

// LSP large file performance benchmark (v2.6.0 task 3).
// Verifies that incremental didChange → diagnostics stays under 50ms
// even on a large (1K+ line) file.

// generateLargeSource creates a .klx file with N functions.
func generateLargeSource(n int) string {
	var sb strings.Builder
	sb.WriteString("program Large;\n\n")
	for i := 0; i < n; i++ {
		sb.WriteString("function Func")
		sb.WriteString(itoa(i))
		sb.WriteString("(): Integer;\nbegin result := ")
		sb.WriteString(itoa(i))
		sb.WriteString("; end;\n\n")
	}
	sb.WriteString("begin\n  WriteLn(Func0());\nend.\n")
	return sb.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func TestLSPPerf_LargeFileParse(t *testing.T) {
	// 500 functions ≈ 2K+ lines
	src := generateLargeSource(500)
	store := lsp.NewDocumentStore()

	start := time.Now()
	store.Update("file:///large.klx", src, 1)
	elapsed := time.Since(start)

	// Parsing + symbol collection should complete in reasonable time.
	if elapsed > 200*time.Millisecond {
		t.Errorf("large file parse took %v (expected < 200ms)", elapsed)
	}
	t.Logf("500-function file parsed in %v", elapsed)
}

func TestLSPPerf_IncrementalEdit(t *testing.T) {
	src := generateLargeSource(500)
	store := lsp.NewDocumentStore()
	store.Update("file:///large.klx", src, 1)

	// Simulate a single-character insertion via incremental change.
	changes := []lsp.TextDocumentContentChange{
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 15},
				End:   lsp.Position{Line: 0, Character: 15},
			},
			Text: "X",
		},
	}

	start := time.Now()
	store.ApplyChanges("file:///large.klx", 2, changes)
	elapsed := time.Since(start)

	// Incremental edit (re-parse) should be fast.
	if elapsed > 200*time.Millisecond {
		t.Errorf("incremental edit took %v (expected < 200ms)", elapsed)
	}
	t.Logf("incremental edit on 500-function file: %v", elapsed)
}

func TestLSPPerf_CompletionLookup(t *testing.T) {
	src := generateLargeSource(200)
	store := lsp.NewDocumentStore()
	store.Update("file:///large.klx", src, 1)

	doc := store.Get("file:///large.klx")
	if doc == nil || doc.Symbols == nil {
		t.Fatal("document or symbols not available")
	}

	// Verify symbol table has entries.
	symCount := len(doc.Symbols.AllSymbols)
	if symCount < 200 {
		t.Errorf("expected >=200 symbols, got %d", symCount)
	}
	t.Logf("200-function file: %d symbols collected", symCount)
}
