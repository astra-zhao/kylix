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
	// VarDecl has Names []string, handle first name only for now
	if len(s.Names) == 0 {
		return nil
	}
	name := s.Names[0]

	// Array type: dispatch to dedicated handler (Milestone 2).
	if arrT, ok := s.Type.(*ast.ArrayType); ok {
		g.emitArrayVarDecl(name, arrT)
		return nil
	}

	// Interface-typed local: reserve { vtable, data } pair allocas.
	if s.Type != nil {
		if tname := typeExprName(s.Type); tname != "" {
			if _, isIface := g.interfaces[tname]; isIface {
				g.emitInterfaceVarDecl(name)
				g.localTypes[name] = tname
				return nil
			}
		}
	}

	llvmT := "i64"
	suffix := "_int"
	kylixType := ""
	if s.Type != nil {
		tname := typeExprName(s.Type)
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
		return fmt.Errorf("complex lvalue not supported yet")
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

	g.line(fmt.Sprintf("%s:", bodyLbl))
	if err := g.emitStatement(s.Body); err != nil {
		return err
	}
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

	// Body
	g.line(fmt.Sprintf("%s:", bodyLbl))
	if err := g.emitStatement(s.Body); err != nil {
		return err
	}

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

	g.line(fmt.Sprintf("  br label %%%s", bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))

	if err := g.emitStatement(s.Body); err != nil {
		return err
	}

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
