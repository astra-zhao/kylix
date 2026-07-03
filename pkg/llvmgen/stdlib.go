package llvmgen

import (
	"fmt"

	"kylix/ast"
)

// stdlib.go — stdlib module-function dispatch for the LLVM backend.
//
// Kylix source `sysutil.ReadFile(path)` parses to a CallExpression whose
// Function is a MemberExpression{Object: Identifier{sysutil}, Member: ReadFile}.
// The Go backend maps this to `stdlib.ReadFile(...)` (Go code), but the LLVM
// backend generates native code and cannot call Go — so each stdlib function
// is lowered to a module-level @__kylix_<Module>_<Func> define that calls libc
// (see stdlib_sysutil.go). This file is the dispatch layer.
//
// Known stdlib modules mirror generator/generator_stdlib.go's stdlibModuleFuncs
// keys. Only modules with an LLVM IR implementation are wired up here; others
// fall through to the method-call path (and produce an unsupported-receiver
// stub, same as before).

// knownStdlibModules lists stdlib module names that the LLVM backend has an IR
// implementation for. A `module.Func()` call is dispatched here only when the
// module is in this set AND was imported via `uses`.
var knownStdlibModules = map[string]bool{
	"sysutil":  true,
	"regex":    true,
	"datetime": true,
	// Future: jsonutil, net, crypto, encoding, ...
}

// stdlibModuleFuncs maps each known stdlib module to the function names it
// exports (mirrors generator/generator_stdlib.go's stdlibModuleFuncs for the
// modules the LLVM backend implements). Used to resolve bare-name calls like
// `ReadFile(...)` (no `sysutil.` qualifier) when the module is `uses`-imported.
var stdlibModuleFuncs = map[string]map[string]bool{
	"sysutil": {
		"ReadFile": true, "WriteFile": true, "FileExists": true,
		"PathJoin": true, "PathBase": true,
	},
	"regex": {
		"IsEmail": true, "IsURL": true, "IsNumeric": true,
		"IsAlpha": true, "IsAlphaNumeric": true, "IsIP": true,
	},
	"datetime": {
		"Now": true, "Today": true, "MakeDate": true, "MakeTime": true,
		"ParseDate": true, "ParseDateTime": true, "FreeArena": true,
	},
}

// resolveStdlibBareCall reports whether funcName is a bare-name stdlib call
// (e.g. `ReadFile(...)`) resolvable to an imported module. Returns the module
// name and ok=true if so. A user-defined function of the same name takes
// precedence (checked against g.funcSigs).
func (g *Generator) resolveStdlibBareCall(funcName string) (module string, ok bool) {
	if g.program == nil {
		return "", false
	}
	used := make(map[string]bool, len(g.program.Uses))
	for _, u := range g.program.Uses {
		used[u] = true
	}
	for mod, funcs := range stdlibModuleFuncs {
		if knownStdlibModules[mod] && used[mod] && funcs[funcName] {
			return mod, true
		}
	}
	return "", false
}

// isStdlibModule reports whether name is a stdlib module the LLVM backend can
// lower to IR, AND it appears in the program's `uses` clause.
func (g *Generator) isStdlibModule(name string) bool {
	if !knownStdlibModules[name] {
		return false
	}
	if g.program == nil {
		return false
	}
	for _, used := range g.program.Uses {
		if used == name {
			return true
		}
	}
	return false
}

// emitStdlibCall dispatches a `module.Func(args)` call to the per-module IR
// emitter. It emits the `call` instruction at the call site and queues the
// function body (deduped) for module-end emission.
func (g *Generator) emitStdlibCall(module, funcName string, args []ast.Expression) (string, string, error) {
	switch module {
	case "sysutil":
		return g.emitSysutilCall(funcName, args)
	case "regex":
		return g.emitRegexCall(funcName, args)
	case "datetime":
		return g.emitDatetimeCall(funcName, args)
	default:
		// Not yet implemented for LLVM — fall back to a stub so IR stays legal.
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; stdlib %s.%s not implemented (LLVM)", r, module, funcName))
		return r, "i64", nil
	}
}

// enqueueStdlib queues a stdlib function body for module-end emission, deduped
// by key. argCount is used for variadic functions (PathJoin is monomorphized
// per arity); pass 0 for fixed-arity functions. bodyKey disambiguates variants
// (e.g. "PathJoin_3").
func (g *Generator) enqueueStdlib(module, name, bodyKey string, argCount int) bool {
	key := module + "." + bodyKey
	if g.stdlibEmitted[key] {
		return false
	}
	g.stdlibEmitted[key] = true
	g.stdlibQueue = append(g.stdlibQueue, stdlibFunc{
		module: module, name: name, key: key, argCount: argCount,
	})
	return true
}

// emitPendingStdlib emits the deferred stdlib module-function bodies as
// module-level defines. Called once at the end of emitProgram (after lambdas,
// before string constants). Each emitter writes its own `define ... { ... }`.
func (g *Generator) emitPendingStdlib() {
	for _, sf := range g.stdlibQueue {
		switch sf.module {
		case "sysutil":
			g.emitSysutilBody(sf.name, sf.argCount)
		case "regex":
			g.emitRegexBody(sf.name)
		case "datetime":
			g.emitDatetimeBody(sf.name, sf.argCount)
		}
	}
}
