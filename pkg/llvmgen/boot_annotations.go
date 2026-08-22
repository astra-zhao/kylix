// boot_annotations.go — KylixBoot annotation scanning + auto-wiring for the
// LLVM backend (v6.1.0).
//
// Mirrors the Go backend's generator/generator_boot_annotations.go: the
// [Controller] / [Service] / [Component] / [Inject] / [Get] / [Post] / [Put] /
// [Delete] / [Authenticated] / [Role] attributes are collected into
// Generator.bootRoutes/bootComponents/bootInjects, then lowered to auto-wiring
// IR before the user's main statements. There is no HTTP server in the LLVM
// backend, so route registration is a no-op — the handlers are wrapped in
// named @__kylix_boot_handler_<i> functions (the "closure" equivalent) that
// load the controller instance from a module global and call the method. The
// wrappers are only ever *defined*, never invoked, which is exactly what the
// tutorial examples need (their main just WriteLn's an "OK" marker).
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
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
	Method      string // GET / POST / PUT / DELETE
	Path        string // normalized, e.g. "/api/hello"
	ClassName   string
	MethodName  string
	HandlerKind bootHandlerKind
	Security    bootSecurity
	SourceLine  int
}

type bootComponent struct {
	ClassName  string
	Kind       string // Service / Component
	SourceLine int
}

type bootInject struct {
	OwnerClass string
	FieldName  string
	FieldType  string
	SourceLine int
}

// bootWrapper pairs a route with the module global holding its controller
// instance so the wrapper define can load it (the no-closure capture).
type bootWrapper struct {
	Index      int
	Route      bootRoute
	CtrlGlobal string // "@__kylix_boot_ctrl_<Class>"
}

// scanBootAnnotations walks a parsed program and collects KylixBoot annotations.
// Called once per program during emitProgram (after class names are pre-registered).
func (g *Generator) scanBootAnnotations(prog *ast.Program) {
	for _, decl := range prog.Declarations {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			if cd, ok := d.Type.(*ast.ClassDecl); ok {
				cd.Name = d.Name
				g.scanBootClass(d.Name, mergeAttributes(d.Attributes, cd.Attributes), cd)
			}
		case *ast.ClassDecl:
			g.scanBootClass(d.Name, d.Attributes, d)
		}
	}
}

func (g *Generator) scanBootClass(className string, attrs []*ast.Attribute, cd *ast.ClassDecl) {
	if className == "" || cd == nil {
		return
	}
	if svc := findAttribute(attrs, "Service", "Component"); svc != nil {
		g.bootComponents = append(g.bootComponents, bootComponent{
			ClassName: className, Kind: svc.Name, SourceLine: cd.Token.Line,
		})
	}
	for _, field := range cd.Fields {
		if findAttribute(field.Attributes, "Inject") == nil {
			continue
		}
		ft, ok := fieldTypeName(field.Type)
		if !ok {
			continue
		}
		for _, name := range field.Names {
			g.bootInjects = append(g.bootInjects, bootInject{
				OwnerClass: className, FieldName: name, FieldType: ft,
				SourceLine: field.Token.Line,
			})
		}
	}
	g.scanBootController(className, attrs, cd)
}

func (g *Generator) scanBootController(className string, attrs []*ast.Attribute, cd *ast.ClassDecl) {
	if className == "" || cd == nil {
		return
	}
	controller := findAttribute(attrs, "Controller")
	if controller == nil {
		return
	}
	basePath, _ := attributeStringArg(controller, "")
	for _, method := range cd.Methods {
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
			methodPath, _ := attributeStringArg(attr, "/")
			g.bootRoutes = append(g.bootRoutes, bootRoute{
				Method:      httpMethod,
				Path:        normalizeBootPath(basePath, methodPath),
				ClassName:   className,
				MethodName:  method.Name,
				HandlerKind: kind,
				Security:    collectBootSecurity(method.Attributes),
				SourceLine:  method.Token.Line,
			})
		}
	}
}

// emitBootGlobals emits the module-level controller instance globals. Called
// once per program before collectGlobals so the symbols exist at module scope.
func (g *Generator) emitBootGlobals() {
	seen := map[string]bool{}
	for _, route := range g.bootRoutes {
		if seen[route.ClassName] {
			continue
		}
		seen[route.ClassName] = true
		g.line(fmt.Sprintf("@__kylix_boot_ctrl_%s = global ptr null", route.ClassName))
	}
	// v6.6.0: the runtime route table that @__kylix_boot_Boot<M> writes and
	// @__kylix_boot_BootRun dispatches against. Fixed 64-entry capacity
	// (enough for the tutorial controllers; overflows are dropped).
	if len(g.bootRoutes) > 0 {
		g.line("@__kylix_boot_routes = global [64 x { ptr, ptr, ptr }] zeroinitializer")
		g.line("@__kylix_boot_nroutes = global i64 0")
	}
}

// emitBootAutoWiring emits the auto-wiring instructions at the top of main(),
// before any user statement (mirrors the Go backend, which generates
// `ctrl := &THelloController{}` + stdlib.BootGET(path, handler) there). The
// route handlers are named wrapper functions (see emitPendingBootDefines),
// passed to the no-op Boot<M> registration calls.
func (g *Generator) emitBootAutoWiring() error {
	if len(g.bootRoutes) == 0 && len(g.bootComponents) == 0 && len(g.bootInjects) == 0 {
		return nil
	}
	g.line("  ; --- KylixBoot auto-wiring (v6.1.0) ---")

	// 1. Instantiate components ([Service]/[Component]) and register them under
	//    their full class name and short name (mirrors the Go backend).
	compRegs := make(map[string]string) // className → SSA instance register
	for _, comp := range g.bootComponents {
		inst, err := g.emitConstructor(comp.ClassName)
		if err != nil {
			return err
		}
		compRegs[comp.ClassName] = inst
		g.enqueueStdlib("boot", "BootRegisterInstance", "BootRegisterInstance", 0)
		g.line(fmt.Sprintf("  call void @__kylix_boot_BootRegisterInstance(ptr %s, ptr %s)",
			g.addString(comp.ClassName), inst))
		short := shortComponentName(comp.ClassName)
		if short != comp.ClassName {
			g.line(fmt.Sprintf("  call void @__kylix_boot_BootRegisterInstance(ptr %s, ptr %s)",
				g.addString(short), inst))
		}
	}

	// 2. Instantiate controllers and stash them in the module globals so the
	//    wrapper functions can load them (no-closure capture).
	ctrlGlobals := make(map[string]string) // className → "@__kylix_boot_ctrl_<Class>"
	for _, route := range g.bootRoutes {
		if _, ok := ctrlGlobals[route.ClassName]; ok {
			continue
		}
		inst, err := g.emitConstructor(route.ClassName)
		if err != nil {
			return err
		}
		glob := fmt.Sprintf("@__kylix_boot_ctrl_%s", route.ClassName)
		ctrlGlobals[route.ClassName] = glob
		g.line(fmt.Sprintf("  store ptr %s, ptr %s", inst, glob))
	}

	// 3. [Inject] field assignments: owner.Field = dependency instance.
	for _, inj := range g.bootInjects {
		depReg := compRegs[inj.FieldType]
		if depReg == "" {
			continue
		}
		ownerGlob := ctrlGlobals[inj.OwnerClass]
		if ownerGlob == "" {
			ownerGlob = compRegs[inj.OwnerClass] // inject into a component
		}
		if ownerGlob == "" {
			continue
		}
		ownerInst := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", ownerInst, ownerGlob))
		slot, _, err := g.emitFieldStore(inj.OwnerClass, ownerInst, inj.FieldName)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("  store ptr %s, ptr %s", depReg, slot))
	}

	// 4. Route registration: Boot<METHOD>(path, wrapper). The wrapper's own
	//    Boot* dependencies (BootText for procedure handlers, BootEnforceAuth/
	//    BootEnforceRole for security) must be queued HERE — before
	//    emitPendingStdlib runs — so their defines exist when the wrappers are
	//    emitted at module end.
	for i, route := range g.bootRoutes {
		if route.HandlerKind == bootHandlerProcedure {
			g.enqueueStdlib("boot", "BootText", "BootText", 0)
		}
		if route.Security.RequireAuth {
			g.enqueueStdlib("boot", "BootEnforceAuth", "BootEnforceAuth", 0)
		}
		for range route.Security.Roles {
			g.enqueueStdlib("boot", "BootEnforceRole", "BootEnforceRole", 0)
		}

		wrapper := fmt.Sprintf("@__kylix_boot_handler_%d", i)
		g.enqueueStdlib("boot", "Boot"+route.Method, "Boot"+route.Method, 0)
		g.line(fmt.Sprintf("  call void @__kylix_boot_Boot%s(ptr %s, ptr %s) ; route %s %s -> %s.%s",
			route.Method, g.addString(route.Path), wrapper,
			route.Method, route.Path, route.ClassName, route.MethodName))
		g.bootWrappers = append(g.bootWrappers, bootWrapper{
			Index: i, Route: route, CtrlGlobal: ctrlGlobals[route.ClassName],
		})
	}
	return nil
}

// emitPendingBootDefines emits the route handler wrappers as module-level
// defines. Called after emitPendingStdlib so the Boot* and BootText defines are
// already in the IR; class methods (@Class_Method) are emitted before main, so
// there are no forward references.
func (g *Generator) emitPendingBootDefines() {
	for _, w := range g.bootWrappers {
		g.emitBootWrapper(w)
	}
}

func (g *Generator) emitBootWrapper(w bootWrapper) {
	route := w.Route
	g.line(fmt.Sprintf("define ptr @__kylix_boot_handler_%d(ptr %%req) {", w.Index))
	g.line("entry:")
	ctrl := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", ctrl, w.CtrlGlobal))

	// Security guards: BootEnforceAuth/BootEnforceRole return null = pass.
	if route.Security.RequireAuth {
		g.enqueueStdlib("boot", "BootEnforceAuth", "BootEnforceAuth", 0)
		authRes := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_BootEnforceAuth(ptr %%req)", authRes))
		authOk := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", authOk, authRes))
		g.line(fmt.Sprintf("  br i1 %s, label %%__kix_auth_ok_%d, label %%__kix_auth_deny_%d", authOk, w.Index, w.Index))
		g.line(fmt.Sprintf("__kix_auth_deny_%d:", w.Index))
		g.line(fmt.Sprintf("  ret ptr %s", authRes))
		g.line(fmt.Sprintf("__kix_auth_ok_%d:", w.Index))
	}
	for ri, role := range route.Security.Roles {
		g.enqueueStdlib("boot", "BootEnforceRole", "BootEnforceRole", 0)
		roleRes := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_BootEnforceRole(ptr %%req, ptr %s)", roleRes, g.addString(role)))
		roleOk := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", roleOk, roleRes))
		g.line(fmt.Sprintf("  br i1 %s, label %%__kix_role_ok_%d_%d, label %%__kix_role_deny_%d_%d", roleOk, w.Index, ri, w.Index, ri))
		g.line(fmt.Sprintf("__kix_role_deny_%d_%d:", w.Index, ri))
		g.line(fmt.Sprintf("  ret ptr %s", roleRes))
		g.line(fmt.Sprintf("__kix_role_ok_%d_%d:", w.Index, ri))
	}

	if route.HandlerKind == bootHandlerProcedure {
		// procedure handler: BootText(200,"") pre-creates the response, then the
		// handler mutates it in place; return it.
		g.enqueueStdlib("boot", "BootText", "BootText", 0)
		res := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_BootText(i64 200, ptr %s)", res, g.ptrTo(g.addString(""), 1)))
		g.line(fmt.Sprintf("  call void @%s_%s(ptr %s, ptr %%req, ptr %s)", route.ClassName, route.MethodName, ctrl, res))
		g.line(fmt.Sprintf("  ret ptr %s", res))
	} else {
		res := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @%s_%s(ptr %s, ptr %%req)", res, route.ClassName, route.MethodName, ctrl))
		g.line(fmt.Sprintf("  ret ptr %s", res))
	}
	g.line("}")
	g.line("")
}

// --- annotation helpers (ported from generator_boot_annotations.go) ---

func shortComponentName(className string) string {
	if len(className) > 1 && strings.HasPrefix(className, "T") {
		return className[1:]
	}
	return className
}

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
