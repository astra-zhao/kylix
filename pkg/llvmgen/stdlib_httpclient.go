package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_httpclient.go — LLVM IR stubs for the `httpclient` stdlib module.
//
// The Go backend's httpclient wraps net/http. The LLVM backend has no HTTP
// client yet, so all functions return empty/default stubs. NewHttpClient
// returns an empty string ptr as a stand-in "handle"; method calls on it
// (SetHeader) and field access (BaseURL) are handled via the method-not-found
// / field-not-found stubs in class.go, which return 0/empty.

const httpClientTypeName = "THttpClient"

func (g *Generator) emitHttpclientCall(funcName string, args []ast.Expression) (string, string, error) {
	// Evaluate all args for side effects.
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	switch funcName {
	case "NewHttpClient":
		// Return a ptr stub (empty string as handle).
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), httpClientTypeName, nil
	default:
		// HttpGet/HttpPost/etc. return empty string.
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	}
}

func (g *Generator) emitHttpclientBody(funcName string) {
	// All stubs inlined at call site.
}

// emitHttpclientMethodCall handles THttpClient method calls (SetHeader etc).
func (g *Generator) emitHttpclientMethodCall(receiver string, method string, args []ast.Expression) (string, string, error) {
	// Evaluate args for side effects.
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	switch method {
	case "SetHeader":
		return "0", "void", nil // no-op stub
	case "Get", "Post", "Put", "Delete":
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil // empty response body
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; THttpClient.%s stub", r, method))
		return r, "i64", nil
	}
}
