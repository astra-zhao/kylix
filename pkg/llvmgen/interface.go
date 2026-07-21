// interface.go — LLVM IR code generation for Kylix interfaces.
//
// An interface is represented at runtime as a fat pointer:
//
//	%IFoo_iface = type { ptr, ptr }   ; { vtable, data }
//
// where `vtable` points at a per-class constant:
//
//	@TFoo_IFoo_vtable = constant [N x ptr] [ ptr @TFoo_IFoo_thunk_m0, ... ]
//
// The vtable slots are in **interface declaration order**. Thunks adapt the
// concrete method's `self` (concrete struct ptr) to the interface slot's
// signature. For the MVP, signatures are assumed identical and the vtable
// points directly at the concrete method @Class_Method.
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// InterfaceInfo holds metadata about a compiled interface.
type InterfaceInfo struct {
	Name    string
	Methods []InterfaceMethodInfo
}

// InterfaceMethodInfo describes one method slot in an interface vtable.
type InterfaceMethodInfo struct {
	Name      string
	VtableIdx int
	RetType   string
	Params    []string
}

// interfaceVtableName returns the per-class interface vtable global name.
func interfaceVtableName(className, ifaceName string) string {
	return fmt.Sprintf("@%s_%s_vtable", className, ifaceName)
}

// emitInterfaceDecl emits the interface vtable type and registers InterfaceInfo.
func (g *Generator) emitInterfaceDecl(decl *ast.InterfaceDecl) error {
	info := &InterfaceInfo{Name: decl.Name}
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
		info.Methods = append(info.Methods, InterfaceMethodInfo{
			Name:      m.Name,
			VtableIdx: i,
			RetType:   retType,
			Params:    paramTypes,
		})
	}
	g.interfaces[decl.Name] = info

	// Fat-pointer type: { ptr vtable, ptr data }
	g.line(fmt.Sprintf("; interface %s — fat pointer { ptr vtable, ptr data }", decl.Name))
	if len(info.Methods) > 0 {
		g.line(fmt.Sprintf("%%%s_vtable = type { %s }",
			decl.Name, strings.Join(repeatPtr(len(info.Methods)), ", ")))
	} else {
		g.line(fmt.Sprintf("%%%s_vtable = type {}", decl.Name))
	}
	g.line(fmt.Sprintf("%%%s_iface = type { ptr, ptr }", decl.Name))
	g.line("")
	return nil
}

func repeatPtr(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "ptr"
	}
	return out
}

// emitInterfaceVtables emits per-interface vtable constants for a class that
// implements one or more interfaces. Called from emitClassDecl.
func (g *Generator) emitInterfaceVtables(info *ClassInfo) {
	for _, ifaceName := range info.Interfaces {
		iface, ok := g.interfaces[ifaceName]
		if !ok || len(iface.Methods) == 0 {
			continue
		}
		var slots []string
		for _, im := range iface.Methods {
			// MVP: concrete method signature matches interface slot, point
			// the vtable directly at @Class_Method.
			slots = append(slots, fmt.Sprintf("ptr @%s_%s", info.Name, im.Name))
		}
		g.line(fmt.Sprintf("%s = constant { %s } [ %s ]",
			interfaceVtableName(info.Name, ifaceName),
			strings.Join(repeatPtr(len(slots)), ", "),
			strings.Join(slots, ", ")))
	}
}

// classImplementsInterface reports whether a class is registered as
// implementing the named interface.
func (g *Generator) classImplementsInterface(className, ifaceName string) bool {
	info, ok := g.classes[className]
	if !ok {
		return false
	}
	for _, i := range info.Interfaces {
		if i == ifaceName {
			return true
		}
	}
	return false
}

// emitBoxInterface boxes a concrete object register into an interface fat
// pointer, returning two registers: the vtable ptr and the data ptr. The
// caller stores them into the interface-typed local's two allocas.
func (g *Generator) emitBoxInterface(className, ifaceName, objReg string) (vtableReg, dataReg string) {
	// The fat pointer stores a pointer to the per-class interface vtable
	// constant, plus the object data pointer.
	return interfaceVtableName(className, ifaceName), objReg
}

func (g *Generator) interfaceMethodCount(ifaceName string) int {
	if iface, ok := g.interfaces[ifaceName]; ok {
		return len(iface.Methods)
	}
	return 0
}

// receiverKind returns ("class"|"interface"|"") and the type name for a
// receiver expression — currently supports plain identifiers tracked in
// g.localTypes.
func (g *Generator) receiverKind(obj ast.Expression) (kind, typeName string) {
	ident, ok := obj.(*ast.Identifier)
	if !ok {
		// v5.4.0: handle member-access chains (X.Y.Z) — resolve the object's
		// Kylix type via exprKylixType (which recurses through field accesses),
		// then classify as class/interface. Previously non-Identifier receivers
		// returned ""/"", breaking `X.Y.Field[i]` and `X.Y.Method()`.
		t := g.exprKylixType(obj)
		if t == "" {
			return "", ""
		}
		if _, isClass := g.classes[t]; isClass {
			return "class", t
		}
		if _, isIface := g.interfaces[t]; isIface {
			return "interface", t
		}
		return "", t
	}
	t, ok := g.localTypes[ident.Value]
	if !ok {
		return "", ""
	}
	if _, isClass := g.classes[t]; isClass {
		return "class", t
	}
	if _, isIface := g.interfaces[t]; isIface {
		return "interface", t
	}
	return "", t
}

// interfaceLocalNames returns the per-interface-var alloca names for the
// vtable and data slots.
func interfaceLocalNames(varName string) (vtableAlloca, dataAlloca string) {
	return fmt.Sprintf("%%v_%s_iface_vt", varName), fmt.Sprintf("%%v_%s_iface_data", varName)
}

// emitInterfaceVarDecl reserves two pointer-sized slots for an interface-typed
// local. The caller has already recorded the kylix type into g.localTypes.
func (g *Generator) emitInterfaceVarDecl(name string) {
	vt, data := interfaceLocalNames(name)
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", vt))
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", data))
	g.line(fmt.Sprintf("  store ptr null, ptr %s", vt))
	g.line(fmt.Sprintf("  store ptr null, ptr %s", data))
	g.locals[name] = vt // primary alloca for legacy lookups (data accessor uses interfaceLocalNames)
}

// emitInterfaceAssign stores the given vtable+data pair into an interface-typed
// local's two allocas.
func (g *Generator) emitInterfaceAssign(varName, vtableReg, dataReg string) {
	vt, data := interfaceLocalNames(varName)
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", vtableReg, vt))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", dataReg, data))
}

// evalInterfaceRHS evaluates the right-hand side of an `iface := X` assignment
// and returns the (vtable, data) pair to store. Reports ok=false if the RHS
// cannot be boxed into the target interface.
func (g *Generator) evalInterfaceRHS(value ast.Expression, ifaceName string) (vtableReg, dataReg string, ok bool) {
	switch rhs := value.(type) {
	case *ast.TypeCastExpression:
		target := typeExprName(rhs.TargetType)
		if target != ifaceName {
			return "", "", false
		}
		kind, typeName := g.receiverKind(rhs.Expression)
		if kind == "class" && g.classImplementsInterface(typeName, ifaceName) {
			objReg, _, err := g.loadObjectPtr(rhs.Expression, typeName)
			if err != nil {
				return "", "", false
			}
			vt, data := g.emitBoxInterface(typeName, ifaceName, objReg)
			return vt, data, true
		}
		return "", "", false
	case *ast.Identifier:
		kind, typeName := g.receiverKind(rhs)
		if kind == "class" && g.classImplementsInterface(typeName, ifaceName) {
			objReg, _, err := g.loadObjectPtr(rhs, typeName)
			if err != nil {
				return "", "", false
			}
			vt, data := g.emitBoxInterface(typeName, ifaceName, objReg)
			return vt, data, true
		}
		if kind == "interface" && typeName == ifaceName {
			// iface := otherIface — copy slots.
			otherVT, otherData := interfaceLocalNames(rhs.Value)
			vtReg := g.tmp()
			g.line(fmt.Sprintf("  %s = load ptr, ptr %s", vtReg, otherVT))
			dataReg := g.tmp()
			g.line(fmt.Sprintf("  %s = load ptr, ptr %s", dataReg, otherData))
			return vtReg, dataReg, true
		}
		return "", "", false
	}
	return "", "", false
}


