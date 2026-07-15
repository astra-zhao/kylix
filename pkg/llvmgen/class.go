// class.go — LLVM IR code generation for Kylix classes and interfaces.
//
// Kylix class → LLVM struct layout:
//   %TFoo = type { ptr, i64, ... }   ; first field = vtable pointer
//
// Vtable layout:
//   @TFoo_vtable = constant [N x ptr] [ ptr @TFoo_Method1, ptr @TFoo_Method2, ... ]
//
// Interface fat pointer (two-word representation):
//   { ptr vtable, ptr data }
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// ClassInfo holds metadata about a compiled class for code generation.
type ClassInfo struct {
	Name       string
	Fields     []FieldInfo
	Methods    []MethodInfo
	Parent     string
	Interfaces []string
}

// FieldInfo describes a class field.
type FieldInfo struct {
	Name      string
	LLVMType  string
	KylixType string // original Kylix type name (e.g. "TUserRepository", "Integer")
	Index     int
	ArrayType *ast.ArrayType // v4.8.0: set when the field is a static array (array[lo..hi] of T); enables self.Items[i] GEP
}

// MethodInfo describes a class method in the vtable.
type MethodInfo struct {
	Name           string
	VtableIdx      int
	RetType        string
	Params         []string
	DefiningClass  string // class where this method's implementation lives (for vtable emit)
}

// emitClassDecl generates LLVM type + vtable + method definitions for a class.
func (g *Generator) emitClassDecl(decl *ast.ClassDecl) error {
	info := g.buildClassInfo(decl)
	g.classes[decl.Name] = info

	// Emit struct type: first field is vtable ptr
	g.emitStructType(info)

	// Emit vtable constant
	g.emitVtable(info, decl)

	// Emit per-interface vtable constants (interface fat-pointer support).
	g.emitInterfaceVtables(info)

	// Emit method functions
	for _, method := range decl.Methods {
		if method.IsExternal {
			continue
		}
		if err := g.emitMethod(decl.Name, method); err != nil {
			return err
		}
	}

	return nil
}

// buildClassInfo extracts field/method metadata from a ClassDecl.
// Inherited fields from the parent class are prepended so that subclass
// instances include the parent's layout (e.g. TFooError inherits Exception.Message).
func (g *Generator) buildClassInfo(decl *ast.ClassDecl) *ClassInfo {
	info := &ClassInfo{
		Name:       decl.Name,
		Parent:     decl.Parent,
		Interfaces: decl.Interfaces,
	}

	// Fields: index 0 = vtable ptr. Inherited parent fields come first
	// (preserving the parent's layout), then this class's own fields.
	idx := 1
	if decl.Parent != "" {
		if parent, ok := g.classes[decl.Parent]; ok {
			for _, f := range parent.Fields {
				info.Fields = append(info.Fields, FieldInfo{
					Name:      f.Name,
					LLVMType:  f.LLVMType,
					KylixType: f.KylixType,
					ArrayType: f.ArrayType,
					Index:     idx,
				})
				idx++
			}
		}
	}
	for _, f := range decl.Fields {
		if len(f.Names) == 0 {
			continue
		}
		llvmT := "i64"
		kylixT := ""
		// v4.8.0: capture the ArrayType for static-array fields so
		// self.Items[i] can GEP into the embedded [N x T] storage. The
		// field's LLVM type is [N x T] (not the fallback i64 that
		// llvmTypeFor gives for "array" name).
		var arrT *ast.ArrayType
		if at, ok := f.Type.(*ast.ArrayType); ok && !at.Dynamic {
			arrT = at
			elemT := "i64"
			if at.ElementType != nil {
				elemT = LLVMType(typeExprName(at.ElementType))
			}
			size := int64(0)
			if at.Size != nil {
				size = evalConstInt(at.Size)
			}
			if size <= 0 {
				size = 1
			}
			llvmT = fmt.Sprintf("[%d x %s]", size, elemT)
		} else if f.Type != nil {
			kylixT = typeExprName(f.Type)
			llvmT = g.llvmTypeFor(kylixT)
		}
		for _, name := range f.Names {
			info.Fields = append(info.Fields, FieldInfo{
				Name:      name,
				LLVMType:  llvmT,
				KylixType: kylixT,
				ArrayType: arrT,
				Index:     idx,
			})
			idx++
		}
	}

	// Methods: build vtable. Inherited parent methods come first (so child
	// vtable is a superset of parent's), then the child's own methods. A child
	// method that overrides a parent method reuses the parent's vtable slot
	// (the slot points to the child implementation).
	if decl.Parent != "" {
		if parent, ok := g.classes[decl.Parent]; ok {
			for _, pm := range parent.Methods {
				mi := MethodInfo{
					Name:          pm.Name,
					VtableIdx:     pm.VtableIdx,
					RetType:       pm.RetType,
					Params:        pm.Params,
					DefiningClass: pm.DefiningClass, // inherited — still points to original definer
				}
				info.Methods = append(info.Methods, mi)
			}
		}
	}
	for _, m := range decl.Methods {
		retType := "void"
		if m.ReturnType != nil {
			retType = g.llvmTypeFor(typeExprName(m.ReturnType))
		}
		var paramTypes []string
		for _, p := range m.Parameters {
			pt := "i64"
			if p.Type != nil {
				pt = g.llvmTypeFor(typeExprName(p.Type))
			}
			paramTypes = append(paramTypes, pt)
		}
		// Override: if a parent method with the same name exists, reuse its
		// vtable slot but point to this child's implementation.
		overrode := false
		for i := range info.Methods {
			if info.Methods[i].Name == m.Name {
				info.Methods[i].RetType = retType
				info.Methods[i].Params = paramTypes
				info.Methods[i].DefiningClass = decl.Name
				overrode = true
				break
			}
		}
		if !overrode {
			info.Methods = append(info.Methods, MethodInfo{
				Name:          m.Name,
				VtableIdx:     len(info.Methods),
				RetType:       retType,
				Params:        paramTypes,
				DefiningClass: decl.Name,
			})
		}
	}

	return info
}

// emitStructType emits:  %TPoint = type { ptr, i64, i64, ... }
func (g *Generator) emitStructType(info *ClassInfo) {
	var fieldTypes []string
	fieldTypes = append(fieldTypes, "ptr") // vtable pointer
	for _, f := range info.Fields {
		fieldTypes = append(fieldTypes, f.LLVMType)
	}
	// Use class name directly (Kylix convention: names already start with T)
	g.line(fmt.Sprintf("%%%s = type { %s }", info.Name, strings.Join(fieldTypes, ", ")))
}

// emitVtable emits:  @TFoo_vtable = constant [N x ptr] [ ptr @TFoo_MethodA, ... ]
// Inherited method slots point to the parent's implementation; overridden
// slots point to the child's implementation (DefiningClass tracks this).
func (g *Generator) emitVtable(info *ClassInfo, decl *ast.ClassDecl) {
	if len(info.Methods) == 0 {
		return
	}
	// Build vtable in vtable-index order.
	slots := make([]string, len(info.Methods))
	for _, m := range info.Methods {
		if m.VtableIdx >= len(slots) {
			continue
		}
		defClass := m.DefiningClass
		if defClass == "" {
			defClass = info.Name
		}
		slots[m.VtableIdx] = fmt.Sprintf("ptr @%s_%s", defClass, m.Name)
	}
	g.line(fmt.Sprintf("@%s_vtable = constant [%d x ptr] [ %s ]",
		info.Name, len(slots), strings.Join(slots, ", ")))
}

// emitMethod emits a class method:  define <ret> @TFoo_Bar(ptr %self, ...) { ... }
func (g *Generator) emitMethod(className string, method *ast.FunctionDecl) error {
	retType := "void"
	if method.ReturnType != nil {
		retType = g.llvmTypeFor(typeExprName(method.ReturnType))
	}

	// Annotation-generated methods (ORM [Query], [Repository]) have no body —
	// only a signature. Emit a stub define so the vtable symbol resolves.
	// (No debug info for stubs: they have no source body to step through.)
	if method.Body == nil {
		var params []string
		params = append(params, "ptr %self")
		for _, p := range method.Parameters {
			llvmT := "i64"
			if p.Type != nil {
				llvmT = g.llvmTypeFor(typeExprName(p.Type))
			}
			params = append(params, fmt.Sprintf("%s %%%s", llvmT, p.Name))
		}
		g.line(fmt.Sprintf("define %s @%s_%s(%s) {", retType, className, method.Name, strings.Join(params, ", ")))
		switch retType {
		case "void":
			g.line("  ret void")
		case "ptr":
			emptyStr := g.addString("")
			g.line(fmt.Sprintf("  ret ptr %s", g.ptrTo(emptyStr, 1)))
		case "i1":
			g.line("  ret i1 false")
		case "double":
			g.line("  ret double 0.0")
		default:
			g.line("  ret i64 0")
		}
		g.line("}")
		g.line("")
		return nil
	}

	// Build parameter list — first param is always `ptr %self`
	var params []string
	params = append(params, fmt.Sprintf("ptr %%self"))
	for _, p := range method.Parameters {
		llvmT := "i64"
		if p.Type != nil {
			llvmT = g.llvmTypeFor(typeExprName(p.Type))
		}
		params = append(params, fmt.Sprintf("%s %%%s", llvmT, p.Name))
	}

	defineLine := fmt.Sprintf("define %s @%s_%s(%s) {", retType, className, method.Name, strings.Join(params, ", "))
	// v4.9.0: register a DISubprogram for the method so OOP methods get
	// per-line stepping + variable inspection (same pattern as emitFunctionDecl).
	var methodSpID int
	if g.debugInfo {
		methodSpID = g.registerSubprogram(className+"_"+method.Name, method.Token.Line)
		defineLine = g.defineLineWithDbg(defineLine, methodSpID)
	}
	g.line(defineLine)
	g.line("entry:")

	savedLocals := g.locals
	savedTypes := g.localTypes
	savedVarSeq := g.varNameSeq
	savedFunc := g.funcName
	savedClass := g.curClassName
	savedMethod := g.curMethodName
	g.locals = make(map[string]string)
	g.localTypes = make(map[string]string)
	g.varNameSeq = make(map[string]int)
	g.funcName = className + "_" + method.Name
	g.curClassName = className
	g.curMethodName = method.Name

	// v4.9.0: scope + position for DILocations emitted inside this method.
	if g.debugInfo {
		g.setDbgScope(methodSpID)
		g.setDbgNode(method)
	}

	// Register `self` pointer
	g.locals["self"] = "%self"
	g.localTypes["self"] = className
	if g.debugInfo {
		// `self` is the function's first param (ptr %self), not an alloca —
		// #dbg_declare on the param register itself associates it with the
		// `self` source variable so LLDB shows the receiver object.
		g.emitDbgDeclare("self", method.Token.Line, "ptr", "%self")
	}

	// Register method params
	for _, p := range method.Parameters {
		llvmT := "i64"
		kylixType := ""
		if p.Type != nil {
			kylixType = typeExprName(p.Type)
			llvmT = LLVMType(kylixType)
		}
		// Suffix by type so emitIdentLoad infers the load type correctly.
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
		// v4.9.0: declare the parameter as a debug local.
		if g.debugInfo {
			declLine := method.Token.Line
			if p.Token.Line > 0 {
				declLine = p.Token.Line
			}
			g.emitDbgDeclare(p.Name, declLine, llvmT, allocaReg)
		}
	}

	// Result variable
	if retType != "void" {
		g.line(fmt.Sprintf("  %%result = alloca %s, align 8", retType))
		g.locals["result"] = "%result"
		if g.debugInfo {
			g.emitDbgDeclare("result", method.Token.Line, retType, "%result")
		}
	}

	// Emit body
	if method.Body != nil {
		for _, stmt := range method.Body.Statements {
			if err := g.emitStatement(stmt); err != nil {
				return err
			}
		}
	}

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
	g.funcName = savedFunc
	g.curClassName = savedClass
	g.curMethodName = savedMethod
	// Leaving this method: clear the debug scope + position so subsequent
	// module-level code doesn't attach a stale !dbg.
	if g.debugInfo {
		g.setDbgScope(0)
		g.clearDbgPos()
	}
	return nil
}

// llvmTypeFor returns the LLVM type for a Kylix type name, taking into account
// user-defined classes (which are pointers to heap-allocated structs → "ptr").
func (g *Generator) llvmTypeFor(typeName string) string {
	if typeName == "" {
		return "i64"
	}
	// Class types are always ptr (heap-allocated).
	if _, ok := g.classes[typeName]; ok {
		return "ptr"
	}
	return LLVMType(typeName)
}

// emitFieldAccess generates a getelementptr + load for field access.
// selfReg is the pointer to the struct, fieldName is the Kylix field name.
func (g *Generator) emitFieldAccess(className, selfReg, fieldName string) (string, string, error) {
	info, ok := g.classes[className]
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; unknown class %s", r, className))
		return r, "i64", nil
	}

	for _, f := range info.Fields {
		if f.Name == fieldName {
			gepReg := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds %%%s, ptr %s, i32 0, i32 %d",
				gepReg, className, selfReg, f.Index))
			loadReg := g.tmp()
			g.line(fmt.Sprintf("  %s = load %s, ptr %s", loadReg, f.LLVMType, gepReg))
			return loadReg, f.LLVMType, nil
		}
	}

	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, 0 ; field %s not found in %s", r, fieldName, className))
	return r, "i64", nil
}

// emitFieldStore generates a getelementptr + store for writing to a field.
// selfReg is the pointer to the struct, fieldName is the Kylix field name.
// Returns the gep register (pointing to the field slot) and the field's LLVM
// type so the caller can coerce and store the value.
func (g *Generator) emitFieldStore(className, selfReg, fieldName string) (gepReg, fieldType string, err error) {
	info, ok := g.classes[className]
	if !ok {
		return "", "i64", fmt.Errorf("unknown class %s", className)
	}
	for _, f := range info.Fields {
		if f.Name == fieldName {
			gep := g.tmp()
			g.line(fmt.Sprintf("  %s = getelementptr inbounds %%%s, ptr %s, i32 0, i32 %d",
				gep, className, selfReg, f.Index))
			return gep, f.LLVMType, nil
		}
	}
	return "", "i64", fmt.Errorf("field %s not found in %s", fieldName, className)
}

// emitConstructor generates a constructor call that allocates and initializes a class.
func (g *Generator) emitConstructor(className string) (string, error) {
	info, ok := g.classes[className]
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; unknown class %s", r, className))
		return r, nil
	}

	// Calculate struct size: 8 bytes per field + 8 for vtable ptr
	size := 8 * (1 + len(info.Fields))
	allocReg := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", allocReg, size))

	// Store vtable pointer at offset 0
	if len(info.Methods) > 0 {
		vtablePtr := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds %%%s, ptr %s, i32 0, i32 0",
			vtablePtr, className, allocReg))
		g.line(fmt.Sprintf("  store ptr @%s_vtable, ptr %s", className, vtablePtr))
	}

	return allocReg, nil
}

// initConstructorArgs applies constructor arguments to a freshly-allocated
// object. Currently handles the common Pascal pattern where the first arg
// initializes a String field named "Message" (e.g. Exception.Create('msg')).
// Other args/fields are ignored — a conservative default that produces a valid
// object pointer without a full constructor-method call.
func (g *Generator) initConstructorArgs(className, objReg string, args []ast.Expression) {
	if len(args) == 0 {
		return
	}
	info, ok := g.classes[className]
	if !ok {
		return
	}
	// Find a String-typed Message field (case-insensitive).
	msgIdx := -1
	for i, f := range info.Fields {
		if strings.EqualFold(f.Name, "Message") && f.LLVMType == "ptr" {
			msgIdx = i
			break
		}
	}
	if msgIdx < 0 {
		return
	}
	// Evaluate the first argument as a string pointer and store it.
	argReg, argType, err := g.emitExpr(args[0])
	if err != nil {
		return
	}
	if argType != "ptr" {
		return // not a string; skip rather than emit a bad store
	}
	fieldPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds %%%s, ptr %s, i32 0, i32 %d",
		fieldPtr, className, objReg, info.Fields[msgIdx].Index))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", argReg, fieldPtr))
}

// emitVirtualCall generates a vtable method dispatch.
func (g *Generator) emitVirtualCall(className, objReg, methodName string, argRegs []string, argTypes []string) (string, string, error) {
	info, ok := g.classes[className]
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; unknown class %s", r, className))
		return r, "i64", nil
	}

	// Find method in vtable
	var meth *MethodInfo
	for i := range info.Methods {
		if info.Methods[i].Name == methodName {
			meth = &info.Methods[i]
			break
		}
	}
	if meth == nil {
		// Annotation-generated methods (IsValid/Validate from [Required]/
		// [Email] etc.) don't exist in the LLVM backend (no KylixBoot
		// codegen pass). Return a stub so the code at least compiles:
		// IsValid/Validate → true (validation always passes).
		if methodName == "IsValid" || methodName == "Validate" {
			r := g.tmp()
			g.line(fmt.Sprintf("  %s = add i1 0, 1 ; %s (annotation stub: true)", r, methodName))
			return r, "i1", nil
		}
		// ORM [Query('...')] / [Repository] generates methods like FindAll,
		// FindById, Save, DeleteById, and the query-specific method (e.g.
		// All). Stub them as empty-string (ptr) or 0 (i64) depending on
		// likely return type — collection methods return ptr (empty string
		// as a safe default), single-entity methods return ptr.
		if methodName == "FindAll" || methodName == "All" ||
			methodName == "FindById" || methodName == "ByEmail" {
			emptyStr := g.addString("")
			return g.ptrTo(emptyStr, 1), "ptr", nil
		}
		if methodName == "Save" || methodName == "DeleteById" {
			r := g.tmp()
			g.line(fmt.Sprintf("  %s = add i64 0, 0 ; %s (ORM stub)", r, methodName))
			return r, "i64", nil
		}
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; method %s not found", r, methodName))
		return r, "i64", nil
	}

	// Load vtable pointer from struct[0]
	vtablePtrLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds %%%s, ptr %s, i32 0, i32 0",
		vtablePtrLoc, className, objReg))
	vtablePtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", vtablePtr, vtablePtrLoc))

	// Load function pointer from vtable[idx]
	fnPtrLoc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x ptr], ptr %s, i32 0, i32 %d",
		fnPtrLoc, len(info.Methods), vtablePtr, meth.VtableIdx))
	fnPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", fnPtr, fnPtrLoc))

	// Call via function pointer
	var callArgs []string
	callArgs = append(callArgs, "ptr "+objReg) // self
	for i, r := range argRegs {
		callArgs = append(callArgs, argTypes[i]+" "+r)
	}

	// Build function type signature for indirect call.
	var paramTypes []string
	paramTypes = append(paramTypes, "ptr") // self
	paramTypes = append(paramTypes, meth.Params...)
	fnType := fmt.Sprintf("%s (%s)", meth.RetType, strings.Join(paramTypes, ", "))

	if meth.RetType == "void" {
		g.line(fmt.Sprintf("  call %s %s(%s)", fnType, fnPtr, strings.Join(callArgs, ", ")))
		return "0", "void", nil
	}
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = call %s %s(%s)", result, fnType, fnPtr, strings.Join(callArgs, ", ")))
	return result, meth.RetType, nil
}

// findMethodInHierarchy walks the parent chain from className looking for a
// method named methodName. Returns the defining class (where the
// implementation actually lives, accounting for inherited vtable slots) and
// method info, or ("", nil) if not found.
func (g *Generator) findMethodInHierarchy(className, methodName string) (string, *MethodInfo) {
	visited := map[string]bool{}
	for c := className; c != "" && !visited[c]; c = g.classes[c].Parent {
		visited[c] = true
		info, ok := g.classes[c]
		if !ok {
			break
		}
		for i := range info.Methods {
			if info.Methods[i].Name == methodName {
				m := &info.Methods[i]
				defClass := m.DefiningClass
				if defClass == "" {
					defClass = c
				}
				return defClass, m
			}
		}
	}
	return "", nil
}

// emitInherited handles `inherited;` and `inherited MethodName(args)`.
// It calls the parent class's method implementation directly (no vtable
// dispatch), passing `self` as the receiver.
func (g *Generator) emitInherited(s *ast.InheritedStatement) error {
	methodName := g.curMethodName
	var argExprs []ast.Expression

	if s.Expr != nil {
		// `inherited MethodName(args)` — Expr is a CallExpression.
		if call, ok := s.Expr.(*ast.CallExpression); ok {
			if ident, ok := call.Function.(*ast.Identifier); ok {
				methodName = ident.Value
			}
			argExprs = call.Arguments
		}
	}

	// Find the method in the parent chain (skip the current class itself).
	parentClass := ""
	if info, ok := g.classes[g.curClassName]; ok {
		parentClass = info.Parent
	}
	defClass, meth := g.findMethodInHierarchy(parentClass, methodName)
	if meth == nil {
		g.line(fmt.Sprintf("  ; inherited: method %s not found in parent chain of %s",
			methodName, g.curClassName))
		return nil
	}

	// Evaluate arguments, coercing to the method's declared param types.
	var argRegs []string
	var argTypes []string
	for i, arg := range argExprs {
		r, t, err := g.emitExpr(arg)
		if err != nil {
			return err
		}
		if i < len(meth.Params) && meth.Params[i] != t {
			r, t = g.coerceValue(r, t, meth.Params[i])
		}
		argRegs = append(argRegs, r)
		argTypes = append(argTypes, t)
	}

	// Direct call to @ParentClass_MethodName(ptr %self, args).
	var callArgs []string
	callArgs = append(callArgs, "ptr %self")
	for i, r := range argRegs {
		callArgs = append(callArgs, argTypes[i]+" "+r)
	}
	fnName := fmt.Sprintf("@%s_%s", defClass, methodName)
	if meth.RetType == "void" {
		g.line(fmt.Sprintf("  call void %s(%s)", fnName, strings.Join(callArgs, ", ")))
		return nil
	}
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = call %s %s(%s)", result, meth.RetType, fnName, strings.Join(callArgs, ", ")))
	// Store the result into %result so the surrounding method returns it.
	if g.locals["result"] != "" {
		g.line(fmt.Sprintf("  store %s %s, ptr %%result", meth.RetType, result))
	}
	return nil
}
