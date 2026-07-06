// Package llvmgen translates Kylix AST directly to LLVM IR text format (.ll).
//
// Design:
//   - Produces human-readable LLVM IR (text format)
//   - Uses SSA (Static Single Assignment) via alloca/load/store for simplicity
//   - llc compiles .ll → .o; clang links .o → native binary
//   - Go backend remains the default; LLVM backend is opt-in via --backend=llvm
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// Generator holds all state for LLVM IR generation.
type Generator struct {
	b          strings.Builder // output IR buffer
	module     string          // LLVM module name
	tmpCount   int             // SSA register counter (%t0, %t1, …)
	labelCount int             // basic block label counter
	locals           map[string]string         // var name → alloca register (%v_name)
	constants        map[string]ast.Expression // const name → value expression (literal)
	funcSigs         map[string]*ast.FunctionDecl // function name → declaration (param/return types)
	multiRetTypes    map[string][]string       // function name → LLVM types for multi-return (nil for single/void)
	classes          map[string]*ClassInfo     // class name → compiled class metadata
	interfaces       map[string]*InterfaceInfo // interface name → metadata
	genericTemplates map[string]*ast.ClassDecl // base name → generic class template
	instantiations   map[string]bool           // mangled name → already specialized
	localTypes       map[string]string         // var name → Kylix type name (class/interface)
	arrayInfo        map[string]*arrayInfo     // var name → array metadata
	varNameSeq       map[string]int            // Kylix var name → count of allocas emitted so far
	program          *ast.Program              // current AST root (for cross-pass access)
	funcName         string                    // current function being generated
	strings          []stringConst             // string constants (emitted at module level)

	// Exception handling (M3): global exception slot + setjmp/longjmp.
	exceptionTypeIDs    map[string]int // exception class name → runtime type ID (Exception=1)
	nextExcTypeID       int            // next ID to assign (starts at 2; 1 reserved for Exception)
	inExceptHandler     bool           // true while emitting an except/on handler body (bare raise)
	exceptionInjected   bool           // guards against double-injecting the Exception class

	// Loop control: break/continue targets for the innermost loop.
	breakLabel    string // label to jump to on 'break'
	continueLabel string // label to jump to on 'continue'

	// Lambda/closure support (M4): lambdas are lowered to named functions with
	// an environment struct. lambdaQueue collects bodies to emit at module end.
	lambdaCount   int                    // next lambda id (@__lambda_0, _1, ...)
	lambdaQueue   []pendingLambda        // deferred lambda function bodies
	closureLocals map[string]bool        // local var names holding a closure value
	closureSigs   map[string]string      // closure local var name → LLVM return type
	closureParams map[string][]string    // closure local var name → LLVM param types

	// inherited: tracks the method currently being generated so `inherited`
	// can resolve the parent-class method to call.
	curClassName  string
	curMethodName string

	// stdlib (v4.2.0 Phase 1): stdlib module functions (e.g. sysutil.ReadFile)
	// are lowered to module-level @__kylix_<Module>_<Func> defines that call
	// libc. Bodies are queued during expression emission and emitted at module
	// end (like lambdas) — they can't be defined inline mid-expression.
	stdlibEmitted map[string]bool // function key ("sysutil.ReadFile") → body already queued
	stdlibQueue   []stdlibFunc    // deferred stdlib function bodies to emit

	// base64TableEmitted guards the @__kylix_b64_table global (emitted once
	// per module, on first Base64Encode/Decode use).
	base64TableEmitted bool

	// hashtabEmitted guards the @__kylix_htab_* runtime helpers (emitted once
	// per module, on first cache/map use).
	hashtabEmitted bool

	// needHashtab is set when any stdlib module (cache, future map) references
	// the hash-table runtime. emitProgram checks it at module end and emits
	// the helpers only if actually needed (avoids bloating every module).
	needHashtab bool

	// needLibcrypto is set when crypto module functions are used; the compile
	// driver checks for crypto symbols in the IR and adds -lcrypto at link.
	needLibcrypto bool

	// needLibsqlite is set when db module functions are used; the compile
	// driver checks for db symbols in the IR and adds -lsqlite3 at link.
	needLibsqlite bool

	// mapVars tracks local variables declared as map[K]V — their alloca
	// holds a ptr to an @__kylix_htab_* table. Indexing/assignment on these
	// routes to htab_get/htab_put instead of the array-index path.
	mapVars map[string]bool
}

type stringConst struct {
	reg  string // @.str.N
	val  string // literal value
	size int    // byte length including \00
}

// stdlibFunc is a deferred stdlib module-function body (e.g. sysutil.ReadFile)
// queued during expression emission and emitted as a module-level define at
// the end of emitProgram (see emitPendingStdlib).
type stdlibFunc struct {
	module   string // "sysutil"
	name     string // "ReadFile" (or "PathJoin")
	key      string // dedup key ("sysutil.ReadFile", or "sysutil.PathJoin_3")
	argCount int    // arg count for variadic functions (PathJoin); 0 otherwise
}

// NewGenerator creates a new LLVM IR generator.
func NewGenerator(moduleName string) *Generator {
	return &Generator{
		module:            moduleName,
		locals:            make(map[string]string),
		constants:         make(map[string]ast.Expression),
		funcSigs:          make(map[string]*ast.FunctionDecl),
		multiRetTypes:     make(map[string][]string),
		classes:           make(map[string]*ClassInfo),
		interfaces:        make(map[string]*InterfaceInfo),
		genericTemplates:  make(map[string]*ast.ClassDecl),
		instantiations:    make(map[string]bool),
		localTypes:        make(map[string]string),
		arrayInfo:         make(map[string]*arrayInfo),
		varNameSeq:        make(map[string]int),
		exceptionTypeIDs:  make(map[string]int),
		nextExcTypeID:     2, // 1 reserved for Exception itself
		closureLocals:     make(map[string]bool),
		closureSigs:       make(map[string]string),
		closureParams:     make(map[string][]string),
		stdlibEmitted:     make(map[string]bool),
		mapVars:           make(map[string]bool),
	}
}

// Generate translates a Kylix AST program to LLVM IR text.
func Generate(prog *ast.Program) (string, error) {
	g := NewGenerator(prog.Name)
	if err := g.emitProgram(prog); err != nil {
		return "", err
	}
	return g.b.String(), nil
}

// ===== Module-level emission =====

func (g *Generator) emitProgram(prog *ast.Program) error {
	g.program = prog
	g.emitHeader()

	// Emit runtime declarations (libc functions we'll call)
	g.emitRuntimeDecls()

	// Inject the built-in Exception class before user decls so that user
	// exception classes (Parent="Exception") resolve against it, and so
	// `on E: Exception do` / `E.Message` work without special-casing.
	g.injectExceptionClass()

	// Pre-register all function/const signatures for forward references.
	// Multi-return struct types are emitted here (in declaration order) so
	// map iteration order never affects generated IR.
	for _, decl := range prog.Declarations {
		if cd, ok := decl.(*ast.ConstDecl); ok {
			g.constants[cd.Name] = cd.Value
		} else if fd, ok := decl.(*ast.FunctionDecl); ok && !fd.IsExternal {
			g.funcSigs[fd.Name] = fd
			if len(fd.ReturnTypes) > 0 {
				var llvmTypes []string
				for _, rt := range fd.ReturnTypes {
					llvmTypes = append(llvmTypes, LLVMType(typeExprName(rt)))
				}
				g.multiRetTypes[fd.Name] = llvmTypes
				g.line(fmt.Sprintf("%%__ret_%s = type { %s }", fd.Name, strings.Join(llvmTypes, ", ")))
			}
		}
	}

	// Emit declarations and function bodies
	for _, decl := range prog.Declarations {
		if err := g.emitDecl(decl); err != nil {
			return err
		}
	}

	// After non-generic decls are emitted (so templates are registered),
	// walk the program and specialize each unique generic instantiation.
	if err := g.collectInstantiations(prog); err != nil {
		return err
	}

	// Exception runtime: assign type IDs, emit the global exception slot and
	// the __kylix_is_subtype helper. Done after decls so all exception classes
	// (user + injected) are registered in g.classes.
	g.collectExceptionTypes()
	g.emitExceptionGlobals()
	g.emitExceptionRuntime()

	// Emit main function from top-level statements
	if len(prog.Statements) > 0 || prog.Name != "" {
		if err := g.emitMain(prog.Statements); err != nil {
			return err
		}
	}

	// Emit deferred lambda function bodies (collected during expression
	// emission — they can't be defined inline mid-expression).
	if err := g.emitPendingLambdas(); err != nil {
		return err
	}

	// Emit deferred stdlib module-function bodies (e.g. sysutil.ReadFile),
	// collected during expression emission. Module-level defines, like lambdas.
	g.emitPendingStdlib()

	// Emit the internal hash-table runtime (used by cache / map) if any
	// module referenced it. Idempotent.
	if g.needHashtab {
		g.emitHashtabBodies()
	}

	// Emit string constants at the end
	g.emitStringConsts()

	return nil
}

func (g *Generator) emitHeader() {
	g.line(fmt.Sprintf("; Kylix LLVM IR — module: %s", g.module))
	g.line(fmt.Sprintf("source_filename = \"%s.klx\"", g.module))
	g.line("target datalayout = \"e-m:o-i64:64-i128:128-n32:64-S128\"")
	g.line("target triple = \"arm64-apple-macosx15.0.0\"")
	g.line("")
}

func (g *Generator) emitRuntimeDecls() {
	g.line("; ===== Runtime declarations (libc) =====")
	g.line("declare i32 @printf(ptr noundef, ...)")
	g.line("declare i32 @puts(ptr noundef)")
	g.line("declare ptr @malloc(i64 noundef)")
	g.line("declare void @free(ptr noundef)")
	g.line("declare i64 @strlen(ptr noundef)")
	g.line("declare ptr @strcpy(ptr noundef, ptr noundef)")
	g.line("declare ptr @strcat(ptr noundef, ptr noundef)")
	g.line("declare i32 @strcmp(ptr noundef, ptr noundef)")
	g.line("declare ptr @memcpy(ptr noundef, ptr noundef, i64 noundef)")
	g.line("declare i64 @atoll(ptr noundef)")
	g.line("declare i32 @snprintf(ptr noundef, i64 noundef, ptr noundef, ...)")
	g.line("; ===== Exception handling runtime (setjmp/longjmp) =====")
	g.line("declare i32 @setjmp(ptr)")
	g.line("declare void @longjmp(ptr, i32)")
	g.line("declare void @exit(i32)")
	g.line("; ===== File I/O (libc, used by stdlib sysutil) =====")
	g.line("declare ptr @fopen(ptr noundef, ptr noundef)")
	g.line("declare i32 @fclose(ptr noundef)")
	g.line("declare i64 @fread(ptr noundef, i64 noundef, i64 noundef, ptr noundef)")
	g.line("declare i64 @fwrite(ptr noundef, i64 noundef, i64 noundef, ptr noundef)")
	g.line("declare i32 @fputs(ptr noundef, ptr noundef)")
	g.line("declare i32 @fseek(ptr noundef, i64 noundef, i32 noundef)")
	g.line("declare i64 @ftell(ptr noundef)")
	g.line("declare i32 @access(ptr noundef, i32 noundef)")
	g.line("; ===== BSD sockets (used by stdlib net) =====")
	g.line("declare i32 @socket(i32 noundef, i32 noundef, i32 noundef)")
	g.line("declare i32 @connect(i32 noundef, ptr noundef, i32 noundef)")
	g.line("declare i32 @bind(i32 noundef, ptr noundef, i32 noundef)")
	g.line("declare i32 @listen(i32 noundef, i32 noundef)")
	g.line("declare i32 @accept(i32 noundef, ptr, ptr)")
	g.line("declare i64 @send(i32 noundef, ptr noundef, i64 noundef, i32 noundef)")
	g.line("declare i64 @recv(i32 noundef, ptr noundef, i64 noundef, i32 noundef)")
	g.line("declare i32 @close(i32 noundef)")
	g.line("declare i32 @setsockopt(i32 noundef, i32 noundef, i32 noundef, ptr noundef, i32 noundef)")
	g.line("declare i32 @inet_pton(i32 noundef, ptr noundef, ptr noundef)")
	g.line("; ===== OpenSSL libcrypto (used by stdlib crypto) =====")
	g.line("declare ptr @SHA256(ptr noundef, i64 noundef, ptr noundef)")
	g.line("declare ptr @MD5(ptr noundef, i64 noundef, ptr noundef)")
	g.line("declare ptr @strncpy(ptr noundef, ptr noundef, i64 noundef)")
	g.line("; ===== SQLite (used by stdlib db) =====")
	g.line("declare i32 @sqlite3_open(ptr noundef, ptr noundef)")
	g.line("declare i32 @sqlite3_close(ptr noundef)")
	g.line("declare i32 @sqlite3_prepare_v2(ptr noundef, ptr noundef, i32 noundef, ptr noundef, ptr noundef)")
	g.line("declare i32 @sqlite3_bind_text(ptr noundef, i32 noundef, ptr noundef, i32 noundef, i64 noundef)")
	g.line("declare i32 @sqlite3_bind_int64(ptr noundef, i32 noundef, i64 noundef)")
	g.line("declare i32 @sqlite3_step(ptr noundef)")
	g.line("declare ptr @sqlite3_column_text(ptr noundef, i32 noundef)")
	g.line("declare i32 @sqlite3_finalize(ptr noundef)")
	g.line("; ===== POSIX regex (used by stdlib regex) =====")
	g.line("declare i32 @regcomp(ptr noundef, ptr noundef, i32 noundef)")
	g.line("declare i32 @regexec(ptr noundef, ptr noundef, i64 noundef, ptr, i32 noundef)")
	g.line("declare void @regfree(ptr noundef)")
	g.line("; ===== time.h (used by stdlib datetime) =====")
	g.line("declare i64 @time(ptr)")
	g.line("declare ptr @localtime(ptr)")
	g.line("declare ptr @localtime_r(ptr, ptr)")
	g.line("declare i64 @mktime(ptr)")
	g.line("declare i64 @strftime(ptr, i64, ptr, ptr)")
	g.line("; ===== LLVM intrinsics =====")
	g.line("declare void @llvm.memset.p0.i64(ptr nocapture writeonly, i8, i64, i1 immarg)")
	g.line("declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg)")
	g.line("")
	g.line("; ===== datetime arena allocator =====")
	g.line("@__kylix_datetime_arena = internal global [1048576 x i8] zeroinitializer, align 8")
	g.line("@__kylix_datetime_arena_ptr = internal global ptr @__kylix_datetime_arena, align 8")
	g.line("")
}

func (g *Generator) emitMain(stmts []ast.Statement) error {
	g.line("; ===== Entry point =====")
	g.line("define i32 @main() {")
	g.line("entry:")
	g.funcName = "main"
	g.locals = make(map[string]string)
	g.varNameSeq = make(map[string]int)

	// Emit top-level VarDecl as local allocas inside main().
	// (Top-level `var x: T;` declarations live in prog.Declarations.)
	for _, decl := range g.program.Declarations {
		if vd, ok := decl.(*ast.VarDecl); ok {
			if err := g.emitVarDecl(vd); err != nil {
				return err
			}
		}
	}

	for _, stmt := range stmts {
		if err := g.emitStatement(stmt); err != nil {
			return err
		}
	}

	g.line("  ret i32 0")
	g.line("}")
	g.line("")
	return nil
}

func (g *Generator) emitStringConsts() {
	if len(g.strings) == 0 {
		return
	}
	g.line("; ===== String constants =====")
	for _, s := range g.strings {
		escaped := llvmEscapeString(s.val)
		g.line(fmt.Sprintf("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\", align 1",
			s.reg, s.size, escaped))
	}
}

// ===== Declarations =====

func (g *Generator) emitDecl(node ast.Node) error {
	switch d := node.(type) {
	case *ast.ConstDecl:
		// Already pre-registered in emitProgram; nothing else to emit.
		return nil
	case *ast.FunctionDecl:
		if d.IsExternal {
			return nil
		}
		return g.emitFunctionDecl(d)
	case *ast.ClassDecl:
		if isGenericTemplate(d) {
			g.registerGenericTemplate(d)
			return nil
		}
		return g.emitClassDecl(d)
	case *ast.TypeDecl:
		// Unwrap type alias declarations — ClassDecl lives inside TypeDecl.Type
		if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
			classDecl.Name = d.Name // ensure the name is set from TypeDecl
			if isGenericTemplate(classDecl) {
				g.registerGenericTemplate(classDecl)
				return nil
			}
			return g.emitClassDecl(classDecl)
		}
		if ifaceDecl, ok := d.Type.(*ast.InterfaceDecl); ok {
			ifaceDecl.Name = d.Name
			return g.emitInterfaceDecl(ifaceDecl)
		}
	case *ast.InterfaceDecl:
		return g.emitInterfaceDecl(d)
	}
	return nil
}

// ===== Helper methods =====

func (g *Generator) line(s string) {
	g.b.WriteString(s)
	g.b.WriteByte('\n')
}

func (g *Generator) tmp() string {
	r := fmt.Sprintf("%%t%d", g.tmpCount)
	g.tmpCount++
	return r
}

func (g *Generator) label() string {
	l := fmt.Sprintf("lbl%d", g.labelCount)
	g.labelCount++
	return l
}

// freshVarReg returns a fresh, function-scope-unique LLVM register name for a
// new alloca backing the Kylix local variable `name`, with the given type
// suffix (e.g. "_int", "_str", "_bool", "_real", or "" for untyped/ptr-class
// locals). LLVM local identifiers live in a single function-wide namespace —
// unlike Kylix's block-scoped `var`, sibling blocks (if/if, try/except,
// foreach/foreach, on-clauses, ...) can legally declare the same name. The
// first declaration keeps the plain "%v_name_suffix" form (matching existing
// IR snapshots/tests); subsequent declarations of the same name get a "_N"
// disambiguator appended after the suffix so the type-inference logic in
// emitIdentLoad (which matches on suffix) keeps working unchanged.
func (g *Generator) freshVarReg(name, suffix string) string {
	n := g.varNameSeq[name]
	g.varNameSeq[name] = n + 1
	if n == 0 {
		return fmt.Sprintf("%%v_%s%s", name, suffix)
	}
	// Disambiguator goes BEFORE the type suffix, not after — emitIdentLoad and
	// friends infer the LLVM type by matching a HasSuffix("_bool"/"_real"/"_str")
	// on the alloca name, so the suffix must remain the literal trailing text.
	return fmt.Sprintf("%%v_%s_%d%s", name, n, suffix)
}

// addString adds a string constant and returns its global register name.
func (g *Generator) addString(val string) string {
	reg := fmt.Sprintf("@.str.%d", len(g.strings))
	g.strings = append(g.strings, stringConst{
		reg:  reg,
		val:  val,
		size: len(val) + 1, // +1 for \00
	})
	return reg
}

// ptrTo returns a getelementptr instruction to get a pointer to a string constant.
func (g *Generator) ptrTo(strReg string, size int) string {
	t := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0", t, size, strReg))
	return t
}

// llvmEscapeString escapes a Go string for LLVM IR string constants.
func llvmEscapeString(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		switch c {
		case '\n':
			b.WriteString(`\0A`)
		case '\r':
			b.WriteString(`\0D`)
		case '\t':
			b.WriteString(`\09`)
		case '"':
			b.WriteString(`\22`)
		case '\\':
			b.WriteString(`\5C`)
		default:
			if c < 32 || c > 126 {
				b.WriteString(fmt.Sprintf(`\%02X`, c))
			} else {
				b.WriteByte(c)
			}
		}
	}
	return b.String()
}
