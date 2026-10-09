package llvmgen

import (
	"fmt"

	"kylix/ast"
)

// stdlib_wasi.go — LLVM lowering of `uses wasi` on wasm32-unknown-wasi.
//
// These bodies call the preview1 import table directly. Args/Environ/Stdin/
// ReadFile/WriteFile stay unimplemented here: the Go package (pkg/wasi,
// GOOS=wasip1) covers them, and a program that needs files on the LLVM
// target gets a compile error instead of a silent zero.

func stdlibWasiFuncSig(name string) (retType string, params []string, ok bool) {
	switch name {
	case "Stdout", "Stderr":
		return "void", []string{"ptr"}, true
	case "WasiExit":
		return "void", []string{"i64"}, true
	case "ClockMonotonic", "ClockWalltime":
		return "i64", nil, true
	case "Getenv":
		return "ptr", []string{"ptr"}, true
	default:
		return "", nil, false
	}
}

func (g *Generator) emitWasiCall(funcName string, args []ast.Expression) (string, string, error) {
	retType, paramTypes, ok := stdlibWasiFuncSig(funcName)
	if !ok {
		return "", "", fmt.Errorf("wasi.%s is not in the LLVM wasm32-wasi pure-logic subset (Stdout, Stderr, Getenv, ClockMonotonic, ClockWalltime, WasiExit). File and argument helpers are implemented in pkg/wasi under GOOS=wasip1", funcName)
	}
	if !g.wasiTarget() {
		return "", "", fmt.Errorf("wasi.%s requires --target wasi/wasm32 (LLVM wasm32-unknown-wasi)", funcName)
	}
	var argList []string
	for i, arg := range args {
		r, _, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		pt := "ptr"
		if i < len(paramTypes) {
			pt = paramTypes[i]
		}
		argList = append(argList, pt+" "+r)
	}
	g.enqueueStdlib("wasi", funcName, funcName, 0)
	fn := "@__kylix_wasi_api_" + funcName
	callArgs := ""
	if len(argList) > 0 {
		callArgs = joinComma(argList)
	}
	if retType == "void" {
		g.line(fmt.Sprintf("  call void %s(%s)", fn, callArgs))
		return "0", "void", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call %s %s(%s)", r, retType, fn, callArgs))
	return r, retType, nil
}

func joinComma(parts []string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out += ", " + p
	}
	return out
}

func (g *Generator) emitWasiBody(name string) {
	switch name {
	case "Stdout":
		g.emitWasiAPIWrite(1, "Stdout")
	case "Stderr":
		g.emitWasiAPIWrite(2, "Stderr")
	case "WasiExit":
		g.line("define void @__kylix_wasi_api_WasiExit(i64 %code) {")
		g.line("entry:")
		g.line("  %c = trunc i64 %code to i32")
		g.line("  call void @exit(i32 %c)")
		g.line("  unreachable")
		g.line("}")
		g.line("")
	case "ClockMonotonic":
		g.emitWasiAPIClock(1, "ClockMonotonic", false)
	case "ClockWalltime":
		g.emitWasiAPIClock(0, "ClockWalltime", true)
	case "Getenv":
		g.emitWasiAPIGetenv()
	}
}

func (g *Generator) emitWasiAPIWrite(fd int, name string) {
	g.line(fmt.Sprintf("define void @__kylix_wasi_api_%s(ptr %%s) {", name))
	g.line("entry:")
	g.line("  %n = call i64 @strlen(ptr %s)")
	g.line("  %n32 = trunc i64 %n to i32")
	g.line(fmt.Sprintf("  call void @__kylix_wasi_write(i32 %d, ptr %%s, i32 %%n32)", fd))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

func (g *Generator) emitWasiAPIClock(id int, name string, asSeconds bool) {
	g.line(fmt.Sprintf("define i64 @__kylix_wasi_api_%s() {", name))
	g.line("entry:")
	g.line("  %slot = alloca i64, align 8")
	g.line(fmt.Sprintf("  %%e = call i32 @__kylix_wasi_clock_time_get(i32 %d, i64 1, ptr %%slot)", id))
	g.line("  %ns = load i64, ptr %slot")
	if asSeconds {
		g.line("  %sec = sdiv i64 %ns, 1000000000")
		g.line("  ret i64 %sec")
	} else {
		g.line("  ret i64 %ns")
	}
	g.line("}")
	g.line("")
}

func (g *Generator) emitWasiAPIGetenv() {
	// Scan the WASI environment block for KEY=VALUE. The block is malloc'd
	// and never freed (same ownership as other LLVM strings). Miss → "".
	g.line("define ptr @__kylix_wasi_api_Getenv(ptr %name) {")
	g.line("entry:")
	g.line("  %countSlot = alloca i32, align 4")
	g.line("  %szSlot = alloca i32, align 4")
	g.line("  %e1 = call i32 @__kylix_wasi_environ_sizes_get(ptr %countSlot, ptr %szSlot)")
	g.line("  %count = load i32, ptr %countSlot")
	g.line("  %sz = load i32, ptr %szSlot")
	g.line("  %count64 = zext i32 %count to i64")
	g.line("  %rawN = mul i64 %count64, 4")
	g.line("  %raw = call ptr @malloc(i64 %rawN)")
	g.line("  %sz64 = zext i32 %sz to i64")
	g.line("  %buf = call ptr @malloc(i64 %sz64)")
	g.line("  %e2 = call i32 @__kylix_wasi_environ_get(ptr %raw, ptr %buf)")
	g.line("  %nameLen = call i64 @strlen(ptr %name)")
	g.line("  %nameLen32 = trunc i64 %nameLen to i32")
	g.line("  br label %loop")
	g.line("loop:")
	g.line("  %i = phi i32 [ 0, %entry ], [ %i2, %next ]")
	g.line("  %more = icmp ult i32 %i, %count")
	g.line("  br i1 %more, label %body, label %miss")
	g.line("body:")
	g.line("  %ip = getelementptr i32, ptr %raw, i32 %i")
	g.line("  %off = load i32, ptr %ip")
	g.line("  %ent = inttoptr i32 %off to ptr")
	g.line("  %eq = call i32 @strncmp(ptr %ent, ptr %name, i64 %nameLen)")
	g.line("  %isEq = icmp eq i32 %eq, 0")
	g.line("  br i1 %isEq, label %checkeq, label %next")
	g.line("checkeq:")
	g.line("  %ep = getelementptr i8, ptr %ent, i32 %nameLen32")
	g.line("  %ech = load i8, ptr %ep")
	g.line("  %isSep = icmp eq i8 %ech, 61")
	g.line("  br i1 %isSep, label %hit, label %next")
	g.line("hit:")
	g.line("  %val = getelementptr i8, ptr %ep, i32 1")
	g.line("  ret ptr %val")
	g.line("next:")
	g.line("  %i2 = add i32 %i, 1")
	g.line("  br label %loop")
	g.line("miss:")
	g.line("  ret ptr @__kylix_emptystr")
	g.line("}")
	g.line("")
}
