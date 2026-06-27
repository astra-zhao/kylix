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
	Name     string
	LLVMType string
	Index    int
}

// MethodInfo describes a class method in the vtable.
type MethodInfo struct {
	Name      string
	VtableIdx int
	RetType   string
	Params    []string
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
		if method.IsExternal || method.Body == nil {
			continue
		}
		if err := g.emitMethod(decl.Name, method); err != nil {
			return err
		}
	}

	return nil
}

// buildClassInfo extracts field/method metadata from a ClassDecl.
func (g *Generator) buildClassInfo(decl *ast.ClassDecl) *ClassInfo {
	info := &ClassInfo{
		Name:       decl.Name,
		Parent:     decl.Parent,
		Interfaces: decl.Interfaces,
	}

	// Fields: index 0 = vtable ptr
	idx := 1
	for _, f := range decl.Fields {
		if len(f.Names) == 0 {
			continue
		}
		llvmT := "i64"
		if f.Type != nil {
			llvmT = LLVMType(typeExprName(f.Type))
		}
		for _, name := range f.Names {
			info.Fields = append(info.Fields, FieldInfo{
				Name:     name,
				LLVMType: llvmT,
				Index:    idx,
			})
			idx++
		}
	}

	// Methods: build vtable index
	for i, m := range decl.Methods {
		retType := "void"
		if m.ReturnType != nil {
			retType = LLVMType(typeExprName(m.ReturnType))
		}
		var paramTypes []string
		for _, p := range m.Parameters {
			pt := "i64"
			if p.Type != nil {
				pt = LLVMType(typeExprName(p.Type))
			}
			paramTypes = append(paramTypes, pt)
		}
		info.Methods = append(info.Methods, MethodInfo{
			Name:      m.Name,
			VtableIdx: i,
			RetType:   retType,
			Params:    paramTypes,
		})
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
func (g *Generator) emitVtable(info *ClassInfo, decl *ast.ClassDecl) {
	if len(info.Methods) == 0 {
		return
	}
	var ptrs []string
	for _, m := range info.Methods {
		ptrs = append(ptrs, fmt.Sprintf("ptr @%s_%s", info.Name, m.Name))
	}
	g.line(fmt.Sprintf("@%s_vtable = constant [%d x ptr] [ %s ]",
		info.Name, len(ptrs), strings.Join(ptrs, ", ")))
}

// emitMethod emits a class method:  define <ret> @TFoo_Bar(ptr %self, ...) { ... }
func (g *Generator) emitMethod(className string, method *ast.FunctionDecl) error {
	retType := "void"
	if method.ReturnType != nil {
		retType = LLVMType(typeExprName(method.ReturnType))
	}

	// Build parameter list — first param is always `ptr %self`
	var params []string
	params = append(params, fmt.Sprintf("ptr %%self"))
	for _, p := range method.Parameters {
		llvmT := "i64"
		if p.Type != nil {
			llvmT = LLVMType(typeExprName(p.Type))
		}
		params = append(params, fmt.Sprintf("%s %%%s", llvmT, p.Name))
	}

	g.line(fmt.Sprintf("define %s @%s_%s(%s) {", retType, className, method.Name, strings.Join(params, ", ")))
	g.line("entry:")

	savedLocals := g.locals
	savedTypes := g.localTypes
	savedFunc := g.funcName
	g.locals = make(map[string]string)
	g.localTypes = make(map[string]string)
	g.funcName = className + "_" + method.Name

	// Register `self` pointer
	g.locals["self"] = "%self"

	// Register method params
	for _, p := range method.Parameters {
		llvmT := "i64"
		if p.Type != nil {
			llvmT = LLVMType(typeExprName(p.Type))
		}
		allocaReg := fmt.Sprintf("%%v_%s_int", p.Name)
		g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, llvmT))
		g.line(fmt.Sprintf("  store %s %%%s, ptr %s", llvmT, p.Name, allocaReg))
		g.locals[p.Name] = allocaReg
		if p.Type != nil {
			g.localTypes[p.Name] = typeExprName(p.Type)
		}
	}

	// Result variable
	if retType != "void" {
		g.line(fmt.Sprintf("  %%result = alloca %s, align 8", retType))
		g.locals["result"] = "%result"
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
	g.funcName = savedFunc
	return nil
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

	result := g.tmp()
	if meth.RetType == "void" {
		g.line(fmt.Sprintf("  call void %s(%s)", fnPtr, strings.Join(callArgs, ", ")))
		return "0", "void", nil
	}

	// Build function type signature for indirect call
	var paramTypes []string
	paramTypes = append(paramTypes, "ptr") // self
	paramTypes = append(paramTypes, meth.Params...)
	fnType := fmt.Sprintf("%s (%s)", meth.RetType, strings.Join(paramTypes, ", "))
	g.line(fmt.Sprintf("  %s = call %s %s(%s)", result, fnType, fnPtr, strings.Join(callArgs, ", ")))
	return result, meth.RetType, nil
}
