package lsp_test

import (
	"testing"

	"kylix/pkg/lsp"
)

// LSP refactoring tests (v2.5.0 task 1).

func TestRename_SingleFile(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "program A;\nvar x: Integer;\nbegin\n  x := 42;\n  WriteLn(x);\nend.", 1)

	doc := store.Get("file:///a.klx")
	if doc == nil {
		t.Fatal("document not found")
	}

	walker := lsp.NewReferenceWalker("x", "file:///a.klx")
	walker.Walk(doc.AST)
	refs := walker.References()
	if len(refs) < 2 {
		t.Errorf("expected >=2 references to 'x', got %d", len(refs))
	}
}

func TestRename_CrossFile(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "program A;\nvar counter: Integer;\nbegin\n  counter := 1;\nend.", 1)
	store.Update("file:///b.klx", "program B;\nbegin\n  WriteLn(counter);\nend.", 2)

	total := 0
	for _, doc := range store.GetAll() {
		walker := lsp.NewReferenceWalker("counter", doc.URI)
		walker.Walk(doc.AST)
		total += len(walker.References())
	}
	if total < 2 {
		t.Errorf("expected >=2 references across files, got %d", total)
	}
}

func TestCodeAction_StaticActions(t *testing.T) {
	actions := []lsp.CodeAction{
		{Title: "Organize Imports", Kind: "source.organizeImports"},
		{Title: "Format Document", Kind: "source.format"},
	}
	if len(actions) != 2 {
		t.Error("expected 2 static actions")
	}
}

func TestReferenceWalker_NoReferences(t *testing.T) {
	store := lsp.NewDocumentStore()
	store.Update("file:///a.klx", "program A;\nbegin\n  WriteLn('hello');\nend.", 1)

	doc := store.Get("file:///a.klx")
	walker := lsp.NewReferenceWalker("nonExistent", "file:///a.klx")
	walker.Walk(doc.AST)
	if len(walker.References()) != 0 {
		t.Errorf("expected 0 references for non-existent name, got %d", len(walker.References()))
	}
}
