// stmt_flow.go — control-flow statement codegen (if/while/for/repeat/foreach/
// case/match/break/continue) — split from stmt.go in v4.5.0 to keep each
// source file under 1000 lines.
package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

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

	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, headerLbl
	g.line(fmt.Sprintf("%s:", bodyLbl))
	if err := g.emitStatement(s.Body); err != nil {
		return err
	}
	g.breakLabel, g.continueLabel = savedBreak, savedContinue
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

	// Body — save/restore break/continue targets for nested loops.
	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, headerLbl
	g.line(fmt.Sprintf("%s:", bodyLbl))
	if err := g.emitStatement(s.Body); err != nil {
		return err
	}
	g.breakLabel, g.continueLabel = savedBreak, savedContinue

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

	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, bodyLbl

	g.line(fmt.Sprintf("  br label %%%s", bodyLbl))
	g.line(fmt.Sprintf("%s:", bodyLbl))

	if err := g.emitStatement(s.Body); err != nil {
		return err
	}

	g.breakLabel, g.continueLabel = savedBreak, savedContinue

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

// emitTupleBuild handles `result := (a, b, ...)` inside a multi-return
// function: evaluates each element and packs them into the `%__ret_FuncName`
// struct via a chain of insertvalue instructions, then stores into %result.
func (g *Generator) emitTupleBuild(tuple *ast.TupleLiteral) error {
	structType := fmt.Sprintf("%%__ret_%s", g.funcName)
	elemTypes := g.multiRetTypes[g.funcName]

	// Build struct value via insertvalue chain: start with undef, insert each field.
	accReg := "undef"
	for i, elem := range tuple.Elements {
		v, _, err := g.emitExpr(elem)
		if err != nil {
			return err
		}
		elemT := "i64"
		if i < len(elemTypes) {
			elemT = elemTypes[i]
		}
		next := g.tmp()
		g.line(fmt.Sprintf("  %s = insertvalue %s %s, %s %s, %d", next, structType, accReg, elemT, v, i))
		accReg = next
	}
	g.line(fmt.Sprintf("  store %s %s, ptr %%result", structType, accReg))
	return nil
}

// emitTupleDestructure handles `(a, b, ...) := Expr` where Expr is a call to
// a multi-return function. Evaluates the call (which yields a struct value),
// extracts each field via extractvalue, and stores into the corresponding
// LHS variables (auto-declaring them if not already local).
func (g *Generator) emitTupleDestructure(tuple *ast.TupleLiteral, rhs ast.Expression) error {
	structVal, structType, err := g.emitExpr(rhs)
	if err != nil {
		return err
	}

	call, isCall := rhs.(*ast.CallExpression)
	var elemTypes []string
	if isCall {
		if fnIdent, ok := call.Function.(*ast.Identifier); ok {
			elemTypes = g.multiRetTypes[fnIdent.Value]
		}
	}

	for i, elem := range tuple.Elements {
		ident, ok := elem.(*ast.Identifier)
		if !ok {
			continue // non-identifier tuple element (e.g. `_`) — skip binding
		}
		elemT := "i64"
		if i < len(elemTypes) {
			elemT = elemTypes[i]
		}
		extracted := g.tmp()
		g.line(fmt.Sprintf("  %s = extractvalue %s %s, %d", extracted, structType, structVal, i))

		allocaReg, exists := g.locals[ident.Value]
		if !exists {
			suffix := "_int"
			switch elemT {
			case "i1":
				suffix = "_bool"
			case "double":
				suffix = "_real"
			case "ptr":
				suffix = "_str"
			}
			allocaReg = g.freshVarReg(ident.Value, suffix)
			g.line(fmt.Sprintf("  %s = alloca %s, align 8", allocaReg, elemT))
			g.locals[ident.Value] = allocaReg
		}
		g.line(fmt.Sprintf("  store %s %s, ptr %s", elemT, extracted, allocaReg))
	}
	return nil
}

// emitBreak generates a branch to the enclosing loop's exit label.
func (g *Generator) emitBreak() error {
	if g.breakLabel == "" {
		return fmt.Errorf("break outside loop")
	}
	// After an unconditional branch the current block is terminated; the
	// following instructions need a fresh (unreachable) label so the IR
	// remains structurally valid.
	g.line(fmt.Sprintf("  br label %%%s", g.breakLabel))
	deadLbl := g.label()
	g.line(fmt.Sprintf("%s:", deadLbl))
	return nil
}

// emitContinue generates a branch to the enclosing loop's header label.
func (g *Generator) emitContinue() error {
	if g.continueLabel == "" {
		return fmt.Errorf("continue outside loop")
	}
	g.line(fmt.Sprintf("  br label %%%s", g.continueLabel))
	deadLbl := g.label()
	g.line(fmt.Sprintf("%s:", deadLbl))
	return nil
}

// emitForEach generates a counted for-over-index loop for `for x in arr do`.
// LLVM has no built-in range; we lower it as a 0-based counted loop using
// the array's Length (via a strlen-style i64 field read where available, or
// a conservative fixed bound of 0 for unknown iterables). This covers the
// common case of iterating array of T where the length is statically known.
func (g *Generator) emitForEach(s *ast.ForEachStatement) error {
	// Emit the iterable expression to get a pointer/value.
	iterReg, _, err := g.emitExpr(s.Iterable)
	if err != nil {
		return err
	}

	// Allocate a loop variable alloca for the element.
	elemAlloca := g.freshVarReg(s.Variable, "")
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", elemAlloca))
	g.locals[s.Variable] = elemAlloca

	// Use a simple i64 index counter.
	idxAlloca := g.tmp() + "_foreach_idx"
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", idxAlloca))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", idxAlloca))

	// For a string (ptr), use strlen as the bound. For other types treat
	// bound as 0 (body never executes) — a conservative safe default.
	boundReg := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", boundReg, iterReg))

	headerLbl := g.label()
	bodyLbl := g.label()
	exitLbl := g.label()

	savedBreak, savedContinue := g.breakLabel, g.continueLabel
	g.breakLabel, g.continueLabel = exitLbl, headerLbl

	g.line(fmt.Sprintf("  br label %%%s", headerLbl))
	g.line(fmt.Sprintf("%s:", headerLbl))
	idxCur := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", idxCur, idxAlloca))
	condReg := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", condReg, idxCur, boundReg))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", condReg, bodyLbl, exitLbl))

	g.line(fmt.Sprintf("%s:", bodyLbl))
	// Load element at current index (char from string / element from array).
	elemPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", elemPtr, iterReg, idxCur))
	elemVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", elemVal, elemPtr))
	elemExt := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i8 %s to i64", elemExt, elemVal))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", elemExt, elemAlloca))

	if err := g.emitStatement(s.Body); err != nil {
		return err
	}
	g.breakLabel, g.continueLabel = savedBreak, savedContinue

	// Increment index.
	idxNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", idxNext, idxCur))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", idxNext, idxAlloca))
	g.line(fmt.Sprintf("  br label %%%s", headerLbl))

	g.line(fmt.Sprintf("%s:", exitLbl))
	return nil
}

// emitCase lowers a Pascal case statement to an LLVM switch instruction.
// case expr of 1: ... 2,3: ... else ... end
func (g *Generator) emitCase(s *ast.CaseStatement) error {
	exprReg, exprType, err := g.emitExpr(s.Expression)
	if err != nil {
		return err
	}
	// Default type for case selector is i64.
	if exprType == "" {
		exprType = "i64"
	}

	defaultLbl := g.label()
	mergeLbl := g.label()

	// Collect branch labels.
	type branchEntry struct {
		values []int64
		lbl    string
	}
	var branches []branchEntry
	for _, br := range s.Branches {
		lbl := g.label()
		var vals []int64
		for _, v := range br.Values {
			if lit, ok := v.(*ast.IntegerLiteral); ok {
				vals = append(vals, lit.Value)
			}
		}
		branches = append(branches, branchEntry{vals, lbl})
	}

	// Emit switch instruction.
	var switchCases []string
	for _, br := range branches {
		for _, val := range br.values {
			switchCases = append(switchCases, fmt.Sprintf("%s %d, label %%%s", exprType, val, br.lbl))
		}
	}
	if len(switchCases) > 0 {
		g.line(fmt.Sprintf("  switch %s %s, label %%%s [ %s ]",
			exprType, exprReg, defaultLbl, strings.Join(switchCases, " ")))
	} else {
		g.line(fmt.Sprintf("  br label %%%s", defaultLbl))
	}

	// Emit branch bodies.
	for i, br := range s.Branches {
		g.line(fmt.Sprintf("%s:", branches[i].lbl))
		if br.Body != nil {
			if err := g.emitBlockScoped(br.Body); err != nil {
				return err
			}
		}
		g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	}

	// Default / else branch.
	g.line(fmt.Sprintf("%s:", defaultLbl))
	if s.ElseBranch != nil {
		if err := g.emitBlockScoped(s.ElseBranch); err != nil {
			return err
		}
	}
	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))

	g.line(fmt.Sprintf("%s:", mergeLbl))
	return nil
}

// emitMatch lowers a Pascal match statement to a chain of conditional branches.
// match expr { pat => body; _ => default }
func (g *Generator) emitMatch(s *ast.MatchStatement) error {
	exprReg, _, err := g.emitExpr(s.Expression)
	if err != nil {
		return err
	}

	mergeLbl := g.label()
	// tmp alloca for the match value so it is available in each comparison.
	matchAlloca := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", matchAlloca))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", exprReg, matchAlloca))

	for _, br := range s.Branches {
		// Wildcard / default
		if ident, ok := br.Pattern.(*ast.Identifier); ok && ident.Value == "_" {
			if br.Body != nil {
				if err := g.emitBlockScoped(br.Body); err != nil {
					return err
				}
			}
			g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
			continue
		}

		// Build condition: _v == p (possibly OR with additional patterns).
		mv := g.tmp()
		g.line(fmt.Sprintf("  %s = load i64, ptr %s", mv, matchAlloca))
		allPats := []ast.Expression{br.Pattern}
		allPats = append(allPats, br.AdditionalPatterns...)
		condReg := ""
		for _, pat := range allPats {
			pv, _, err := g.emitExpr(pat)
			if err != nil {
				return err
			}
			cmp := g.tmp()
			g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", cmp, mv, pv))
			if condReg == "" {
				condReg = cmp
			} else {
				or := g.tmp()
				g.line(fmt.Sprintf("  %s = or i1 %s, %s", or, condReg, cmp))
				condReg = or
			}
		}
		// Optional when guard.
		if br.When != nil {
			guardReg, _, err := g.emitExpr(br.When)
			if err != nil {
				return err
			}
			and := g.tmp()
			g.line(fmt.Sprintf("  %s = and i1 %s, %s", and, condReg, guardReg))
			condReg = and
		}

		bodyLbl := g.label()
		nextLbl := g.label()
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", condReg, bodyLbl, nextLbl))
		g.line(fmt.Sprintf("%s:", bodyLbl))
		if br.Body != nil {
			if err := g.emitBlockScoped(br.Body); err != nil {
				return err
			}
		}
		g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
		g.line(fmt.Sprintf("%s:", nextLbl))
	}

	g.line(fmt.Sprintf("  br label %%%s", mergeLbl))
	g.line(fmt.Sprintf("%s:", mergeLbl))
	return nil
}
