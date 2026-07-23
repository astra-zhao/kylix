// stmt.go — LLVM IR code generation for Kylix statements.
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// emitStatement generates code for a single statement.
func (g *Generator) emitStatement(node ast.Statement) error {
	// v4.6.0: record the source position of this statement so every IR
	// instruction emitted while processing it carries a !dbg DILocation.
	// Cleared on return so synthetic trailing instructions (ret) don't claim
	// a stale source line. save/restore lets nested dispatch (e.g. Block →
	// If → Assign) each set their own position without clobbering the parent's
	// on the way out.
	savedLine, savedCol := 0, 0
	if g.debugInfo {
		savedLine, savedCol = g.dbg.curLine, g.dbg.curCol
		g.setDbgNode(node)
	}
	defer func() {
		if g.debugInfo {
			g.dbg.curLine, g.dbg.curCol = savedLine, savedCol
		}
	}()
	switch s := node.(type) {
	case *ast.AssignmentStatement:
		return g.emitAssign(s)
	case *ast.ExpressionStatement:
		// v5.4.0: statement-style `append(slice, elem)` is a mutating call —
		// the new slice must be stored back to the original variable/field.
		// Without this, `append(Files, x)` discards the result and Files stays
		// empty → GenerateMulti gets no input → empty output.
		if call, ok := s.Expression.(*ast.CallExpression); ok && len(call.Arguments) == 2 {
			if ident, ok := call.Function.(*ast.Identifier); ok && ident.Value == "append" {
				result, _, err := g.emitAppend(call.Arguments[0], call.Arguments[1])
				if err != nil {
					return err
				}
				// Store result back to the first argument.
				if member, ok := call.Arguments[0].(*ast.MemberExpression); ok {
					kind, typeName := g.receiverKind(member.Object)
					if kind == "class" {
						objReg, _, err := g.loadObjectPtr(member.Object, typeName)
						if err == nil {
							gepReg, _, err := g.emitFieldStore(typeName, objReg, member.Member)
							if err == nil {
								g.line(fmt.Sprintf("  store { ptr, i64, i64 } %s, ptr %s", result, gepReg))
							}
						}
						return nil
					}
				}
				if leftIdent, ok := call.Arguments[0].(*ast.Identifier); ok {
					if allocaReg, ok := g.locals[leftIdent.Value]; ok {
						g.line(fmt.Sprintf("  store { ptr, i64, i64 } %s, ptr %s", result, allocaReg))
					}
				}
				return nil
			}
		}
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

// stmtDbgRestore is a no-op helper kept to document the save/restore pattern
// in emitStatement's defer (kept inline there for clarity). Retained as a
// placeholder so future statement-level debug instrumentation has a home.
func (g *Generator) stmtDbgRestore(savedLine, savedCol int) {
	if g.debugInfo {
		g.dbg.curLine, g.dbg.curCol = savedLine, savedCol
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

	// v4.9.0: open a DILexicalBlock for this nested scope so locals declared
	// inside (and the instructions emitted here) are scoped to the block,
	// not the whole function. Only when we're actually inside a function
	// (curScope != 0) and debug info is on; otherwise this is a no-op.
	savedDbgScope := 0
	if g.debugInfo && g.dbg != nil && g.dbg.curScope != 0 {
		savedDbgScope = g.registerLexicalBlock()
	}

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
	// Restore the enclosing scope (subprogram or outer block) on block exit.
	if g.debugInfo && g.dbg != nil && savedDbgScope != 0 {
		g.setDbgScope(savedDbgScope)
	}
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
	decl.Parameters = normalizeParams(decl.Parameters) // v5.4.0: `level, msg: String`

	// v5.4.0: external method definition `procedure ClassName.MethodName` —
	// the name contains a dot. Lower to @ClassName_MethodName with a leading
	// `ptr %self` parameter (mirroring emitMethod), and register self's class
	// so `self.Field` / `self.Method()` resolve. Without this, external methods
	// (most of the bootstrap's TLexer/TParser/TGenerator methods) had no self.
	extClassName, extMethodName := "", ""
	funcSymbol := decl.Name
	if idx := strings.Index(decl.Name, "."); idx >= 0 {
		extClassName = decl.Name[:idx]
		extMethodName = decl.Name[idx+1:]
		funcSymbol = extClassName + "_" + extMethodName
	}

	// Determine return type: check multi-return first, then single, else void.
	retType := "void"
	isMultiRet := false
	if multiTypes := g.multiRetTypes[decl.Name]; len(multiTypes) > 0 {
		retType = fmt.Sprintf("%%__ret_%s", decl.Name)
		isMultiRet = true
	} else if decl.ReturnType != nil {
		retType = g.llvmTypeOfExpr(decl.ReturnType)
	}

	// Build parameter list
	var params []string
	if extClassName != "" {
		params = append(params, "ptr %self") // v5.4.0: external method receiver
	}
	for _, p := range decl.Parameters {
		llvmT := "i64"
		if p.Type != nil {
			llvmT = g.llvmTypeOfExpr(p.Type)
		}
		params = append(params, fmt.Sprintf("%s %%%s", llvmT, p.Name))
	}

	defineLine := fmt.Sprintf("define %s @%s(%s) {", retType, funcSymbol, strings.Join(params, ", "))
	var funcSpID int
	if g.debugInfo {
		funcSpID = g.registerSubprogram(decl.Name, decl.Token.Line)
		defineLine = g.defineLineWithDbg(defineLine, funcSpID)
	}
	g.line(defineLine)
	g.line("entry:")
	g.funcName = decl.Name
	savedLocals := g.locals
	savedTypes := g.localTypes
	savedVarSeq := g.varNameSeq
	g.locals = make(map[string]string)
	g.localTypes = make(map[string]string)
	g.varNameSeq = make(map[string]int)
	g.registerGlobalsInScope() // v5.4.0: make globals visible in this function
	// v5.4.0: external method receiver — register self as the class instance.
	if extClassName != "" {
		g.locals["self"] = "%self"
		g.localTypes["self"] = extClassName
	}
	// v4.6.0: scope for DILocations inside this function = its subprogram.
	// Position the entry-block setup at the function's declaration line so the
	// %result alloca + parameter stores carry a valid !dbg before the body
	// statements set their own positions.
	if g.debugInfo {
		g.setDbgScope(funcSpID)
		g.setDbgNode(decl)
	}

	// Allocate result variable for functions
	if retType != "void" {
		g.line(fmt.Sprintf("  %%result = alloca %s, align 8", retType))
		g.locals["result"] = "%result"
		g.resultLLVMType = retType // v5.4.0: so emitIdentLoad loads `result` as the right type
		if isMultiRet {
			// Mark result as a tuple so assignment can detect it.
			g.localTypes["result"] = "__tuple__"
		} else if decl.ReturnType != nil {
			// v5.5.0: set localTypes["result"] to the Kylix return type name so
			// receiverKind(result) resolves it for `result.Field := x` on
			// class/record return types (e.g. TLexer.ReadNumber returns TToken,
			// and `result.TokenType := tokType` needs to GEP into TToken).
			retKylix := typeExprName(decl.ReturnType)
			if _, isClass := g.classes[retKylix]; isClass {
				g.localTypes["result"] = retKylix
				// v5.5.0: for record/class return types, malloc the struct and
				// store its pointer into %result so `result.Field := x` GEPs
				// into valid memory (not null). Records have no Create method.
				if g.records[retKylix] {
					size := int64(8) // vtable ptr
					for _, f := range g.classes[retKylix].Fields {
						size += llvmTypeSize(f.LLVMType)
					}
					recReg := g.tmp()
					g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", recReg, size))
					// Store vtable for is/as (records have [0 x ptr] vtable).
					g.line(fmt.Sprintf("  store ptr @%s_vtable, ptr %s", retKylix, recReg))
					g.line(fmt.Sprintf("  store ptr %s, ptr %%result", recReg))
				}
			}
		}
		// v4.6.0: declare `result` as a debug local so LLDB can show its
		// value while stepping through the function body (it's the implicit
		// return slot — a real alloca, so dbg.declare applies directly).
		if g.debugInfo {
			g.emitDbgDeclare("result", decl.Token.Line, retType, "%result")
		}
	}

	// Allocate parameters as locals
	for _, p := range decl.Parameters {
		llvmT := "i64"
		kylixType := ""
		isSlice := false
		var elemT, elemKylixT string
		if p.Type != nil {
			kylixType = typeExprName(p.Type)
			llvmT = g.llvmTypeOfExpr(p.Type)
			if at, ok := p.Type.(*ast.ArrayType); ok && at.Dynamic {
				isSlice = true
				elemKylixT = typeExprName(at.ElementType)
				elemT = g.llvmTypeOfExpr(at.ElementType)
			}
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
		case "{ ptr, i64, i64 }":
			suffix = "_dyn"
		}
		allocaReg := fmt.Sprintf("%%v_%s%s", p.Name, suffix)
		g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, llvmT))
		g.line(fmt.Sprintf("  store %s %%%s, ptr %s", llvmT, p.Name, allocaReg))
		g.locals[p.Name] = allocaReg
		if kylixType != "" {
			g.localTypes[p.Name] = kylixType
		}
		// v5.4.0: register slice params in arrayInfo so Length(p)/p[i] work
		// (previously function array-of-T params weren't tracked, causing
		// "variable progs is not an array" when indexing them).
		if isSlice {
			g.arrayInfo[p.Name] = &arrayInfo{IsDynamic: true, ElementType: elemT, ElementKylixType: elemKylixT}
		}
		// v4.6.0: declare the parameter as a debug local so LLDB can show it.
		if g.debugInfo {
			declLine := decl.Token.Line
			if p.Token.Line > 0 {
				declLine = p.Token.Line
			}
			g.emitDbgDeclare(p.Name, declLine, llvmT, allocaReg)
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
	// Leaving this function: clear the debug scope + position so subsequent
	// module-level code (other functions, stdlib defines, metadata) doesn't
	// attach a stale !dbg.
	if g.debugInfo {
		g.setDbgScope(0)
		g.clearDbgPos()
	}
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

		// Constructor inference: `var x := TFoo.Create` or `var x := TFoo.Create()`
		// produces a ptr, but we must record the class name so later
		// `x.Field` / `x.Method` resolve. Covers both the bare-member form
		// (MemberExpression) and the call form (CallExpression wrapping a
		// MemberExpression), and the generic variants TStack<Integer>.Create.
		inferredClass := ""
		if name := constructorClassName(s.Value, g); name != "" {
			inferredClass = name
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
		if g.records[ident.Value] {
			// v5.4.0: record local — heap-allocate the struct now (records have
			// no Create method, so `var tok: TToken` must malloc the storage
			// up front so `tok.Field := x` GEPs into valid memory).
			// v5.5.0: use llvmTypeSize for correct sizing (slice/map fields are
			// >8 bytes; the old 8×count formula under-allocated).
			allocaReg := g.freshVarReg(name, "")
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
			size := int64(8) // vtable ptr
			for _, f := range g.classes[ident.Value].Fields {
				size += llvmTypeSize(f.LLVMType)
			}
			rec := g.tmp()
			g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", rec, size))
			g.line(fmt.Sprintf("  store ptr %s, ptr %s", rec, allocaReg))
			g.locals[name] = allocaReg
			g.localTypes[name] = ident.Value
			return nil
		}
		if _, isClass := g.classes[ident.Value]; isClass {
			allocaReg := g.freshVarReg(name, "")
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
			g.line(fmt.Sprintf("  store ptr null, ptr %s", allocaReg))
			g.locals[name] = allocaReg
			g.localTypes[name] = ident.Value
			return nil
		}
	}

	// v5.0.0: Variant-typed local. `var v: Variant` parses as Identifier
	// (capital V is not the discriminated-union keyword `variant`). The slot
	// holds a pointer to a boxed {i32 tag, i64 payload} value. Must come
	// BEFORE the generic fallback below (which would allocate a ptr slot with
	// the wrong _int suffix and crash on load).
	if isVariantTypeExpr(varType) {
		allocaReg := g.freshVarReg(name, "_var")
		g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
		g.line(fmt.Sprintf("  store ptr null, ptr %s", allocaReg))
		g.locals[name] = allocaReg
		g.localTypes[name] = "Variant"
		g.needVariantRuntime = true
		if g.debugInfo {
			g.emitDbgDeclare(name, varDeclLine(varType), "ptr", allocaReg)
		}
		return nil
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
	// v4.6.0: declare the local as a debug variable (DILocalVariable +
	// llvm.dbg.declare) so LLDB can resolve its name + value at breakpoints.
	if g.debugInfo {
		g.emitDbgDeclare(name, varDeclLine(varType), llvmT, allocaReg)
	}
	return nil
}

// emitDbgDeclare records a DILocalVariable for `name` (scoped to the current
// function) and emits a `#dbg_declare` intrinsic record next to the alloca,
// so the debugger can map the alloca to a source variable. No-op when debug
// info is off.
//
// LLVM 22 deprecated the `call void @llvm.dbg.declare(...)` intrinsic in favor
// of the `#dbg_declare` record syntax:
//
//	#dbg_declare(ptr <alloca>, !<DILocalVariable>, !DIExpression(), !<DILocation>)
//
// The record takes 4 operands: the storage address (an SSA value), the
// DILocalVariable metadata, a DIExpression (empty = direct addressing — "the
// value at this address IS the variable"), and a DILocation (the source
// position; we use the current position so the record is associated with the
// declaration line). The `declare void @llvm.dbg.declare` is NOT emitted —
// the record is standalone.
//
// DILocalVariable ID is allocated here (during codegen); the empty-expression
// is inlined as `!DIExpression()` (no separate node needed in LLVM 22 record
// syntax).
func (g *Generator) emitDbgDeclare(name string, line int, llvmType, allocaReg string) {
	if g.dbg == nil {
		return
	}
	if line == 0 {
		line = g.dbg.curLine
	}
	vID := g.registerLocalVariable(name, line, llvmType, allocaReg)
	locID := g.curDbgLocID()
	// If no current position (shouldn't happen inside a function), fall back
	// to a position synthesized from the variable's declaration line.
	if locID == 0 {
		g.dbg.curLine = line
		g.dbg.curCol = 1
		locID = g.curDbgLocID()
	}
	g.line(fmt.Sprintf("  #dbg_declare(ptr %s, %s, !DIExpression(), %s)",
		allocaReg, dbgRef(vID), dbgRef(locID)))
}

// varDeclLine best-effort extracts a source line from a var type AST node for
// the debug declaration. Returns 0 if unavailable (emitDbgDeclare then falls
// back to the current position).
func varDeclLine(varType ast.Expression) int {
	if varType == nil {
		return 0
	}
	// Identifier-typed vars carry their Token; others fall back to 0.
	if ident, ok := varType.(*ast.Identifier); ok {
		return ident.Token.Line
	}
	return 0
}

// constructorClassName inspects a var-decl initializer expression and, if it
// is a constructor call (TFoo.Create, TFoo.Create(args), or the generic
// TStack<Integer>.Create / TStack<Integer>.Create(args)), returns the class
// name the variable holds (the specialized mangled name for generics). This
// lets `var x := TFoo.Create()` record the receiver type so later
// `x.Method()` dispatches correctly. Returns "" if the initializer is not a
// recognized constructor form or the target class isn't registered.
//
// Handles both forms:
//   - *ast.MemberExpression{Member:"Create"}         (bare `TFoo.Create`)
//   - *ast.CallExpression{Function: MemberExpression} (`TFoo.Create(...)`)
//
// and both receiver kinds:
//   - *ast.Identifier (TFoo)                         → ident.Value
//   - *ast.GenericType (TStack<Integer>)            → mangleGeneric(...)
func constructorClassName(value ast.Expression, g *Generator) string {
	var member *ast.MemberExpression
	switch v := value.(type) {
	case *ast.MemberExpression:
		member = v
	case *ast.CallExpression:
		m, ok := v.Function.(*ast.MemberExpression)
		if !ok {
			return ""
		}
		member = m
	default:
		return ""
	}
	if member.Member != "Create" {
		return ""
	}
	if ident, ok := member.Object.(*ast.Identifier); ok {
		if _, known := g.classes[ident.Value]; known {
			return ident.Value
		}
	}
	if gt, ok := member.Object.(*ast.GenericType); ok {
		mangled := mangleGeneric(gt.Base, gt.TypeParams)
		if mangled != "" {
			if _, known := g.classes[mangled]; known {
				return mangled
			}
		}
	}
	return ""
}

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
		// v5.0.0: for `array of Variant`, box the RHS into a Variant before
		// storing (the element slot holds a box pointer). A Variant RHS (e.g.
		// arr[0] := arr[1]) is passed through; a scalar RHS is boxed by type.
		if leftIdent, ok := idx.Left.(*ast.Identifier); ok {
			if info, hasInfo := g.arrayInfo[leftIdent.Value]; hasInfo && info.IsVariant && t != variantT {
				v = g.emitVariantBox(v, t)
				elemType = "ptr"
			}
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
			// v5.4.0: a class/record-typed field's RHS may carry the Kylix type
			// name (emitMember returns it for receiver resolution). Coerce to
			// ptr (the actual LLVM type) so the store is well-typed.
			if _, isClass := g.classes[fieldType]; isClass {
				fieldType = "ptr"
			}
			if _, isClass := g.classes[t]; isClass {
				t = "ptr"
			}
			// v5.6.0: record field assignment (e.g. self.CurToken := self.PeekToken)
			// must deep-copy the record struct (value semantics, matching Go's
			// `self.CurToken = self.PeekToken` which copies the entire TToken struct).
			// Without this, CurToken and PeekToken share the same TToken object —
			// subsequent PeekToken updates corrupt CurToken → lexer state corruption.
			if fieldKylixT := fieldKylixType(typeName, member.Member, g); fieldKylixT != "" && g.records[fieldKylixT] {
				// Branch: if RHS is null, store null; else malloc+memcpy+store.
				nullCk := g.tmp()
				g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", nullCk, v))
				storeNullLbl := g.label()
				copyLbl := g.label()
				doneLbl := g.label()
				g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nullCk, storeNullLbl, copyLbl))
				// Null branch: store null.
				g.line(fmt.Sprintf("%s:", storeNullLbl))
				g.line(fmt.Sprintf("  store ptr null, ptr %s", gepReg))
				g.line(fmt.Sprintf("  br label %%%s", doneLbl))
				// Copy branch: malloc + memcpy + store.
				g.line(fmt.Sprintf("%s:", copyLbl))
				recSize := int64(8) // vtable ptr
				for _, f := range g.classes[fieldKylixT].Fields {
					recSize += llvmTypeSize(f.LLVMType)
				}
				newRec := g.tmp()
				g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", newRec, recSize))
				g.needMemcpy = true
				g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %d)", newRec, v, recSize))
				g.line(fmt.Sprintf("  store ptr %s, ptr %s", newRec, gepReg))
				g.line(fmt.Sprintf("  br label %%%s", doneLbl))
				g.line(fmt.Sprintf("%s:", doneLbl))
				return nil
			}
			// v5.4.0: slice field — RHS is an SSA struct value (from call or
			// ArrayLiteral insertvalue), store directly.
			if fieldType == "{ ptr, i64, i64 }" && t == "{ ptr, i64, i64 }" {
				g.line(fmt.Sprintf("  store { ptr, i64, i64 } %s, ptr %s", v, gepReg))
				return nil
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
		// v5.4.0: type-inferred local — choose the alloca type from the RHS's
		// LLVM type (t), so a ptr/string RHS gets a _str slot (not the default
		// i64, which type-mismatches on store). Also try exprKylixType for
		// class-typed RHS so receiverKind resolves for later field/method access.
		elemKylix := g.exprKylixType(s.Value)
		if _, isClass := g.classes[elemKylix]; isClass {
			allocaReg = g.freshVarReg(varName, "_str")
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
			g.locals[varName] = allocaReg
			g.localTypes[varName] = elemKylix
			t = "ptr"
		} else if t == "{ ptr, i64, i64 }" {
			allocaReg = g.freshVarReg(varName, "_dyn")
			g.line(fmt.Sprintf("  %s = alloca { ptr, i64, i64 }, align 8", allocaReg))
			g.line(fmt.Sprintf("  store { ptr, i64, i64 } zeroinitializer, ptr %s", allocaReg))
			g.locals[varName] = allocaReg
			g.arrayInfo[varName] = &arrayInfo{IsDynamic: true, ElementType: "ptr"}
		} else if t == "ptr" {
			// v5.4.0: string/ptr RHS (e.g. map lookup, LowerCase, ReadFile) →
			// _str slot so the value isn't truncated to i64.
			allocaReg = g.freshVarReg(varName, "_str")
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", allocaReg))
			g.locals[varName] = allocaReg
		} else if t == "double" {
			allocaReg = g.freshVarReg(varName, "_real")
			g.line(fmt.Sprintf("  %s = alloca double, align 8", allocaReg))
			g.locals[varName] = allocaReg
		} else if t == "i1" {
			allocaReg = g.freshVarReg(varName, "_bool")
			g.line(fmt.Sprintf("  %s = alloca i1, align 1", allocaReg))
			g.locals[varName] = allocaReg
		} else {
			// Auto-declare as i64
			allocaReg = g.freshVarReg(varName, "_int")
			g.line(fmt.Sprintf("  %s = alloca i64, align 8", allocaReg))
			g.locals[varName] = allocaReg
			t = "i64"
		}
	}

	// v4.9.0: dynamic-array assignment. `arr := JsonGetArray(...)` (or any RHS
	// yielding a {ptr, i64, i64} slice) must copy the whole slice struct into
	// arr's alloca — the generic scalar path below would store only the first
	// word and leave len/cap stale. If the LHS is a _dyn alloca and the RHS is
	// a slice struct value (carried as a pointer to a temporary alloca), load
	// the struct from the RHS alloca and store it into the LHS alloca.
	if strings.HasSuffix(allocaReg, "_dyn") && t == "{ ptr, i64, i64 }" {
		// v5.4.0: slice RHS is an SSA struct value (from call or ArrayLiteral
		// insertvalue) — store directly, no load needed.
		g.line(fmt.Sprintf("  store { ptr, i64, i64 } %s, ptr %s", v, allocaReg))
		return nil
	}

	// Infer actual type from alloca name
	actualType := "i64"
	isVariantSlot := strings.HasSuffix(allocaReg, "_var")
	if strings.HasSuffix(allocaReg, "_bool") {
		actualType = "i1"
	} else if strings.HasSuffix(allocaReg, "_real") {
		actualType = "double"
	} else if strings.HasSuffix(allocaReg, "_str") {
		actualType = "ptr"
	} else if strings.HasSuffix(allocaReg, "_map") {
		actualType = "ptr"
	} else if isVariantSlot {
		// v5.0.0: a Variant slot stores a box pointer (ptr). The RHS is
		// boxed below before the store, so treat the slot as a ptr sink.
		actualType = "ptr"
	} else if allocaReg == "%result" && t != "" {
		actualType = t
	} else if gt, ok := g.globalTypes[varName]; ok {
		// v5.4.0: global variable — use its declared LLVM type for the store.
		actualType = gt
	} else if kylixT, ok := g.localTypes[varName]; ok {
		// Class-typed local (alloca %v_name, no suffix) holds a ptr.
		if _, isClass := g.classes[kylixT]; isClass {
			actualType = "ptr"
		}
	}

	// v5.0.0: Variant assignment. Box the RHS into a Variant when storing
	// into a Variant slot. A Variant RHS (v := otherVariant) is passed
	// through unchanged; a scalar RHS is boxed by its evaluated type.
	if isVariantSlot && t != variantT {
		v = g.emitVariantBox(v, t)
		t = "ptr" // the value is now a box pointer; skip coercion below.
	}

	// Type coercion: if RHS type doesn't match the alloca type, cast it.
	// v5.1.0: a Variant RHS unboxes to the concrete slot type via coerceValue
	// (variant→i64/double/ptr/i1 dispatch on tag); this also covers the
	// legacy i1↔i64 / i64↔double casts.
	if t != actualType {
		v, t = g.coerceValue(v, t, actualType)
	}

	g.line(fmt.Sprintf("  store %s %s, ptr %s", actualType, v, allocaReg))
	return nil
}

// fieldKylixType returns the Kylix type name of a class field, or "" if not
// found. v5.6.0.
func fieldKylixType(className, fieldName string, g *Generator) string {
	info, ok := g.classes[className]
	if !ok {
		return ""
	}
	for _, f := range info.Fields {
		if f.Name == fieldName {
			return f.KylixType
		}
	}
	return ""
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
