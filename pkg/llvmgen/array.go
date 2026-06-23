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
	Size        int64  // for static arrays only
	LowerBound  int64  // Pascal range lower bound (default 1)
}

// emitArrayVarDecl allocates an array variable.
// Returns true if the type was an array; false otherwise.
func (g *Generator) emitArrayVarDecl(name string, arr *ast.ArrayType) bool {
	elemType := "i64"
	if arr.ElementType != nil {
		elemType = LLVMType(typeExprName(arr.ElementType))
	}

	if arr.Dynamic {
		// Dynamic array: { ptr data; i64 len; i64 cap }
		// For Milestone 2, represent as a stack-allocated struct of 3 words.
		allocaReg := fmt.Sprintf("%%v_%s_dyn", name)
		g.line(fmt.Sprintf("  %s = alloca { ptr, i64, i64 }, align 8", allocaReg))
		// Zero-init: data=null, len=0, cap=0
		g.line(fmt.Sprintf("  store { ptr, i64, i64 } zeroinitializer, ptr %s", allocaReg))
		g.locals[name] = allocaReg
		g.arrayInfo[name] = &arrayInfo{IsDynamic: true, ElementType: elemType}
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

	allocaReg := fmt.Sprintf("%%v_%s_arr", name)
	g.line(fmt.Sprintf("  %s = alloca [%d x %s], align 8", allocaReg, size, elemType))
	// Zero-init the whole array
	g.line(fmt.Sprintf("  store [%d x %s] zeroinitializer, ptr %s", size, elemType, allocaReg))

	g.locals[name] = allocaReg
	g.arrayInfo[name] = &arrayInfo{
		IsDynamic:   false,
		ElementType: elemType,
		Size:        size,
		LowerBound:  1, // Pascal default
	}
	return true
}

// emitArrayIndex generates GEP for array[index] access.
// Returns (resultReg, elementType, error). For assignment context, returns
// the pointer register (use g.line to emit a store yourself).
func (g *Generator) emitArrayIndex(idx *ast.IndexExpression, asLValue bool) (string, string, error) {
	// Resolve the array variable
	leftIdent, ok := idx.Left.(*ast.Identifier)
	if !ok {
		return "", "", fmt.Errorf("array index target must be an identifier")
	}
	allocaReg, ok := g.locals[leftIdent.Value]
	if !ok {
		return "", "", fmt.Errorf("undefined array variable: %s", leftIdent.Value)
	}
	info, hasInfo := g.arrayInfo[leftIdent.Value]
	if !hasInfo {
		return "", "", fmt.Errorf("variable %s is not an array", leftIdent.Value)
	}

	// Compute index (Pascal 1-based → LLVM 0-based for static arrays)
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
		return ptr, info.ElementType, nil
	}
	loaded := g.tmp()
	g.line(fmt.Sprintf("  %s = load %s, ptr %s", loaded, info.ElementType, ptr))
	return loaded, info.ElementType, nil
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
