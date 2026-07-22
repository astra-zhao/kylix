// rtti.go — v5.4.0: runtime class hierarchy type-info for `is`/`as` class casts.
//
// The exception runtime (exc.go) uses i32 type IDs for the Exception subtree;
// this is the general counterpart for ALL user classes, keyed by vtable pointer
// (each class's @TFoo_vtable is a unique global constant, so pointer equality
// identifies a class, and a child→parent vtable edge table supports subtype
// queries). Used by emitIsExpr/emitAsExpr for `obj is TClass` / `obj as TClass`.
package llvmgen

import (
	"fmt"
	"sort"
	"strings"
)

// collectClassHierarchy builds the child→parent vtable edge list for all user
// classes (including records, which have a null vtable and are excluded). Runs
// after all classes are registered (emitProgram end, like collectExceptionTypes).
func (g *Generator) collectClassHierarchy() []classEdge {
	var edges []classEdge
	// Deterministic order by class name.
	names := make([]string, 0, len(g.classes))
	for name := range g.classes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		info := g.classes[name]
		if info == nil || info.Parent == "" {
			continue
		}
		if g.records[name] {
			continue // records have no vtable
		}
		edges = append(edges, classEdge{child: name, parent: info.Parent})
	}
	return edges
}

type classEdge struct{ child, parent string }

// emitClassRuntime emits the class subtype edge table and the
// @__kylix_class_is_a helper. Idempotent. v5.4.0.
func (g *Generator) emitClassRuntime() {
	edges := g.collectClassHierarchy()

	// v5.4.0: ensure every class (not just those in the edge table) has a
	// vtable constant. Classes with no methods get an empty [0 x ptr] vtable
	// from emitVtable, but if it wasn't emitted for any reason, emit a
	// fallback here so all @TFoo_vtable references resolve.
	for name, info := range g.classes {
		if g.records[name] {
			continue
		}
		if info != nil && len(info.Methods) == 0 {
			g.line(fmt.Sprintf("@%s_vtable = constant [0 x ptr] []", name))
		}
	}

	if len(edges) == 0 {
		// No class hierarchy — is/as class checks are trivially false (no edges
		// means no subclassing); emit an empty table + a helper that returns
		// false unless child==parent.
		g.line("@__kylix_classtab = constant [0 x { ptr, ptr }] zeroinitializer")
		g.line("define i1 @__kylix_class_is_a(ptr %child, ptr %parent) {")
		g.line("entry:")
		g.line("  %eq = icmp eq ptr %child, %parent")
		g.line("  ret i1 %eq")
		g.line("}")
		g.line("")
		return
	}

	// Named struct type for one edge (LLVM IR requires named struct types for
	// array-of-struct constants; anonymous { ptr, ptr } triggers parse errors).
	g.line("%__kylix_class_edge = type { ptr, ptr }")
	var parts []string
	for _, e := range edges {
		parts = append(parts, fmt.Sprintf("%%__kylix_class_edge { ptr @%s_vtable, ptr @%s_vtable }", e.child, e.parent))
	}
	g.line(fmt.Sprintf("@__kylix_classtab = constant [%d x %%__kylix_class_edge] [ %s ]",
		len(edges), strings.Join(parts, ", ")))
	g.line("")

	// define i1 @__kylix_class_is_a(ptr %child, ptr %parent)
	// Walks child→parent vtable chain by scanning the edge table for the
	// current child's parent, until it reaches %parent (true) or a child with
	// no outgoing edge (false). Bounded by table length.
	n := len(edges)
	g.line(fmt.Sprintf(`define i1 @__kylix_class_is_a(ptr %%child, ptr %%parent) {
entry:
  %%eq = icmp eq ptr %%child, %%parent
  br i1 %%eq, label %%ret_true, label %%loop
loop:
  %%c = phi ptr [ %%child, %%entry ], [ %%c_next, %%loop_next ]
  %%i = phi i64 [ 0, %%entry ], [ %%i_next, %%loop_next ]
  %%oob = icmp eq i64 %%i, %d
  br i1 %%oob, label %%ret_false, label %%body
body:
  %%slot = getelementptr inbounds [%d x %%__kylix_class_edge], ptr @__kylix_classtab, i64 0, i64 %%i
  %%cvt_ptr = getelementptr inbounds %%__kylix_class_edge, ptr %%slot, i32 0, i32 0
  %%cvt = load ptr, ptr %%cvt_ptr
  %%iscur = icmp eq ptr %%cvt, %%c
  br i1 %%iscur, label %%found, label %%loop_next
found:
  %%pvt_ptr = getelementptr inbounds %%__kylix_class_edge, ptr %%slot, i32 0, i32 1
  %%par = load ptr, ptr %%pvt_ptr
  %%match = icmp eq ptr %%par, %%parent
  br i1 %%match, label %%ret_true, label %%update
update:
  br label %%loop_next
loop_next:
  %%c_next = phi ptr [ %%c, %%body ], [ %%par, %%update ]
  %%i_next = add i64 %%i, 1
  br label %%loop
ret_true:
  ret i1 true
ret_false:
  ret i1 false
}
`, n, n))
}

// classIsACall emits a call to @__kylix_class_is_a(objVT, targetVT) and returns
// the i1 result register. objReg is a pointer to the object (its [0] slot is
// the vtable pointer). targetClassName names the class to test against.
func (g *Generator) classIsACall(objReg, targetClassName string) (string, error) {
	g.needClassRTTI = true
	// v5.4.0: guard against null object pointers — `decl is TClass` where decl
	// is nil should return false, not crash on vtable load.
	nullCk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", nullCk, objReg))
	trueLbl := g.label()
	falseLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nullCk, falseLbl, trueLbl))
	g.line(fmt.Sprintf("%s:", trueLbl))
	// Object's first word (offset 0) is the vtable pointer.
	objVT := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", objVT, objReg))
	trueR := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_class_is_a(ptr %s, ptr @%s_vtable)",
		trueR, objVT, targetClassName))
	g.line(fmt.Sprintf("  br label %%%s", exitLbl))
	g.line(fmt.Sprintf("%s:", falseLbl))
	falseR := g.tmp()
	g.line(fmt.Sprintf("  %s = add i1 0, 0 ; null is %s = false", falseR, targetClassName))
	g.line(fmt.Sprintf("  br label %%%s", exitLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	// Phi to merge the two branches.
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = phi i1 [ %s, %%%s ], [ %s, %%%s ]", r, trueR, trueLbl, falseR, falseLbl))
	return r, nil
}
