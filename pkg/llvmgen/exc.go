// exc.go — LLVM IR generation for exception handling (M3).
//
// Implements try/except/finally + raise via a global exception slot and
// setjmp/longjmp (route C in the design doc). The Itanium C++ EH ABI was
// rejected because it requires libc++abi linkage and is infeasible to emit by
// hand-written IR text; setjmp/longjmp keep all IR as call/load/store/br/icmp.
//
// Runtime model:
//
//	@__kylix_exc_obj    global ptr   — the raised object (a class instance ptr)
//	@__kylix_exc_type   global i32   — type ID of the raised object
//	@__kylix_exc_active global i1    — an exception is in flight (uncaught)
//	@__kylix_jmpbuf     global ptr   — current handler's setjmp buffer
//
//	raise writes the slot then longjmps to @__kylix_jmpbuf. try installs a
//	new handler (saving the old one for nesting), and on longjmp return reads
//	the type ID and matches it against on-clauses via @__kylix_is_subtype.
package llvmgen

import (
	"fmt"
	"sort"
	"strings"

	"kylix/ast"
)

// excJmpBufSize is the alloca size for a setjmp buffer. arm64 macOS needs ~272
// bytes; 288 aligned to 16 is a safe conservative bound. Over-allocating stack
// is harmless.
const excJmpBufSize = 288

// raiseExceptionTypeName extracts the exception class name from a raise
// expression. Handles `T.Create(...)` (constructor, both no-arg MemberExpression
// and arg-bearing CallExpression forms) and a bare class instance variable.
// Returns "" if the type cannot be determined (→ generic ID 0).
func raiseExceptionTypeName(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	// Unwrap call: raise T.Create('msg') → CallExpression{Function: MemberExpression}.
	if call, ok := expr.(*ast.CallExpression); ok {
		return raiseExceptionTypeName(call.Function)
	}
	if m, ok := expr.(*ast.MemberExpression); ok && m.Member == "Create" {
		if ident, ok := m.Object.(*ast.Identifier); ok {
			return ident.Value
		}
	}
	return typeExprName(expr)
}

// emitRaise generates IR for `raise <expr>` or bare `raise`.
//
//	raise Exc.Create('msg')  →  store obj+type into the global slot, longjmp
//	raise                     →  re-throw the in-flight exception (longjmp outer)
func (g *Generator) emitRaise(s *ast.RaiseStatement) error {
	if s.Exception == nil {
		// Bare raise: only valid inside an except handler. If we're not in one,
		// fall back to raising a generic Exception (matches Go backend behavior).
		if !g.inExceptHandler {
			return g.emitRaiseGeneric()
		}
		// Re-throw: the global slot still holds the current exception. Keep
		// exc_active=true and longjmp to the outer handler.
		return g.emitLongjmpToHandler()
	}

	// Evaluate the exception expression → object pointer.
	objReg, _, err := g.emitExpr(s.Exception)
	if err != nil {
		return err
	}
	typeName := raiseExceptionTypeName(s.Exception)
	tid := g.excTypeID(typeName)

	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_exc_obj", objReg))
	g.line(fmt.Sprintf("  store i32 %d, ptr @__kylix_exc_type", tid))
	g.line("  store i1 true, ptr @__kylix_exc_active")
	return g.emitLongjmpToHandler()
}

// emitRaiseGeneric raises a synthetic Exception with a default message, used
// when bare `raise` appears outside any except handler.
func (g *Generator) emitRaiseGeneric() error {
	msg := g.addString("exception")
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_exc_obj", msg))
	g.line("  store i32 1, ptr @__kylix_exc_type") // Exception = ID 1
	g.line("  store i1 true, ptr @__kylix_exc_active")
	return g.emitLongjmpToHandler()
}

// emitLongjmpToHandler loads the current handler's jmpbuf and longjmps to it.
// If no handler is installed (jmpbuf is null), the program exits with status 70
// (EX_SOFTWARE) — an uncaught exception.
func (g *Generator) emitLongjmpToHandler() error {
	jb := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_jmpbuf", jb))
	nz := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", nz, jb))
	hasLbl := g.label()
	noLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nz, hasLbl, noLbl))

	g.line(fmt.Sprintf("%s:", hasLbl))
	g.line(fmt.Sprintf("  call void @longjmp(ptr %s, i32 1)", jb))
	g.line("  unreachable")

	g.line(fmt.Sprintf("%s:", noLbl))
	g.line("  call void @exit(i32 70)")
	g.line("  unreachable")
	return nil
}

// emitTry generates IR for try/except/finally.
//
// Control-flow shape:
//
//	setjmp → try_body (install handler, run body, pop, → finally_normal)
//	       ↘ except_dispatch (pop, match on-clauses by type ID, → finally_exc
//	                          or finally_reraise if uncaught)
//
//	finally is emitted up to three times (normal/exc/reraise) so it always
//	runs — longjmp would otherwise skip cleanup. Nesting is supported by
//	saving/restoring @__kylix_jmpbuf (a stack of handlers via an alloca slot).
func (g *Generator) emitTry(s *ast.TryStatement) error {
	// alloca for the setjmp buffer and for saving the outer handler pointer.
	bufReg := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [%d x i8], align 16", bufReg, excJmpBufSize))
	bufptr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr [%d x i8], ptr %s, i64 0, i64 0", bufptr, excJmpBufSize, bufReg))
	oldJBSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", oldJBSlot))

	// setjmp: returns 0 on first call, non-zero when longjmp returns here.
	rc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @setjmp(ptr %s)", rc, bufptr))
	isHandler := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", isHandler, rc))
	tryBodyLbl := g.label()
	exceptLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isHandler, exceptLbl, tryBodyLbl))

	// ── try body ──────────────────────────────────────────────
	g.line(fmt.Sprintf("%s:", tryBodyLbl))
	oldJB := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_jmpbuf", oldJB))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", oldJB, oldJBSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_jmpbuf", bufptr))
	g.line("  store i1 false, ptr @__kylix_exc_active")
	if s.Body != nil {
		for _, st := range s.Body.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	// Pop handler, clear active, fall through to finally (normal path).
	restoredJB := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", restoredJB, oldJBSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_jmpbuf", restoredJB))
	g.line("  store i1 false, ptr @__kylix_exc_active")

	finallyNormalLbl := g.label()
	finallyExcLbl := g.label()
	finallyReraiseLbl := g.label()
	endLbl := g.label()

	g.line(fmt.Sprintf("  br label %%%s", finallyNormalLbl))

	// ── except dispatch ───────────────────────────────────────
	g.line(fmt.Sprintf("%s:", exceptLbl))
	// Pop the handler installed by try_body (restore outer).
	restoredJB2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", restoredJB2, oldJBSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_jmpbuf", restoredJB2))

	tid := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr @__kylix_exc_type", tid))

	// Match on-clauses in order. Each emits a subtype check and branches to
	// its body or the next check.
	nextCheck := exceptLbl
	matched := false
	for _, on := range s.OnClauses {
		onBodyLbl := g.label()
		thisCheck := nextCheck
		nextCheck = g.label()
		wantID := g.excTypeID(typeExprName(on.Type))

		m := g.tmp()
		g.line(fmt.Sprintf("  %s = call i1 @__kylix_is_subtype(i32 %s, i32 %d)", m, tid, wantID))
		// The check is emitted under the current "thisCheck" label (the first
		// one reuses exceptLbl which we already emitted above).
		if thisCheck != exceptLbl {
			g.line(fmt.Sprintf("%s:", thisCheck))
		}
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", m, onBodyLbl, nextCheck))

		// on-body: bind E to the exception object, run body, clear active.
		g.line(fmt.Sprintf("%s:", onBodyLbl))
		if on.Variable != "" {
			eAlloca := fmt.Sprintf("%%v_%s", on.Variable)
			g.line(fmt.Sprintf("  %s = alloca ptr, align 8", eAlloca))
			obj := g.tmp()
			g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_exc_obj", obj))
			g.line(fmt.Sprintf("  store ptr %s, ptr %s", obj, eAlloca))
			g.locals[on.Variable] = eAlloca
			if t := typeExprName(on.Type); t != "" {
				g.localTypes[on.Variable] = t
			}
		}
		g.inExceptHandler = true
		if on.Body != nil {
			for _, st := range on.Body.Statements {
				if err := g.emitStatement(st); err != nil {
					g.inExceptHandler = false
					return err
				}
			}
		}
		g.inExceptHandler = false
		g.line("  store i1 false, ptr @__kylix_exc_active")
		g.line(fmt.Sprintf("  br label %%%s", finallyExcLbl))
		matched = true
	}

	// After the last on-clause check, emit the fall-through label.
	if len(s.OnClauses) > 0 {
		g.line(fmt.Sprintf("%s:", nextCheck))
	}

	// No on-clause matched (or none present):
	//   - a plain ExceptBlock handles everything → run it, → finally_exc
	//   - otherwise the exception stays active → finally_reraise
	if s.ExceptBlock != nil {
		g.inExceptHandler = true
		for _, st := range s.ExceptBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				g.inExceptHandler = false
				return err
			}
		}
		g.inExceptHandler = false
		g.line("  store i1 false, ptr @__kylix_exc_active")
		g.line(fmt.Sprintf("  br label %%%s", finallyExcLbl))
	} else {
		// Uncaught: keep exc_active=true so the reraise path re-throws.
		g.line(fmt.Sprintf("  br label %%%s", finallyReraiseLbl))
	}
	_ = matched

	// ── finally: normal path (try body completed) ─────────────
	g.line(fmt.Sprintf("%s:", finallyNormalLbl))
	if s.FinallyBlock != nil {
		for _, st := range s.FinallyBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	g.line(fmt.Sprintf("  br label %%%s", endLbl))

	// ── finally: except handled path ──────────────────────────
	g.line(fmt.Sprintf("%s:", finallyExcLbl))
	if s.FinallyBlock != nil {
		for _, st := range s.FinallyBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	g.line(fmt.Sprintf("  br label %%%s", endLbl))

	// ── finally: reraise path (uncaught exception) ────────────
	g.line(fmt.Sprintf("%s:", finallyReraiseLbl))
	if s.FinallyBlock != nil {
		for _, st := range s.FinallyBlock.Statements {
			if err := g.emitStatement(st); err != nil {
				return err
			}
		}
	}
	// Re-throw: longjmp to the outer handler (current @__kylix_jmpbuf).
	outerJB := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_jmpbuf", outerJB))
	g.line(fmt.Sprintf("  call void @longjmp(ptr %s, i32 1)", outerJB))
	g.line("  unreachable")

	g.line(fmt.Sprintf("%s:", endLbl))
	return nil
}

// injectExceptionClass emits a built-in Exception class so that user code can
// reference Exception / E.Message without the Go stdlib type being present in
// the LLVM backend. It builds a ClassDecl and routes it through the normal
// emitClassDecl path (struct type + registration in g.classes).
func (g *Generator) injectExceptionClass() {
	if g.exceptionInjected {
		return
	}
	g.exceptionInjected = true
	if _, exists := g.classes["Exception"]; exists {
		return // user already declared an Exception class
	}
	msgType := &ast.Identifier{Value: "String"}
	decl := &ast.ClassDecl{
		Name:   "Exception",
		Fields: []*ast.VarDecl{{Names: []string{"Message"}, Type: msgType}},
	}
	// emitClassDecl registers the class and emits its struct type. Errors are
	// not returned by emitClassDecl; a failure here would surface as missing
	// struct type downstream, which tests catch.
	_ = g.emitClassDecl(decl)
}

// collectExceptionTypes assigns a runtime type ID to every exception class
// (Exception and its subclasses). Exception itself is ID 1; subclasses get
// increasing IDs from g.nextExcTypeID. Called after all class decls are emitted.
func (g *Generator) collectExceptionTypes() {
	g.exceptionTypeIDs["Exception"] = 1
	// Deterministic order across runs: iterate sorted class names.
	names := make([]string, 0, len(g.classes))
	for n := range g.classes {
		if n == "Exception" {
			continue
		}
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		if g.isExceptionSubclass(name) {
			g.exceptionTypeIDs[name] = g.nextExcTypeID
			g.nextExcTypeID++
		}
	}
}

// isExceptionSubclass reports whether name's parent chain reaches "Exception".
// Exception itself returns false (it is the root, handled separately).
func (g *Generator) isExceptionSubclass(name string) bool {
	visited := map[string]bool{}
	current := name
	for current != "" && current != "Exception" {
		if visited[current] {
			return false // cycle guard
		}
		visited[current] = true
		info, ok := g.classes[current]
		if !ok {
			return false
		}
		current = info.Parent
	}
	return current == "Exception" && name != "Exception"
}

// excTypeID returns the runtime type ID for an exception class name, or 0 if
// the name is not a known exception class (0 = generic/untyped exception).
func (g *Generator) excTypeID(name string) int {
	if id, ok := g.exceptionTypeIDs[name]; ok {
		return id
	}
	return 0
}

// isExceptionClass reports whether name is part of the Exception hierarchy
// (Exception itself or a subclass). Used to validate on-clause types.
func (g *Generator) isExceptionClass(name string) bool {
	if name == "Exception" {
		return true
	}
	return g.isExceptionSubclass(name)
}

// emitExceptionGlobals emits the four module-level globals that form the
// exception slot.
func (g *Generator) emitExceptionGlobals() {
	g.line("; ===== Exception handling globals =====")
	g.line("@__kylix_exc_obj = global ptr null")
	g.line("@__kylix_exc_type = global i32 0")
	g.line("@__kylix_exc_active = global i1 false")
	g.line("@__kylix_jmpbuf = global ptr null")
	g.line("")
}

// emitExceptionRuntime emits the subtype table and the @__kylix_is_subtype
// helper used by on-clause matching. The table is a [N x { i32, i32 }] array
// of (childID, parentID) edges; the helper walks the child→parent chain.
func (g *Generator) emitExceptionRuntime() {
	g.line("; ===== Exception subtype table =====")

	// Build edge list: for each exception class with a known parent in the
	// hierarchy, emit (child, parent). Exception (ID 1) has parent 0 (none).
	type edge struct{ child, parent int }
	var edges []edge
	for name, id := range g.exceptionTypeIDs {
		if name == "Exception" {
			continue
		}
		info, ok := g.classes[name]
		if !ok || info.Parent == "" {
			continue
		}
		parentID, ok := g.exceptionTypeIDs[info.Parent]
		if !ok {
			continue
		}
		edges = append(edges, edge{id, parentID})
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].child != edges[j].child {
			return edges[i].child < edges[j].child
		}
		return edges[i].parent < edges[j].parent
	})

	// struct type for one edge: { i32 child, i32 parent }
	g.line("%__kylix_edge = type { i32, i32 }")
	if len(edges) == 0 {
		g.line("@__kylix_exctab = constant [0 x %__kylix_edge] zeroinitializer")
	} else {
		var parts []string
		for _, e := range edges {
			parts = append(parts, fmt.Sprintf("i32 %d, i32 %d", e.child, e.parent))
		}
		g.line(fmt.Sprintf("@__kylix_exctab = constant [%d x %%__kylix_edge] [ %s ]",
			len(edges), joinEdgeParts(parts)))
	}
	g.line("")

	// define i1 @__kylix_is_subtype(i32 %child, i32 %parent)
	// Walks the child→parent chain by repeatedly scanning the edge table for
	// the current node's parent, until it reaches %parent (return true) or a
	// node with no outgoing edge (return false). Cycles are bounded by the
	// table length (each iteration consumes one table slot at most).
	n := len(edges)
	g.line(fmt.Sprintf(`define i1 @__kylix_is_subtype(i32 %%child, i32 %%parent) {
entry:
  %%eq = icmp eq i32 %%child, %%parent
  br i1 %%eq, label %%ret_true, label %%loop
loop:
  %%c = phi i32 [ %%child, %%entry ], [ %%c_next, %%loop_next ]
  %%i = phi i64 [ 0, %%entry ], [ %%i_next, %%loop_next ]
  %%oob = icmp eq i64 %%i, %d
  br i1 %%oob, label %%ret_false, label %%body
body:
  %%slot = getelementptr inbounds [%d x %%__kylix_edge], ptr @__kylix_exctab, i64 0, i64 %%i
  %%cid_ptr = getelementptr inbounds %%__kylix_edge, ptr %%slot, i64 0, i32 0
  %%cid = load i32, ptr %%cid_ptr
  %%iscur = icmp eq i32 %%cid, %%c
  br i1 %%iscur, label %%found, label %%loop_next
found:
  %%pid_ptr = getelementptr inbounds %%__kylix_edge, ptr %%slot, i64 0, i32 1
  %%par = load i32, ptr %%pid_ptr
  %%match = icmp eq i32 %%par, %%parent
  br i1 %%match, label %%ret_true, label %%update
update:
  br label %%loop_next
loop_next:
  %%c_next = phi i32 [ %%c, %%body ], [ %%par, %%update ]
  %%i_next = add i64 %%i, 1
  br label %%loop
ret_true:
  ret i1 true
ret_false:
  ret i1 false
}
`, n, n))
}

// joinEdgeParts wraps each "i32 a, i32 b" part in { } so the array literal is
// well-formed: [ { i32 2, i32 1 }, { i32 3, i32 1 } ].
func joinEdgeParts(parts []string) string {
	wrapped := make([]string, len(parts))
	for i, p := range parts {
		wrapped[i] = "{ " + p + " }"
	}
	return strings.Join(wrapped, ", ")
}
