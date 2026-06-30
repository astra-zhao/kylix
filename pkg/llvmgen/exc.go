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
  %%c = phi i32 [ %%child, %%entry ], [ %%par, %%loop_next ]
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
  br i1 %%match, label %%ret_true, label %%loop_next
loop_next:
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
