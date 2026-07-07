package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_boot.go — LLVM IR stubs for the `boot` (KylixBoot) stdlib module.
//
// The KylixBoot framework (pkg/boot) depends on Go's net/http and reflect
// packages. The LLVM backend does not have an HTTP server or RTTI yet, so
// all KylixBoot runtime functions return empty/default values — enough for
// the tutorial examples (41-47, 49-51) to compile and run without crashing.

func (g *Generator) emitBootCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "BootText":
		return g.emitBootStubCall(funcName, args, "ptr")
	case "BootJSON":
		return g.emitBootStubCall(funcName, args, "ptr")
	case "BootRegisterJwtAuth":
		// void function — evaluate args for side effects, return void.
		for _, a := range args {
			if _, _, err := g.emitExpr(a); err != nil {
				return "", "", err
			}
		}
		return "0", "void", nil
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; boot.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

func (g *Generator) emitBootBody(funcName string) {
	// All boot functions are stubs inlined at call site, no separate defines.
}

func (g *Generator) emitBootStubCall(funcName string, args []ast.Expression, retType string) (string, string, error) {
	// Evaluate all arguments for side effects.
	for _, a := range args {
		if _, _, err := g.emitExpr(a); err != nil {
			return "", "", err
		}
	}
	g.enqueueStdlib("boot", funcName, funcName, 0)
	if retType == "ptr" {
		emptyStr := g.addString("")
		return g.ptrTo(emptyStr, 1), "ptr", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0 ; boot.%s stub", r, funcName))
	return r, retType, nil
}
