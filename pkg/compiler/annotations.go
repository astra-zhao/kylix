package compiler

import (
	"fmt"
	"kylix/ast"
	"strings"
)

type bootAnnotationClass struct {
	File      string
	ClassName string
	Attrs     []*ast.Attribute
	Class     *ast.ClassDecl
}

type bootAnnotationRouteKey struct {
	Method string
	Path   string
}

type bootAnnotationRouteInfo struct {
	File      string
	ClassName string
	Method    string
	Path      string
	Line      int
}

// CheckBootAnnotations validates KylixBoot annotations across a project before
// code generation. It catches framework contract errors that the generator would
// otherwise have to skip conservatively.
func CheckBootAnnotations(programs []*ast.Program, files []string) []Diagnostic {
	classes := collectBootAnnotationClasses(programs, files)
	return validateBootAnnotationClasses(classes)
}

func checkBootAnnotations(program *ast.Program, file string) []Diagnostic {
	return CheckBootAnnotations([]*ast.Program{program}, []string{file})
}

func collectBootAnnotationClasses(programs []*ast.Program, files []string) []bootAnnotationClass {
	var classes []bootAnnotationClass
	for i, program := range programs {
		file := ""
		if i < len(files) {
			file = files[i]
		}
		for _, decl := range program.Declarations {
			switch d := decl.(type) {
			case *ast.TypeDecl:
				if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
					classes = append(classes, bootAnnotationClass{
						File:      file,
						ClassName: d.Name,
						Attrs:     mergeBootAnnotationAttrs(d.Attributes, classDecl.Attributes),
						Class:     classDecl,
					})
				}
			case *ast.ClassDecl:
				classes = append(classes, bootAnnotationClass{
					File:      file,
					ClassName: d.Name,
					Attrs:     d.Attributes,
					Class:     d,
				})
			}
		}
	}
	return classes
}

func validateBootAnnotationClasses(classes []bootAnnotationClass) []Diagnostic {
	var diags []Diagnostic
	components := map[string]bool{}
	for _, class := range classes {
		if findBootAnnotation(class.Attrs, "Service", "Component") != nil {
			components[class.ClassName] = true
		}
	}

	routes := map[bootAnnotationRouteKey]bootAnnotationRouteInfo{}
	for _, class := range classes {
		controller := findBootAnnotation(class.Attrs, "Controller")
		if controller != nil {
			basePath, ok := bootAnnotationStringArg(controller, "")
			if !ok {
				diags = append(diags, NewError(class.File, controller.Token.Line, controller.Token.Column,
					ErrInvalidAnnotation, "[Controller] path argument must be a string literal"))
				basePath = ""
			}
			for _, method := range class.Class.Methods {
				if method.Body == nil {
					continue
				}
				for _, attr := range method.Attributes {
					httpMethod, ok := bootAnnotationRouteMethod(attr.Name)
					if !ok {
						continue
					}
					methodPath, ok := bootAnnotationStringArg(attr, "/")
					if !ok {
						diags = append(diags, NewError(class.File, attr.Token.Line, attr.Token.Column,
							ErrInvalidAnnotation, fmt.Sprintf("[%s] path argument must be a string literal", attr.Name)))
						methodPath = "/"
					}
					if !bootAnnotationSupportedHandler(method) {
						diags = append(diags, NewErrorHint(class.File, method.Token.Line, method.Token.Column,
							ErrUnsupportedHandler,
							fmt.Sprintf("KylixBoot route handler %s.%s must be function(req: TRequest): TResponse or procedure(req: TRequest; res: TResponse)", class.ClassName, method.Name),
							"Use either: function Method(req: TRequest): TResponse; or procedure Method(req: TRequest; res: TResponse);"))
					}
					for _, mattr := range method.Attributes {
						if !strings.EqualFold(mattr.Name, "Body") {
							continue
						}
						if len(mattr.Args) == 0 {
							diags = append(diags, NewError(class.File, mattr.Token.Line, mattr.Token.Column,
								ErrBodyBinding, "[Body] requires an entity class argument"))
						} else if _, ok := mattr.Args[0].(*ast.Identifier); !ok {
							diags = append(diags, NewError(class.File, mattr.Token.Line, mattr.Token.Column,
								ErrBodyBinding, "[Body] argument must be an entity class identifier"))
						}
					}
					path := bootAnnotationNormalizePath(basePath, methodPath)
					key := bootAnnotationRouteKey{Method: httpMethod, Path: path}
					if prev, exists := routes[key]; exists {
						diags = append(diags, NewError(class.File, attr.Token.Line, attr.Token.Column,
							ErrDuplicateRoute,
							fmt.Sprintf("duplicate KylixBoot route %s %s; already registered by %s.%s in %s:%d",
								httpMethod, path, prev.ClassName, prev.Method, prev.File, prev.Line)))
					} else {
						routes[key] = bootAnnotationRouteInfo{
							File:      class.File,
							ClassName: class.ClassName,
							Method:    method.Name,
							Path:      path,
							Line:      method.Token.Line,
						}
					}
				}
			}
		}

		if controller == nil && findBootAnnotation(class.Attrs, "Service", "Component") == nil {
			continue
		}
		for _, field := range class.Class.Fields {
			if findBootAnnotation(field.Attributes, "Inject") == nil {
				continue
			}
			fieldType, ok := bootAnnotationFieldTypeName(field.Type)
			if !ok || !components[fieldType] {
				for _, name := range field.Names {
					diags = append(diags, NewErrorHint(class.File, field.Token.Line, field.Token.Column,
						ErrMissingInjectTarget,
						fmt.Sprintf("[Inject] field %s.%s has no matching [Service] or [Component] for type %s", class.ClassName, name, fieldType),
						fmt.Sprintf("Annotate type %s with [Service] or [Component], or remove [Inject].", fieldType)))
				}
			}
		}
	}
	return diags
}

func mergeBootAnnotationAttrs(a, b []*ast.Attribute) []*ast.Attribute {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	merged := make([]*ast.Attribute, 0, len(a)+len(b))
	merged = append(merged, a...)
	merged = append(merged, b...)
	return merged
}

func findBootAnnotation(attrs []*ast.Attribute, names ...string) *ast.Attribute {
	for _, attr := range attrs {
		for _, name := range names {
			if strings.EqualFold(attr.Name, name) {
				return attr
			}
		}
	}
	return nil
}

func bootAnnotationStringArg(attr *ast.Attribute, fallback string) (string, bool) {
	if attr == nil || len(attr.Args) == 0 {
		return fallback, true
	}
	if s, ok := attr.Args[0].(*ast.StringLiteral); ok {
		return s.Value, true
	}
	return fallback, false
}

func bootAnnotationRouteMethod(name string) (string, bool) {
	switch strings.ToLower(name) {
	case "get":
		return "GET", true
	case "post":
		return "POST", true
	case "put":
		return "PUT", true
	case "delete":
		return "DELETE", true
	}
	return "", false
}

func bootAnnotationNormalizePath(base, sub string) string {
	base = strings.TrimSpace(base)
	sub = strings.TrimSpace(sub)
	if base == "" {
		if sub == "" {
			return "/"
		}
		if strings.HasPrefix(sub, "/") {
			return sub
		}
		return "/" + sub
	}
	if sub == "" || sub == "/" {
		if strings.HasPrefix(base, "/") {
			return strings.TrimRight(base, "/")
		}
		return "/" + strings.TrimRight(base, "/")
	}
	return "/" + strings.Trim(strings.TrimRight(base, "/")+"/"+strings.TrimLeft(sub, "/"), "/")
}

func bootAnnotationSupportedHandler(method *ast.FunctionDecl) bool {
	if method == nil {
		return false
	}
	if len(method.Parameters) == 1 && method.ReturnType != nil &&
		bootAnnotationIsRequestType(method.Parameters[0].Type) && bootAnnotationIsResponseType(method.ReturnType) {
		return true
	}
	if len(method.Parameters) == 2 && method.ReturnType == nil && len(method.ReturnTypes) == 0 &&
		bootAnnotationIsRequestType(method.Parameters[0].Type) && bootAnnotationIsResponseType(method.Parameters[1].Type) {
		return true
	}
	return false
}

func bootAnnotationIsRequestType(expr ast.Expression) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Value == "TRequest" || ident.Value == "BootRequest"
	}
	return false
}

func bootAnnotationIsResponseType(expr ast.Expression) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Value == "TResponse" || ident.Value == "BootResponse"
	}
	return false
}

func bootAnnotationFieldTypeName(expr ast.Expression) (string, bool) {
	switch t := expr.(type) {
	case *ast.Identifier:
		return t.Value, true
	case *ast.GenericType:
		return t.Base, true
	default:
		return "", false
	}
}
