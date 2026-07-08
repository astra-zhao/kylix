package llvmgen

import (
	"fmt"
	"path/filepath"
	"strings"
)

// debug.go — DWARF debug-info emission for the LLVM backend (v4.5.0 Phase C).
//
// When CompileOpts.DebugInfo is set (kylix build --backend=llvm -g), the
// generator attaches DWARF metadata to the emitted IR so that LLDB/GDB can
// resolve function names, source files, and starting line numbers — enabling
// `break <function>`, function-level backtraces, and source correlation.
//
// Scope (MVP): function-level debug info. Each user-defined function (and
// `main`) gets a DISubprogram referenced from its `define` line via `!dbg !N`.
// Per-instruction DILocation (line stepping within a function) is not emitted
// in this cut — stdlib pre-generated IR has no source lines, and attaching
// stale locations would mislead the debugger. A follow-up can thread source
// positions through emitStatement/emitExpr for per-line DILocation.
//
// Metadata layout (IDs are assigned deterministically):
//
//	!0 = distinct !DICompileUnit(...)
//	!1 = !{i32 7, !"Dwarf Version", i32 4}
//	!2 = !{i32 2, !"Debug Info Version", i32 3}
//	!3 = !DIFile(filename: "...", directory: "...")
//	!4.. = distinct !DISubprogram(...)   ; one per function, in registration order
//	!N = !DISubroutineType(types: !{null})   ; shared void() type
//	!N+1 = !{}                              ; empty retainedNodes list

// dbgMeta collects DWARF metadata during codegen and emits it at module end.
type dbgMeta struct {
	srcFilename string // DIFile filename (e.g. "main.klx")
	srcDir      string // DIFile directory (absolute)
	progs       []dbgSubprogram
	nextID      int // next metadata ID to allocate
}

type dbgSubprogram struct {
	id   int    // metadata ID for this DISubprogram (e.g. 4 → !4)
	name string // function name
	line int    // source line of the function declaration
}

// initDbgMeta prepares the metadata collector with the source file info.
func (g *Generator) initDbgMeta(srcFile string) {
	dir, name := filepath.Split(srcFile)
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	g.dbg = &dbgMeta{
		srcFilename: name,
		srcDir:      dir,
		nextID:      4, // !0=CU, !1,!2=flags, !3=DIFile → subprograms start at !4
	}
}

// registerSubprogram records a DISubprogram for a function and returns its
// metadata ID (so the caller can append `!dbg !N` to the define line).
func (g *Generator) registerSubprogram(name string, line int) int {
	if g.dbg == nil {
		return 0
	}
	id := g.dbg.nextID
	g.dbg.nextID++
	g.dbg.progs = append(g.dbg.progs, dbgSubprogram{id: id, name: name, line: line})
	return id
}

// dbgRef formats a metadata ID as an IR reference (!N).
func dbgRef(id int) string {
	return fmt.Sprintf("!%d", id)
}

// emitDbgMetadata appends the DWARF metadata block to the module output.
// Called once at the end of emitProgram when debugInfo is on.
func (g *Generator) emitDbgMetadata() {
	if g.dbg == nil {
		return
	}
	d := g.dbg
	g.line("")
	g.line("; ===== DWARF debug info (kylix -g) =====")
	// Named metadata anchors.
	g.line("!llvm.dbg.cu = !{!0}")
	g.line("!llvm.module.flags = !{!1, !2}")
	// !0 = DICompileUnit
	g.line(fmt.Sprintf("!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: \"kylix\", isOptimized: false, runtimeVersion: 0, emissionKind: FullDebug)"))
	// !1, !2 = module flags (Dwarf Version 4, Debug Info Version 3)
	g.line(`!1 = !{i32 7, !"Dwarf Version", i32 4}`)
	g.line(`!2 = !{i32 2, !"Debug Info Version", i32 3}`)
	// !3 = DIFile
	g.line(fmt.Sprintf("!3 = !DIFile(filename: %q, directory: %q)", d.srcFilename, d.srcDir))
	// Shared subroutine type: void(). Two nodes: the types list (!{null}) and
	// the DISubroutineType referencing it.
	typeListID := d.nextID
	d.nextID++
	subrTypeID := d.nextID
	d.nextID++
	emptyListID := d.nextID
	d.nextID++
	// Subprograms (one per registered function) — reference the subroutine type.
	for _, sp := range d.progs {
		g.line(fmt.Sprintf(
			"!%d = distinct !DISubprogram(name: %q, scope: !3, file: !3, line: %d, type: %s, scopeLine: %d, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: %s)",
			sp.id, sp.name, sp.line, dbgRef(subrTypeID), sp.line, dbgRef(emptyListID),
		))
	}
	// Types list: a single null (void return, no params).
	g.line(fmt.Sprintf("%s = !{null}", dbgRef(typeListID)))
	// Subroutine type referencing the types list.
	g.line(fmt.Sprintf("%s = !DISubroutineType(types: %s)", dbgRef(subrTypeID), dbgRef(typeListID)))
	// Empty retainedNodes list.
	g.line(fmt.Sprintf("%s = !{}", dbgRef(emptyListID)))
}

// defineLineWithDbg returns a `define ...` line, appending `!dbg !N` when
// debug info is active and a subprogram was registered for this function.
func (g *Generator) defineLineWithDbg(defineLine string, subprogID int) string {
	if g.debugInfo && subprogID > 0 {
		// Insert before the trailing " {": "define ... @name(...) !dbg !N {"
		return strings.TrimSuffix(defineLine, " {") + " " + fmt.Sprintf("!dbg %s {", dbgRef(subprogID))
	}
	return defineLine
}
