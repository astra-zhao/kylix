// array.go — LLVM IR codegen for Kylix arrays (Milestone 2).
//
// Static arrays:  var nums: array[1..5] of Integer;
//   → %v_nums = alloca [5 x i64], align 8
//   nums[1] → getelementptr [5 x i64], ptr %v_nums, i64 0, i64 0
//   (Pascal 1-indexed → LLVM 0-indexed)
//
// Dynamic arrays:  var nums: array of Integer;
//   → %v_nums = alloca %struct.kylix_slice, align 8
//   {ptr data, i64 len, i64 cap}
//
// SetLength(arr, n) and append(arr, x) call libc realloc.
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// arrayInfo tracks compile-time metadata for an array variable.
type arrayInfo struct {
	IsDynamic   bool
	ElementType string // LLVM type, e.g. "i64", "double", "ptr"
	// ElementKylixType is the Kylix type name of elements (v5.4.0), so that
	// `arr[i].Method()` / `arr[i] is TFoo` can resolve the receiver type when
	// the element is a class (e.g. array of TNode → elements are TNode).
	ElementKylixType string
	Size        int64  // for static arrays only
	LowerBound  int64  // Pascal range lower bound (default 1)
	// v5.0.0: IsVariant marks `array of Variant` — elements are box pointers
	// (ptr), and index reads return the "variant" pseudo-type so downstream
	// comparisons/printing dispatch on the tag.
	IsVariant bool
}

// emitArrayVarDecl allocates an array variable.
// Returns true if the type was an array; false otherwise.
func (g *Generator) emitArrayVarDecl(name string, arr *ast.ArrayType) bool {
	elemType := "i64"
	elemKylix := ""
	isVariant := false
	if arr.ElementType != nil {
		elemKylix = typeExprName(arr.ElementType)
		// v5.0.0: detect Variant element → elements are box pointers (ptr),
		// and flag IsVariant so index reads return the "variant" pseudo-type.
		if isVariantTypeExpr(arr.ElementType) {
			isVariant = true
			elemType = "ptr"
			g.needVariantRuntime = true
		} else {
			// v5.4.0: resolve element type via the AST (class elements → ptr,
			// instead of the old LLVMType(typeExprName(...)) which fell back
			// to i64 for every class/nested-array element).
			elemType = g.llvmTypeOfExpr(arr.ElementType)
		}
	}

	if arr.Dynamic {
		// Dynamic array: { ptr data; i64 len; i64 cap }
		// For Milestone 2, represent as a stack-allocated struct of 3 words.
		allocaReg := g.freshVarReg(name, "_dyn")
		g.line(fmt.Sprintf("  %s = alloca { ptr, i64, i64 }, align 8", allocaReg))
		// Zero-init: data=null, len=0, cap=0
		g.line(fmt.Sprintf("  store { ptr, i64, i64 } zeroinitializer, ptr %s", allocaReg))
		g.locals[name] = allocaReg
		g.arrayInfo[name] = &arrayInfo{IsDynamic: true, ElementType: elemType, ElementKylixType: elemKylix, IsVariant: isVariant}
		return true
	}

	// Static array: alloca [N x T]
	size := int64(0)
	if arr.Size != nil {
		size = evalConstInt(arr.Size)
	}
	if size <= 0 {
		size = 1 // safety
	}
	// Pascal range lower bound for index adjustment (array[0..N] → 0,
	// array[1..N] → 1, array[5..N] → 5). nil for the single-value form
	// array[N] → default 0. emitArrayIndex uses this to shift source indices
	// to LLVM 0-based (sub idx, LowerBound). The previous hardcoded 1 broke
	// array[0..N] (0-1 underflowed → wild GEP → segfault on example23).
	lb := int64(0)
	if arr.LowerBound != nil {
		lb = evalConstInt(arr.LowerBound)
	}

	allocaReg := g.freshVarReg(name, "_arr")
	g.line(fmt.Sprintf("  %s = alloca [%d x %s], align 8", allocaReg, size, elemType))
	// Zero-init the whole array
	g.line(fmt.Sprintf("  store [%d x %s] zeroinitializer, ptr %s", size, elemType, allocaReg))

	g.locals[name] = allocaReg
	g.arrayInfo[name] = &arrayInfo{
		IsDynamic:        false,
		ElementType:      elemType,
		ElementKylixType: elemKylix,
		Size:             size,
		LowerBound:       lb,
		IsVariant:        isVariant,
	}
	return true
}

// emitArrayIndex generates GEP for array[index] access.
// Returns (resultReg, elementType, error). For assignment context, returns
// the pointer register (use g.line to emit a store yourself).
func (g *Generator) emitArrayIndex(idx *ast.IndexExpression, asLValue bool) (string, string, error) {
	// Class field array: self.Items[i] / rec.Fields[i] — Left is a
	// MemberExpression whose Object is a class-typed identifier and Member is
	// the array field name. Resolve the field address (the [N x T] static
	// storage or the {ptr,len,cap} dynamic slice embedded in the struct) via
	// emitFieldStore (GEP without load), then index into it.
	// v5.4.0: handles dynamic array fields (slice GEP) in addition to static,
	// and uses llvmTypeOfExpr so class elements resolve to ptr (not i64).
	if member, ok := idx.Left.(*ast.MemberExpression); ok {
		kind, typeName := g.receiverKind(member.Object)
		if kind == "class" {
			objReg, _, err := g.loadObjectPtr(member.Object, typeName)
			if err != nil {
				return "", "", err
			}
			fieldAddr, _, err := g.emitFieldStore(typeName, objReg, member.Member)
			if err != nil {
				return "", "", err
			}
			// Look up the field's array/map metadata (recorded by buildClassInfo).
			var at *ast.ArrayType
			var mt *ast.MapType
			if info, ok := g.classes[typeName]; ok {
				for _, f := range info.Fields {
					if f.Name == member.Member {
						at = f.ArrayType
						mt = f.MapType
						break
					}
				}
			}
			if mt != nil {
				// v5.4.0: map field obj.Field[key] → load the htab handle from the
				// field slot, then htab_get(handle, key).
				return g.emitMapFieldIndexGet(typeName, objReg, member.Member, mt, idx)
			}
			if at == nil {
				return "", "", fmt.Errorf("class field %s.%s is not an array or map", typeName, member.Member)
			}
			elemT := "i64"
			if at.ElementType != nil {
				elemT = g.llvmTypeOfExpr(at.ElementType)
			}
			if at.Dynamic {
				// Dynamic slice field {ptr,len,cap} embedded in the struct:
				// load data ptr (offset 0), then GEP element.
				idxReg, _, err := g.emitExpr(idx.Index)
				if err != nil {
					return "", "", err
				}
				zeroIdx := g.tmp()
				lb := int64(0)
				if at.LowerBound != nil {
					lb = evalConstInt(at.LowerBound)
				}
				if lb != 0 {
					g.line(fmt.Sprintf("  %s = sub i64 %s, %d", zeroIdx, idxReg, lb))
				} else {
					g.line(fmt.Sprintf("  %s = add i64 %s, 0", zeroIdx, idxReg))
				}
				dataPtrLoc := g.tmp()
				g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 0",
					dataPtrLoc, fieldAddr))
				dataPtr := g.tmp()
				g.line(fmt.Sprintf("  %s = load ptr, ptr %s", dataPtr, dataPtrLoc))
				ptr := g.tmp()
				g.line(fmt.Sprintf("  %s = getelementptr inbounds %s, ptr %s, i64 %s",
					ptr, elemT, dataPtr, zeroIdx))
				if asLValue {
					return ptr, elemT, nil
				}
				loaded := g.tmp()
				g.line(fmt.Sprintf("  %s = load %s, ptr %s", loaded, elemT, ptr))
				return loaded, elemT, nil
			}
			// Static array field [N x T] embedded in the struct.
			arrSize := int64(1)
			if at.Size != nil {
				arrSize = evalConstInt(at.Size)
			}
			if arrSize <= 0 {
				arrSize = 1
			}
			lb := int64(0)
			if at.LowerBound != nil {
				lb = evalConstInt(at.LowerBound)
			}
			return g.emitStaticArrayGEP(fieldAddr, arrSize, elemT, lb, idx.Index, asLValue)
		}
	}

	// Resolve the array variable
	leftIdent, ok := idx.Left.(*ast.Identifier)
	if !ok {
		// v5.4.0 diagnostic: identify why the class-field branch above didn't
		// handle this MemberExpression index (e.g. receiver not recognized as a
		// class, or the Object is a non-Identifier expression).
		if m, mok := idx.Left.(*ast.MemberExpression); mok {
			kind, tn := g.receiverKind(m.Object)
			return "", "", fmt.Errorf("array index target must be an identifier (member %q, object type %T, receiverKind=%q/%q)",
				m.Member, m.Object, kind, tn)
		}
		return "", "", fmt.Errorf("array index target must be an identifier (left type %T)", idx.Left)
	}

	// Map variable? Route to htab_get (map indexing reuses the hash-table
	// runtime; see stdlib_map.go).
	if g.mapVars[leftIdent.Value] {
		return g.emitMapIndexGet(idx)
	}

	allocaReg, ok := g.locals[leftIdent.Value]
	if !ok {
		return "", "", fmt.Errorf("undefined array variable: %s", leftIdent.Value)
	}
	info, hasInfo := g.arrayInfo[leftIdent.Value]
	if !hasInfo {
		return "", "", fmt.Errorf("variable %s is not an array", leftIdent.Value)
	}

	// Compute index (Pascal range → LLVM 0-based).
	idxReg, _, err := g.emitExpr(idx.Index)
	if err != nil {
		return "", "", err
	}
	zeroIdx := g.tmp()
	if !info.IsDynamic && info.LowerBound != 0 {
		g.line(fmt.Sprintf("  %s = sub i64 %s, %d", zeroIdx, idxReg, info.LowerBound))
	} else {
		g.line(fmt.Sprintf("  %s = add i64 %s, 0", zeroIdx, idxReg))
	}

	var ptr string
	if info.IsDynamic {
		// Load data pointer from the {ptr,len,cap} struct, then GEP into it.
		dataPtrLoc := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 0",
			dataPtrLoc, allocaReg))
		dataPtr := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", dataPtr, dataPtrLoc))
		ptr = g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds %s, ptr %s, i64 %s",
			ptr, info.ElementType, dataPtr, zeroIdx))
	} else {
		ptr = g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x %s], ptr %s, i64 0, i64 %s",
			ptr, info.Size, info.ElementType, allocaReg, zeroIdx))
	}

	if asLValue {
		// asLValue element type: for Variant arrays the slot stores a box
		// pointer (ptr); for typed arrays it's the element type.
		if info.IsVariant {
			return ptr, "ptr", nil
		}
		return ptr, info.ElementType, nil
	}
	// v5.0.0: for Variant arrays, load the box pointer and return the
	// "variant" pseudo-type so downstream comparisons/printing dispatch.
	if info.IsVariant {
		loaded := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", loaded, ptr))
		return loaded, variantT, nil
	}
	loaded := g.tmp()
	g.line(fmt.Sprintf("  %s = load %s, ptr %s", loaded, info.ElementType, ptr))
	return loaded, info.ElementType, nil
}

// emitStaticArrayGEP emits the index-compute + GEP for a static array given
// its storage base pointer, element count, element type, and Pascal lower
// bound. Shared by the class-field-array path (local arrays stay inlined in
// emitArrayIndex above for the dynamic-array branch). Returns the element
// pointer (asLValue) or the loaded value.
func (g *Generator) emitStaticArrayGEP(baseReg string, size int64, elemType string, lowerBound int64, indexExpr ast.Expression, asLValue bool) (string, string, error) {
	idxReg, _, err := g.emitExpr(indexExpr)
	if err != nil {
		return "", "", err
	}
	zeroIdx := g.tmp()
	if lowerBound != 0 {
		g.line(fmt.Sprintf("  %s = sub i64 %s, %d", zeroIdx, idxReg, lowerBound))
	} else {
		g.line(fmt.Sprintf("  %s = add i64 %s, 0", zeroIdx, idxReg))
	}
	ptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [%d x %s], ptr %s, i64 0, i64 %s",
		ptr, size, elemType, baseReg, zeroIdx))
	if asLValue {
		return ptr, elemType, nil
	}
	loaded := g.tmp()
	g.line(fmt.Sprintf("  %s = load %s, ptr %s", loaded, elemType, ptr))
	return loaded, elemType, nil
}

// emitArrayLength returns the length of an array.
//
//	Length(arr) for static → constant N
//	Length(arr) for dynamic → load len from struct
func (g *Generator) emitArrayLength(arr ast.Expression) (string, string, error) {
	ident, ok := arr.(*ast.Identifier)
	if !ok {
		return "", "", fmt.Errorf("Length() target must be an identifier")
	}
	info, hasInfo := g.arrayInfo[ident.Value]
	if !hasInfo {
		return "", "", fmt.Errorf("Length() target is not an array: %s", ident.Value)
	}
	allocaReg := g.locals[ident.Value]

	if info.IsDynamic {
		lenLoc := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, i64, i64 }, ptr %s, i32 0, i32 1",
			lenLoc, allocaReg))
		lenVal := g.tmp()
		g.line(fmt.Sprintf("  %s = load i64, ptr %s", lenVal, lenLoc))
		return lenVal, "i64", nil
	}

	r := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 0, %d ; static array length", r, info.Size))
	return r, "i64", nil
}

// arrayHelpersUsed indicates whether the program triggered any array codegen,
// so the IR header can include forward declarations of malloc/realloc/free.
func (g *Generator) arrayHelpersUsed() bool {
	return len(g.arrayInfo) > 0
}

// Force unused import to satisfy gofmt during partial development.
var _ = strings.HasPrefix

// evalConstInt evaluates a compile-time integer expression (literals and
// integer arithmetic) for use in array size resolution. Pascal arrays use
// range syntax 'array[1..N]' which the parser desugars to ((N-1)+1).
func evalConstInt(e ast.Expression) int64 {
	switch n := e.(type) {
	case *ast.IntegerLiteral:
		return n.Value
	case *ast.PrefixExpression:
		v := evalConstInt(n.Right)
		if n.Operator == "-" {
			return -v
		}
		return v
	case *ast.InfixExpression:
		l := evalConstInt(n.Left)
		r := evalConstInt(n.Right)
		switch n.Operator {
		case "+":
			return l + r
		case "-":
			return l - r
		case "*":
			return l * r
		case "/", "div":
			if r == 0 {
				return 0
			}
			return l / r
		case "mod":
			if r == 0 {
				return 0
			}
			return l % r
		}
	}
	return 0
}

// indexElementKylixType returns the Kylix type name of an array-index
// expression's element (e.g. rec.Fields[i] → "TVarDecl"), or "" if unknown.
// Used by emitAssign to type-infer `field := rec.Fields[i]` locals so a later
// `field.Names[j]` resolves the receiver class. v5.4.0.
func (g *Generator) indexElementKylixType(idx *ast.IndexExpression) string {
	if idx == nil {
		return ""
	}
	// Local/param slice: arrayInfo carries ElementKylixType.
	if left, ok := idx.Left.(*ast.Identifier); ok {
		if info, ok := g.arrayInfo[left.Value]; ok && info.ElementKylixType != "" {
			return info.ElementKylixType
		}
	}
	// Class field array: obj.Field[i] — look up the field's array element type.
	if member, ok := idx.Left.(*ast.MemberExpression); ok {
		kind, typeName := g.receiverKind(member.Object)
		if kind == "class" {
			if info, ok := g.classes[typeName]; ok {
				for _, f := range info.Fields {
					if f.Name == member.Member && f.ArrayType != nil && f.ArrayType.ElementType != nil {
						return typeExprName(f.ArrayType.ElementType)
					}
				}
			}
		}
	}
	return ""
}

// callReturnKylixType returns the Kylix type name returned by a call, or "" if
// unknown. Used by exprKylixType for `x := func()`/`x := obj.Method()`/`x :=
// TFoo.Create()` type-inferred locals. v5.4.0.
func (g *Generator) callReturnKylixType(call *ast.CallExpression) string {
	if call == nil {
		return ""
	}
	// MemberExpression func: obj.Method() or TFoo.Create() constructor.
	if member, ok := call.Function.(*ast.MemberExpression); ok {
		// Constructor: TFoo.Create() — Object is a class name.
		if ident, ok := member.Object.(*ast.Identifier); ok {
			if _, isClass := g.classes[ident.Value]; isClass {
				return ident.Value
			}
		}
		// Method: obj.Method() — find the method's RetKylixType in hierarchy.
		kind, tn := g.receiverKind(member.Object)
		if kind == "class" {
			if _, meth := g.findMethodInHierarchy(tn, member.Member); meth != nil {
				return meth.RetKylixType
			}
		}
		return ""
	}
	// Identifier func: top-level function (e.g. NewLexer) — funcSigs.
	if ident, ok := call.Function.(*ast.Identifier); ok {
		if sig, ok := g.funcSigs[ident.Value]; ok && sig.ReturnType != nil {
			return typeExprName(sig.ReturnType)
		}
	}
	return ""
}

// exprKylixType returns the Kylix type name of an expression, for type-inferring
// `x := <expr>` locals so later field/array/method access resolves the
// receiver class. Covers identifier, array index, `as TClass` cast, function
// call, and field access. v5.4.0.
func (g *Generator) exprKylixType(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		return g.localTypes[e.Value]
	case *ast.IndexExpression:
		return g.indexElementKylixType(e)
	case *ast.TypeCastExpression:
		// `x as TFoo` → TFoo (runtime check is emitAsExpr's job; the static
		// type for inference is the cast target).
		if e.TargetType != nil {
			return typeExprName(e.TargetType)
		}
	case *ast.CallExpression:
		return g.callReturnKylixType(e)
	case *ast.MemberExpression:
		// Constructor pattern: TFoo.Create (no parens, used as a value) → TFoo.
		if e.Member == "Create" {
			if ident, ok := e.Object.(*ast.Identifier); ok {
				if _, isClass := g.classes[ident.Value]; isClass {
					return ident.Value
				}
			}
		}
		// obj.Field → field's declared Kylix type (for `x := obj.Field` where
		// Field is a class-typed field). Recursive on e.Object so chained access
		// (X.Y.Z) resolves through each field's class type. v5.4.0.
		objType := g.exprKylixType(e.Object)
		if objType != "" {
			if info, ok := g.classes[objType]; ok {
				for _, f := range info.Fields {
					if f.Name == e.Member && f.KylixType != "" {
						return f.KylixType
					}
				}
			}
		}
	}
	return ""
}
