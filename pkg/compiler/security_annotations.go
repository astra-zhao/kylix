package compiler

import (
	"fmt"
	"kylix/ast"
	"strings"
)

// CheckSecurityAnnotations validates [Authenticated] / [Role] usage on
// controller methods before code generation. It catches missing/invalid role
// arguments and security annotations applied to non-route methods.
func CheckSecurityAnnotations(programs []*ast.Program, files []string) []Diagnostic {
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
					attrs := mergeBootAnnotationAttrs(d.Attributes, classDecl.Attributes)
					diags = append(diags, checkSecurityClass(file, d.Name, attrs, classDecl)...)
				}
			case *ast.ClassDecl:
				diags = append(diags, checkSecurityClass(file, d.Name, d.Attributes, d)...)
			}
		}
	}
	return diags
}

func checkSecurityAnnotations(program *ast.Program, file string) []Diagnostic {
	return CheckSecurityAnnotations([]*ast.Program{program}, []string{file})
}

func checkSecurityClass(file, className string, classAttrs []*ast.Attribute, classDecl *ast.ClassDecl) []Diagnostic {
	if className == "" || classDecl == nil {
		return nil
	}
	var diags []Diagnostic
	isController := findBootAnnotation(classAttrs, "Controller") != nil
	for _, method := range classDecl.Methods {
		hasRoute := false
		for _, attr := range method.Attributes {
			if _, ok := bootAnnotationRouteMethod(attr.Name); ok {
				hasRoute = true
				break
			}
		}
		for _, attr := range method.Attributes {
			name := strings.ToLower(attr.Name)
			if name != "authenticated" && name != "role" {
				continue
			}
			if !isController || !hasRoute {
				diags = append(diags, NewErrorHint(file, attr.Token.Line, attr.Token.Column,
					ErrInvalidSecurity,
					fmt.Sprintf("[%s] on %s.%s requires a route method inside a [Controller] class", attr.Name, className, method.Name),
					"Apply [Authenticated]/[Role] alongside [Get]/[Post]/[Put]/[Delete] on a controller method."))
				continue
			}
			if name == "role" {
				if len(attr.Args) == 0 {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidSecurity, "[Role] requires a string role name argument"))
					continue
				}
				if _, ok := attr.Args[0].(*ast.StringLiteral); !ok {
					diags = append(diags, NewError(file, attr.Token.Line, attr.Token.Column,
						ErrInvalidSecurity, "[Role] argument must be a string literal"))
				}
			}
		}
	}
	return diags
}
