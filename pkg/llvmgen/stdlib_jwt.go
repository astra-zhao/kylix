package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_jwt.go — LLVM IR stubs for the `jwt` stdlib module.
// All functions return empty strings / false (no JWT implementation in LLVM backend).

func (g *Generator) emitJwtCall(funcName string, args []ast.Expression) (string, string, error) {
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	switch funcName {
	case "JwtSign":
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	case "JwtVerify":
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i1 0, 0 ; JwtVerify stub", r))
		return r, "i1", nil
	case "JwtSubject":
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; jwt.%s stub", r, funcName))
		return r, "i64", nil
	}
}

func (g *Generator) emitJwtBody(funcName string) {}
