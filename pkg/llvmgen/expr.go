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
	case "variant":
		// v5.0.0: a Variant storage slot holds a pointer to a boxed
		// {i32 tag, i64 payload} value. The "variant" pseudo-type (returned
		// by emitExpr for Variant *values*) is handled in emitInfix/WriteLn,
		// not via LLVMType.
		return "ptr"
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
	// v4.6.0: record the source position so IR emitted for this expression
	// carries a !dbg DILocation (per-line stepping). Save/restore so the
	// caller's position is unaffected when this returns — a sub-expression's
	// position only applies to its own instructions, not subsequent ones in
	// the caller's node. Statements set their own position in emitStatement.
	if g.debugInfo {
		savedLine, savedCol := g.dbg.curLine, g.dbg.curCol
		g.setDbgNode(node)
		defer func() {
			g.dbg.curLine, g.dbg.curCol = savedLine, savedCol
		}()
	}
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
		// Lambdas lower to a named function + env struct + closure pair.
		// See lambda.go for the full lowering.
		return g.emitLambda(e)

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
		// String slice s[start:end] — allocate a new buffer, memcpy the
		// substring [start, end) from the base pointer, null-terminate, and
		// return the new ptr. For now only string slicing is supported (array
		// slicing would need to track element size and return a slice struct).
		base, baseType, err := g.emitExpr(e.Left)
		if err != nil {
			return "", "", err
		}
		if baseType != "ptr" {
			// Non-string slice (e.g. array[lo:hi]) not yet implemented.
			return base, baseType, nil
		}

		low, _, err := g.emitExpr(e.Low)
		if err != nil {
			return "", "", err
		}
		high, _, err := g.emitExpr(e.High)
		if err != nil {
			return "", "", err
		}

		// length = high - low
		length := g.tmp()
		g.line(fmt.Sprintf("  %s = sub i64 %s, %s", length, high, low))

		// Allocate length + 1 (for null terminator)
		allocSize := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, 1", allocSize, length))
		buf := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, allocSize))

		// src = base + low
		src := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", src, base, low))

		// memcpy(buf, src, length)
		g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", buf, src, length))

		// Write null terminator: buf[length] = '\0'
		termPtr := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, buf, length))
		g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))

		return buf, "ptr", nil

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
	// Check if it's a constant first — constants are resolved by re-evaluating
	// their literal value expression inline (no storage allocated).
	if constExpr, ok := g.constants[name]; ok {
		return g.emitExpr(constExpr)
	}

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
	} else if strings.HasSuffix(allocaReg, "_map") {
		llvmT = "ptr"
	} else if strings.HasSuffix(allocaReg, "_var") {
		// v5.0.0: Variant local — load the box pointer; downstream
		// emitInfix/emitWriteLn recognize the "variant" pseudo-type.
		llvmT = variantT
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load %s, ptr %s", r, llvmLoadType(llvmT), allocaReg))
	return r, llvmT, nil
}

// llvmLoadType returns the real LLVM IR type to use in a load instruction for
// the given (possibly pseudo) codegen type. The "variant" pseudo-type maps to
// ptr (a box pointer is what's actually stored in the slot).
func llvmLoadType(t string) string {
	if t == variantT {
		return "ptr"
	}
	return t
}

func (g *Generator) emitInfix(e *ast.InfixExpression) (string, string, error) {
	lv, lt, err := g.emitExpr(e.Left)
	if err != nil {
		return "", "", err
	}
	rv, rt, err := g.emitExpr(e.Right)
	if err != nil {
		return "", "", err
	}

	// v5.0.0: Variant-aware comparison. If either operand is a Variant box
	// (pseudo-type "variant") and the operator is relational, box the other
	// operand and dispatch to the runtime comparator (numeric coercion + tag
	// dispatch). Placed BEFORE the string-concat/coerce/numeric paths so a
	// Variant never reaches the ptr/i64 strcmp/icmp paths (which would
	// miscompare the box pointer or type-mismatch).
	if isVariantOperand(lt) || isVariantOperand(rt) {
		if isComparisonOp(e.Operator) {
			return g.emitVariantCompare(e.Operator, lv, lt, rv, rt)
		}
		// v5.1.0: Variant arithmetic (+,-,*,/) dispatches to the runtime
		// (str concat for +, int/double per tag). Returns a Variant box.
		if isArithOp(e.Operator) {
			return g.emitVariantArith(e.Operator, lv, lt, rv, rt)
		}
		// div/mod on Variants still unsupported — safe zero stub.
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; variant op %q unsupported", r, e.Operator))
		return r, "i64", nil
	}

	// String concatenation: `+` on two ptr (string) operands → malloc + strcat.
	// (Pascal `+` on strings concatenates; numeric `+` adds.)
	if e.Operator == "+" && (lt == "ptr" || rt == "ptr") {
		return g.emitStringConcat(lv, rv), "ptr", nil
	}

	// Coerce mixed int/double operands to a common type (double wins) so
	// arithmetic/comparison ops never mix i64 and double operands.
	if lt != rt {
		if lt == "double" && rt == "i64" {
			rv, rt = g.coerceValue(rv, rt, "double")
		} else if lt == "i64" && rt == "double" {
			lv, lt = g.coerceValue(lv, lt, "double")
		}
	}

	// String comparison: `=`/`<>`/`<`/`<=`/`>`/`>=` on ptr (string) operands
	// use strcmp (lexicographic), not icmp (which would compare the pointer
	// addresses, not the string contents).
	//
	// Exception: pointer-vs-nil comparisons (e.g. `if c <> nil`) must use icmp
	// on the raw pointer value, NOT strcmp — calling strcmp against a null
	// pointer would dereference it and segfault. Detect nil by AST node type
	// since by this point the nil literal has already been lowered to a plain
	// ptr register indistinguishable from a string ptr.
	isStringCmp := lt == "ptr" && rt == "ptr" && !isNilNode(e.Left) && !isNilNode(e.Right)
	switch e.Operator {
	case "=", "<>", "<", "<=", ">", ">=":
		if isStringCmp {
			return g.emitStringCompare(e.Operator, lv, rv)
		}
		// Pointer-vs-nil (or any ptr-ptr that isn't a string cmp): use icmp
		// on the raw pointers.
		if lt == "ptr" && rt == "ptr" {
			return g.emitPtrCompare(e.Operator, lv, rv)
		}
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

// isNilNode reports whether e is a `nil` literal AST node. Used to steer
// ptr-ptr comparisons away from strcmp (which would dereference null) toward
// raw pointer icmp.
func isNilNode(e ast.Expression) bool {
	_, ok := e.(*ast.NilLiteral)
	return ok
}

// emitPtrCompare lowers =/<> on two ptr operands as a raw pointer equality
// test (icmp eq/ne ptr). Ordering comparisons (<, <=, >, >=) on pointers are
// also lowered with icmp slt/sle/sgt/sge — well-defined for pointers in LLVM.
func (g *Generator) emitPtrCompare(op, lv, rv string) (string, string, error) {
	r := g.tmp()
	switch op {
	case "=":
		g.line(fmt.Sprintf("  %s = icmp eq ptr %s, %s", r, lv, rv))
	case "<>":
		g.line(fmt.Sprintf("  %s = icmp ne ptr %s, %s", r, lv, rv))
	case "<":
		g.line(fmt.Sprintf("  %s = icmp slt ptr %s, %s", r, lv, rv))
	case "<=":
		g.line(fmt.Sprintf("  %s = icmp sle ptr %s, %s", r, lv, rv))
	case ">":
		g.line(fmt.Sprintf("  %s = icmp sgt ptr %s, %s", r, lv, rv))
	case ">=":
		g.line(fmt.Sprintf("  %s = icmp sge ptr %s, %s", r, lv, rv))
	}
	return r, "i1", nil
}

// emitStringCompare lowers a Pascal comparison operator applied to two String
// (ptr) operands via libc strcmp, which returns <0/0/>0 for lexicographic
// less-than/equal/greater-than. icmp on the raw pointers would compare
// addresses, not string contents, so this must not go through emitInfix's
// normal numeric-comparison path.
func (g *Generator) emitStringCompare(op, lv, rv string) (string, string, error) {
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, lv, rv))
	r := g.tmp()
	switch op {
	case "=":
		g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", r, cmp))
	case "<>":
		g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", r, cmp))
	case "<":
		g.line(fmt.Sprintf("  %s = icmp slt i32 %s, 0", r, cmp))
	case "<=":
		g.line(fmt.Sprintf("  %s = icmp sle i32 %s, 0", r, cmp))
	case ">":
		g.line(fmt.Sprintf("  %s = icmp sgt i32 %s, 0", r, cmp))
	case ">=":
		g.line(fmt.Sprintf("  %s = icmp sge i32 %s, 0", r, cmp))
	}
	return r, "i1", nil
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

// emitStringConcat concatenates two string pointers (ptr operands) into a
// freshly malloc'd buffer via strcpy + strcat, returning the result ptr.
func (g *Generator) emitStringConcat(lv, rv string) string {
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 512)", buf))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %s)", buf, lv))
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", buf, rv))
	return buf
}

// emitCall generates a function call expression.
func (g *Generator) emitCall(e *ast.CallExpression) (string, string, error) {
	// stdlib module function call: `sysutil.ReadFile(path)` — MemberExpression
	// whose Object is an imported stdlib module name. Dispatch to libc-backed
	// IR before the generic method-call path treats `sysutil` as a receiver.
	if member, ok := e.Function.(*ast.MemberExpression); ok {
		if ident, ok := member.Object.(*ast.Identifier); ok && g.isStdlibModule(ident.Value) {
			return g.emitStdlibCall(ident.Value, member.Member, e.Arguments)
		}
		return g.emitMethodCall(member, e.Arguments)
	}

	funcName := ""
	if ident, ok := e.Function.(*ast.Identifier); ok {
		funcName = ident.Value
	}

	// Bare-name stdlib call: `ReadFile(...)` (no `sysutil.` qualifier) resolved
	// to an imported module. User-defined functions (in funcSigs) take
	// precedence, matching the Go backend's resolveStdlibFunc behavior.
	if funcName != "" && g.funcSigs[funcName] == nil {
		if mod, ok := g.resolveStdlibBareCall(funcName); ok {
			return g.emitStdlibCall(mod, funcName, e.Arguments)
		}
	}

	// Closure call: callee is a local variable holding a closure value.
	// Indirect-call through {func_ptr, env_ptr} (see lambda.go).
	if funcName != "" && g.closureLocals[funcName] {
		return g.emitClosureCall(funcName, e.Arguments)
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
		// v5.0.0: Length(arr) for a dynamic/static array must read the array's
		// length (slice len word / static count), not strlen the data pointer.
		// emitArrayLength already exists but was never wired up (dead code).
		if ident, ok := e.Arguments[0].(*ast.Identifier); ok {
			if _, hasInfo := g.arrayInfo[ident.Value]; hasInfo {
				return g.emitArrayLength(e.Arguments[0])
			}
		}
		return g.emitLength(e.Arguments[0])
	}

	// Generic function call
	sig := g.funcSigs[funcName]

	var argRegs []string
	var argTypes []string
	for i, arg := range e.Arguments {
		r, t, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		// Coerce the argument to match the declared parameter type (e.g. an
		// Integer literal passed to a Real parameter needs sitofp).
		if sig != nil && i < len(sig.Parameters) && sig.Parameters[i].Type != nil {
			wantT := LLVMType(typeExprName(sig.Parameters[i].Type))
			if wantT != t {
				r, t = g.coerceValue(r, t, wantT)
			}
		}
		argRegs = append(argRegs, r)
		argTypes = append(argTypes, t)
	}

	retType := "i64"
	if multiTypes := g.multiRetTypes[funcName]; len(multiTypes) > 0 {
		// Multi-return function: returns a struct.
		retType = fmt.Sprintf("%%__ret_%s", funcName)
	} else if sig != nil && sig.ReturnType != nil {
		retType = LLVMType(typeExprName(sig.ReturnType))
	} else if sig != nil {
		retType = "void"
	}

	var argList []string
	for i, r := range argRegs {
		argList = append(argList, argTypes[i]+" "+r)
	}
	if retType == "void" {
		g.line(fmt.Sprintf("  call void @%s(%s)", funcName, strings.Join(argList, ", ")))
		return "0", "void", nil
	}
	result := g.tmp()
	g.line(fmt.Sprintf("  %s = call %s @%s(%s)", result, retType, funcName, strings.Join(argList, ", ")))
	return result, retType, nil
}

// coerceValue converts a value register from one LLVM type to another,
// emitting the appropriate cast instruction. Used when an argument's
// evaluated type doesn't match the declared parameter type (e.g. an integer
// literal passed to a Real parameter).
func (g *Generator) coerceValue(reg, fromT, toT string) (string, string) {
	if fromT == toT {
		return reg, fromT
	}
	// v5.1.0: Variant→concrete unbox needs the Variant runtime.
	if fromT == variantT {
		g.needVariantRuntime = true
	}
	cast := g.tmp()
	switch {
	// v5.1.0: Variant→concrete unbox (dispatches on tag at runtime). Enables
	// `n := v` (Variant→Integer), `s := v`, and call-arg coercion. Without
	// this, a variant value stored into a concrete slot would type-mismatch.
	case fromT == variantT && toT == "i64":
		g.line(fmt.Sprintf("  %s = call i64 @__kylix_variant_as_int(ptr %s)", cast, reg))
		return cast, "i64"
	case fromT == variantT && toT == "double":
		g.line(fmt.Sprintf("  %s = call double @__kylix_variant_as_double(ptr %s)", cast, reg))
		return cast, "double"
	case fromT == variantT && toT == "ptr":
		g.line(fmt.Sprintf("  %s = call ptr @__kylix_variant_as_str(ptr %s)", cast, reg))
		return cast, "ptr"
	case fromT == variantT && toT == "i1":
		g.line(fmt.Sprintf("  %s = call i1 @__kylix_variant_as_bool(ptr %s)", cast, reg))
		return cast, "i1"
	case fromT == "i64" && toT == "double":
		g.line(fmt.Sprintf("  %s = sitofp i64 %s to double", cast, reg))
		return cast, "double"
	case fromT == "double" && toT == "i64":
		g.line(fmt.Sprintf("  %s = fptosi double %s to i64", cast, reg))
		return cast, "i64"
	case fromT == "i1" && toT == "i64":
		g.line(fmt.Sprintf("  %s = zext i1 %s to i64", cast, reg))
		return cast, "i64"
	case fromT == "i64" && toT == "i1":
		g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", cast, reg))
		return cast, "i1"
	default:
		// No known conversion (e.g. ptr↔i64 for class/interface args) — pass
		// through unchanged rather than emitting an invalid cast.
		return reg, fromT
	}
}

// emitWriteLn generates a puts() or printf() call for WriteLn.
func (g *Generator) emitWriteLn(arg ast.Expression) (string, string, error) {
	v, t, err := g.emitExpr(arg)
	if err != nil {
		return "", "", err
	}

	switch t {
	case "variant":
		// v5.0.0: Variant box — dispatch to the runtime println (tag-aware).
		return g.emitVariantPrint(v, true)
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
		fmtReg := g.addString("%.15g\n")
		fmtPtr := g.ptrTo(fmtReg, len("%.15g\n")+1)
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
	case "variant":
		// v5.0.0: Variant box — dispatch to the runtime print (no newline).
		return g.emitVariantPrint(v, false)
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
func (g *Generator) emitWriteLnMulti(args []ast.Expression) (string, string, error) {
	const bufSize = 512
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", buf, bufSize))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", buf))

	ldFmt := g.addString("%ld")
	ldFmtPtr := g.ptrTo(ldFmt, 4)

	for i, arg := range args {
		// Insert a space between consecutive arguments (matching fmt.Println).
		if i > 0 {
			spaceStr := g.addString(" ")
			spacePtr := g.ptrTo(spaceStr, 2)
			g.line(fmt.Sprintf("  %s = call ptr @strcat(ptr %s, ptr %s)", g.tmp(), buf, spacePtr))
		}

		reg, t, err := g.emitExpr(arg)
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
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @puts(ptr noundef %s)", r, buf))
	return "0", "void", nil
}

// isTDateTimeReceiver checks if the expression is a TDateTime instance variable.
func (g *Generator) isTDateTimeReceiver(obj ast.Expression) bool {
	ident, ok := obj.(*ast.Identifier)
	if !ok {
		return false
	}
	t, ok := g.localTypes[ident.Value]
	return ok && t == "TDateTime"
}
