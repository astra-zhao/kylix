// expr_access.go — member/method/interface/closure access expression codegen
// (split from expr.go in v4.5.0 to keep each source file under 1000 lines).
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

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

	// THttpClient (libcurl-backed handle): field access (e.g. c.BaseURL) is
	// lowered to a real GEP+load on the 32-byte handle struct. See
	// emitHttpclientFieldAccess in stdlib_httpclient.go.
	if typeName == "THttpClient" {
		return g.emitHttpclientFieldAccess(e.Object, e.Member)
	}

	if kind != "class" {
		// v5.4.0: emit a null ptr (not i64 0) so downstream comparisons with
		// ptr/string operands stay type-consistent and llc accepts the IR.
		// This is a conservative fallback for member access whose receiver
		// type is unknown (e.g. record-typed, not yet supported); it may
		// produce wrong values but keeps the module compilable.
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; member access on non-class %s.%s", r, typeName, e.Member))
		return r, "ptr", nil
	}
	objReg, objType, err := g.loadObjectPtr(e.Object, typeName)
	if err != nil {
		return "", "", err
	}
	reg, llvmT, err := g.emitFieldAccess(typeName, objReg, e.Member)
	if err != nil {
		return "", "", err
	}
	// If the field is a class-typed field, return the Kylix class name so
	// downstream method dispatch (obj.Method()) can resolve the receiver class.
	if llvmT == "ptr" {
		if classInfo, ok := g.classes[typeName]; ok {
			for _, f := range classInfo.Fields {
				if f.Name == e.Member && f.KylixType != "" {
					if _, isClass := g.classes[f.KylixType]; isClass {
						return reg, f.KylixType, nil
					}
					break
				}
			}
		}
	}
	_ = objType
	return reg, llvmT, nil
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
	// `self` is registered as "%self" (a function parameter, already a ptr);
	// other locals are "%v_name" allocas that need a load to dereference.
	// v5.4.0: a global (`@__kylix_g_*`) stores the object pointer — load it
	// into a fresh register so downstream GEP/load use the object, not the
	// global address (which llc could mis-optimize by forward-substituting the
	// global's stored value and conflating object/vtable offsets).
	if strings.HasPrefix(alloca, "@__kylix_g_") {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, alloca))
		return r, "ptr", nil
	}
	if !strings.HasPrefix(alloca, "%v_") {
		return alloca, "ptr", nil
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

	// Special case: stdlib opaque-handle types (TDateTime, TCache, ...) are
	// not registered in g.classes, so receiverKind returns kind="". Dispatch
	// their methods via the per-type stdlib emitter.
	if typeName == "TDateTime" {
		objReg, _, err := g.emitExpr(member.Object)
		if err != nil {
			return "", "", err
		}
		return g.emitDatetimeMethodCall(objReg, member.Member, args)
	}
	if typeName == "TCache" {
		objReg, _, err := g.emitExpr(member.Object)
		if err != nil {
			return "", "", err
		}
		return g.emitCacheMethodCall(objReg, member.Member, args)
	}
	if typeName == "THttpClient" {
		objReg, _, err := g.emitExpr(member.Object)
		if err != nil {
			return "", "", err
		}
		return g.emitHttpclientMethodCall(objReg, member.Member, args)
	}

	if kind == "" {
		// Unknown receiver type — evaluate it to check if it's a stdlib
		// opaque-handle type (TDateTime via chained call, TCache, ...) or
		// a class-typed field access (e.g. self.Repo where Repo: TUserRepository).
		objReg, objType, err := g.emitExpr(member.Object)
		if err != nil {
			return "", "", err
		}
		switch objType {
		case "TDateTime":
			return g.emitDatetimeMethodCall(objReg, member.Member, args)
		case "TCache":
			return g.emitCacheMethodCall(objReg, member.Member, args)
		case "THttpClient":
			return g.emitHttpclientMethodCall(objReg, member.Member, args)
		}
		// If objType is a known class name (field access returned the Kylix
		// class type), dispatch via emitVirtualCall directly.
		if _, isClass := g.classes[objType]; isClass {
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
			return g.emitVirtualCall(objType, objReg, member.Member, argRegs, argTypes)
		}
		// Not a known stdlib handle, fall through to unsupported receiver
	}

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

// emitClosureCall indirect-calls a closure stored in a local variable. The
// closure value is { func_ptr, env_ptr }; the call passes env as the first
// argument. Param/return types come from the lambda's recorded signature.
func (g *Generator) emitClosureCall(varName string, args []ast.Expression) (string, string, error) {
	allocaReg, ok := g.locals[varName]
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; undefined closure %s", r, varName))
		return r, "i64", nil
	}

	// Load the closure pair { ptr, ptr }.
	closureVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load { ptr, ptr }, ptr %s", closureVal, allocaReg))
	fptr := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, ptr } %s, 0", fptr, closureVal))
	eptr := g.tmp()
	g.line(fmt.Sprintf("  %s = extractvalue { ptr, ptr } %s, 1", eptr, closureVal))

	// Evaluate arguments, coercing each to the lambda's declared param type.
	paramTypes := g.closureParams[varName]
	retType := g.closureSigs[varName]
	if retType == "" {
		retType = "void"
	}

	var argRegs []string
	var argTypes []string
	for i, arg := range args {
		r, t, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		if i < len(paramTypes) && paramTypes[i] != t {
			r, t = g.coerceValue(r, t, paramTypes[i])
		}
		argRegs = append(argRegs, r)
		argTypes = append(argTypes, t)
	}

	// Function type signature: retType (ptr env, paramTypes...).
	sigParams := []string{"ptr"}
	sigParams = append(sigParams, paramTypes...)
	fnType := fmt.Sprintf("%s (%s)", retType, strings.Join(sigParams, ", "))

	callArgs := []string{"ptr " + eptr}
	for i, r := range argRegs {
		callArgs = append(callArgs, argTypes[i]+" "+r)
	}

	if retType == "void" {
		// Indirect call: `call <fnptrty> %ptr(args)` — fnptrty includes the return type.
		g.line(fmt.Sprintf("  call %s %s(%s)", fnType, fptr, strings.Join(callArgs, ", ")))
		return "0", "void", nil
	}
	result := g.tmp()
	// Indirect call: `call <fnptrty> %ptr(args)` — fnptrty includes the return type.
	g.line(fmt.Sprintf("  %s = call %s %s(%s)", result, fnType, fptr, strings.Join(callArgs, ", ")))
	return result, retType, nil
}

// emitIsExpr lowers `obj is IFoo` to a compile-time i1: 1 if the object's
// concrete class implements the target interface, else 0. Dynamic checks on
// already-boxed interface values would require runtime type IDs (deferred).
func (g *Generator) emitIsExpr(e *ast.IsExpression) (string, string, error) {
	target := typeExprName(e.TargetType)
	kind, typeName := g.receiverKind(e.Expression)
	// Interface target: old compile-time check (classImplementsInterface).
	if kind == "class" && g.classImplementsInterface(typeName, target) {
		val := 1
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i1 0, %d ; %s is %s", r, val, typeName, target))
		return r, "i1", nil
	}
	// v5.4.0: class target — runtime subtype check via vtable hierarchy.
	// `obj is TClass` where obj's compile-time type is a base class (or the
	// same class): load obj's vtable ptr and walk the class edge table to see
	// if TClass is an ancestor. This is what makes the bootstrap's ~95
	// `decl is TClassDecl` type-dispatch sites actually branch correctly.
	if _, isClassTarget := g.classes[target]; isClassTarget {
		objReg, _, err := g.loadObjectPtr(e.Expression, typeName)
		if err != nil {
			return "", "", err
		}
		r, err := g.classIsACall(objReg, target)
		if err != nil {
			return "", "", err
		}
		return r, "i1", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i1 0, 0 ; %s is %s (unsupported)", r, typeName, target))
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
	// v5.4.0: `obj as TConcreteClass` — the cast target is a class (not an
	// interface), so the instance is already the right opaque ptr (subclass
	// prefix layout is compatible). Return the object pointer. Runtime subtype
	// validation (classID + __kylix_class_is_a) is added in Stage 2 proper;
	// the bootstrap always guards `as` with `is`, so unconditional return is
	// safe for self-compilation.
	if _, isClass := g.classes[target]; isClass {
		objReg, _, err := g.loadObjectPtr(e.Expression, typeName)
		if err != nil {
			return "", "", err
		}
		return objReg, "ptr", nil
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
			case "variant":
				// v5.0.0: unbox the Variant to a string and strcat it.
				strReg := g.emitVariantAsStr(reg)
				g.line(fmt.Sprintf("  %s = call ptr @strcat(ptr %s, ptr %s)", g.tmp(), buf, strReg))
			case "ptr":
				g.line(fmt.Sprintf("  %s = call ptr @strcat(ptr %s, ptr %s)", g.tmp(), buf, reg))
			case "i64":
				pos := g.tmp()
				g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", pos, buf))
				dst := g.tmp()
				g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
				rest := g.tmp()
				g.line(fmt.Sprintf("  %s = sub i64 %d, %s", rest, bufSize, pos))
				g.line(fmt.Sprintf("  %s = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 %s, ptr %s, i64 %s)",
					g.tmp(), dst, rest, ldFmtPtr, reg))
			case "double":
				fFmt := g.addString("%.15g")
				fFmtPtr := g.ptrTo(fFmt, 6)
				pos := g.tmp()
				g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", pos, buf))
				dst := g.tmp()
				g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
				rest := g.tmp()
				g.line(fmt.Sprintf("  %s = sub i64 %d, %s", rest, bufSize, pos))
				g.line(fmt.Sprintf("  %s = call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 %s, ptr %s, double %s)",
					g.tmp(), dst, rest, fFmtPtr, reg))
			case "i1":
				// Boolean: append "true" or "false" string
				trueStr := g.addString("true")
				falseStr := g.addString("false")
				truePtr := g.ptrTo(trueStr, 5)
				falsePtr := g.ptrTo(falseStr, 6)
				selected := g.tmp()
				g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", selected, reg, truePtr, falsePtr))
				g.line(fmt.Sprintf("  %s = call ptr @strcat(ptr %s, ptr %s)", g.tmp(), buf, selected))
			}
		}
	}
	return buf, "ptr", nil
}

// emitWriteLnMulti prints multiple arguments (like WriteLn('x=', n, '!')) by
// building a string buffer via the interpolation infrastructure and puts-ing it.
