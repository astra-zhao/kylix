// stmt.go — LLVM IR code generation for Kylix statements.
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// emitStatement generates code for a single statement.
func (g *Generator) emitStatement(node ast.Statement) error {
	switch s := node.(type) {
	case *ast.AssignmentStatement:
		return g.emitAssign(s)
	case *ast.ExpressionStatement:
		_, _, err := g.emitExpr(s.Expression)
		return err
	case *ast.BlockStatement:
		for _, stmt := range s.Statements {
			if err := g.emitStatement(stmt); err != nil {
				return err
			}
		}
		return nil
	case *ast.IfStatement:
		return g.emitIf(s)
	case *ast.WhileStatement:
		return g.emitWhile(s)
	case *ast.ForStatement:
		return g.emitFor(s)
	case *ast.RepeatStatement:
		return g.emitRepeat(s)
	case *ast.VarDecl:
		return g.emitVarDecl(s)
	case *ast.ReturnStatement:
		return g.emitReturn(s)
	case *ast.TryStatement:
		return g.emitTry(s)
	case *ast.RaiseStatement:
		return g.emitRaise(s)
	case *ast.ForEachStatement:
		return g.emitForEach(s)
	case *ast.CaseStatement:
		return g.emitCase(s)
	case *ast.MatchStatement:
		return g.emitMatch(s)
	case *ast.BreakStatement:
		return g.emitBreak()
	case *ast.ContinueStatement:
		return g.emitContinue()
	default:
		return nil
	}
}

// emitFunctionDecl generates an LLVM function definition.
func (g *Generator) emitFunctionDecl(decl *ast.FunctionDecl) error {
	if decl.Body == nil {
		return nil // forward declaration, skip
	}

	// Determine return type
	retType := "void"
	if decl.ReturnType != nil {
		retType = LLVMType(typeExprName(decl.ReturnType))
	}

	// Build parameter list
	var params []string
	for _, p := range decl.Parameters {
		llvmT := "i64"
		if p.Type != nil {
			llvmT = LLVMType(typeExprName(p.Type))
		}
		params = append(params, fmt.Sprintf("%s %%%s", llvmT, p.Name))
	}

	g.line(fmt.Sprintf("define %s @%s(%s) {", retType, decl.Name, strings.Join(params, ", ")))
	g.line("entry:")
	g.funcName = decl.Name
	savedLocals := g.locals
	savedTypes := g.localTypes
	g.locals = make(map[string]string)
	g.localTypes = make(map[string]string)

	// Allocate result variable for functions
	if retType != "void" {
		g.line(fmt.Sprintf("  %%result = alloca %s, align 8", retType))
		g.locals["result"] = "%result"
	}

	// Allocate parameters as locals
	for _, p := range decl.Parameters {
		llvmT := "i64"
		if p.Type != nil {
			llvmT = LLVMType(typeExprName(p.Type))
		}
		allocaReg := fmt.Sprintf("%%v_%s", p.Name)
		g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, llvmT))
		g.line(fmt.Sprintf("  store %s %%%s, ptr %s", llvmT, p.Name, allocaReg))
		g.locals[p.Name] = allocaReg
		if p.Type != nil {
			g.localTypes[p.Name] = typeExprName(p.Type)
		}
	}

	// Emit local declarations
	for _, ld := range decl.LocalDecls {
		if vd, ok := ld.(*ast.VarDecl); ok {
			if err := g.emitVarDecl(vd); err != nil {
				return err
			}
		}
	}

	// Emit body
	if decl.Body != nil {
		for _, stmt := range decl.Body.Statements {
			if err := g.emitStatement(stmt); err != nil {
				return err
			}
		}
	}

	// Return result
	if retType != "void" {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load %s, ptr %%result", r, retType))
		g.line(fmt.Sprintf("  ret %s %s", retType, r))
	} else {
		g.line("  ret void")
	}

	g.line("}")
	g.line("")
	g.locals = savedLocals
	g.localTypes = savedTypes
	return nil
}

// emitVarDecl allocates stack space for a variable.
func (g *Generator) emitVarDecl(s *ast.VarDecl) error {
	// VarDecl has Names []string — handle all names (e.g., x, y: Integer).
	if len(s.Names) == 0 {
		return nil
	}

	// Emit alloca for each variable in the declaration.
	for _, name := range s.Names {
		if err := g.emitVarDeclSingle(name, s.Type); err != nil {
			return err
		}
	}
	return nil
}

// emitVarDeclSingle allocates stack space for a single variable.
func (g *Generator) emitVarDeclSingle(name string, varType ast.Expression) error {
	// Array type: dispatch to dedicated handler (Milestone 2).
	if arrT, ok := varType.(*ast.ArrayType); ok {
		g.emitArrayVarDecl(name, arrT)
		return nil
	}

	// Interface-typed local: reserve { vtable, data } pair allocas.
	if varType != nil {
		if tname := typeExprName(varType); tname != "" {
			if _, isIface := g.interfaces[tname]; isIface {
				g.emitInterfaceVarDecl(name)
				g.localTypes[name] = tname
				return nil
			}
		}
	}

	// Generic instantiation: TBox<Integer> → record the mangled type, then
	// allocate a pointer slot (class instances are heap-allocated and the
	// local holds a ptr to the struct).
	if gt, ok := varType.(*ast.GenericType); ok {
		mangled := mangleGeneric(gt.Base, gt.TypeParams)
		if mangled != "" {
			allocaReg := fmt.Sprintf("%%v_%s", name)
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
			g.line(fmt.Sprintf("  store ptr null, ptr %s", allocaReg))
			g.locals[name] = allocaReg
			g.localTypes[name] = mangled
			return nil
		}
	}

	// Plain class-typed local: hold a ptr to the heap-allocated instance.
	if ident, ok := varType.(*ast.Identifier); ok {
		if _, isClass := g.classes[ident.Value]; isClass {
			allocaReg := fmt.Sprintf("%%v_%s", name)
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
			g.line(fmt.Sprintf("  store ptr null, ptr %s", allocaReg))
			g.locals[name] = allocaReg
			g.localTypes[name] = ident.Value
			return nil
		}
	}

	llvmT := "i64"
	suffix := "_int"
	kylixType := ""
	if varType != nil {
		tname := typeExprName(varType)
		kylixType = tname
		llvmT = LLVMType(tname)
		switch strings.ToLower(tname) {
		case "boolean", "bool":
			suffix = "_bool"
		case "real", "double":
			suffix = "_real"
		case "string":
			suffix = "_str"
		}
	}
	allocaReg := fmt.Sprintf("%%v_%s%s", name, suffix)
	g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, llvmT))

	// Zero-initialize
	switch llvmT {
	case "ptr":
		g.line(fmt.Sprintf("  store ptr null, ptr %s", allocaReg))
	case "i1":
		g.line(fmt.Sprintf("  store i1 0, ptr %s", allocaReg))
	case "double":
		g.line(fmt.Sprintf("  store double 0.0, ptr %s", allocaReg))
	default:
		g.line(fmt.Sprintf("  store i64 0, ptr %s", allocaReg))
	}

	g.locals[name] = allocaReg
	if kylixType != "" {
		g.localTypes[name] = kylixType
	}
	return nil
}

// emitAssign generates a store instruction.
func (g *Generator) emitAssign(s *ast.AssignmentStatement) error {
	// LHS may be an interface-typed local — handle boxing before evaluating value
	// so we can pick the right per-class vtable.
	if ident, ok := s.Name.(*ast.Identifier); ok {
		if ifaceName, isIface := g.localTypes[ident.Value]; isIface {
			if _, known := g.interfaces[ifaceName]; known {
				if vtableReg, dataReg, ok := g.evalInterfaceRHS(s.Value, ifaceName); ok {
					g.emitInterfaceAssign(ident.Value, vtableReg, dataReg)
					return nil
				}
			}
		}
	}

	v, t, err := g.emitExpr(s.Value)
	if err != nil {
		return err
	}

	// Handle array element assignment: arr[i] := value
	if idx, ok := s.Name.(*ast.IndexExpression); ok {
		ptrReg, elemType, err := g.emitArrayIndex(idx, true)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("  store %s %s, ptr %s", elemType, v, ptrReg))
		return nil
	}

	// s.Name is Expression, extract identifier name
	varName := ""
	if ident, ok := s.Name.(*ast.Identifier); ok {
		varName = ident.Value
	} else {
		// Complex lvalue (tuple destructuring: (q, r) := func()) or other
		// unsupported LHS forms. Multi-return lowering in LLVM requires
		// struct return types + extractvalue — deferred. Emit a comment stub
		// so the IR remains valid.
		g.line(fmt.Sprintf("  ; complex lvalue assignment (tuple destructuring) deferred"))
		return nil
	}

	allocaReg, ok := g.locals[varName]
	if !ok {
		// Auto-declare as i64
		allocaReg = fmt.Sprintf("%%v_%s_int", varName)
		g.line(fmt.Sprintf("  %s = alloca i64, align 8", allocaReg))
		g.locals[varName] = allocaReg
		t = "i64"
	}

	// Infer actual type from alloca name
	actualType := "i64"
	if strings.HasSuffix(allocaReg, "_bool") {
		actualType = "i1"
	} else if strings.HasSuffix(allocaReg, "_real") {
		actualType = "double"
	} else if strings.HasSuffix(allocaReg, "_str") {
		actualType = "ptr"
	} else if allocaReg == "%result" && t != "" {
		actualType = t
	}

	// Type coercion: if RHS type doesn't match the alloca type, cast it.
	if t != actualType {
		cast := g.tmp()
		switch {
		case t == "i1" && actualType == "i64":
			g.line(fmt.Sprintf("  %s = zext i1 %s to i64", cast, v))
			v = cast
		case t == "i64" && actualType == "i1":
			// i64 → i1: truncate or compare to zero
			cmp := g.tmp()
			g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", cmp, v))
			v = cmp
		case t == "i64" && actualType == "double":
			g.line(fmt.Sprintf("  %s = sitofp i64 %s to double", cast, v))
			v = cast
		case t == "double" && actualType == "i64":
			g.line(fmt.Sprintf("  %s = fptosi double %s to i64", cast, v))
			v = cast
		}
	}

	g.line(fmt.Sprintf("  store %s %s, ptr %s", actualType, v, allocaReg))
	return nil
}

// emitReturn generates a return via the result variable.
func (g *Generator) emitReturn(s *ast.ReturnStatement) error {
	if s.Value != nil {
		v, t, err := g.emitExpr(s.Value)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("  store %s %s, ptr %%result", t, v))
	}
	// Jump to exit label (we use a single exit block approach)
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", exitLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	return nil
}

// emitIf generates if/then/else as LLVM conditional branches.
func (g *Generator) emitIf(s *ast.IfStatement) error {
	cond, _, err := g.emitExpr(s.Condition)
	if err != nil {
		return err
	}

	thenLbl := g.label()
	mergeLbl := g.label()
	elseLbl := mergeLbl
	if s.Alternative != nil {
		elseLbl = g.label()
	}

	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", cond, thenLbl, elseLbl))

	// Then block
	g.line(fmt.Sprintf("%s:", thenLbl))
	if err := g.emitStatement(s.Consequence); err != nil {
		return err
	}
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))

	// Else block
	if s.Alternative != nil {
		g.line(fmt.Sprintf("%s:", elseLbl))
		if err := g.emitStatement(s.Alternative); err != nil {
			return err
		}
		g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	}

	// Merge block
	g.line(fmt.Sprintf("%s:", mergeLbl))
	return nil
}

// emitWhile generates a while loop using a header/body/exit pattern.
func (g *Generator) emitWhile(s *ast.WhileStatement) error {
	headerLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()

	g.line(fmt.Sprintf("  br label %%%s", headerLbl))
	g.line(fmt.Sprintf("%s:", headerLbl))

	cond, _, err := g.emitExpr(s.Condition)
	if err != nil {
		return err
	}
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", cond, bodyLbl, exitLbl))

	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, headerLbl
	g.line(fmt.Sprintf("%s:", bodyLbl))
	if err := g.emitStatement(s.Body); err != nil {
		return err
	}
	g.breakLabel, g.continueLabel = savedBreak, savedContinue
	g.line(fmt.Sprintf("  br label %%%s", headerLbl))

	g.line(fmt.Sprintf("%s:", exitLbl))
	return nil
}

// emitFor generates a counted for loop.
func (g *Generator) emitFor(s *ast.ForStatement) error {
	// Allocate loop variable
	counterReg := fmt.Sprintf("%%v_%s_int", s.Variable)
	if _, exists := g.locals[s.Variable]; !exists {
		g.line(fmt.Sprintf("  %s = alloca i64, align 8", counterReg))
		g.locals[s.Variable] = counterReg
	} else {
		counterReg = g.locals[s.Variable]
	}

	// Initialize
	startV, _, err := g.emitExpr(s.From)
	if err != nil {
		return err
	}
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", startV, counterReg))

	headerLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()

	g.line(fmt.Sprintf("  br label %%%s", headerLbl))
	g.line(fmt.Sprintf("%s:", headerLbl))

	// Condition: counter <= end (DownTo: counter >= end)
	curV := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curV, counterReg))
	endV, _, err := g.emitExpr(s.To)
	if err != nil {
		return err
	}
	condV := g.tmp()
	if s.DownTo {
		g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", condV, curV, endV))
	} else {
		g.line(fmt.Sprintf("  %s = icmp sle i64 %s, %s", condV, curV, endV))
	}
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", condV, bodyLbl, exitLbl))

	// Body — save/restore break/continue targets for nested loops.
	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, headerLbl
	g.line(fmt.Sprintf("%s:", bodyLbl))
	if err := g.emitStatement(s.Body); err != nil {
		return err
	}
	g.breakLabel, g.continueLabel = savedBreak, savedContinue

	// Increment/decrement
	stepV := g.tmp()
	curV2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curV2, counterReg))
	if s.DownTo {
		g.line(fmt.Sprintf("  %s = sub i64 %s, 1", stepV, curV2))
	} else {
		g.line(fmt.Sprintf("  %s = add i64 %s, 1", stepV, curV2))
	}
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", stepV, counterReg))
	g.line(fmt.Sprintf("  br label %%%s", headerLbl))

	g.line(fmt.Sprintf("%s:", exitLbl))
	return nil
}

// emitRepeat generates a repeat...until loop.
func (g *Generator) emitRepeat(s *ast.RepeatStatement) error {
	bodyLbl := g.label()
	exitLbl := g.label()

	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, bodyLbl

	g.line(fmt.Sprintf("  br label %%%s", bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))

	if err := g.emitStatement(s.Body); err != nil {
		return err
	}

	g.breakLabel, g.continueLabel = savedBreak, savedContinue

	cond, _, err := g.emitExpr(s.Condition)
	if err != nil {
		return err
	}
	// repeat until cond → loop while !cond
	notCond := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i1 %s, 1", notCond, cond))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", notCond, bodyLbl, exitLbl))

	g.line(fmt.Sprintf("%s:", exitLbl))
	return nil
}

// raiseExceptionTypeName extracts the exception class name from a raise
// expression. Handles `T.Create(...)` (constructor, both no-arg MemberExpression
// and arg-bearing CallExpression forms) and a bare class instance variable.
// Returns "" if the type cannot be determined (→ generic ID 0).
func raiseExceptionTypeName(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	// Unwrap call: raise T.Create('msg') → CallExpression{Function: MemberExpression}.
	if call, ok := expr.(*ast.CallExpression); ok {
		return raiseExceptionTypeName(call.Function)
	}
	if m, ok := expr.(*ast.MemberExpression); ok && m.Member == "Create" {
		if ident, ok := m.Object.(*ast.Identifier); ok {
			return ident.Value
		}
	}
	return typeExprName(expr)
}

// emitRaise generates IR for `raise <expr>` or bare `raise`.
//
//	raise Exc.Create('msg')  →  store obj+type into the global slot, longjmp
//	raise                     →  re-throw the in-flight exception (longjmp outer)
func (g *Generator) emitRaise(s *ast.RaiseStatement) error {
	if s.Exception == nil {
		// Bare raise: only valid inside an except handler. If we're not in one,
		// fall back to raising a generic Exception (matches Go backend behavior).
		if !g.inExceptHandler {
			return g.emitRaiseGeneric()
		}
		// Re-throw: the global slot still holds the current exception. Keep
		// exc_active=true and longjmp to the outer handler.
		return g.emitLongjmpToHandler()
	}

	// Evaluate the exception expression → object pointer.
	objReg, _, err := g.emitExpr(s.Exception)
	if err != nil {
		return err
	}
	typeName := raiseExceptionTypeName(s.Exception)
	tid := g.excTypeID(typeName)

	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_exc_obj", objReg))
	g.line(fmt.Sprintf("  store i32 %d, ptr @__kylix_exc_type", tid))
	g.line("  store i1 true, ptr @__kylix_exc_active")
	return g.emitLongjmpToHandler()
}

// emitRaiseGeneric raises a synthetic Exception with a default message, used
// when bare `raise` appears outside any except handler.
func (g *Generator) emitRaiseGeneric() error {
	msg := g.addString("exception")
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_exc_obj", msg))
	g.line("  store i32 1, ptr @__kylix_exc_type") // Exception = ID 1
	g.line("  store i1 true, ptr @__kylix_exc_active")
	return g.emitLongjmpToHandler()
}

// emitLongjmpToHandler loads the current handler's jmpbuf and longjmps to it.
// If no handler is installed (jmpbuf is null), the program exits with status 70
// (EX_SOFTWARE) — an uncaught exception.
func (g *Generator) emitLongjmpToHandler() error {
	jb := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_jmpbuf", jb))
	nz := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", nz, jb))
	hasLbl := g.label()
	noLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nz, hasLbl, noLbl))

	g.line(fmt.Sprintf("%s:", hasLbl))
	g.line(fmt.Sprintf("  call void @longjmp(ptr %s, i32 1)", jb))
	g.line("  unreachable")

	g.line(fmt.Sprintf("%s:", noLbl))
	g.line("  call void @exit(i32 70)")
	g.line("  unreachable")
	return nil
}

// emitTry generates IR for try/except/finally.
//
// Control-flow shape:
//
//	setjmp → try_body (install handler, run body, pop, → finally_normal)
//	       ↘ except_dispatch (pop, match on-clauses by type ID, → finally_exc
//	                          or finally_reraise if uncaught)
//
//	finally is emitted up to three times (normal/exc/reraise) so it always
//	runs — longjmp would otherwise skip cleanup. Nesting is supported by
//	saving/restoring @__kylix_jmpbuf (a stack of handlers via an alloca slot).
func (g *Generator) emitTry(s *ast.TryStatement) error {
	// alloca for the setjmp buffer and for saving the outer handler pointer.
	bufReg := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [%d x i8], align 16", bufReg, excJmpBufSize))
	bufptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr [%d x i8], ptr %s, i64 0, i64 0", bufptr, excJmpBufSize, bufReg))
	oldJBSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", oldJBSlot))

	// setjmp: returns 0 on first call, non-zero when longjmp returns here.
	rc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @setjmp(ptr %s)", rc, bufptr))
	isHandler := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", isHandler, rc))
	tryBodyLbl := g.label()
	exceptLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isHandler, exceptLbl, tryBodyLbl))

	// ── try body ──────────────────────────────────────────────
	g.line(fmt.Sprintf("%s:", tryBodyLbl))
	oldJB := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_jmpbuf", oldJB))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", oldJB, oldJBSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_jmpbuf", bufptr))
	g.line("  store i1 false, ptr @__kylix_exc_active")
	if s.Body != nil {
		for _, st := range s.Body.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	// Pop handler, clear active, fall through to finally (normal path).
	restoredJB := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", restoredJB, oldJBSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_jmpbuf", restoredJB))
	g.line("  store i1 false, ptr @__kylix_exc_active")

	finallyNormalLbl := g.label()
	finallyExcLbl := g.label()
	finallyReraiseLbl := g.label()
	endLbl := g.label()

	g.line(fmt.Sprintf("  br label %%%s", finallyNormalLbl))

	// ── except dispatch ───────────────────────────────────────
	g.line(fmt.Sprintf("%s:", exceptLbl))
	// Pop the handler installed by try_body (restore outer).
	restoredJB2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", restoredJB2, oldJBSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_jmpbuf", restoredJB2))

	tid := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr @__kylix_exc_type", tid))

	// Match on-clauses in order. Each emits a subtype check and branches to
	// its body or the next check.
	nextCheck := exceptLbl
	matched := false
	for _, on := range s.OnClauses {
		onBodyLbl := g.label()
		thisCheck := nextCheck
		nextCheck = g.label()
		wantID := g.excTypeID(typeExprName(on.Type))

		m := g.tmp()
		g.line(fmt.Sprintf("  %s = call i1 @__kylix_is_subtype(i32 %s, i32 %d)", m, tid, wantID))
		// The check is emitted under the current "thisCheck" label (the first
		// one reuses exceptLbl which we already emitted above).
		if thisCheck != exceptLbl {
			g.line(fmt.Sprintf("%s:", thisCheck))
		}
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", m, onBodyLbl, nextCheck))

		// on-body: bind E to the exception object, run body, clear active.
		g.line(fmt.Sprintf("%s:", onBodyLbl))
		if on.Variable != "" {
			eAlloca := fmt.Sprintf("%%v_%s", on.Variable)
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", eAlloca))
			obj := g.tmp()
			g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_exc_obj", obj))
			g.line(fmt.Sprintf("  store ptr %s, ptr %s", obj, eAlloca))
			g.locals[on.Variable] = eAlloca
			if t := typeExprName(on.Type); t != "" {
				g.localTypes[on.Variable] = t
			}
		}
		g.inExceptHandler = true
		if on.Body != nil {
			for _, st := range on.Body.Statements {
				if err := g.emitStatement(st); err != nil {
					g.inExceptHandler = false
					return err
				}
			}
		}
		g.inExceptHandler = false
		g.line("  store i1 false, ptr @__kylix_exc_active")
		g.line(fmt.Sprintf("  br label %%%s", finallyExcLbl))
		matched = true
	}

	// After the last on-clause check, emit the fall-through label.
	if len(s.OnClauses) > 0 {
		g.line(fmt.Sprintf("%s:", nextCheck))
	}

	// No on-clause matched (or none present):
	//   - a plain ExceptBlock handles everything → run it, → finally_exc
	//   - otherwise the exception stays active → finally_reraise
	if s.ExceptBlock != nil {
		g.inExceptHandler = true
		for _, st := range s.ExceptBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				g.inExceptHandler = false
				return err
			}
		}
		g.inExceptHandler = false
		g.line("  store i1 false, ptr @__kylix_exc_active")
		g.line(fmt.Sprintf("  br label %%%s", finallyExcLbl))
	} else {
		// Uncaught: keep exc_active=true so the reraise path re-throws.
		g.line(fmt.Sprintf("  br label %%%s", finallyReraiseLbl))
	}
	_ = matched

	// ── finally: normal path (try body completed) ─────────────
	g.line(fmt.Sprintf("%s:", finallyNormalLbl))
	if s.FinallyBlock != nil {
		for _, st := range s.FinallyBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	g.line(fmt.Sprintf("  br label %%%s", endLbl))

	// ── finally: except handled path ──────────────────────────
	g.line(fmt.Sprintf("%s:", finallyExcLbl))
	if s.FinallyBlock != nil {
		for _, st := range s.FinallyBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	g.line(fmt.Sprintf("  br label %%%s", endLbl))

	// ── finally: reraise path (uncaught exception) ────────────
	g.line(fmt.Sprintf("%s:", finallyReraiseLbl))
	if s.FinallyBlock != nil {
		for _, st := range s.FinallyBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	// Re-throw: longjmp to the outer handler (current @__kylix_jmpbuf).
	outerJB := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_jmpbuf", outerJB))
	g.line(fmt.Sprintf("  call void @longjmp(ptr %s, i32 1)", outerJB))
	g.line("  unreachable")

	g.line(fmt.Sprintf("%s:", endLbl))
	return nil
}

// emitBreak generates a branch to the enclosing loop's exit label.
func (g *Generator) emitBreak() error {
	if g.breakLabel == "" {
		return fmt.Errorf("break outside loop")
	}
	// After an unconditional branch the current block is terminated; the
	// following instructions need a fresh (unreachable) label so the IR
	// remains structurally valid.
	g.line(fmt.Sprintf("  br label %%%s", g.breakLabel))
	deadLbl := g.label()
	g.line(fmt.Sprintf("%s:", deadLbl))
	return nil
}

// emitContinue generates a branch to the enclosing loop's header label.
func (g *Generator) emitContinue() error {
	if g.continueLabel == "" {
		return fmt.Errorf("continue outside loop")
	}
	g.line(fmt.Sprintf("  br label %%%s", g.continueLabel))
	deadLbl := g.label()
	g.line(fmt.Sprintf("%s:", deadLbl))
	return nil
}

// emitForEach generates a counted for-over-index loop for `for x in arr do`.
// LLVM has no built-in range; we lower it as a 0-based counted loop using
// the array's Length (via a strlen-style i64 field read where available, or
// a conservative fixed bound of 0 for unknown iterables). This covers the
// common case of iterating array of T where the length is statically known.
func (g *Generator) emitForEach(s *ast.ForEachStatement) error {
	// Emit the iterable expression to get a pointer/value.
	iterReg, _, err := g.emitExpr(s.Iterable)
	if err != nil {
		return err
	}

	// Allocate a loop variable alloca for the element.
	elemAlloca := fmt.Sprintf("%%v_%s", s.Variable)
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", elemAlloca))
	g.locals[s.Variable] = elemAlloca

	// Use a simple i64 index counter.
	idxAlloca := g.tmp() + "_foreach_idx"
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", idxAlloca))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", idxAlloca))

	// For a string (ptr), use strlen as the bound. For other types treat
	// bound as 0 (body never executes) — a conservative safe default.
	boundReg := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", boundReg, iterReg))

	headerLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()

	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, headerLbl

	g.line(fmt.Sprintf("  br label %%%s", headerLbl))
	g.line(fmt.Sprintf("%s:", headerLbl))
	idxCur := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", idxCur, idxAlloca))
	condReg := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", condReg, idxCur, boundReg))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", condReg, bodyLbl, exitLbl))

	g.line(fmt.Sprintf("%s:", bodyLbl))
	// Load element at current index (char from string / element from array).
	elemPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", elemPtr, iterReg, idxCur))
	elemVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", elemVal, elemPtr))
	elemExt := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i8 %s to i64", elemExt, elemVal))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", elemExt, elemAlloca))

	if err := g.emitStatement(s.Body); err != nil {
		return err
	}
	g.breakLabel, g.continueLabel = savedBreak, savedContinue

	// Increment index.
	idxNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", idxNext, idxCur))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", idxNext, idxAlloca))
	g.line(fmt.Sprintf("  br label %%%s", headerLbl))

	g.line(fmt.Sprintf("%s:", exitLbl))
	return nil
}

// emitCase lowers a Pascal case statement to an LLVM switch instruction.
// case expr of 1: ... 2,3: ... else ... end
func (g *Generator) emitCase(s *ast.CaseStatement) error {
	exprReg, exprType, err := g.emitExpr(s.Expression)
	if err != nil {
		return err
	}
	// Default type for case selector is i64.
	if exprType == "" {
		exprType = "i64"
	}

	defaultLbl := g.label()
	mergeLbl := g.label()

	// Collect branch labels.
	type branchEntry struct {
		values []int64
		lbl    string
	}
	var branches []branchEntry
	for _, br := range s.Branches {
		lbl := g.label()
		var vals []int64
		for _, v := range br.Values {
			if lit, ok := v.(*ast.IntegerLiteral); ok {
				vals = append(vals, lit.Value)
			}
		}
		branches = append(branches, branchEntry{vals, lbl})
	}

	// Emit switch instruction.
	var switchCases []string
	for _, br := range branches {
		for _, val := range br.values {
			switchCases = append(switchCases, fmt.Sprintf("%s %d, label %%%s", exprType, val, br.lbl))
		}
	}
	if len(switchCases) > 0 {
		g.line(fmt.Sprintf("  switch %s %s, label %%%s [ %s ]",
			exprType, exprReg, defaultLbl, strings.Join(switchCases, " ")))
	} else {
		g.line(fmt.Sprintf("  br label %%%s", defaultLbl))
	}

	// Emit branch bodies.
	for i, br := range s.Branches {
		g.line(fmt.Sprintf("%s:", branches[i].lbl))
		if br.Body != nil {
			for _, st := range br.Body.Statements {
				if err := g.emitStatement(st); err != nil {
					return err
				}
			}
		}
		g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	}

	// Default / else branch.
	g.line(fmt.Sprintf("%s:", defaultLbl))
	if s.ElseBranch != nil {
		for _, st := range s.ElseBranch.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))

	g.line(fmt.Sprintf("%s:", mergeLbl))
	return nil
}

// emitMatch lowers a Pascal match statement to a chain of conditional branches.
// match expr { pat => body; _ => default }
func (g *Generator) emitMatch(s *ast.MatchStatement) error {
	exprReg, _, err := g.emitExpr(s.Expression)
	if err != nil {
		return err
	}

	mergeLbl := g.label()
	// tmp alloca for the match value so it is available in each comparison.
	matchAlloca := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", matchAlloca))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", exprReg, matchAlloca))

	for _, br := range s.Branches {
		// Wildcard / default
		if ident, ok := br.Pattern.(*ast.Identifier); ok && ident.Value == "_" {
			if br.Body != nil {
				for _, st := range br.Body.Statements {
					if err := g.emitStatement(st); err != nil {
						return err
					}
				}
			}
			g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
			continue
		}

		// Build condition: _v == p (possibly OR with additional patterns).
		mv := g.tmp()
		g.line(fmt.Sprintf("  %s = load i64, ptr %s", mv, matchAlloca))
		allPats := []ast.Expression{br.Pattern}
		allPats = append(allPats, br.AdditionalPatterns...)
		condReg := ""
		for _, pat := range allPats {
			pv, _, err := g.emitExpr(pat)
			if err != nil {
				return err
			}
			cmp := g.tmp()
			g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", cmp, mv, pv))
			if condReg == "" {
				condReg = cmp
			} else {
				or := g.tmp()
				g.line(fmt.Sprintf("  %s = or i1 %s, %s", or, condReg, cmp))
				condReg = or
			}
		}
		// Optional when guard.
		if br.When != nil {
			guardReg, _, err := g.emitExpr(br.When)
			if err != nil {
				return err
			}
			and := g.tmp()
			g.line(fmt.Sprintf("  %s = and i1 %s, %s", and, condReg, guardReg))
			condReg = and
		}

		bodyLbl := g.label()
		nextLbl := g.label()
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", condReg, bodyLbl, nextLbl))
		g.line(fmt.Sprintf("%s:", bodyLbl))
		if br.Body != nil {
			for _, st := range br.Body.Statements {
				if err := g.emitStatement(st); err != nil {
					return err
				}
			}
		}
		g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
		g.line(fmt.Sprintf("%s:", nextLbl))
	}

	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	g.line(fmt.Sprintf("%s:", mergeLbl))
	return nil
}
