package llvmgen

import (
	"kylix/ast"
)

// export.go — C ABI export mechanism & shared library runtime helpers (v0.14.0).
//
// Kylix functions and procedures annotated with [Export] or [Export('c_symbol')]
// are exposed as un-mangled, global C ABI symbols in the generated LLVM IR.
//
// Memory Ownership Contract:
// Any dynamic string allocated by Kylix and returned across the C ABI boundary
// must be freed by the caller invoking the compiler-injected @kylix_free(ptr %p)
// function, preventing leaks across runtime/FFI boundaries.

// getExportSymbol reports whether attrs contains an [Export] attribute, returning
// the specified C export symbol (or defaultName if no custom name was given).
func getExportSymbol(attrs []*ast.Attribute, defaultName string) (string, bool) {
	attr := findAttribute(attrs, "Export")
	if attr == nil {
		return "", false
	}
	if len(attr.Args) > 0 {
		if lit, ok := attr.Args[0].(*ast.StringLiteral); ok && lit.Value != "" {
			return lit.Value, true
		}
	}
	return defaultName, true
}

// calleeSymbol returns the symbol name to use when generating a call instruction.
// If the target function was annotated with [Export('c_name')], returns c_name.
func (g *Generator) calleeSymbol(name string) string {
	if g.exportedFuncs != nil {
		if sym, ok := g.exportedFuncs[name]; ok && sym != "" {
			return sym
		}
	}
	return name
}

// emitKylixFree emits the standard C ABI memory deallocation function.
// External callers (C, JNI, Swift) call kylix_free(ptr) to return ownership
// of strings and heap allocations produced by exported Kylix functions.
func (g *Generator) emitKylixFree() {
	g.line("; ===== Kylix C ABI runtime export: kylix_free =====")
	if g.gc == "boehm" {
		g.line("declare void @GC_free(ptr noundef)")
	}
	g.line("define void @kylix_free(ptr %p) {")
	g.line("entry:")
	g.line("  %isnull = icmp eq ptr %p, null")
	g.line("  br i1 %isnull, label %exit, label %do_free")
	g.line("do_free:")
	if g.gc == "boehm" {
		// Boehm GC mode: memory is garbage collected, but explicit GC_free is safe
		g.line("  call void @GC_free(ptr %p)")
	} else {
		// Default malloc mode: route to libc free
		g.line("  call void @free(ptr %p)")
	}
	g.line("  br label %exit")
	g.line("exit:")
	g.line("  ret void")
	g.line("}")
	g.line("")
}
