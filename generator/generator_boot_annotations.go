package generator

import (
	"fmt"
	"kylix/ast"
	"strings"
)

type bootHandlerKind int

const (
	bootHandlerFunction bootHandlerKind = iota
	bootHandlerProcedure
)

type bootSecurity struct {
	RequireAuth bool
	Roles       []string
}

type bootRoute struct {
	Method      string
	Path        string
	ClassName   string
	MethodName  string
	HandlerKind bootHandlerKind
	Security    bootSecurity
	SourceLine  int
}

type bootComponent struct {
	ClassName  string
	Kind       string
	SourceLine int
}

type bootInjectField struct {
	OwnerClass string
	FieldName  string
	FieldType  string
	SourceLine int
}

func (g *Generator) scanBootAnnotations(program *ast.Program) {
	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
				classDecl.Name = d.Name
				g.scanBootClass(d.Name, mergeAttributes(d.Attributes, classDecl.Attributes), classDecl)
			}
		case *ast.ClassDecl:
			g.scanBootClass(d.Name, d.Attributes, d)
		}
	}
	if len(g.bootRoutes) > 0 || len(g.bootComponents) > 0 || len(g.bootInjects) > 0 {
		g.imports["kylix/stdlib"] = true
	}
}

func (g *Generator) scanBootClass(className string, attrs []*ast.Attribute, classDecl *ast.ClassDecl) {
	if className == "" || classDecl == nil {
		return
	}
	if service := findAttribute(attrs, "Service", "Component"); service != nil {
		g.bootComponents = append(g.bootComponents, bootComponent{
			ClassName:  className,
			Kind:       service.Name,
			SourceLine: classDecl.Token.Line,
		})
	}
	for _, field := range classDecl.Fields {
		if findAttribute(field.Attributes, "Inject") == nil {
			continue
		}
		fieldType, ok := fieldTypeName(field.Type)
		if !ok {
			continue
		}
		for _, name := range field.Names {
			g.bootInjects = append(g.bootInjects, bootInjectField{
				OwnerClass: className,
				FieldName:  name,
				FieldType:  fieldType,
				SourceLine: field.Token.Line,
			})
		}
	}
	g.scanBootController(className, attrs, classDecl)
}

func (g *Generator) scanBootController(className string, attrs []*ast.Attribute, classDecl *ast.ClassDecl) {
	if className == "" || classDecl == nil {
		return
	}
	controller := findAttribute(attrs, "Controller")
	if controller == nil {
		return
	}
	basePath, ok := attributeStringArg(controller, "")
	if !ok {
		basePath = ""
	}
	for _, method := range classDecl.Methods {
		if method.Body == nil {
			continue
		}
		for _, attr := range method.Attributes {
			httpMethod, ok := isBootRouteAttr(attr.Name)
			if !ok {
				continue
			}
			kind, ok := bootHandlerKindFor(method)
			if !ok {
				continue
			}
			methodPath, ok := attributeStringArg(attr, "/")
			if !ok {
				methodPath = "/"
			}
			security := collectBootSecurity(method.Attributes)
			g.bootRoutes = append(g.bootRoutes, bootRoute{
				Method:      httpMethod,
				Path:        normalizeBootPath(basePath, methodPath),
				ClassName:   className,
				MethodName:  method.Name,
				HandlerKind: kind,
				Security:    security,
				SourceLine:  method.Token.Line,
			})
		}
	}
}

func shortComponentName(className string) string {
	if len(className) > 1 && strings.HasPrefix(className, "T") {
		return className[1:]
	}
	return className
}

func componentVarName(className string) string { return fmt.Sprintf("__kylix_svc_%s", className) }

func controllerVarName(className string) string { return fmt.Sprintf("__kylix_ctrl_%s", className) }

func fieldTypeName(expr ast.Expression) (string, bool) {
	switch t := expr.(type) {
	case *ast.Identifier:
		return t.Value, true
	case *ast.GenericType:
		return t.Base, true
	default:
		return "", false
	}
}

func mergeAttributes(a, b []*ast.Attribute) []*ast.Attribute {
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

func findAttribute(attrs []*ast.Attribute, names ...string) *ast.Attribute {
	for _, attr := range attrs {
		for _, name := range names {
			if strings.EqualFold(attr.Name, name) {
				return attr
			}
		}
	}
	return nil
}

func attributeStringArg(attr *ast.Attribute, fallback string) (string, bool) {
	if attr == nil || len(attr.Args) == 0 {
		return fallback, true
	}
	if s, ok := attr.Args[0].(*ast.StringLiteral); ok {
		return s.Value, true
	}
	return fallback, false
}

func isBootRouteAttr(name string) (string, bool) {
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

func normalizeBootPath(base, sub string) string {
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

func bootHandlerKindFor(method *ast.FunctionDecl) (bootHandlerKind, bool) {
	if method == nil {
		return bootHandlerFunction, false
	}
	if len(method.Parameters) == 1 && method.ReturnType != nil &&
		isBootRequestType(method.Parameters[0].Type) && isBootResponseType(method.ReturnType) {
		return bootHandlerFunction, true
	}
	if len(method.Parameters) == 2 && method.ReturnType == nil && len(method.ReturnTypes) == 0 &&
		isBootRequestType(method.Parameters[0].Type) && isBootResponseType(method.Parameters[1].Type) {
		return bootHandlerProcedure, true
	}
	return bootHandlerFunction, false
}

func isBootRequestType(expr ast.Expression) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Value == "TRequest" || ident.Value == "BootRequest"
	}
	return false
}

func isBootResponseType(expr ast.Expression) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Value == "TResponse" || ident.Value == "BootResponse"
	}
	return false
}

// collectBootSecurity returns auth/role requirements declared on a route
// method via [Authenticated] / [Role('name')] attributes. [Role] implies
// authentication.
func collectBootSecurity(attrs []*ast.Attribute) bootSecurity {
	var s bootSecurity
	for _, attr := range attrs {
		switch strings.ToLower(attr.Name) {
		case "authenticated":
			s.RequireAuth = true
		case "role":
			if role, ok := attributeStringArg(attr, ""); ok && role != "" {
				s.Roles = append(s.Roles, role)
				s.RequireAuth = true
			}
		}
	}
	return s
}

func (g *Generator) emitBootAutoWiring() {
	componentVars := map[string]string{}
	for _, component := range g.bootComponents {
		varName := componentVarName(component.ClassName)
		componentVars[component.ClassName] = varName
		g.writeLine(fmt.Sprintf("%s := &%s{}", varName, component.ClassName))
		g.writeLine(fmt.Sprintf("stdlib.BootRegisterInstance(%q, %s)", component.ClassName, varName))
		shortName := shortComponentName(component.ClassName)
		if shortName != component.ClassName {
			g.writeLine(fmt.Sprintf("stdlib.BootRegisterInstance(%q, %s)", shortName, varName))
		}
	}

	controllerVars := map[string]string{}
	for _, className := range g.uniqueBootControllerClasses() {
		varName := controllerVarName(className)
		controllerVars[className] = varName
		g.writeLine(fmt.Sprintf("%s := &%s{}", varName, className))
	}

	g.emitBootInjectAssignments(componentVars, componentVars)
	g.emitBootInjectAssignments(controllerVars, componentVars)

	for _, route := range g.bootRoutes {
		ctrlVar := controllerVars[route.ClassName]
		if ctrlVar == "" {
			ctrlVar = controllerVarName(route.ClassName)
		}
		g.writeLine(fmt.Sprintf("stdlib.Boot%s(%q, func(req *stdlib.BootRequest) *stdlib.BootResponse {", route.Method, route.Path))
		g.indent++
		if route.Security.RequireAuth {
			g.writeLine("if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }")
		}
		for _, role := range route.Security.Roles {
			g.writeLine(fmt.Sprintf("if __r := stdlib.BootEnforceRole(req, %q); __r != nil { return __r }", role))
		}
		if route.HandlerKind == bootHandlerProcedure {
			g.writeLine("res := stdlib.BootText(200, \"\")")
			g.writeLine(fmt.Sprintf("%s.%s(req, res)", ctrlVar, route.MethodName))
			g.writeLine("return res")
		} else {
			g.writeLine(fmt.Sprintf("return %s.%s(req)", ctrlVar, route.MethodName))
		}
		g.indent--
		g.writeLine("})")
	}
}

func (g *Generator) uniqueBootControllerClasses() []string {
	seen := map[string]bool{}
	var classes []string
	for _, route := range g.bootRoutes {
		if route.ClassName == "" || seen[route.ClassName] {
			continue
		}
		seen[route.ClassName] = true
		classes = append(classes, route.ClassName)
	}
	return classes
}

func (g *Generator) emitBootInjectAssignments(ownerVars, componentVars map[string]string) {
	for _, inject := range g.bootInjects {
		ownerVar := ownerVars[inject.OwnerClass]
		depVar := componentVars[inject.FieldType]
		if ownerVar == "" || depVar == "" {
			continue
		}
		g.writeLine(fmt.Sprintf("%s.%s = %s", ownerVar, inject.FieldName, depVar))
	}
}
