package entitymetaapi

import (
	"testing"

	"kylix/ast"
)

// encoding_test.go — the shared derivation both backends call. If these rules
// change, the Go and LLVM emitters change together (they do not reimplement
// them), so this file is the single place the encoding is pinned.

func attr(name string, args ...ast.Expression) *ast.Attribute {
	return &ast.Attribute{Name: name, Args: args}
}

func str(s string) ast.Expression { return &ast.StringLiteral{Value: s} }

func intLit(v int64) ast.Expression { return &ast.IntegerLiteral{Value: v} }

func TestKind(t *testing.T) {
	cases := []struct{ field, typ, want string }{
		{"Id", "Integer", "number"},
		{"IsActive", "Boolean", "checkbox"},
		{"Name", "String", "text"},
		{"Score", "Real", "number"},
		{"Password", "String", "password"},
		{"password_hash", "String", "password"}, // name match, case-insensitive
		{"UserPassword", "String", "password"},
		{"Notes", "String", "text"},
	}
	for _, c := range cases {
		if got := Kind(c.field, c.typ); got != c.want {
			t.Errorf("Kind(%q, %q) = %q, want %q", c.field, c.typ, got, c.want)
		}
	}
}

func TestFieldFlags(t *testing.T) {
	cases := []struct {
		name  string
		attrs []*ast.Attribute
		kind  string
		want  string
	}{
		{"plain", nil, "text", ""},
		{"required", []*ast.Attribute{attr("Required")}, "text", "required"},
		{"email", []*ast.Attribute{attr("Email")}, "text", "email"},
		{
			"lengths",
			[]*ast.Attribute{attr("MinLen", intLit(3)), attr("MaxLen", intLit(32))},
			"text", "minlen=3,maxlen=32",
		},
		{
			"numbers",
			[]*ast.Attribute{attr("Min", intLit(1)), attr("Max", intLit(9))},
			"number", "min=1,max=9",
		},
		{
			"markers",
			[]*ast.Attribute{attr("Searchable"), attr("Nullable")},
			"text", "nullable,searchable",
		},
		{
			"hidden",
			[]*ast.Attribute{attr("Hidden")},
			"text", "hidden",
		},
		{
			"default seeds the create form",
			[]*ast.Attribute{attr("Default", str("1"))},
			"checkbox", "default=1",
		},
		{
			"password is always secret",
			[]*ast.Attribute{attr("Required")},
			"password", "required,secret",
		},
		{
			"validation order is fixed",
			[]*ast.Attribute{attr("MaxLen", intLit(8)), attr("Required"), attr("MinLen", intLit(2))},
			"text", "required,minlen=2,maxlen=8",
		},
	}
	for _, c := range cases {
		if got := FieldFlags(c.attrs, c.kind); got != c.want {
			t.Errorf("%s: FieldFlags = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestListableAndFormable(t *testing.T) {
	if !Listable(nil, "text") {
		t.Error("a plain text field must be listable by default")
	}
	if Listable(nil, "password") {
		t.Error("password fields must never be listable")
	}
	hidden := []*ast.Attribute{attr("Hidden")}
	if Listable(hidden, "text") {
		t.Error("[Hidden] must remove the field from lists")
	}
	if Formable(hidden) {
		t.Error("[Hidden] must remove the field from forms")
	}
	if !Formable(nil) {
		t.Error("a plain field must be formable by default")
	}
}

func TestHelpers(t *testing.T) {
	a := attr("Label", str("Users"))
	if got, ok := StringArg(a); !ok || got != "Users" {
		t.Errorf("StringArg = (%q, %v)", got, ok)
	}
	if _, ok := StringArg(attr("Label")); ok {
		t.Error("StringArg must report false for a missing argument")
	}
	if got, ok := IntArg(attr("Min", intLit(5))); !ok || got != 5 {
		t.Errorf("IntArg = (%d, %v)", got, ok)
	}
	if FindAttribute([]*ast.Attribute{a}, "label") == nil {
		t.Error("FindAttribute must match case-insensitively")
	}
	if FindAttribute([]*ast.Attribute{a}, "Other") != nil {
		t.Error("FindAttribute must return nil for an absent attribute")
	}
}
