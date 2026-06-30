// expr.go — LLVM IR code generation for Kylix expressions.
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// LLVMType maps a Kylix type name to its LLVM IR type.
func LLVMType(typeName string) string {
	switch strings.ToLower(typeName) {
	case "integer", "int64", "int":
		return "i64"
	case "boolean", "bool":
		return "i1"
	case "real", "double", "float":
		return "double"
	case "string":
		return "ptr" // pointer to i8 (null-terminated)
	case "char":
		return "i8"
	default:
		return "i64" // fallback
	}
}

// typeExprName extracts the string name from an AST type expression.
// ast.Identifier.Value is the authoritative type name; TokenLiteral() may return
// the wrong token (e.g. ";") depending on parser position.
func typeExprName(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Value
	}
	// Fallback for other expression types
	return expr.TokenLiteral()
}

// emitExpr generates code for an expression, returning the SSA register holding the result.
// Returns ("", type, error).
func (g *Generator) emitExpr(node ast.Expression) (reg string, llvmType string, err error) {
	switch e := node.(type) {
	case *ast.IntegerLiteral:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, %d", r, e.Value))
		return r, "i64", nil

	case *ast.FloatLiteral:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = fadd double 0.0, %f", r, e.Value))
		return r, "double", nil

	case *ast.BooleanLiteral:
		r := g.tmp()
		val := 0
		if e.Value {
			val = 1
		}
		g.line(fmt.Sprintf("  %s = add i1 0, %d", r, val))
		return r, "i1", nil

	case *ast.StringLiteral:
		strReg := g.addString(e.Value)
		size := len(e.Value) + 1
		ptr := g.ptrTo(strReg, size)
		return ptr, "ptr", nil

	case *ast.NilLiteral:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr", r))
		return r, "ptr", nil

	case *ast.Identifier:
		return g.emitIdentLoad(e.Value)

	case *ast.InfixExpression:
		return g.emitInfix(e)

	case *ast.PrefixExpression:
		return g.emitPrefix(e)

	case *ast.CallExpression:
		return g.emitCall(e)

	case *ast.IndexExpression:
		return g.emitArrayIndex(e, false)

	case *ast.MemberExpression:
		return g.emitMember(e)

	case *ast.IsExpression:
		return g.emitIsExpr(e)

	case *ast.TypeCastExpression:
		return g.emitAsExpr(e)

	case *ast.StringInterpolation:
		return g.emitStringInterpolation(e)

	case *ast.LambdaExpression:
		// Closures require an environment struct + boxing — out of scope for
		// LLVM M3. Emit a null pointer stub so surrounding code still compiles.
		// The Go backend fully supports lambdas; LLVM backend support is tracked
		// in ROADMAP under LLVM M3 (closures).
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; lambda/closure unsupported in LLVM backend", r))
		return r, "ptr", nil

	case *ast.ArrayLiteral:
		// Array literals — allocate a heap buffer and store each element.
		// Conservative: uses i64 element type for all literals.
		n := len(e.Elements)
		size := int64(8 * (n + 1)) // +1 for length word
		buf := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", buf, size))
		// store length at index 0
		lenPtr := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 0", lenPtr, buf))
		g.line(fmt.Sprintf("  store i64 %d, ptr %s", int64(n), lenPtr))
		for i, elem := range e.Elements {
			v, _, err := g.emitExpr(elem)
			if err != nil {
				return "", "", err
			}
			ep := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds i64, ptr %s, i64 %d", ep, buf, int64(i+1)))
			g.line(fmt.Sprintf("  store i64 %s, ptr %s", v, ep))
		}
		return buf, "ptr", nil

	case *ast.SliceExpression:
		// Slice expressions (arr[lo..hi]) — emit the base pointer for now;
		// proper slice struct construction is deferred.
		base, _, err := g.emitExpr(e.Left)
		if err != nil {
			return "", "", err
		}
		return base, "ptr", nil

	case *ast.TupleLiteral:
		// Tuple literals for multi-return — return the first element as a
		// conservative fallback (multi-return lowering is complex in LLVM SSA).
		if len(e.Elements) == 0 {
			r := g.tmp()
			g.line(fmt.Sprintf("  %s = add i64 0, 0 ; empty tuple", r))
			return r, "i64", nil
		}
		return g.emitExpr(e.Elements[0])

	case *ast.AwaitExpression:
		// Async/await is not supported in the LLVM backend. Emit the inner
		// expression synchronously so surrounding code still type-checks.
		return g.emitExpr(e.Expression)

	default:
		// Unknown expression — emit zero
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; unhandled expr %T", r, node))
		return r, "i64", nil
	}
}

func (g *Generator) emitIdentLoad(name string) (string, string, error) {
	allocaReg, ok := g.locals[name]
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; undefined var %s", r, name))
		return r, "i64", nil
	}
	// Determine type from alloca name convention: %v_name_TYPE
	llvmT := "i64" // default
	if strings.HasSuffix(allocaReg, "_bool") {
		llvmT = "i1"
	} else if strings.HasSuffix(allocaReg, "_real") {
		llvmT = "double"
	} else if strings.HasSuffix(allocaReg, "_str") {
		llvmT = "ptr"
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load %s, ptr %s", r, llvmT, allocaReg))
	return r, llvmT, nil
}

func (g *Generator) emitInfix(e *ast.InfixExpression) (string, string, error) {
	lv, lt, err := g.emitExpr(e.Left)
	if err != nil {
		return "", "", err
	}
	rv, _, err := g.emitExpr(e.Right)
	if err != nil {
		return "", "", err
	}

	r := g.tmp()

	isFloat := lt == "double"
	switch e.Operator {
	case "+":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fadd double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = add i64 %s, %s", r, lv, rv))
		}
		return r, lt, nil
	case "-":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fsub double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = sub i64 %s, %s", r, lv, rv))
		}
		return r, lt, nil
	case "*":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fmul double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = mul i64 %s, %s", r, lv, rv))
		}
		return r, lt, nil
	case "/", "div":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fdiv double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = sdiv i64 %s, %s", r, lv, rv))
		}
		return r, lt, nil
	case "mod":
		g.line(fmt.Sprintf("  %s = srem i64 %s, %s", r, lv, rv))
		return r, "i64", nil
	case "=":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fcmp oeq double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", r, lv, rv))
		}
		return r, "i1", nil
	case "<>":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fcmp one double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = icmp ne i64 %s, %s", r, lv, rv))
		}
		return r, "i1", nil
	case "<":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fcmp olt double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", r, lv, rv))
		}
		return r, "i1", nil
	case "<=":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fcmp ole double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = icmp sle i64 %s, %s", r, lv, rv))
		}
		return r, "i1", nil
	case ">":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fcmp ogt double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = icmp sgt i64 %s, %s", r, lv, rv))
		}
		return r, "i1", nil
	case ">=":
		if isFloat {
			g.line(fmt.Sprintf("  %s = fcmp oge double %s, %s", r, lv, rv))
		} else {
			g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", r, lv, rv))
		}
		return r, "i1", nil
	case "and":
		g.line(fmt.Sprintf("  %s = and i1 %s, %s", r, lv, rv))
		return r, "i1", nil
	case "or":
		g.line(fmt.Sprintf("  %s = or i1 %s, %s", r, lv, rv))
		return r, "i1", nil
	default:
		g.line(fmt.Sprintf("  %s = add i64 %s, 0 ; unhandled op %s", r, lv, e.Operator))
		return r, lt, nil
	}
}

func (g *Generator) emitPrefix(e *ast.PrefixExpression) (string, string, error) {
	v, t, err := g.emitExpr(e.Right)
	if err != nil {
		return "", "", err
	}
	r := g.tmp()
	switch e.Operator {
	case "not":
		g.line(fmt.Sprintf("  %s = xor i1 %s, 1", r, v))
		return r, "i1", nil
	case "-":
		if t == "double" {
			g.line(fmt.Sprintf("  %s = fneg double %s", r, v))
		} else {
			g.line(fmt.Sprintf("  %s = sub i64 0, %s", r, v))
		}
		return r, t, nil
	default:
		return v, t, nil
	}
}

// emitCall generates a function call expression.
func (g *Generator) emitCall(e *ast.CallExpression) (string, string, error) {
	// obj.Method(args) — method dispatch on a class or interface receiver.
	if member, ok := e.Function.(*ast.MemberExpression); ok {
		return g.emitMethodCall(member, e.Arguments)
	}

	funcName := ""
	if ident, ok := e.Function.(*ast.Identifier); ok {
		funcName = ident.Value
	}

	// Built-in: WriteLn — 0, 1, or multiple arguments.
	if funcName == "WriteLn" {
		if len(e.Arguments) == 0 {
			// Empty WriteLn → puts("") → newline.
			emptyReg := g.addString("")
			emptyPtr := g.ptrTo(emptyReg, 1)
			r := g.tmp()
			g.line(fmt.Sprintf("  %s = call i32 @puts(ptr noundef %s)", r, emptyPtr))
			return "0", "void", nil
		}
		if len(e.Arguments) == 1 {
			return g.emitWriteLn(e.Arguments[0])
		}
		// Multiple args: build interpolation buffer and puts it.
		return g.emitWriteLnMulti(e.Arguments)
	}

	// Built-in: Write — 1 or multiple arguments (no newline).
	if funcName == "Write" {
		if len(e.Arguments) == 1 {
			return g.emitWrite(e.Arguments[0])
		}
		if len(e.Arguments) > 1 {
			for _, arg := range e.Arguments {
				if _, _, err := g.emitWrite(arg); err != nil {
					return "", "", err
				}
			}
			return "0", "void", nil
		}
	}

	// Built-in: IntToStr(n)
	if funcName == "IntToStr" && len(e.Arguments) == 1 {
		return g.emitIntToStr(e.Arguments[0])
	}

	// Built-in: Length(s)
	if funcName == "Length" && len(e.Arguments) == 1 {
		return g.emitLength(e.Arguments[0])
	}

	// Generic function call
	var argRegs []string
	var argTypes []string
	for _, arg := range e.Arguments {
		r, t, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		argRegs = append(argRegs, r)
		argTypes = append(argTypes, t)
	}

	result := g.tmp()
	var argList []string
	for i, r := range argRegs {
		argList = append(argList, argTypes[i]+" "+r)
	}
	g.line(fmt.Sprintf("  %s = call i64 @%s(%s)", result, funcName, strings.Join(argList, ", ")))
	return result, "i64", nil
}

// emitWriteLn generates a puts() or printf() call for WriteLn.
func (g *Generator) emitWriteLn(arg ast.Expression) (string, string, error) {
	v, t, err := g.emitExpr(arg)
	if err != nil {
		return "", "", err
	}

	switch t {
	case "ptr":
		// puts() prints the string + newline
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 @puts(ptr noundef %s)", r, v))
		return "0", "void", nil
	case "i64":
		// printf("%lld\n", n)
		fmtReg := g.addString("%lld\n")
		fmtPtr := g.ptrTo(fmtReg, len("%lld\n")+1)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 (ptr, ...) @printf(ptr noundef %s, i64 %s)", r, fmtPtr, v))
		return "0", "void", nil
	case "double":
		fmtReg := g.addString("%f\n")
		fmtPtr := g.ptrTo(fmtReg, len("%f\n")+1)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 (ptr, ...) @printf(ptr noundef %s, double %s)", r, fmtPtr, v))
		return "0", "void", nil
	case "i1":
		// Print "true" or "false"
		trueReg := g.addString("true\n")
		falseReg := g.addString("false\n")
		truePtr := g.ptrTo(trueReg, len("true\n")+1)
		falsePtr := g.ptrTo(falseReg, len("false\n")+1)
		selected := g.tmp()
		g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", selected, v, truePtr, falsePtr))
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 @puts(ptr noundef %s)", r, selected))
		return "0", "void", nil
	default:
		fmtReg := g.addString("%lld\n")
		fmtPtr := g.ptrTo(fmtReg, len("%lld\n")+1)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 (ptr, ...) @printf(ptr noundef %s, i64 %s)", r, fmtPtr, v))
		return "0", "void", nil
	}
}

// emitWrite generates a printf call without newline.
func (g *Generator) emitWrite(arg ast.Expression) (string, string, error) {
	v, t, err := g.emitExpr(arg)
	if err != nil {
		return "", "", err
	}
	switch t {
	case "ptr":
		fmtReg := g.addString("%s")
		fmtPtr := g.ptrTo(fmtReg, len("%s")+1)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 (ptr, ...) @printf(ptr noundef %s, ptr %s)", r, fmtPtr, v))
		return "0", "void", nil
	default:
		fmtReg := g.addString("%lld")
		fmtPtr := g.ptrTo(fmtReg, len("%lld")+1)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 (ptr, ...) @printf(ptr noundef %s, i64 %s)", r, fmtPtr, v))
		return "0", "void", nil
	}
}

// emitIntToStr converts i64 to ptr via snprintf.
func (g *Generator) emitIntToStr(arg ast.Expression) (string, string, error) {
	v, _, err := g.emitExpr(arg)
	if err != nil {
		return "", "", err
	}
	// Allocate 24 bytes on stack for the number string
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [24 x i8], align 1", buf))
	bufPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [24 x i8], ptr %s, i64 0, i64 0", bufPtr, buf))
	fmtReg := g.addString("%lld")
	fmtPtr := g.ptrTo(fmtReg, len("%lld")+1)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 (ptr, i64, ptr, ...) @snprintf(ptr noundef %s, i64 24, ptr noundef %s, i64 %s)",
		r, bufPtr, fmtPtr, v))
	return bufPtr, "ptr", nil
}

// emitLength returns the length of a string via strlen.
func (g *Generator) emitLength(arg ast.Expression) (string, string, error) {
	v, t, err := g.emitExpr(arg)
	if err != nil {
		return "", "", err
	}
	if t != "ptr" {
		// For non-strings, return 0
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; Length of non-string", r))
		return r, "i64", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr noundef %s)", r, v))
	return r, "i64", nil
}

// emitMember lowers obj.Field — field access on a class-typed receiver.
// Interface receivers don't currently expose fields directly.
func (g *Generator) emitMember(e *ast.MemberExpression) (string, string, error) {
	// Constructor pattern: TFoo.Create or TBox<Integer>.Create — return a
	// fresh heap-allocated instance of the (specialized) class.
	if e.Member == "Create" {
		if ident, ok := e.Object.(*ast.Identifier); ok {
			if _, known := g.classes[ident.Value]; known {
				reg, err := g.emitConstructor(ident.Value)
				return reg, "ptr", err
			}
		}
		if gt, ok := e.Object.(*ast.GenericType); ok {
			mangled := mangleGeneric(gt.Base, gt.TypeParams)
			if mangled != "" {
				if _, known := g.classes[mangled]; known {
					reg, err := g.emitConstructor(mangled)
					return reg, "ptr", err
				}
			}
		}
	}

	kind, typeName := g.receiverKind(e.Object)
	if kind != "class" {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; member access on non-class %s.%s", r, typeName, e.Member))
		return r, "i64", nil
	}
	objReg, objType, err := g.loadObjectPtr(e.Object, typeName)
	if err != nil {
		return "", "", err
	}
	_ = objType
	return g.emitFieldAccess(typeName, objReg, e.Member)
}

// loadObjectPtr loads the underlying object pointer for a receiver expression.
// For class-typed locals this is the loaded ptr from the alloca; for interface
// locals it's the data slot.
func (g *Generator) loadObjectPtr(obj ast.Expression, typeName string) (string, string, error) {
	ident, ok := obj.(*ast.Identifier)
	if !ok {
		return g.emitExpr(obj)
	}
	if _, isIface := g.interfaces[typeName]; isIface {
		_, dataAlloca := interfaceLocalNames(ident.Value)
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, dataAlloca))
		return r, "ptr", nil
	}
	alloca, ok := g.locals[ident.Value]
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; unknown receiver %s", r, ident.Value))
		return r, "ptr", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, alloca))
	return r, "ptr", nil
}

// emitMethodCall lowers obj.Method(args). Concrete class receivers dispatch
// directly to @Class_Method via emitVirtualCall (which already loads the class
// vtable). Interface receivers indirect through the per-class interface vtable
// stored in the fat pointer.
func (g *Generator) emitMethodCall(member *ast.MemberExpression, args []ast.Expression) (string, string, error) {
	// Constructor call with arguments: TFoo.Create('msg') or TBox<Integer>.Create(x).
	// The no-arg form is handled in emitMember; this branch covers the call form
	// (CallExpression wrapping a MemberExpression). Routes through emitConstructor
	// so a real object pointer is produced instead of an "unsupported receiver" stub.
	if member.Member == "Create" {
		if ident, ok := member.Object.(*ast.Identifier); ok {
			if _, known := g.classes[ident.Value]; known {
				reg, err := g.emitConstructor(ident.Value)
				if err != nil {
					return "", "", err
				}
				g.initConstructorArgs(ident.Value, reg, args)
				return reg, "ptr", nil
			}
		}
		if gt, ok := member.Object.(*ast.GenericType); ok {
			mangled := mangleGeneric(gt.Base, gt.TypeParams)
			if mangled != "" {
				if _, known := g.classes[mangled]; known {
					reg, err := g.emitConstructor(mangled)
					if err != nil {
						return "", "", err
					}
					g.initConstructorArgs(mangled, reg, args)
					return reg, "ptr", nil
				}
			}
		}
	}

	kind, typeName := g.receiverKind(member.Object)
	argRegs := make([]string, 0, len(args))
	argTypes := make([]string, 0, len(args))
	for _, a := range args {
		r, t, err := g.emitExpr(a)
		if err != nil {
			return "", "", err
		}
		argRegs = append(argRegs, r)
		argTypes = append(argTypes, t)
	}
	switch kind {
	case "class":
		objReg, _, err := g.loadObjectPtr(member.Object, typeName)
		if err != nil {
			return "", "", err
		}
		return g.emitVirtualCall(typeName, objReg, member.Member, argRegs, argTypes)
	case "interface":
		ident := member.Object.(*ast.Identifier)
		return g.emitInterfaceCall(ident.Value, typeName, member.Member, argRegs, argTypes)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; unsupported receiver for %s", r, member.Member))
		return r, "i64", nil
	}
}

// emitInterfaceCall loads (vtable, data) from the interface local and indirect-calls
// the method slot resolved by interface declaration order.
func (g *Generator) emitInterfaceCall(varName, ifaceName, methodName string, argRegs, argTypes []string) (string, string, error) {
	iface, ok := g.interfaces[ifaceName]
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; unknown interface %s", r, ifaceName))
		return r, "i64", nil
	}
	var slot *InterfaceMethodInfo
	for i := range iface.Methods {
		if iface.Methods[i].Name == methodName {
			slot = &iface.Methods[i]
			break
		}
	}
	if slot == nil {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; %s.%s not found", r, ifaceName, methodName))
		return r, "i64", nil
	}
	vtAlloca, dataAlloca := interfaceLocalNames(varName)
	vtablePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", vtablePtr, vtAlloca))
	dataPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", dataPtr, dataAlloca))

	slotPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x ptr], ptr %s, i32 0, i32 %d",
		slotPtr, len(iface.Methods), vtablePtr, slot.VtableIdx))
	fnPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", fnPtr, slotPtr))

	callArgs := []string{"ptr " + dataPtr}
	for i, r := range argRegs {
		callArgs = append(callArgs, argTypes[i]+" "+r)
	}
	paramSig := []string{"ptr"}
	paramSig = append(paramSig, slot.Params...)
	fnType := fmt.Sprintf("%s (%s)", slot.RetType, strings.Join(paramSig, ", "))
	if slot.RetType == "void" {
		g.line(fmt.Sprintf("  call void %s(%s)", fnPtr, strings.Join(callArgs, ", ")))
		return "0", "void", nil
	}
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = call %s %s(%s)", result, fnType, fnPtr, strings.Join(callArgs, ", ")))
	return result, slot.RetType, nil
}

// emitIsExpr lowers `obj is IFoo` to a compile-time i1: 1 if the object's
// concrete class implements the target interface, else 0. Dynamic checks on
// already-boxed interface values would require runtime type IDs (deferred).
func (g *Generator) emitIsExpr(e *ast.IsExpression) (string, string, error) {
	target := typeExprName(e.TargetType)
	kind, typeName := g.receiverKind(e.Expression)
	val := 0
	if kind == "class" && g.classImplementsInterface(typeName, target) {
		val = 1
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i1 0, %d ; %s is %s", r, val, typeName, target))
	return r, "i1", nil
}

// emitAsExpr lowers `obj as IFoo` to a fat-pointer construction. The result
// type is reported as "ptr" so callers can store it via emitInterfaceAssign
// when the destination is an interface-typed local. Failure → null fat pointer.
func (g *Generator) emitAsExpr(e *ast.TypeCastExpression) (string, string, error) {
	target := typeExprName(e.TargetType)
	kind, typeName := g.receiverKind(e.Expression)
	if kind == "class" && g.classImplementsInterface(typeName, target) {
		objReg, _, err := g.loadObjectPtr(e.Expression, typeName)
		if err != nil {
			return "", "", err
		}
		vt, data := g.emitBoxInterface(typeName, target, objReg)
		// Bundle: return the data ptr as the canonical "value", leaving the
		// vtable accessible through the class+iface pair. Real boxed storage
		// happens in emitAssign when the LHS is an interface-typed local.
		_ = vt
		return data, "ptr", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; %s as %s — incompatible", r, typeName, target))
	return r, "ptr", nil
}


// emitStringInterpolation lowers a `'...${expr}...'` interpolation to a heap
// buffer built by strcat (for string parts) and snprintf (for integers). The
// result is a ptr to a null-terminated string.
//
// Conservative: only String and Integer expression parts are formatted; other
// types are skipped (the substring is omitted). Buffer is a fixed 256 bytes.
func (g *Generator) emitStringInterpolation(e *ast.StringInterpolation) (string, string, error) {
	const bufSize = 256
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", buf, bufSize))
	// Initialize to empty string (NUL terminator).
	g.line(fmt.Sprintf("  store i8 0, ptr %s", buf))

	// "%ld\0" format constant for integer formatting.
	ldFmt := g.addString("%ld")
	ldFmtPtr := g.ptrTo(ldFmt, 4)

	for _, part := range e.Parts {
		switch p := part.(type) {
		case *ast.StringLiteral:
			strPtr := g.ptrTo(g.addString(p.Value), len(p.Value)+1)
			g.line(fmt.Sprintf("  %s = call ptr @strcat(ptr %s, ptr %s)", g.tmp(), buf, strPtr))
		default:
			reg, t, err := g.emitExpr(part)
			if err != nil {
				return "", "", err
			}
			switch t {
			case "ptr":
				g.line(fmt.Sprintf("  %s = call ptr @strcat(ptr %s, ptr %s)", g.tmp(), buf, reg))
			case "i64":
				pos := g.tmp()
				g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", pos, buf))
				dst := g.tmp()
				g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
				rest := g.tmp()
				g.line(fmt.Sprintf("  %s = sub i64 %d, %s", rest, bufSize, pos))
				g.line(fmt.Sprintf("  %s = call i32 @snprintf(ptr %s, i64 %s, ptr %s, i64 %s)",
					g.tmp(), dst, rest, ldFmtPtr, reg))
			case "double":
				fFmt := g.addString("%f")
				fFmtPtr := g.ptrTo(fFmt, 3)
				pos := g.tmp()
				g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", pos, buf))
				dst := g.tmp()
				g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
				rest := g.tmp()
				g.line(fmt.Sprintf("  %s = sub i64 %d, %s", rest, bufSize, pos))
				g.line(fmt.Sprintf("  %s = call i32 @snprintf(ptr %s, i64 %s, ptr %s, double %s)",
					g.tmp(), dst, rest, fFmtPtr, reg))
			}
		}
	}
	return buf, "ptr", nil
}

// emitWriteLnMulti prints multiple arguments (like WriteLn('x=', n, '!')) by
// building a string buffer via the interpolation infrastructure and puts-ing it.
func (g *Generator) emitWriteLnMulti(args []ast.Expression) (string, string, error) {
	const bufSize = 512
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", buf, bufSize))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", buf))

	ldFmt := g.addString("%ld")
	ldFmtPtr := g.ptrTo(ldFmt, 4)

	for _, arg := range args {
		reg, t, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		switch t {
		case "ptr":
			g.line(fmt.Sprintf("  %s = call ptr @strcat(ptr %s, ptr %s)", g.tmp(), buf, reg))
		case "i64":
			pos := g.tmp()
			g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", pos, buf))
			dst := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
			rest := g.tmp()
			g.line(fmt.Sprintf("  %s = sub i64 %d, %s", rest, bufSize, pos))
			g.line(fmt.Sprintf("  %s = call i32 @snprintf(ptr %s, i64 %s, ptr %s, i64 %s)",
				g.tmp(), dst, rest, ldFmtPtr, reg))
		case "double":
			fFmt := g.addString("%f")
			fFmtPtr := g.ptrTo(fFmt, 3)
			pos := g.tmp()
			g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", pos, buf))
			dst := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
			rest := g.tmp()
			g.line(fmt.Sprintf("  %s = sub i64 %d, %s", rest, bufSize, pos))
			g.line(fmt.Sprintf("  %s = call i32 @snprintf(ptr %s, i64 %s, ptr %s, double %s)",
				g.tmp(), dst, rest, fFmtPtr, reg))
		}
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @puts(ptr noundef %s)", r, buf))
	return "0", "void", nil
}
