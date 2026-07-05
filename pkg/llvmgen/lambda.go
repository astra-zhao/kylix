// lambda.go — LLVM IR generation for lambda/closure (M4).
//
// Lambdas are lowered to a named function @__lambda_N plus an environment
// struct %__env_N holding captured variables. A closure value is the pair
// { ptr func_ptr, ptr env_ptr }.
//
//	lambda body → @__lambda_N(ptr %env, <params>)
//	captures     → %__env_N = type { T1, T2, ... }  (malloc'd at creation)
//	closure value → { ptr, ptr }  (func ptr + env ptr)
//
// Calls through a closure local indirect-call the function pointer, passing
// env as the first argument. See emitClosureCall in expr.go.
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// capture records one captured variable: its name and LLVM type.
type capture struct {
	name     string
	llvmType string
}

// pendingLambda is a lambda body deferred to module-end emission (define
// blocks can't be emitted inline mid-expression).
type pendingLambda struct {
	id       int
	params   []*ast.Parameter
	retType  string // LLVM return type ("void" for procedures)
	retExpr  bool   // body is a bare expression (-> ret <expr>, no result alloca)
	body     ast.Node
	captures []capture
}

// closureType is the LLVM type of a closure value: { func_ptr, env_ptr }.
const closureType = "{ ptr, ptr }"

// lambdaName returns the global function name for a lambda by id.
func lambdaName(id int) string { return fmt.Sprintf("@__lambda_%d", id) }

// envTypeLiteral returns the literal LLVM struct type for a lambda's env,
// e.g. `{ i64, ptr }`. Used inline (no module-level type definition needed),
// avoiding forward-reference issues when the lambda is created mid-function.
func envTypeLiteral(caps []capture) string {
	if len(caps) == 0 {
		return "ptr" // no env; treat as opaque ptr (null)
	}
	var parts []string
	for _, c := range caps {
		parts = append(parts, c.llvmType)
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// collectCaptures walks a lambda body and returns the variables it references
// that are NOT its own parameters or locally-declared variables — i.e. the
// outer-scope variables it captures. ownParams is the set of parameter names;
// localDecls is the set of names declared inside the body.
func (g *Generator) collectCaptures(body ast.Node, ownParams, localDecls map[string]bool) []capture {
	var caps []capture
	seen := map[string]bool{}
	var visitExpr func(ast.Expression)
	var visitStmt func(ast.Statement)

	visitExpr = func(e ast.Expression) {
		if e == nil {
			return
		}
		switch x := e.(type) {
		case *ast.Identifier:
			n := x.Value
			if ownParams[n] || localDecls[n] || seen[n] {
				return
			}
			// Captured only if it resolves to an outer local variable.
			if allocaReg, ok := g.locals[n]; ok {
				seen[n] = true
				caps = append(caps, capture{name: n, llvmType: allocaLLVMType(allocaReg)})
			}
		case *ast.InfixExpression:
			visitExpr(x.Left)
			visitExpr(x.Right)
		case *ast.PrefixExpression:
			visitExpr(x.Right)
		case *ast.CallExpression:
			visitExpr(x.Function)
			for _, a := range x.Arguments {
				visitExpr(a)
			}
		case *ast.MemberExpression:
			visitExpr(x.Object)
		case *ast.IndexExpression:
			visitExpr(x.Left)
			visitExpr(x.Index)
		case *ast.SliceExpression:
			visitExpr(x.Left)
			visitExpr(x.Low)
			visitExpr(x.High)
		case *ast.StringInterpolation:
			for _, p := range x.Parts {
				if expr, ok := p.(ast.Expression); ok {
					visitExpr(expr)
				}
			}
		case *ast.TupleLiteral:
			for _, el := range x.Elements {
				visitExpr(el)
			}
		case *ast.ArrayLiteral:
			for _, el := range x.Elements {
				visitExpr(el)
			}
		case *ast.TypeCastExpression:
			visitExpr(x.Expression)
		case *ast.AwaitExpression:
			visitExpr(x.Expression)
		}
	}

	visitStmt = func(s ast.Statement) {
		if s == nil {
			return
		}
		switch x := s.(type) {
		case *ast.BlockStatement:
			for _, st := range x.Statements {
				visitStmt(st)
			}
		case *ast.ExpressionStatement:
			visitExpr(x.Expression)
		case *ast.AssignmentStatement:
			visitExpr(x.Name)
			visitExpr(x.Value)
		case *ast.VarDecl:
			for _, n := range x.Names {
				localDecls[n] = true
			}
			visitExpr(x.Value)
		case *ast.IfStatement:
			visitExpr(x.Condition)
			visitStmt(x.Consequence)
			visitStmt(x.Alternative)
		case *ast.WhileStatement:
			visitExpr(x.Condition)
			visitStmt(x.Body)
		case *ast.ForStatement:
			visitStmt(x.Body)
		case *ast.ForEachStatement:
			visitStmt(x.Body)
		case *ast.RepeatStatement:
			visitStmt(x.Body)
			visitExpr(x.Condition)
		case *ast.ReturnStatement:
			visitExpr(x.Value)
		case *ast.CaseStatement:
			visitExpr(x.Expression)
			for _, br := range x.Branches {
				visitStmt(br.Body)
			}
		case *ast.MatchStatement:
			visitExpr(x.Expression)
			for _, br := range x.Branches {
				visitStmt(br.Body)
			}
		case *ast.TryStatement:
			visitStmt(x.Body)
			visitStmt(x.ExceptBlock)
			visitStmt(x.FinallyBlock)
			for _, on := range x.OnClauses {
				visitStmt(on.Body)
			}
		case *ast.RaiseStatement:
			visitExpr(x.Exception)
		}
	}

	visitStmt(asStatement(body))
	return caps
}

// asStatement coerces a lambda body (BlockStatement or Expression) into a
// Statement so collectCaptures can walk it uniformly.
func asStatement(body ast.Node) ast.Statement {
	if s, ok := body.(ast.Statement); ok {
		return s
	}
	if e, ok := body.(ast.Expression); ok {
		return &ast.ExpressionStatement{Expression: e}
	}
	return nil
}

// allocaLLVMType infers the LLVM type of a local from its alloca register
// name, mirroring emitIdentLoad's suffix convention.
func allocaLLVMType(allocaReg string) string {
	switch {
	case strings.HasSuffix(allocaReg, "_bool"):
		return "i1"
	case strings.HasSuffix(allocaReg, "_real"):
		return "double"
	case strings.HasSuffix(allocaReg, "_str"):
		return "ptr"
	default:
		return "i64"
	}
}

// emitLambda handles a LambdaExpression at its creation site: allocates the
// environment struct (filling captured values), constructs the closure pair,
// and queues the function body for module-end emission. Returns the alloca
// register holding the closure value (type closureType).
func (g *Generator) emitLambda(e *ast.LambdaExpression) (string, string, error) {
	id := g.lambdaCount
	g.lambdaCount++

	// Determine return type.
	retType := "void"
	retExpr := false
	if e.ReturnType != nil {
		retType = LLVMType(typeExprName(e.ReturnType))
	}
	// Expression-bodied lambda: body is an Expression, not BlockStatement.
	if _, isBlock := e.Body.(*ast.BlockStatement); !isBlock {
		retExpr = true
		if retType == "void" {
			retType = "i64" // infer i64 if unspecified
		}
	}

	// Own parameter set + local decls (for capture analysis).
	ownParams := map[string]bool{}
	for _, p := range e.Parameters {
		ownParams[p.Name] = true
	}
	localDecls := map[string]bool{}
	caps := g.collectCaptures(e.Body, ownParams, localDecls)

	// Env struct uses a literal LLVM type (e.g. `{ i64, ptr }`) inline so no
	// module-level type definition is needed — avoids forward-reference issues
	// when the lambda is created mid-function.
	envT := envTypeLiteral(caps)

	// Allocate + fill environment (if captures).
	var envReg string
	if len(caps) == 0 {
		// No captures: env ptr is null.
		envReg = "null"
	} else {
		size := int64(8 * len(caps)) // conservative: each field <= 8 bytes
		envReg = g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", envReg, size))
		for i, c := range caps {
			// Load the captured value from the outer scope.
			valReg, _, err := g.emitIdentLoad(c.name)
			if err != nil {
				return "", "", err
			}
			slot := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr %s, ptr %s, i32 0, i32 %d", slot, envT, envReg, i))
			g.line(fmt.Sprintf("  store %s %s, ptr %s", c.llvmType, valReg, slot))
		}
	}

	// Construct the closure pair {func_ptr, env_ptr} on the stack.
	closureReg := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca %s, align 8", closureReg, closureType))
	fslot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr %s, ptr %s, i32 0, i32 0", fslot, closureType, closureReg))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", lambdaName(id), fslot))
	eslot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr %s, ptr %s, i32 0, i32 1", eslot, closureType, closureReg))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", envReg, eslot))

	// Queue the body for module-end emission.
	g.lambdaQueue = append(g.lambdaQueue, pendingLambda{
		id:       id,
		params:   e.Parameters,
		retType:  retType,
		retExpr:  retExpr,
		body:     e.Body,
		captures: caps,
	})

	return closureReg, closureType, nil
}

// emitPendingLambdas emits the deferred lambda function bodies at module end.
// Each becomes: define <retType> @__lambda_N(ptr %env, <params>) { ... }.
// Env structs use literal types (no module-level type definitions needed).
func (g *Generator) emitPendingLambdas() error {
	for _, pl := range g.lambdaQueue {
		if err := g.emitLambdaFunc(pl); err != nil {
			return err
		}
	}
	return nil
}

// emitLambdaFunc emits one lambda's function definition.
func (g *Generator) emitLambdaFunc(pl pendingLambda) error {
	// Build parameter list: env first, then lambda params.
	var params []string
	params = append(params, "ptr %env")
	for _, p := range pl.params {
		llvmT := "i64"
		if p.Type != nil {
			llvmT = LLVMType(typeExprName(p.Type))
		}
		params = append(params, fmt.Sprintf("%s %%%s", llvmT, p.Name))
	}

	g.line(fmt.Sprintf("define %s %s(%s) {", pl.retType, lambdaName(pl.id), strings.Join(params, ", ")))
	g.line("entry:")

	// Save + reset scope (lambdas have their own local scope).
	savedLocals := g.locals
	savedTypes := g.localTypes
	savedVarSeq := g.varNameSeq
	savedFuncName := g.funcName
	g.locals = make(map[string]string)
	g.localTypes = make(map[string]string)
	g.varNameSeq = make(map[string]int)
	g.funcName = fmt.Sprintf("__lambda_%d", pl.id)

	envT := envTypeLiteral(pl.captures)

	// Materialize captured variables: load each env field into a local alloca
	// so the body's emitIdentLoad calls work transparently.
	for i, c := range pl.captures {
		slot := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr %s, ptr %%env, i32 0, i32 %d", slot, envT, i))
		val := g.tmp()
		g.line(fmt.Sprintf("  %s = load %s, ptr %s", val, c.llvmType, slot))
		// Allocate a local mirroring the outer naming convention so
		// emitIdentLoad infers the type from the suffix.
		suffix := "_int"
		switch c.llvmType {
		case "i1":
			suffix = "_bool"
		case "double":
			suffix = "_real"
		case "ptr":
			suffix = "_str"
		}
		allocaReg := fmt.Sprintf("%%v_%s%s", c.name, suffix)
		g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, c.llvmType))
		g.line(fmt.Sprintf("  store %s %s, ptr %s", c.llvmType, val, allocaReg))
		g.locals[c.name] = allocaReg
	}

	// Allocate parameters as locals (same logic as emitFunctionDecl).
	for _, p := range pl.params {
		llvmT := "i64"
		kylixType := ""
		if p.Type != nil {
			kylixType = typeExprName(p.Type)
			llvmT = LLVMType(kylixType)
		}
		suffix := "_int"
		switch llvmT {
		case "i1":
			suffix = "_bool"
		case "double":
			suffix = "_real"
		case "ptr":
			suffix = "_str"
		}
		allocaReg := fmt.Sprintf("%%v_%s%s", p.Name, suffix)
		g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, llvmT))
		g.line(fmt.Sprintf("  store %s %%%s, ptr %s", llvmT, p.Name, allocaReg))
		g.locals[p.Name] = allocaReg
		if kylixType != "" {
			g.localTypes[p.Name] = kylixType
		}
	}

	// Result alloca for functions with a return type.
	if pl.retType != "void" && !pl.retExpr {
		g.line(fmt.Sprintf("  %%result = alloca %s, align 8", pl.retType))
		g.locals["result"] = "%result"
	}

	// Emit body.
	if pl.retExpr {
		// Expression-bodied: ret <expr>.
		bodyExpr, ok := pl.body.(ast.Expression)
		if !ok {
			g.line("  ret void")
		} else {
			v, t, err := g.emitExpr(bodyExpr)
			if err != nil {
				g.locals = savedLocals
				g.localTypes = savedTypes
				g.varNameSeq = savedVarSeq
				g.funcName = savedFuncName
				return err
			}
			if t != pl.retType {
				v, t = g.coerceValue(v, t, pl.retType)
			}
			g.line(fmt.Sprintf("  ret %s %s", pl.retType, v))
		}
	} else {
		// Block-bodied.
		if blk, ok := pl.body.(*ast.BlockStatement); ok {
			for _, st := range blk.Statements {
				if err := g.emitStatement(st); err != nil {
					g.locals = savedLocals
					g.localTypes = savedTypes
					g.varNameSeq = savedVarSeq
					g.funcName = savedFuncName
					return err
				}
			}
		}
		if pl.retType == "void" {
			g.line("  ret void")
		} else {
			r := g.tmp()
			g.line(fmt.Sprintf("  %s = load %s, ptr %%result", r, pl.retType))
			g.line(fmt.Sprintf("  ret %s %s", pl.retType, r))
		}
	}

	g.line("}")
	g.line("")

	// Restore outer scope.
	g.locals = savedLocals
	g.localTypes = savedTypes
	g.varNameSeq = savedVarSeq
	g.funcName = savedFuncName
	return nil
}
