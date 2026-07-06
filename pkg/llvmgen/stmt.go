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
		return g.emitBlockScoped(s)
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
	case *ast.InheritedStatement:
		return g.emitInherited(s)
	default:
		return nil
	}
}

// emitBlockScoped runs a BlockStatement's statements with Kylix block-scoping
// semantics for locals: `var` declarations inside the block (and anything
// they register — locals/localTypes/arrayInfo/closureLocals/closureSigs/
// closureParams) are visible only within the block and are rolled back once
// it exits, so a sibling block (the other branch of an if, a later loop
// iteration's body reusing the same AST node, an except handler, ...) can
// declare a variable of the same name without seeing — or being seen by —
// this block's binding. The underlying LLVM alloca is NOT removed (it lives
// for the whole function per LLVM's SSA rules and freshVarReg already gave it
// a unique register name); only the Kylix-level name→register visibility is
// scoped. varNameSeq is intentionally NOT rolled back, so a later sibling
// declaration of the same source name still gets a fresh register rather than
// colliding with this block's (now invisible but still-live) alloca.
func (g *Generator) emitBlockScoped(s *ast.BlockStatement) error {
	savedLocals := g.locals
	savedTypes := g.localTypes
	savedArrayInfo := g.arrayInfo
	savedClosureLocals := g.closureLocals
	savedClosureSigs := g.closureSigs
	savedClosureParams := g.closureParams

	g.locals = cloneStringMap(g.locals)
	g.localTypes = cloneStringMap(g.localTypes)
	g.arrayInfo = cloneArrayInfoMap(g.arrayInfo)
	g.closureLocals = cloneBoolMap(g.closureLocals)
	g.closureSigs = cloneStringMap(g.closureSigs)
	g.closureParams = cloneStringSliceMap(g.closureParams)

	var err error
	for _, stmt := range s.Statements {
		if err = g.emitStatement(stmt); err != nil {
			break
		}
	}

	g.locals = savedLocals
	g.localTypes = savedTypes
	g.arrayInfo = savedArrayInfo
	g.closureLocals = savedClosureLocals
	g.closureSigs = savedClosureSigs
	g.closureParams = savedClosureParams
	return err
}

func cloneStringMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneBoolMap(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneStringSliceMap(m map[string][]string) map[string][]string {
	out := make(map[string][]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneArrayInfoMap(m map[string]*arrayInfo) map[string]*arrayInfo {
	out := make(map[string]*arrayInfo, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// emitFunctionDecl generates an LLVM function definition.
func (g *Generator) emitFunctionDecl(decl *ast.FunctionDecl) error {
	if decl.Body == nil {
		return nil // forward declaration, skip
	}

	// Determine return type: check multi-return first, then single, else void.
	retType := "void"
	isMultiRet := false
	if multiTypes := g.multiRetTypes[decl.Name]; len(multiTypes) > 0 {
		retType = fmt.Sprintf("%%__ret_%s", decl.Name)
		isMultiRet = true
	} else if decl.ReturnType != nil {
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
	savedVarSeq := g.varNameSeq
	g.locals = make(map[string]string)
	g.localTypes = make(map[string]string)
	g.varNameSeq = make(map[string]int)

	// Allocate result variable for functions
	if retType != "void" {
		g.line(fmt.Sprintf("  %%result = alloca %s, align 8", retType))
		g.locals["result"] = "%result"
		if isMultiRet {
			// Mark result as a tuple so assignment can detect it.
			g.localTypes["result"] = "__tuple__"
		}
	}

	// Allocate parameters as locals
	for _, p := range decl.Parameters {
		llvmT := "i64"
		kylixType := ""
		if p.Type != nil {
			kylixType = typeExprName(p.Type)
			llvmT = LLVMType(kylixType)
		}
		// Use suffix convention so emitIdentLoad can infer type from alloca name.
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

	// Emit local declarations
	for _, ld := range decl.LocalDecls {
		if vd, ok := ld.(*ast.VarDecl); ok {
			if err := g.emitVarDecl(vd); err != nil {
				return err
			}
		} else if cd, ok := ld.(*ast.ConstDecl); ok {
			// Register local constant (resolved at use site)
			g.constants[cd.Name] = cd.Value
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
	g.varNameSeq = savedVarSeq
	return nil
}

// emitVarDecl allocates stack space for a variable.
func (g *Generator) emitVarDecl(s *ast.VarDecl) error {
	// VarDecl has Names []string — handle all names (e.g., x, y: Integer).
	if len(s.Names) == 0 {
		return nil
	}

	// For type-inferred declarations (var x := expr), we need to evaluate the
	// expression first to determine its LLVM type, then emit the alloca with the
	// correct type, then store the value.
	if s.Type == nil && s.Value != nil && s.Inferred {
		// Lambda inference: emitLambda returns an alloca holding the closure
		// pair directly — use it as the variable's own storage (no separate
		// alloca + store). Mark as a closure for indirect-call codegen.
		if lam, ok := s.Value.(*ast.LambdaExpression); ok {
			closureReg, _, err := g.emitLambda(lam)
			if err != nil {
				return err
			}
			retT := "void"
			if lam.ReturnType != nil {
				retT = LLVMType(typeExprName(lam.ReturnType))
			} else if _, isBlock := lam.Body.(*ast.BlockStatement); !isBlock {
				retT = "i64"
			}
			var paramTypes []string
			for _, p := range lam.Parameters {
				pt := "i64"
				if p.Type != nil {
					pt = LLVMType(typeExprName(p.Type))
				}
				paramTypes = append(paramTypes, pt)
			}
			for _, name := range s.Names {
				g.locals[name] = closureReg
				g.closureLocals[name] = true
				g.closureSigs[name] = retT
				g.closureParams[name] = paramTypes
			}
			return nil
		}

		// Evaluate the RHS to get its type.
		valReg, llvmType, err := g.emitExpr(s.Value)
		if err != nil {
			return err
		}

		// Constructor inference: `var x := TFoo.Create` produces a ptr, but we
		// must record the class name so later `x.Field` / `x.Method` resolve.
		inferredClass := ""
		if member, ok := s.Value.(*ast.MemberExpression); ok && member.Member == "Create" {
			if ident, ok := member.Object.(*ast.Identifier); ok {
				if _, known := g.classes[ident.Value]; known {
					inferredClass = ident.Value
				}
			}
		}

		// Stdlib opaque-type inference: stdlib module functions may return a
		// pseudo-type name (TDateTime, TTcpConn, TTcpListener, ...) that is NOT
		// a real LLVM type. Treat any non-standard type string as an opaque
		// pointer (ptr) and record the Kylix-side name in localTypes so later
		// method-style dispatch (if any) can recognize it.
		isOpaquePtr := false
		switch llvmType {
		case "i1", "i64", "double", "ptr", "void", "TDateTime":
			if llvmType == "TDateTime" {
				inferredClass = "TDateTime"
			}
		default:
			// Non-standard type name → opaque pointer-backed stdlib handle.
			isOpaquePtr = true
			inferredClass = llvmType
			llvmType = "ptr" // normalize so the switch below picks _str
		}
		_ = isOpaquePtr

		// Allocate variables with the inferred type.
		for _, name := range s.Names {
			suffix := "_int"
			actualLLVMType := llvmType
			switch llvmType {
			case "i1":
				suffix = "_bool"
			case "double":
				suffix = "_real"
			case "ptr":
				suffix = "_str"
			case "TDateTime":
				suffix = "_str"
				actualLLVMType = "ptr"
			}
			allocaReg := g.freshVarReg(name, suffix)
			g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, actualLLVMType))
			g.line(fmt.Sprintf("  store %s %s, ptr %s", actualLLVMType, valReg, allocaReg))
			g.locals[name] = allocaReg
			if inferredClass != "" {
				g.localTypes[name] = inferredClass
			}
		}
		return nil
	}

	// Explicit type or no initializer: emit alloca for each variable.
	for _, name := range s.Names {
		if err := g.emitVarDeclSingle(name, s.Type); err != nil {
			return err
		}
	}

	// If an initializer is present (var x: T = expr, or x := expr with explicit
	// type), emit assignment after alloca.
	if s.Value != nil && !s.Inferred {
		for _, name := range s.Names {
			assignStmt := &ast.AssignmentStatement{
				Name:  &ast.Identifier{Value: name},
				Value: s.Value,
			}
			if err := g.emitAssign(assignStmt); err != nil {
				return err
			}
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

	// Map type: allocate a ptr slot, initialize with htab_new().
	if mapT, ok := varType.(*ast.MapType); ok {
		return g.emitMapVarDecl(name, mapT)
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
			allocaReg := g.freshVarReg(name, "")
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
			allocaReg := g.freshVarReg(name, "")
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
	allocaReg := g.freshVarReg(name, suffix)
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
	// Case 1: Tuple destructuring `(a, b) := Func()` — LHS is TupleLiteral.
	if tuple, ok := s.Name.(*ast.TupleLiteral); ok {
		return g.emitTupleDestructure(tuple, s.Value)
	}

	// Case 2: `result := (a, b)` — assigning TupleLiteral to result in multi-return func.
	if ident, ok := s.Name.(*ast.Identifier); ok && ident.Value == "result" {
		if g.localTypes["result"] == "__tuple__" {
			if tupleLit, ok := s.Value.(*ast.TupleLiteral); ok {
				return g.emitTupleBuild(tupleLit)
			}
		}
	}

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
		// Map element assignment: m[k] := v → htab_put
		if leftIdent, ok := idx.Left.(*ast.Identifier); ok && g.mapVars[leftIdent.Value] {
			return g.emitMapIndexPut(idx, v, t)
		}
		ptrReg, elemType, err := g.emitArrayIndex(idx, true)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("  store %s %s, ptr %s", elemType, v, ptrReg))
		return nil
	}

	// Handle object field assignment: obj.Field := value
	if member, ok := s.Name.(*ast.MemberExpression); ok {
		kind, typeName := g.receiverKind(member.Object)
		if kind == "class" {
			objReg, _, err := g.loadObjectPtr(member.Object, typeName)
			if err != nil {
				return err
			}
			gepReg, fieldType, err := g.emitFieldStore(typeName, objReg, member.Member)
			if err != nil {
				return err
			}
			// Coerce the RHS to the field's declared type.
			if t != fieldType {
				v, t = g.coerceValue(v, t, fieldType)
			}
			g.line(fmt.Sprintf("  store %s %s, ptr %s", t, v, gepReg))
			return nil
		}
		// Non-class member assignment — emit a comment and skip.
		g.line(fmt.Sprintf("  ; unhandled member assignment %s.%s", typeName, member.Member))
		return nil
	}

	// s.Name is Expression, extract identifier name
	varName := ""
	if ident, ok := s.Name.(*ast.Identifier); ok {
		varName = ident.Value
	} else {
		// Unknown LHS form (not handled above).
		g.line(fmt.Sprintf("  ; unhandled LHS %T", s.Name))
		return nil
	}

	allocaReg, ok := g.locals[varName]
	if !ok {
		// Auto-declare as i64
		allocaReg = g.freshVarReg(varName, "_int")
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
	} else if strings.HasSuffix(allocaReg, "_map") {
		actualType = "ptr"
	} else if allocaReg == "%result" && t != "" {
		actualType = t
	} else if kylixT, ok := g.localTypes[varName]; ok {
		// Class-typed local (alloca %v_name, no suffix) holds a ptr.
		if _, isClass := g.classes[kylixT]; isClass {
			actualType = "ptr"
		}
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

// emitTupleBuild handles `result := (a, b, ...)` inside a multi-return
// function: evaluates each element and packs them into the `%__ret_FuncName`
// struct via a chain of insertvalue instructions, then stores into %result.
func (g *Generator) emitTupleBuild(tuple *ast.TupleLiteral) error {
	structType := fmt.Sprintf("%%__ret_%s", g.funcName)
	elemTypes := g.multiRetTypes[g.funcName]

	// Build struct value via insertvalue chain: start with undef, insert each field.
	accReg := "undef"
	for i, elem := range tuple.Elements {
		v, _, err := g.emitExpr(elem)
		if err != nil {
			return err
		}
		elemT := "i64"
		if i < len(elemTypes) {
			elemT = elemTypes[i]
		}
		next := g.tmp()
		g.line(fmt.Sprintf("  %s = insertvalue %s %s, %s %s, %d", next, structType, accReg, elemT, v, i))
		accReg = next
	}
	g.line(fmt.Sprintf("  store %s %s, ptr %%result", structType, accReg))
	return nil
}

// emitTupleDestructure handles `(a, b, ...) := Expr` where Expr is a call to
// a multi-return function. Evaluates the call (which yields a struct value),
// extracts each field via extractvalue, and stores into the corresponding
// LHS variables (auto-declaring them if not already local).
func (g *Generator) emitTupleDestructure(tuple *ast.TupleLiteral, rhs ast.Expression) error {
	structVal, structType, err := g.emitExpr(rhs)
	if err != nil {
		return err
	}

	call, isCall := rhs.(*ast.CallExpression)
	var elemTypes []string
	if isCall {
		if fnIdent, ok := call.Function.(*ast.Identifier); ok {
			elemTypes = g.multiRetTypes[fnIdent.Value]
		}
	}

	for i, elem := range tuple.Elements {
		ident, ok := elem.(*ast.Identifier)
		if !ok {
			continue // non-identifier tuple element (e.g. `_`) — skip binding
		}
		elemT := "i64"
		if i < len(elemTypes) {
			elemT = elemTypes[i]
		}
		extracted := g.tmp()
		g.line(fmt.Sprintf("  %s = extractvalue %s %s, %d", extracted, structType, structVal, i))

		allocaReg, exists := g.locals[ident.Value]
		if !exists {
			suffix := "_int"
			switch elemT {
			case "i1":
				suffix = "_bool"
			case "double":
				suffix = "_real"
			case "ptr":
				suffix = "_str"
			}
			allocaReg = g.freshVarReg(ident.Value, suffix)
			g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, elemT))
			g.locals[ident.Value] = allocaReg
		}
		g.line(fmt.Sprintf("  store %s %s, ptr %s", elemT, extracted, allocaReg))
	}
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
	elemAlloca := g.freshVarReg(s.Variable, "")
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
			if err := g.emitBlockScoped(br.Body); err != nil {
				return err
			}
		}
		g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	}

	// Default / else branch.
	g.line(fmt.Sprintf("%s:", defaultLbl))
	if s.ElseBranch != nil {
		if err := g.emitBlockScoped(s.ElseBranch); err != nil {
			return err
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
				if err := g.emitBlockScoped(br.Body); err != nil {
					return err
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
			if err := g.emitBlockScoped(br.Body); err != nil {
				return err
			}
		}
		g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
		g.line(fmt.Sprintf("%s:", nextLbl))
	}

	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	g.line(fmt.Sprintf("%s:", mergeLbl))
	return nil
}
