package entitymetaapi

import (
	"strconv"
	"strings"

	"kylix/ast"
)

// encoding.go — the field-metadata encoding shared by both backends.
//
// The Go generator (generator/generator_entitymeta.go) and the LLVM backend
// (pkg/llvmgen/orm_annotations.go) must derive byte-identical registration
// arguments, otherwise the CRUD engine renders different pages depending on
// which backend compiled it. Both therefore call the helpers below rather than
// reimplementing the rules.

// validationFlagNames are the field annotations that become validation flags,
// in the canonical order they are emitted.
var validationFlagNames = []string{"required", "email", "min", "max", "minlen", "maxlen"}

// Kind maps a field to its form-control kind. Password detection is by field
// name — the column holds a hash, so the control must be a password box.
func Kind(fieldName, fieldType string) string {
	if strings.Contains(strings.ToLower(fieldName), "password") {
		return "password"
	}
	switch fieldType {
	case "Boolean", "boolean":
		return "checkbox"
	case "Integer", "int64", "Int64", "Real", "Double", "Single":
		return "number"
	}
	return "text"
}

// FieldFlags renders the field's flags string: validation rules first (the
// engine's form validator consumes them), then the CRUD markers. The order is
// fixed so both backends emit identical strings.
func FieldFlags(attrs []*ast.Attribute, kind string) string {
	var flags []string
	// Canonical order, not source order: reordering annotations in the class
	// must not change the emitted metadata string.
	for _, name := range validationFlagNames {
		attr := FindAttribute(attrs, name)
		if attr == nil {
			continue
		}
		if name == "required" || name == "email" {
			flags = append(flags, name)
			continue
		}
		if v, ok := IntArg(attr); ok {
			flags = append(flags, name+"="+strconv.FormatInt(v, 10))
		}
	}
	if FindAttribute(attrs, "Nullable") != nil {
		flags = append(flags, "nullable")
	}
	// [Default('v')] seeds the create form's initial value (an unchecked
	// checkbox submits nothing, so a field that should start checked needs one).
	if def, ok := StringArg(FindAttribute(attrs, "Default")); ok && def != "" {
		flags = append(flags, "default="+def)
	}
	if FindAttribute(attrs, "Searchable") != nil {
		flags = append(flags, "searchable")
	}
	if FindAttribute(attrs, "Hidden") != nil {
		flags = append(flags, "hidden")
	}
	// v0.12.0: [Unique] becomes a schema constraint when tables are generated
	// from metadata (and a hint for the generated forms).
	if FindAttribute(attrs, "Unique") != nil {
		flags = append(flags, "unique")
	}
	if kind == "password" {
		flags = append(flags, "secret")
	}
	return strings.Join(flags, ",")
}

// Listable reports whether a field appears in list pages: password fields never
// do, and [Hidden] removes a column from both the list and the form.
func Listable(attrs []*ast.Attribute, kind string) bool {
	return FindAttribute(attrs, "Hidden") == nil && kind != "password"
}

// Formable reports whether a field appears in create/edit forms.
func Formable(attrs []*ast.Attribute) bool {
	return FindAttribute(attrs, "Hidden") == nil
}

// FindAttribute returns the first attribute matching one of names
// (case-insensitive).
func FindAttribute(attrs []*ast.Attribute, names ...string) *ast.Attribute {
	for _, attr := range attrs {
		for _, name := range names {
			if strings.EqualFold(attr.Name, name) {
				return attr
			}
		}
	}
	return nil
}

// StringArg returns the attribute's first string argument.
func StringArg(attr *ast.Attribute) (string, bool) {
	if attr == nil || len(attr.Args) == 0 {
		return "", false
	}
	if s, ok := attr.Args[0].(*ast.StringLiteral); ok {
		return s.Value, true
	}
	return "", false
}

// IntArg returns the attribute's first integer argument.
func IntArg(attr *ast.Attribute) (int64, bool) {
	if attr == nil || len(attr.Args) == 0 {
		return 0, false
	}
	if lit, ok := attr.Args[0].(*ast.IntegerLiteral); ok {
		return lit.Value, true
	}
	return 0, false
}

// Separators are the characters the runtime encoding reserves. Labels carrying
// them are rejected at compile time (pkg/compiler.CheckORMAnnotations) so the
// runtime never has to escape.
const Separators = "|\n\r"
