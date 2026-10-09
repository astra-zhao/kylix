package llvmgen

import (
	"fmt"
	"strings"

	"kylix/internal/wasiapi"
)

// wasi_rt.go — wasm32-unknown-wasi runtime (v0.15).
//
// Pure logic only: the module imports wasi_snapshot_preview1 and nothing
// else (no DOM, no env). libc calls the rest of the backend already emits
// (malloc/printf/puts/strlen/…) are defined here on top of that table.
// setjmp is not available in wasm; @setjmp returns 0 and @longjmp aborts
// via proc_exit(70), so try/except does not catch.
//
// The heap is a fixed bump allocator. @free is a no-op: a WASI command
// drops the instance on proc_exit. Do not treat this as a general GC.

const wasiHeapBytes = 8 << 20

const wasiAttrBase = 200

func (g *Generator) wasiTarget() bool { return g.targetOS == "wasi" }

func wasiSym(name string) string { return "@__kylix_wasi_" + name }

// emitWasiImportDecls emits the preview1 import table. Called from
// emitRuntimeDecls instead of the host libc declares. Bodies are emitted
// later by emitWasiRuntime (a declare and a later define of @malloc would
// clash, so the libc symbols are only defined, never declared).
func (g *Generator) emitWasiImportDecls() {
	g.line("; ===== wasi_snapshot_preview1 import table (no DOM) =====")
	for i, im := range wasiapi.Preview1 {
		ret := "void"
		if len(im.Results) == 1 {
			ret = im.Results[0]
		}
		noreturn := ""
		if im.Noreturn {
			noreturn = " noreturn"
		}
		g.line(fmt.Sprintf("declare %s %s(%s)%s #%d",
			ret, wasiSym(im.Name), strings.Join(im.Params, ", "), noreturn, wasiAttrBase+i))
	}
	g.line("declare void @llvm.va_start(ptr)")
	g.line("declare void @llvm.va_end(ptr)")
	g.line("declare double @llvm.fabs.f64(double)")
	g.line("")
	for i, im := range wasiapi.Preview1 {
		extra := ""
		if im.Noreturn {
			extra = "noreturn "
		}
		g.line(fmt.Sprintf(`attributes #%d = { %s"wasm-import-module"="%s" "wasm-import-name"="%s" }`,
			wasiAttrBase+i, extra, wasiapi.Snapshot, im.Name))
	}
	g.line("")
}

// emitWasiRuntime emits the bump heap, the libc shims, and @_start.
// No-op unless the target is wasi/wasm32. Host IR is untouched.
func (g *Generator) emitWasiRuntime() {
	if !g.wasiTarget() {
		return
	}
	if g.debugInfo {
		g.clearDbgPos()
		g.setDbgScope(0)
	}
	g.line("; ===== wasm32-wasi runtime =====")
	g.line(fmt.Sprintf("@__kylix_wasi_heap = global [%d x i8] zeroinitializer, align 16", wasiHeapBytes))
	g.line("@__kylix_wasi_brk = global i32 0")
	g.line(`@__kylix_wasi_nl = private unnamed_addr constant [1 x i8] c"\0A"`)
	g.line(`@__kylix_wasi_oom = private unnamed_addr constant [20 x i8] c"wasi heap exhausted\0A"`)
	g.line(`@__kylix_wasi_nan = private unnamed_addr constant [4 x i8] c"NaN\00"`)
	g.line(`@__kylix_wasi_inf = private unnamed_addr constant [4 x i8] c"Inf\00"`)

	// Keep every preview1 import live so the linked module's import section
	// is the table, not only the syscalls this particular program called.
	var used []string
	for _, im := range wasiapi.Preview1 {
		used = append(used, "ptr "+wasiSym(im.Name))
	}
	g.line(fmt.Sprintf("@llvm.used = appending global [%d x ptr] [%s]", len(used), strings.Join(used, ", ")))
	g.line("")

	g.emitWasiMemory()
	g.emitWasiIO()
	g.emitWasiStrings()
	g.emitWasiFormat()
	g.emitWasiProcess()
	g.emitWasiStart()
}

func (g *Generator) emitWasiMemory() {
	heapTy := fmt.Sprintf("[%d x i8]", wasiHeapBytes)
	g.line("define ptr @malloc(i64 %n) {")
	g.line("entry:")
	g.line("  %n32 = trunc i64 %n to i32")
	g.line("  %isz = icmp eq i32 %n32, 0")
	g.line("  %sz = select i1 %isz, i32 1, i32 %n32")
	g.line("  %brk0 = load i32, ptr @__kylix_wasi_brk")
	g.line("  %brkA = add i32 %brk0, 15")
	g.line("  %aligned = and i32 %brkA, -16")
	g.line("  %user = add i32 %aligned, 16")
	g.line("  %nextRaw = add i32 %user, %sz")
	g.line("  %nextA = add i32 %nextRaw, 15")
	g.line("  %next = and i32 %nextA, -16")
	g.line(fmt.Sprintf("  %%oom = icmp ugt i32 %%next, %d", wasiHeapBytes))
	g.line("  br i1 %oom, label %fail, label %ok")
	g.line("fail:")
	g.line("  call void @__kylix_wasi_write(i32 2, ptr @__kylix_wasi_oom, i32 20)")
	g.line("  call void @__kylix_wasi_proc_exit(i32 1)")
	g.line("  unreachable")
	g.line("ok:")
	g.line("  store i32 %next, ptr @__kylix_wasi_brk")
	g.line(fmt.Sprintf("  %%base = getelementptr inbounds %s, ptr @__kylix_wasi_heap, i32 0, i32 %%aligned", heapTy))
	g.line("  store i32 %sz, ptr %base")
	g.line("  %payload = getelementptr i8, ptr %base, i32 16")
	g.line("  ret ptr %payload")
	g.line("}")
	g.line("")

	g.line("define void @free(ptr %p) {")
	g.line("entry:")
	g.line("  ret void")
	g.line("}")
	g.line("")

	// llc -O0 lowers llvm.memset.p0.i64 to the wasm32 libcall
	// `i32 @memset(i32, i32, i32)` (size_t is i32). There is no wasi-libc
	// on this link, so the symbol has to be defined. Do not implement it
	// by calling the intrinsic: that becomes a recursive libcall.
	g.line("define i32 @memset(i32 %dst, i32 %val, i32 %n) {")
	g.line("entry:")
	g.line("  %p = inttoptr i32 %dst to ptr")
	g.line("  %b = trunc i32 %val to i8")
	g.line("  call void @__kylix_wasi_fill(ptr %p, i8 %b, i32 %n)")
	g.line("  ret i32 %dst")
	g.line("}")
	g.line("")

	g.line("define void @__kylix_wasi_fill(ptr %dst, i8 %val, i32 %n) {")
	g.line("entry:")
	g.line("  %i = alloca i32, align 4")
	g.line("  store i32 0, ptr %i")
	g.line("  br label %loop")
	g.line("loop:")
	g.line("  %ii = load i32, ptr %i")
	g.line("  %done = icmp uge i32 %ii, %n")
	g.line("  br i1 %done, label %ret, label %body")
	g.line("body:")
	g.line("  %p = getelementptr i8, ptr %dst, i32 %ii")
	g.line("  store i8 %val, ptr %p")
	g.line("  %ii2 = add i32 %ii, 1")
	g.line("  store i32 %ii2, ptr %i")
	g.line("  br label %loop")
	g.line("ret:")
	g.line("  ret void")
	g.line("}")
	g.line("")

	g.line("define ptr @calloc(i64 %nmemb, i64 %size) {")
	g.line("entry:")
	g.line("  %bytes = mul i64 %nmemb, %size")
	g.line("  %p = call ptr @malloc(i64 %bytes)")
	g.line("  %n32 = trunc i64 %bytes to i32")
	g.line("  call void @__kylix_wasi_fill(ptr %p, i8 0, i32 %n32)")
	g.line("  ret ptr %p")
	g.line("}")
	g.line("")

	g.line("define ptr @realloc(ptr %p, i64 %n) {")
	g.line("entry:")
	g.line("  %isnull = icmp eq ptr %p, null")
	g.line("  br i1 %isnull, label %fresh, label %copy")
	g.line("fresh:")
	g.line("  %q0 = call ptr @malloc(i64 %n)")
	g.line("  br label %done")
	g.line("copy:")
	g.line("  %hdr = getelementptr i8, ptr %p, i32 -16")
	g.line("  %old32 = load i32, ptr %hdr")
	g.line("  %old = zext i32 %old32 to i64")
	g.line("  %q1 = call ptr @malloc(i64 %n)")
	g.line("  %oldSmall = icmp ult i64 %old, %n")
	g.line("  %ncopy = select i1 %oldSmall, i64 %old, i64 %n")
	g.line("  call ptr @memcpy(ptr %q1, ptr %p, i64 %ncopy)")
	g.line("  br label %done")
	g.line("done:")
	g.line("  %q = phi ptr [ %q0, %fresh ], [ %q1, %copy ]")
	g.line("  ret ptr %q")
	g.line("}")
	g.line("")
}

func (g *Generator) emitWasiIO() {
	g.line("define void @__kylix_wasi_write(i32 %fd, ptr %buf, i32 %len) {")
	g.line("entry:")
	g.line("  %empty = icmp eq i32 %len, 0")
	g.line("  br i1 %empty, label %ret, label %do")
	g.line("do:")
	g.line("  %iov = alloca { i32, i32 }, align 4")
	g.line("  %off = ptrtoint ptr %buf to i32")
	g.line("  store i32 %off, ptr %iov, align 4")
	g.line("  %lp = getelementptr inbounds { i32, i32 }, ptr %iov, i32 0, i32 1")
	g.line("  store i32 %len, ptr %lp, align 4")
	g.line("  %nw = alloca i32, align 4")
	g.line("  %e = call i32 @__kylix_wasi_fd_write(i32 %fd, ptr %iov, i32 1, ptr %nw)")
	g.line("  br label %ret")
	g.line("ret:")
	g.line("  ret void")
	g.line("}")
	g.line("")

	g.line("define i64 @write(i32 %fd, ptr %buf, i64 %n) {")
	g.line("entry:")
	g.line("  %n32 = trunc i64 %n to i32")
	g.line("  call void @__kylix_wasi_write(i32 %fd, ptr %buf, i32 %n32)")
	g.line("  ret i64 %n")
	g.line("}")
	g.line("")

	g.line("define i32 @puts(ptr %s) {")
	g.line("entry:")
	g.line("  %n = call i64 @strlen(ptr %s)")
	g.line("  %n32 = trunc i64 %n to i32")
	g.line("  call void @__kylix_wasi_write(i32 1, ptr %s, i32 %n32)")
	g.line("  call void @__kylix_wasi_write(i32 1, ptr @__kylix_wasi_nl, i32 1)")
	g.line("  ret i32 0")
	g.line("}")
	g.line("")
}

func (g *Generator) emitWasiStrings() {
	g.line("define i64 @strlen(ptr %s) {")
	g.line("entry:")
	g.line("  %isnull = icmp eq ptr %s, null")
	g.line("  br i1 %isnull, label %zero, label %scan")
	g.line("zero:")
	g.line("  ret i64 0")
	g.line("scan:")
	g.line("  %i = alloca i32, align 4")
	g.line("  store i32 0, ptr %i")
	g.line("  br label %loop")
	g.line("loop:")
	g.line("  %ii = load i32, ptr %i")
	g.line("  %p = getelementptr i8, ptr %s, i32 %ii")
	g.line("  %c = load i8, ptr %p")
	g.line("  %z = icmp eq i8 %c, 0")
	g.line("  br i1 %z, label %done, label %next")
	g.line("next:")
	g.line("  %ii2 = add i32 %ii, 1")
	g.line("  store i32 %ii2, ptr %i")
	g.line("  br label %loop")
	g.line("done:")
	g.line("  %n32 = load i32, ptr %i")
	g.line("  %n = zext i32 %n32 to i64")
	g.line("  ret i64 %n")
	g.line("}")
	g.line("")

	g.line("define ptr @strcpy(ptr %dst, ptr %src) {")
	g.line("entry:")
	g.line("  %null = icmp eq ptr %src, null")
	g.line("  br i1 %null, label %empty, label %scan")
	g.line("empty:")
	g.line("  store i8 0, ptr %dst")
	g.line("  ret ptr %dst")
	g.line("scan:")
	g.line("  %i = alloca i32, align 4")
	g.line("  store i32 0, ptr %i")
	g.line("  br label %loop")
	g.line("loop:")
	g.line("  %ii = load i32, ptr %i")
	g.line("  %sp = getelementptr i8, ptr %src, i32 %ii")
	g.line("  %c = load i8, ptr %sp")
	g.line("  %dp = getelementptr i8, ptr %dst, i32 %ii")
	g.line("  store i8 %c, ptr %dp")
	g.line("  %z = icmp eq i8 %c, 0")
	g.line("  br i1 %z, label %done, label %next")
	g.line("next:")
	g.line("  %ii2 = add i32 %ii, 1")
	g.line("  store i32 %ii2, ptr %i")
	g.line("  br label %loop")
	g.line("done:")
	g.line("  ret ptr %dst")
	g.line("}")
	g.line("")

	g.line("define ptr @strcat(ptr %dst, ptr %src) {")
	g.line("entry:")
	g.line("  %n = call i64 @strlen(ptr %dst)")
	g.line("  %n32 = trunc i64 %n to i32")
	g.line("  %end = getelementptr i8, ptr %dst, i32 %n32")
	g.line("  %r = call ptr @strcpy(ptr %end, ptr %src)")
	g.line("  ret ptr %dst")
	g.line("}")
	g.line("")

	// Byte loops, not llvm.memcpy/memmove. On wasm32 those intrinsics lower
	// to libcalls whose size argument is i32; a body that calls the
	// intrinsic would recurse into this same symbol.
	g.line("define ptr @memcpy(ptr %dst, ptr %src, i64 %n) {")
	g.line("entry:")
	g.line("  %n32 = trunc i64 %n to i32")
	g.line("  %i = alloca i32, align 4")
	g.line("  store i32 0, ptr %i")
	g.line("  br label %loop")
	g.line("loop:")
	g.line("  %ii = load i32, ptr %i")
	g.line("  %done = icmp uge i32 %ii, %n32")
	g.line("  br i1 %done, label %ret, label %body")
	g.line("body:")
	g.line("  %sp = getelementptr i8, ptr %src, i32 %ii")
	g.line("  %c = load i8, ptr %sp")
	g.line("  %dp = getelementptr i8, ptr %dst, i32 %ii")
	g.line("  store i8 %c, ptr %dp")
	g.line("  %ii2 = add i32 %ii, 1")
	g.line("  store i32 %ii2, ptr %i")
	g.line("  br label %loop")
	g.line("ret:")
	g.line("  ret ptr %dst")
	g.line("}")
	g.line("")

	g.line("define ptr @memmove(ptr %dst, ptr %src, i64 %n) {")
	g.line("entry:")
	g.line("  %n32 = trunc i64 %n to i32")
	g.line("  %d = ptrtoint ptr %dst to i32")
	g.line("  %s = ptrtoint ptr %src to i32")
	g.line("  %back = icmp ugt i32 %d, %s")
	g.line("  br i1 %back, label %bprep, label %fwd")
	g.line("fwd:")
	g.line("  call ptr @memcpy(ptr %dst, ptr %src, i64 %n)")
	g.line("  ret ptr %dst")
	g.line("bprep:")
	g.line("  %i = alloca i32, align 4")
	g.line("  store i32 %n32, ptr %i")
	g.line("  br label %bloop")
	g.line("bloop:")
	g.line("  %ii = load i32, ptr %i")
	g.line("  %z = icmp eq i32 %ii, 0")
	g.line("  br i1 %z, label %bret, label %bbody")
	g.line("bbody:")
	g.line("  %ii2 = sub i32 %ii, 1")
	g.line("  store i32 %ii2, ptr %i")
	g.line("  %sp = getelementptr i8, ptr %src, i32 %ii2")
	g.line("  %c = load i8, ptr %sp")
	g.line("  %dp = getelementptr i8, ptr %dst, i32 %ii2")
	g.line("  store i8 %c, ptr %dp")
	g.line("  br label %bloop")
	g.line("bret:")
	g.line("  ret ptr %dst")
	g.line("}")
	g.line("")

	g.line("define i32 @strcmp(ptr %a, ptr %b) {")
	g.line("entry:")
	g.line("  %an = icmp eq ptr %a, null")
	g.line("  %aa = select i1 %an, ptr @__kylix_emptystr, ptr %a")
	g.line("  %bn = icmp eq ptr %b, null")
	g.line("  %bb = select i1 %bn, ptr @__kylix_emptystr, ptr %b")
	g.line("  %i = alloca i32, align 4")
	g.line("  store i32 0, ptr %i")
	g.line("  br label %loop")
	g.line("loop:")
	g.line("  %ii = load i32, ptr %i")
	g.line("  %ap = getelementptr i8, ptr %aa, i32 %ii")
	g.line("  %bp = getelementptr i8, ptr %bb, i32 %ii")
	g.line("  %ca = load i8, ptr %ap")
	g.line("  %cb = load i8, ptr %bp")
	g.line("  %eq = icmp eq i8 %ca, %cb")
	g.line("  br i1 %eq, label %same, label %diff")
	g.line("same:")
	g.line("  %z = icmp eq i8 %ca, 0")
	g.line("  br i1 %z, label %eq0, label %next")
	g.line("next:")
	g.line("  %ii2 = add i32 %ii, 1")
	g.line("  store i32 %ii2, ptr %i")
	g.line("  br label %loop")
	g.line("eq0:")
	g.line("  ret i32 0")
	g.line("diff:")
	g.line("  %ua = zext i8 %ca to i32")
	g.line("  %ub = zext i8 %cb to i32")
	g.line("  %lt = icmp ult i32 %ua, %ub")
	g.line("  %res = select i1 %lt, i32 -1, i32 1")
	g.line("  ret i32 %res")
	g.line("}")
	g.line("")

	g.line("define i32 @strncmp(ptr %a, ptr %b, i64 %n) {")
	g.line("entry:")
	g.line("  %i = alloca i64, align 8")
	g.line("  store i64 0, ptr %i")
	g.line("  br label %loop")
	g.line("loop:")
	g.line("  %ii = load i64, ptr %i")
	g.line("  %done = icmp uge i64 %ii, %n")
	g.line("  br i1 %done, label %eq0, label %body")
	g.line("body:")
	g.line("  %ii32 = trunc i64 %ii to i32")
	g.line("  %ap = getelementptr i8, ptr %a, i32 %ii32")
	g.line("  %bp = getelementptr i8, ptr %b, i32 %ii32")
	g.line("  %ca = load i8, ptr %ap")
	g.line("  %cb = load i8, ptr %bp")
	g.line("  %eq = icmp eq i8 %ca, %cb")
	g.line("  br i1 %eq, label %same, label %diff")
	g.line("same:")
	g.line("  %z = icmp eq i8 %ca, 0")
	g.line("  br i1 %z, label %eq0, label %inc")
	g.line("inc:")
	g.line("  %ii2 = add i64 %ii, 1")
	g.line("  store i64 %ii2, ptr %i")
	g.line("  br label %loop")
	g.line("eq0:")
	g.line("  ret i32 0")
	g.line("diff:")
	g.line("  %ua = zext i8 %ca to i32")
	g.line("  %ub = zext i8 %cb to i32")
	g.line("  %lt = icmp ult i32 %ua, %ub")
	g.line("  %res = select i1 %lt, i32 -1, i32 1")
	g.line("  ret i32 %res")
	g.line("}")
	g.line("")
}

func (g *Generator) emitWasiProcess() {
	g.line("define void @exit(i32 %code) noreturn {")
	g.line("entry:")
	g.line("  call void @__kylix_wasi_proc_exit(i32 %code)")
	g.line("  unreachable")
	g.line("}")
	g.line("")
	g.line("define i32 @setjmp(ptr %buf) {")
	g.line("entry:")
	g.line("  ret i32 0")
	g.line("}")
	g.line("")
	g.line("define i32 @_setjmp(ptr %buf) {")
	g.line("entry:")
	g.line("  ret i32 0")
	g.line("}")
	g.line("")
	g.line("define void @longjmp(ptr %buf, i32 %val) noreturn {")
	g.line("entry:")
	g.line("  call void @__kylix_wasi_proc_exit(i32 70)")
	g.line("  unreachable")
	g.line("}")
	g.line("")
	g.line("define void @_longjmp(ptr %buf, i32 %val) noreturn {")
	g.line("entry:")
	g.line("  call void @__kylix_wasi_proc_exit(i32 70)")
	g.line("  unreachable")
	g.line("}")
	g.line("")
}

func (g *Generator) emitWasiStart() {
	if g.isShared {
		return
	}
	g.line("define void @_start() {")
	g.line("entry:")
	if g.needArgs {
		g.line("  %argcSlot = alloca i32, align 4")
		g.line("  %szSlot = alloca i32, align 4")
		g.line("  %es = call i32 @__kylix_wasi_args_sizes_get(ptr %argcSlot, ptr %szSlot)")
		g.line("  %argc = load i32, ptr %argcSlot")
		g.line("  %sz = load i32, ptr %szSlot")
		g.line("  %argc64 = zext i32 %argc to i64")
		g.line("  %rawN = mul i64 %argc64, 4")
		g.line("  %raw = call ptr @malloc(i64 %rawN)")
		g.line("  %sz64 = zext i32 %sz to i64")
		g.line("  %buf = call ptr @malloc(i64 %sz64)")
		g.line("  %eg = call i32 @__kylix_wasi_args_get(ptr %raw, ptr %buf)")
		g.line("  %argv = call ptr @malloc(i64 %rawN)")
		g.line("  br label %aloop")
		g.line("aloop:")
		g.line("  %i = phi i32 [ 0, %entry ], [ %i2, %abody ]")
		g.line("  %more = icmp ult i32 %i, %argc")
		g.line("  br i1 %more, label %abody, label %call")
		g.line("abody:")
		g.line("  %ip = getelementptr i32, ptr %raw, i32 %i")
		g.line("  %off = load i32, ptr %ip")
		g.line("  %sp = inttoptr i32 %off to ptr")
		g.line("  %dp = getelementptr ptr, ptr %argv, i32 %i")
		g.line("  store ptr %sp, ptr %dp")
		g.line("  %i2 = add i32 %i, 1")
		g.line("  br label %aloop")
		g.line("call:")
		g.line("  %rc = call i32 @main(i32 %argc, ptr %argv)")
		g.line("  call void @__kylix_wasi_proc_exit(i32 %rc)")
		g.line("  unreachable")
	} else {
		g.line("  %rc = call i32 @main()")
		g.line("  call void @__kylix_wasi_proc_exit(i32 %rc)")
		g.line("  unreachable")
	}
	g.line("}")
	g.line("")
}
