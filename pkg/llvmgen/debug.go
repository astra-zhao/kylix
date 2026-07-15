package llvmgen

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"kylix/ast"
	"kylix/token"
)

// debug.go — DWARF debug-info emission for the LLVM backend.
//
// When CompileOpts.DebugInfo is set (kylix build --backend=llvm -g), the
// generator attaches DWARF metadata to the emitted IR so that LLDB/GDB can
// resolve function names, source files, and source lines — enabling
// `break <function>`, per-line stepping (`step`/`next`), function-level
// backtraces, and local-variable inspection.
//
// Scope (v4.6.0): per-instruction DILocation + DILocalVariable.
//   - Each user-defined function (and `main`) gets a DISubprogram referenced
//     from its `define` line via `!dbg !N`.
//   - emitStatement/emitExpr set a "current source position" (line+column)
//     before emitting instructions; the line() helper appends `, !dbg !M` to
//     each instruction-level IR line, where !M is a DILocation node for the
//     current position scoped to the current subprogram.
//   - emitVarDecl emits a DILocalVariable + `call void @llvm.dbg.declare(...)`
//     next to each alloca so LLDB can resolve local variable names + scopes.
//
// Metadata layout (IDs are assigned deterministically):
//
//	!0 = distinct !DICompileUnit(...)
//	!1 = !{i32 7, !"Dwarf Version", i32 4}
//	!2 = !{i32 2, !"Debug Info Version", i32 3}
//	!3 = !DIFile(filename: "...", directory: "...")
//	!4.. = distinct !DISubprogram(...)   ; one per function, in registration order
//	!N   = !DISubroutineType(types: !{null})   ; shared void() type
//	!N+1 = !{}                              ; empty retainedNodes list (base)
//	!N+2.. = !DILocation(...)              ; one per unique (line, col, scope) pair
//	!M..   = !DILocalVariable(...)         ; one per user-local alloca
//	!K..   = !DIBasicType(...)             ; canonical type nodes

// dbgMeta collects DWARF metadata during codegen and emits it at module end.
type dbgMeta struct {
	srcFilename string // DIFile filename (e.g. "main.klx")
	srcDir      string // DIFile directory (absolute)
	progs       []dbgSubprogram
	nextID      int // next metadata ID to allocate

	// Per-instruction DILocation support (v4.6.0).
	// curLoc is the "current source position": the (line,column) of the AST
	// node currently being emitted. line() appends !dbg !<locID> to each
	// instruction-level IR line, where locID resolves to a DILocation scoped
	// to curScope (the subprogram of the function we're inside).
	curLine  int // 0 = no position set (don't attach !dbg)
	curCol   int
	curScope int // subprogram metadata ID (0 outside a function)
	locs     []dbgLocation
	locByKey map[dbgLocKey]int // dedup: (line,col,scope) → metadata ID

	// DILocalVariable support (v4.6.0).
	locals []dbgLocalVar

	// DILexicalBlock support (v4.9.0): nested source scopes inside a
	// subprogram. curScope may point at a lexical-block ID (instead of a
	// subprogram) when emitting inside a block; DILocations + locals emitted
	// there attach to the block, giving LLDB block-scoped variable visibility.
	lexBlocks []dbgLexicalBlock
}

type dbgSubprogram struct {
	id   int    // metadata ID for this DISubprogram (e.g. 4 → !4)
	name string // function name
	line int    // source line of the function declaration
}

// dbgLocKey dedups DILocation nodes by (line, column, scope). Identical
// positions within the same function share one !DILocation metadata node.
type dbgLocKey struct {
	line  int
	col   int
	scope int
}

type dbgLocation struct {
	id    int // metadata ID
	line  int
	col   int
	scope int // subprogram ID
}

// dbgLocalVar records a DILocalVariable for one user-local alloca.
type dbgLocalVar struct {
	id       int    // metadata ID for the DILocalVariable
	name     string // source variable name
	line     int    // declaration line
	scope    int    // subprogram or lexical-block ID
	llvmType string // LLVM type string (for diType mapping)
	alloca   string // the alloca register (ptr to storage)
}

// dbgLexicalBlock records one DILexicalBlock: a nested source scope inside a
// subprogram (e.g. the body of an if/while/for, or a Pascal `begin...end`
// block). Locals declared inside it get this block as their scope so LLDB
// can show block-scoped variables correctly, and stepping reflects nesting.
type dbgLexicalBlock struct {
	id     int // metadata ID for the DILexicalBlock
	parent int // enclosing scope (subprogram or outer lexical block)
	line   int
	col    int
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
		locByKey:    make(map[dbgLocKey]int),
	}
}

// registerSubprogram records a DISubprogram for a function and returns its
// metadata ID (so the caller can append `!dbg !N` to the define line). The
// returned ID also becomes the "scope" for DILocations emitted inside that
// function (set via setDbgScope).
func (g *Generator) registerSubprogram(name string, line int) int {
	if g.dbg == nil {
		return 0
	}
	id := g.dbg.nextID
	g.dbg.nextID++
	g.dbg.progs = append(g.dbg.progs, dbgSubprogram{id: id, name: name, line: line})
	return id
}

// setDbgScope sets the subprogram that subsequent instructions belong to.
// Called when entering a function body (emitMain/emitFunctionDecl/emitMethod).
// Pass 0 to clear (outside any function).
func (g *Generator) setDbgScope(subprogID int) {
	if g.dbg == nil {
		return
	}
	g.dbg.curScope = subprogID
}

// setDbgNode extracts the source line/column from an AST node's Token and
// sets it as the current debug position. This is the main entrypoint called
// at the top of emitStatement/emitExpr: before emitting any IR for the node,
// record where it came from so each instruction carries a !dbg location.
func (g *Generator) setDbgNode(node ast.Node) {
	if g.dbg == nil || node == nil {
		return
	}
	tok := nodeToken(node)
	if tok.Line == 0 {
		return
	}
	g.dbg.curLine = tok.Line
	g.dbg.curCol = tok.Column
}

// clearDbgPos unsets the current source position (synthetic/runtime code).
func (g *Generator) clearDbgPos() {
	if g.dbg == nil {
		return
	}
	g.dbg.curLine = 0
	g.dbg.curCol = 0
}

// registerLexicalBlock allocates a DILexicalBlock scoped to the current scope
// (a subprogram or an enclosing lexical block) and switches curScope to it.
// Returns the new block's metadata ID so the caller can restore the previous
// scope on exit. Called when entering a nested block (emitBlockScoped) under
// -g, so locals declared in that block are scoped to it rather than the whole
// function. The block's source position is taken from the current position.
func (g *Generator) registerLexicalBlock() int {
	if g.dbg == nil {
		return 0
	}
	id := g.dbg.nextID
	g.dbg.nextID++
	line := g.dbg.curLine
	if line == 0 {
		line = 1
	}
	col := g.dbg.curCol
	if col == 0 {
		col = 1
	}
	g.dbg.lexBlocks = append(g.dbg.lexBlocks, dbgLexicalBlock{
		id:     id,
		parent: g.dbg.curScope,
		line:   line,
		col:    col,
	})
	prev := g.dbg.curScope
	g.dbg.curScope = id
	return prev
}

// registerLocalVariable records a DILocalVariable for a local alloca and
// returns its metadata ID. The caller emits a `call void @llvm.dbg.declare`
// next to the alloca to associate the LLVM value with the source variable.
func (g *Generator) registerLocalVariable(name string, line int, llvmType, allocaReg string) int {
	if g.dbg == nil {
		return 0
	}
	id := g.dbg.nextID
	g.dbg.nextID++
	g.dbg.locals = append(g.dbg.locals, dbgLocalVar{
		id:       id,
		name:     name,
		line:     line,
		scope:    g.dbg.curScope,
		llvmType: llvmType,
		alloca:   allocaReg,
	})
	return id
}

// curDbgLocID returns the metadata ID for a DILocation at the current
// position (allocating one if needed), or 0 if no position is set or debug
// info is off. Dedup keeps identical (line,col,scope) positions sharing one node.
func (g *Generator) curDbgLocID() int {
	if g.dbg == nil || g.dbg.curLine == 0 || g.dbg.curScope == 0 {
		return 0
	}
	key := dbgLocKey{line: g.dbg.curLine, col: g.dbg.curCol, scope: g.dbg.curScope}
	if id, ok := g.dbg.locByKey[key]; ok {
		return id
	}
	id := g.dbg.nextID
	g.dbg.nextID++
	g.dbg.locByKey[key] = id
	g.dbg.locs = append(g.dbg.locs, dbgLocation{id: id, line: key.line, col: key.col, scope: key.scope})
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
	baseEmptyListID := d.nextID
	d.nextID++
	// v4.8.0: per-llvmType DIBasicType nodes so LLDB formats values correctly
	// (int64 → DW_ATE_signed, double → DW_ATE_float, ptr → DW_ATE_address,
	// i1 → DW_ATE_boolean). Collect the set of types actually referenced by
	// locals; always include i64 as the fallback. Allocate IDs in sorted
	// order so emission is deterministic and the cache key stays stable.
	typeSet := map[string]bool{}
	for _, lv := range d.locals {
		typeSet[lv.llvmType] = true
	}
	typeSet["i64"] = true // fallback for unknown / untyped locals
	typeKeys := make([]string, 0, len(typeSet))
	for k := range typeSet {
		typeKeys = append(typeKeys, k)
	}
	sort.Strings(typeKeys)
	diTypeIDs := map[string]int{} // llvmType → metadata ID
	for _, k := range typeKeys {
		diTypeIDs[k] = d.nextID
		d.nextID++
	}
	// Subprograms (one per registered function) — reference the subroutine type.
	// Each gets its own retainedNodes list (so its locals can be attached).
	subprogRetained := make(map[int]int) // subprogram ID → its retainedNodes list ID
	for _, sp := range d.progs {
		retainedID := d.nextID
		d.nextID++
		subprogRetained[sp.id] = retainedID
		g.line(fmt.Sprintf(
			"!%d = distinct !DISubprogram(name: %q, scope: !3, file: !3, line: %d, type: %s, scopeLine: %d, spFlags: DISPFlagDefinition, unit: !0, retainedNodes: %s)",
			sp.id, sp.name, sp.line, dbgRef(subrTypeID), sp.line, dbgRef(retainedID),
		))
	}
	// DILexicalBlock nodes (v4.9.0): one per nested block. Each references its
	// enclosing scope (a subprogram or an outer lexical block) so the scope
	// tree reflects source nesting. Locals/DILocations emitted inside the
	// block point at the block ID (set via registerLexicalBlock → curScope).
	for _, lb := range d.lexBlocks {
		g.line(fmt.Sprintf(
			"!%d = distinct !DILexicalBlock(scope: %s, file: !3, line: %d, column: %d)",
			lb.id, dbgRef(lb.parent), lb.line, lb.col,
		))
	}
	// DILocalVariable nodes (one per user-local alloca). Each references its
	// declaring function's subprogram as scope and a basic DI type matching
	// the variable's LLVM type (so LLDB formats the value correctly).
	for _, lv := range d.locals {
		typeID := diTypeIDs[lv.llvmType]
		if typeID == 0 {
			typeID = diTypeIDs["i64"] // fallback
		}
		g.line(fmt.Sprintf(
			"!%d = !DILocalVariable(name: %q, scope: %s, file: !3, line: %d, type: %s)",
			lv.id, lv.name, dbgRef(lv.scope), lv.line, dbgRef(typeID),
		))
	}
	// DILocation nodes (one per unique position). inlinedAt omitted (no
	// inlining in -O0 codegen).
	for _, loc := range d.locs {
		g.line(fmt.Sprintf(
			"!%d = !DILocation(line: %d, column: %d, scope: %s)",
			loc.id, loc.line, loc.col, dbgRef(loc.scope),
		))
	}
	// Types list: a single null (void return, no params).
	g.line(fmt.Sprintf("%s = !{null}", dbgRef(typeListID)))
	// Subroutine type referencing the types list.
	g.line(fmt.Sprintf("%s = !DISubroutineType(types: %s)", dbgRef(subrTypeID), dbgRef(typeListID)))
	// Empty retainedNodes list (base, for subprograms with no locals).
	g.line(fmt.Sprintf("%s = !{}", dbgRef(baseEmptyListID)))
	// Per-llvmType DIBasicType definitions. int64 → DW_ATE_signed (fallback),
	// double → DW_ATE_float, ptr → DW_ATE_address, i1 → DW_ATE_boolean.
	for _, k := range typeKeys {
		id := diTypeIDs[k]
		switch k {
		case "double":
			g.line(fmt.Sprintf("!%d = !DIBasicType(name: \"double\", size: 64, encoding: DW_ATE_float)", id))
		case "ptr":
			g.line(fmt.Sprintf("!%d = !DIBasicType(name: \"ptr\", size: 64, encoding: DW_ATE_address)", id))
		case "i1":
			g.line(fmt.Sprintf("!%d = !DIBasicType(name: \"bool\", size: 8, encoding: DW_ATE_boolean)", id))
		default: // "i64" + any unrecognized → signed 64-bit fallback
			g.line(fmt.Sprintf("!%d = !DIBasicType(name: \"int64\", size: 64, encoding: DW_ATE_signed)", id))
		}
	}
	// Per-subprogram retainedNodes lists (containing their locals, if any).
	// A local's scope may be a subprogram or a lexical block; only locals
	// whose scope IS a subprogram go into that subprogram's retainedNodes.
	// (Lexical-block-scoped locals are reachable via the block's own scope
	// chain — DWARF doesn't require them in the subprogram's list, and LLDB
	// resolves them through the lexical block.)
	subprogLocalIDs := make(map[int][]int)
	for _, lv := range d.locals {
		if _, isSubprog := subprogRetained[lv.scope]; isSubprog {
			subprogLocalIDs[lv.scope] = append(subprogLocalIDs[lv.scope], lv.id)
		}
	}
	for spID, retainedID := range subprogRetained {
		ids := subprogLocalIDs[spID]
		if len(ids) == 0 {
			// Empty list — stays valid.
			g.line(fmt.Sprintf("%s = !{}", dbgRef(retainedID)))
		} else {
			parts := make([]string, len(ids))
			for i, id := range ids {
				parts[i] = dbgRef(id)
			}
			g.line(fmt.Sprintf("%s = !{%s}", dbgRef(retainedID), strings.Join(parts, ", ")))
		}
	}
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

// nodeToken extracts the token.Token from an AST node for its source position.
// Returns a zero token if the node has no Token field (synthetic nodes).
func nodeToken(node ast.Node) token.Token {
	var t token.Token
	// Reflection-free: type switch over concrete node types that carry a Token.
	switch n := node.(type) {
	// Statements
	case *ast.AssignmentStatement:
		t = n.Token
	case *ast.ExpressionStatement:
		t = n.Token
	case *ast.BlockStatement:
		t = n.Token
	case *ast.IfStatement:
		t = n.Token
	case *ast.WhileStatement:
		t = n.Token
	case *ast.ForStatement:
		t = n.Token
	case *ast.RepeatStatement:
		t = n.Token
	case *ast.VarDecl:
		t = n.Token
	case *ast.ReturnStatement:
		t = n.Token
	case *ast.TryStatement:
		t = n.Token
	case *ast.RaiseStatement:
		t = n.Token
	case *ast.ForEachStatement:
		t = n.Token
	case *ast.CaseStatement:
		t = n.Token
	case *ast.MatchStatement:
		t = n.Token
	case *ast.BreakStatement:
		t = n.Token
	case *ast.ContinueStatement:
		t = n.Token
	case *ast.InheritedStatement:
		t = n.Token
	// Declarations (v4.9.0: methods carry a Token for source position)
	case *ast.FunctionDecl:
		t = n.Token
	// Expressions
	case *ast.IntegerLiteral:
		t = n.Token
	case *ast.FloatLiteral:
		t = n.Token
	case *ast.BooleanLiteral:
		t = n.Token
	case *ast.StringLiteral:
		t = n.Token
	case *ast.Identifier:
		t = n.Token
	case *ast.InfixExpression:
		t = n.Token
	case *ast.PrefixExpression:
		t = n.Token
	case *ast.CallExpression:
		t = n.Token
	case *ast.MemberExpression:
		t = n.Token
	case *ast.IndexExpression:
		t = n.Token
	case *ast.SliceExpression:
		t = n.Token
	case *ast.TypeCastExpression:
		t = n.Token
	case *ast.IsExpression:
		t = n.Token
	case *ast.ArrayLiteral:
		t = n.Token
	case *ast.LambdaExpression:
		t = n.Token
	}
	return t
}
