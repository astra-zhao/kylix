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
	locals     map[string]string  // var name → alloca register (%v_name)
	classes    map[string]*ClassInfo // class name → compiled class metadata
	arrayInfo  map[string]*arrayInfo // var name → array metadata
	program    *ast.Program    // current AST root (for cross-pass access)
	funcName   string          // current function being generated
	strings    []stringConst   // string constants (emitted at module level)
}

type stringConst struct {
	reg  string // @.str.N
	val  string // literal value
	size int    // byte length including \00
}

// NewGenerator creates a new LLVM IR generator.
func NewGenerator(moduleName string) *Generator {
	return &Generator{
		module:    moduleName,
		locals:    make(map[string]string),
		classes:   make(map[string]*ClassInfo),
		arrayInfo: make(map[string]*arrayInfo),
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

	// Emit declarations and function bodies
	for _, decl := range prog.Declarations {
		if err := g.emitDecl(decl); err != nil {
			return err
		}
	}

	// Emit main function from top-level statements
	if len(prog.Statements) > 0 || prog.Name != "" {
		if err := g.emitMain(prog.Statements); err != nil {
			return err
		}
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
	g.line("declare i64 @atoll(ptr noundef)")
	g.line("declare i32 @snprintf(ptr noundef, i64 noundef, ptr noundef, ...)")
	g.line("")
}

func (g *Generator) emitMain(stmts []ast.Statement) error {
	g.line("; ===== Entry point =====")
	g.line("define i32 @main() {")
	g.line("entry:")
	g.funcName = "main"
	g.locals = make(map[string]string)

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
	case *ast.FunctionDecl:
		if d.IsExternal {
			return nil
		}
		return g.emitFunctionDecl(d)
	case *ast.ClassDecl:
		return g.emitClassDecl(d)
	case *ast.TypeDecl:
		// Unwrap type alias declarations — ClassDecl lives inside TypeDecl.Type
		if classDecl, ok := d.Type.(*ast.ClassDecl); ok {
			classDecl.Name = d.Name // ensure the name is set from TypeDecl
			return g.emitClassDecl(classDecl)
		}
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
