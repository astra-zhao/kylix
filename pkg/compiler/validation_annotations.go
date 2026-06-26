package compiler

import (
	"fmt"
	"kylix/ast"
	"strings"
)

// CheckValidationAnnotations validates [Required]/[Email]/[Min]/[Max]/[MinLen]/[MaxLen]
// usage across a project before code generation. It catches obvious mis-use such
// as missing integer args or numeric validators on string fields.
func CheckValidationAnnotations(programs []*ast.Program, files []string) []Diagnostic {
	var diags []Diagnostic
	for i, program := range programs {
		file := ""
		if i < len(files) {
			file = files[i]
		}
		for _, decl := range program.Declarations {
			switch d := decl.(type) {
			case *ast.TypeDecl:
				if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
					diags = append(diags, checkValidationClass(file, d.Name, classDecl)...)
				}
			case *ast.ClassDecl:
				diags = append(diags, checkValidationClass(file, d.Name, d)...)
			}
		}
	}
	return diags
}

func checkValidationAnnotations(program *ast.Program, file string) []Diagnostic {
	return CheckValidationAnnotations([]*ast.Program{program}, []string{file})
}

func checkValidationClass(file, className string, classDecl *ast.ClassDecl) []Diagnostic {
	if className == "" || classDecl == nil {
		return nil
	}
	var diags []Diagnostic
	for _, field := range classDecl.Fields {
		fieldType, _ := bootAnnotationFieldTypeName(field.Type)
		for _, attr := range field.Attributes {
			name := strings.ToLower(attr.Name)
			switch name {
			case "required":
				// no arg required; valid on any field type
			case "email":
				if !isValidationStringField(fieldType) {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidValidation,
						fmt.Sprintf("[Email] requires a String field, but %s.%s is %s", className, joinNames(field.Names), fieldType)))
				}
			case "minlen", "maxlen":
				if _, ok := validationIntArg(attr); !ok {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidValidation,
						fmt.Sprintf("[%s] requires an integer argument", attr.Name)))
				} else if !isValidationStringField(fieldType) {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidValidation,
						fmt.Sprintf("[%s] requires a String field, but %s.%s is %s", attr.Name, className, joinNames(field.Names), fieldType)))
				}
			case "min", "max":
				if _, ok := validationIntArg(attr); !ok {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidValidation,
						fmt.Sprintf("[%s] requires an integer argument", attr.Name)))
				} else if !isValidationIntegerField(fieldType) {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidValidation,
						fmt.Sprintf("[%s] requires an Integer field, but %s.%s is %s", attr.Name, className, joinNames(field.Names), fieldType)))
				}
			}
		}
	}
	return diags
}

func validationIntArg(attr *ast.Attribute) (int64, bool) {
	if attr == nil || len(attr.Args) == 0 {
		return 0, false
	}
	if lit, ok := attr.Args[0].(*ast.IntegerLiteral); ok {
		return lit.Value, true
	}
	return 0, false
}

func isValidationStringField(t string) bool { return t == "String" || t == "string" }

func isValidationIntegerField(t string) bool {
	switch t {
	case "Integer", "Int", "int", "int64", "LongInt", "Cardinal", "Word", "Byte":
		return true
	}
	return false
}

func joinNames(names []string) string { return strings.Join(names, ",") }
