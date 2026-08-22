package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// bootOpaqueTypes are the KylixBoot HTTP framework types lowered to opaque
// pointers in LLVM IR (v6.1.0). TRequest/TResponse carry no IR-visible
// structure for the tutorial handlers — they are passed/returned by pointer
// and only their field accesses (resp.Body) are compiled to real GEP loads.
var bootOpaqueTypes = map[string]bool{
	"TRequest":     true,
	"TResponse":    true,
	"BootRequest":  true,
	"BootResponse": true,
}

// stdlib_boot.go — LLVM IR for the `boot` (KylixBoot) stdlib module.
//
// The KylixBoot framework on the Go backend depends on net/http and reflect.
// The LLVM backend has no HTTP server and no RTTI, so most KylixBoot runtime
// functions are typed stubs that evaluate their arguments and return an empty
// default — enough for the tutorial examples (41-47, 49-51) to compile,
// register routes and run without crashing. BootText/BootJSON/BootHTML return
// a real {i64 status, ptr body} response handle so handler bodies lower to
// valid IR (the pre-v6.1.0 catch-all i64 stub produced a store type mismatch
// when their ptr result was stored into an i64 %result).

// bootStubReturnTypes maps each boot function to the LLVM type its call-site
// result register must have. Functions not listed default to i64.
var bootStubReturnTypes = map[string]string{
	"BootConfigGetString":  "ptr",
	"BootResolve":          "ptr",
	"BootEnforceAuth":      "ptr",
	"BootEnforceRole":      "ptr",
	"BootConfigGetInt":     "i64",
	"BootRun":              "i64",
	"BootReadJSON":         "i64",
	"BootRegisterJwtAuth":  "void",
	"BootGET":              "void",
	"BootPOST":             "void",
	"BootPUT":              "void",
	"BootDELETE":           "void",
	"BootUseLogger":        "void",
	"BootUseRecover":       "void",
	"BootUseCORS":          "void",
	"BootUseRequestID":     "void",
	"BootConfigSet":        "void",
	"BootRegisterInstance": "void",
	"BootRegisterAuth":     "void",
	"BootRegisterRoles":    "void",
}

func (g *Generator) emitBootCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "BootText", "BootJSON", "BootHTML":
		// Real response handle {i64 status, ptr body}.
		return g.emitBootResponseCall(funcName, args)
	case "BootRun":
		// v6.6.0: real HTTP server (listen + accept + dispatch).
		if len(args) < 1 {
			return "", "", fmt.Errorf("boot.BootRun expects a port, got %d", len(args))
		}
		portReg, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		g.enqueueStdlib("boot", "BootRun", "BootRun", 0)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i64 @__kylix_boot_BootRun(i64 %s)", r, portReg))
		return r, "i64", nil
	case "BootRegisterJwtAuth":
		// void — evaluate args for side effects, return void.
		for _, a := range args {
			if _, _, err := g.emitExpr(a); err != nil {
				return "", "", err
			}
		}
		return "0", "void", nil
	default:
		retType, ok := bootStubReturnTypes[funcName]
		if !ok {
			retType = "i64"
		}
		return g.emitBootStubCall(funcName, args, retType)
	}
}

// emitBootResponseCall emits a call to a BootText/BootJSON/BootHTML define.
// Status defaults to 200, body to the string arg (BootJSON ignores it — no
// JSON serializer is needed for the tutorials, body is left empty).
func (g *Generator) emitBootResponseCall(funcName string, args []ast.Expression) (string, string, error) {
	statusReg := "200"
	if len(args) >= 1 {
		r, _, err := g.emitExpr(args[0])
		if err != nil {
			return "", "", err
		}
		statusReg = r
	}

	bodyReg := g.ptrTo(g.addString(""), 1)
	if funcName != "BootJSON" && len(args) >= 2 {
		r, _, err := g.emitExpr(args[1])
		if err != nil {
			return "", "", err
		}
		bodyReg = r
	}

	g.enqueueStdlib("boot", funcName, funcName, 2)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_%s(i64 %s, ptr %s)", r, funcName, statusReg, bodyReg))
	return r, "ptr", nil
}

func (g *Generator) emitBootBody(funcName string) {
	switch funcName {
	case "BootText", "BootJSON", "BootHTML":
		g.emitBootResponseBody(funcName)
	case "BootGET", "BootPOST", "BootPUT", "BootDELETE":
		// v6.6.0: real route registration — store {method, path, handler} in
		// the module route table for @__kylix_boot_BootRun to dispatch.
		g.emitBootRegisterRouteBody(funcName)
	case "BootRegisterInstance", "BootUseLogger", "BootUseRecover", "BootUseCORS",
		"BootUseRequestID", "BootConfigSet", "BootRegisterAuth", "BootRegisterRoles":
		// void no-op — registrations are accepted but have no runtime effect
		// on the LLVM backend (no middleware / DI reflection).
		g.line(fmt.Sprintf("define void @__kylix_boot_%s(ptr %%a, ptr %%b) {", funcName))
		g.line("  ret void")
		g.line("}")
	case "BootRun":
		// v6.6.0: real HTTP server (listen + accept + dispatch).
		g.emitBootRunBody()
	case "readheaders":
		g.emitBootReadHeadersBody()
	case "parsereq":
		g.emitBootParseRequestBody()
	case "routelookup":
		g.emitBootRouteLookupBody()
	case "pathmatch":
		g.emitBootPathMatchBody()
	case "BootEnforceAuth":
		// null = pass (authentication is disabled).
		g.line("define ptr @__kylix_boot_BootEnforceAuth(ptr %req) {")
		g.line("  ret ptr null")
		g.line("}")
	case "BootEnforceRole":
		g.line("define ptr @__kylix_boot_BootEnforceRole(ptr %req, ptr %role) {")
		g.line("  ret ptr null")
		g.line("}")
	case "BootReadJSON":
		g.line("define i64 @__kylix_boot_BootReadJSON(ptr %req, ptr %dst) {")
		g.line("  ret i64 0")
		g.line("}")
	default:
		// No separate define — the stub is inlined at the call site.
	}
}

// emitBootResponseBody emits a {i64 status, ptr body} response handle:
//
//	%h = malloc(16); store status @ 0; store body @ 8; ret %h
func (g *Generator) emitBootResponseBody(funcName string) {
	g.line(fmt.Sprintf("define ptr @__kylix_boot_%s(i64 %%status, ptr %%body) {", funcName))
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", h))
	s := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 0", s, h))
	g.line(fmt.Sprintf("  store i64 %%status, ptr %s", s))
	b := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", b, h))
	g.line(fmt.Sprintf("  store ptr %%body, ptr %s", b))
	g.line(fmt.Sprintf("  ret ptr %s", h))
	g.line("}")
}

// emitBootResponseFieldAccess lowers `resp.Body` / `resp.Status` on a
// KylixBoot TResponse handle ({i64 status, ptr body}, 16 bytes) to GEP+load.
// The receiver is a call expression like `ctrl.ListUsers(nil)` whose value is
// the handle returned by BootText/BootJSON (see emitBootResponseCall).
func (g *Generator) emitBootResponseFieldAccess(obj ast.Expression, field string) (string, string, error) {
	handle, _, err := g.emitExpr(obj)
	if err != nil {
		return "", "", err
	}
	switch strings.ToLower(field) {
	case "body":
		slot := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", slot, handle))
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, slot))
		return r, "ptr", nil
	case "status", "statuscode":
		slot := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 0", slot, handle))
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load i64, ptr %s", r, slot))
		return r, "i64", nil
	default:
		// Unknown field — null fallback keeps the IR legal.
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; TResponse.%s unsupported", r, field))
		return r, "ptr", nil
	}
}

func (g *Generator) emitBootStubCall(funcName string, args []ast.Expression, retType string) (string, string, error) {
	// Evaluate all arguments for side effects.
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	if retType == "ptr" {
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	}
	if retType == "void" {
		return "0", "void", nil
	}
	if retType == "i1" {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i1 0, 0 ; boot.%s stub", r, funcName))
		return r, "i1", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0 ; boot.%s stub", r, funcName))
	return r, "i64", nil
}

// emitBootRegisterRouteBody — v6.6.0: `void @__kylix_boot_Boot<M>(ptr %path,
// ptr %wrapper)` writes {method, path, handler} into the module route table
// (@__kylix_boot_routes / @__kylix_boot_nroutes), which @__kylix_boot_BootRun
// dispatches against. Capacity 64; overflow is dropped.
func (g *Generator) emitBootRegisterRouteBody(method string) {
	methodLit := strings.TrimPrefix(method, "Boot")
	methodPtr := g.addString(methodLit)
	g.line(fmt.Sprintf("define void @__kylix_boot_%s(ptr %%path, ptr %%wrapper) {", method))
	g.line("entry:")
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_boot_nroutes", n))
	full := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 64", full, n))
	storeLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", full, doneLbl, storeLbl))
	g.line(fmt.Sprintf("%s:", storeLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x { ptr, ptr, ptr }], ptr @__kylix_boot_routes, i64 0, i64 %s", slot, n))
	mField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 0", mField, slot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", methodPtr, mField))
	pField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 1", pField, slot))
	g.line(fmt.Sprintf("  store ptr %%path, ptr %s", pField))
	hField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 2", hField, slot))
	g.line(fmt.Sprintf("  store ptr %%wrapper, ptr %s", hField))
	nextN := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", nextN, n))
	g.line(fmt.Sprintf("  store i64 %s, ptr @__kylix_boot_nroutes", nextN))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}
