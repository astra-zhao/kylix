package llvmgen

import "fmt"

// stdlib_arena.go — per-iteration bump arena for long-running request loops
// (v0.6.9). The boot HTTP server allocates per-request framework buffers
// (headers, method/path slots, request handle, response buffer, body) that were
// never freed, so a long-running server leaked ~6.7KB per request. With the
// arena, the loop calls @__kylix_arena_reset each iteration and allocates its
// buffers from @__kylix_arena_alloc; the pool is reused and framework buffers
// never accumulate. User objects (Variant boxes, classes, strings) still use
// malloc — the arena only backs the framework loop.
//
// The arena is 1 MiB (like the datetime arena). Sub-pool allocations bump the
// pointer; if a request overruns the arena, arena_alloc falls back to malloc
// (rare large bodies leak, but the common path is leak-free).

// arenaSize is the bump pool capacity in bytes.
const arenaSize = 1048576

// emitArenaBodies emits the arena globals + allocator/reset defines. Called
// once per module from emitProgram when g.needArena is set (v0.6.9).
func (g *Generator) emitArenaBodies() {
	g.line(fmt.Sprintf("@__kylix_arena = internal global [%d x i8] zeroinitializer, align 8", arenaSize))
	g.line("@__kylix_arena_ptr = internal global ptr @__kylix_arena")
	g.line("")
	g.line("define ptr @__kylix_arena_alloc(i64 %size) {")
	g.line("entry:")
	cur := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr @__kylix_arena_ptr", cur))
	newPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %%size", newPtr, cur))
	arenaEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr @__kylix_arena, i64 %d", arenaEnd, arenaSize))
	ok := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ule ptr %s, %s", ok, newPtr, arenaEnd))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ok, okLbl, failLbl))
	g.line(fmt.Sprintf("%s:", okLbl))
	g.line(fmt.Sprintf("  store ptr %s, ptr @__kylix_arena_ptr", newPtr))
	g.line(fmt.Sprintf("  ret ptr %s", cur))
	// Arena exhausted (rare large request) → fall back to plain malloc so the
	// request still works; such buffers are not reclaimed (accepted cost).
	g.line(fmt.Sprintf("%s:", failLbl))
	fallback := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %%size)", fallback))
	g.line(fmt.Sprintf("  ret ptr %s", fallback))
	g.line("}")
	g.line("")
	g.line("define void @__kylix_arena_reset() {")
	g.line("entry:")
	g.line("  store ptr @__kylix_arena, ptr @__kylix_arena_ptr")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// arenaAllocCall emits `%r = call ptr @__kylix_arena_alloc(i64 <sizeReg>)` and
// sets g.needArena. Helper for boot server IR that allocates from the pool.
func (g *Generator) arenaAllocCall(sizeReg string) string {
	g.needArena = true
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_arena_alloc(i64 %s)", r, sizeReg))
	return r
}

// arenaResetCall emits `call void @__kylix_arena_reset()`. Sets g.needArena.
func (g *Generator) arenaResetCall() {
	g.needArena = true
	g.line("  call void @__kylix_arena_reset()")
}