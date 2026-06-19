package lsp_test

import (
	"strings"
	"testing"

	"kylix/pkg/lsp"
)

// LSP incremental synchronization tests (Task 1).

func TestSync_FullReplace(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "program A;\nbegin end.", 1)

	doc := store.Update("file:///a.klx", "program B;\nbegin end.", 2)
	if !strings.Contains(doc.Text, "program B") {
		t.Errorf("expected new text, got: %s", doc.Text)
	}
	if doc.Version != 2 {
		t.Errorf("expected version 2, got %d", doc.Version)
	}
}

func TestSync_RejectStaleVersion(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "v3 text", 3)

	// Stale update with older version — should be rejected.
	doc := store.Update("file:///a.klx", "v1 text", 1)
	if doc.Version != 3 {
		t.Errorf("expected stale update rejected (version stays 3), got %d", doc.Version)
	}
	if doc.Text != "v3 text" {
		t.Errorf("expected text unchanged, got: %s", doc.Text)
	}
}

func TestSync_IncrementalRangeEdit(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "program A;\nbegin\n  WriteLn('hi');\nend.", 1)

	// Replace 'hi' with 'world' on line 2 (0-indexed), characters 11..13
	changes := []lsp.TextDocumentContentChange{
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 2, Character: 11},
				End:   lsp.Position{Line: 2, Character: 13},
			},
			Text: "world",
		},
	}
	doc := store.ApplyChanges("file:///a.klx", 2, changes)

	if !strings.Contains(doc.Text, "WriteLn('world')") {
		t.Errorf("expected 'world' replacement, got: %s", doc.Text)
	}
	if doc.Version != 2 {
		t.Errorf("expected version 2, got %d", doc.Version)
	}
}

func TestSync_IncrementalInsert(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "abc", 1)

	// Insert 'X' at position 1 (between 'a' and 'b')
	changes := []lsp.TextDocumentContentChange{
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 1},
				End:   lsp.Position{Line: 0, Character: 1},
			},
			Text: "X",
		},
	}
	doc := store.ApplyChanges("file:///a.klx", 2, changes)

	if doc.Text != "aXbc" {
		t.Errorf("expected 'aXbc', got %q", doc.Text)
	}
}

func TestSync_IncrementalDelete(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "hello world", 1)

	// Delete characters 5..6 (the space + 'w')
	changes := []lsp.TextDocumentContentChange{
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 5},
				End:   lsp.Position{Line: 0, Character: 7},
			},
			Text: "",
		},
	}
	doc := store.ApplyChanges("file:///a.klx", 2, changes)

	if doc.Text != "helloorld" {
		t.Errorf("expected 'helloorld', got %q", doc.Text)
	}
}

func TestSync_MultipleChangesAtomic(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "abc\ndef\nghi", 1)

	// Two edits in one didChange — should apply sequentially atomically.
	changes := []lsp.TextDocumentContentChange{
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 3},
			},
			Text: "ABC",
		},
		{
			Range: &lsp.Range{
				Start: lsp.Position{Line: 1, Character: 0},
				End:   lsp.Position{Line: 1, Character: 3},
			},
			Text: "DEF",
		},
	}
	doc := store.ApplyChanges("file:///a.klx", 2, changes)

	expected := "ABC\nDEF\nghi"
	if doc.Text != expected {
		t.Errorf("expected %q, got %q", expected, doc.Text)
	}
}

func TestSync_FullReplaceViaApplyChanges(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "old", 1)

	// Range == nil → full document replace
	changes := []lsp.TextDocumentContentChange{
		{Text: "completely new content"},
	}
	doc := store.ApplyChanges("file:///a.klx", 2, changes)

	if doc.Text != "completely new content" {
		t.Errorf("expected full replace, got %q", doc.Text)
	}
}

func TestSync_VersionUnsetWhenNegative(t *testing.T) {
	// version=-1 means "no version provided" (legacy clients). Should still work.
	store := lsp.NewDocumentStore()
	doc := store.Update("file:///a.klx", "hello", -1)
	if doc.Version != 0 {
		t.Errorf("expected default version 0 for unset, got %d", doc.Version)
	}
}
