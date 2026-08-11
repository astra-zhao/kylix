package llvmgen

import (
	"fmt"
	"strings"

	"kylix/ast"
)

// stdlib_sysutil.go — LLVM IR implementation of the `sysutil` stdlib module.
//
// Each Kylix `sysutil.Func(args)` call lowers to a `call @__kylix_sysutil_Func`
// at the call site (emitSysutilCall), with the function body emitted once at
// module end (emitSysutilBody, queued via enqueueStdlib). Bodies call libc
// directly — the LLVM backend has no Go dependency.
//
// String = ptr (null-terminated char array), matching the rest of the backend.
// Memory is malloc'd and never freed (no GC) — documented limitation, same as
// string/array heap allocation elsewhere.

// stdlibSysutilFuncSig returns the LLVM (retType, paramTypes) for a sysutil
// function, or ok=false if unknown.
func stdlibSysutilFuncSig(name string) (retType string, params []string, ok bool) {
	switch name {
	case "ReadFile":
		return "ptr", []string{"ptr"}, true // (path) → String
	case "WriteFile":
		return "void", []string{"ptr", "ptr"}, true // (path, content) → void
	case "FileExists":
		return "i1", []string{"ptr"}, true // (path) → Boolean
	case "PathJoin":
		return "ptr", nil, true // variadic (...String) → String; params built per-call
	case "PathBase":
		return "ptr", []string{"ptr"}, true // (path) → String
	// v6.2.0: file/dir/env operations (POSIX).
	case "DirExists":
		return "i1", []string{"ptr"}, true
	case "CreateDir":
		return "void", []string{"ptr"}, true
	case "DeleteFile":
		return "void", []string{"ptr"}, true
	case "AppendFile":
		return "void", []string{"ptr", "ptr"}, true
	case "CopyFile":
		return "void", []string{"ptr", "ptr"}, true
	case "GetWorkingDir":
		return "ptr", nil, true
	case "SetWorkingDir":
		return "void", []string{"ptr"}, true
	case "GetTempDir":
		return "ptr", nil, true
	case "GetFileSize":
		return "i64", []string{"ptr"}, true
	}
	return "", nil, false
}

// emitSysutilCall emits the call instruction for a sysutil function and queues
// its body for module-end emission (deduped). Returns the result register and
// LLVM type. For variadic PathJoin, params are typed per actual argument.
func (g *Generator) emitSysutilCall(funcName string, args []ast.Expression) (string, string, error) {
	retType, paramTypes, ok := stdlibSysutilFuncSig(funcName)
	if !ok {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; sysutil.%s not implemented", r, funcName))
		return r, "i64", nil
	}

	// Emit each argument expression. PathJoin is variadic — its paramTypes are
	// built from the actual arg count (all ptr / String).
	argRegs := make([]string, 0, len(args))
	argTypes := make([]string, 0, len(args))
	for i, arg := range args {
		r, t, err := g.emitExpr(arg)
		if err != nil {
			return "", "", err
		}
		// String args come through as ptr; coerce Integer/Boolean to ptr is not
		// meaningful for sysutil — all sysutil params are String (ptr). If an
		// arg isn't ptr, best-effort: keep as-is (let llc surface type errors).
		_ = i
		argRegs = append(argRegs, r)
		argTypes = append(argTypes, t)
	}
	if funcName == "PathJoin" {
		paramTypes = make([]string, len(args))
		for i := range paramTypes {
			paramTypes[i] = "ptr"
		}
	}

	// Queue the function body (deduped). PathJoin needs the arg count baked
	// into its body, so it is keyed by arg count too.
	bodyKey := funcName
	argCount := 0
	if funcName == "PathJoin" {
		argCount = len(args)
		bodyKey = fmt.Sprintf("PathJoin_%d", len(args))
	}
	g.enqueueStdlib("sysutil", funcName, bodyKey, argCount)

	fn := sysutilFuncName(funcName, len(args))
	var argList []string
	for i, r := range argRegs {
		pt := "ptr"
		if i < len(paramTypes) {
			pt = paramTypes[i]
		}
		argList = append(argList, pt+" "+r)
	}
	callArgs := strings.Join(argList, ", ")

	if retType == "void" {
		g.line(fmt.Sprintf("  call void %s(%s)", fn, callArgs))
		return "0", "void", nil
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call %s %s(%s)", r, retType, fn, callArgs))
	// FileExists returns i1; callers comparing in `if` need the value as-is.
	return r, retType, nil
}

// sysutilFuncName returns the LLVM symbol for a sysutil function. PathJoin is
// monomorphized by arg count (each arity gets its own define).
func sysutilFuncName(funcName string, argCount int) string {
	if funcName == "PathJoin" {
		return fmt.Sprintf("@__kylix_sysutil_PathJoin_%d", argCount)
	}
	return "@__kylix_sysutil_" + funcName
}

// emitSysutilBody emits the module-level `define` for one sysutil function.
// Called from emitPendingStdlib. argCount is the arity for PathJoin.
func (g *Generator) emitSysutilBody(name string, argCount int) {
	switch name {
	case "ReadFile":
		g.emitSysutilReadFile()
	case "WriteFile":
		g.emitSysutilWriteFile()
	case "FileExists":
		g.emitSysutilFileExists()
	case "PathJoin":
		g.emitSysutilPathJoin(argCount)
	case "PathBase":
		g.emitSysutilPathBase()
	case "DirExists":
		g.emitSysutilDirExists()
	case "CreateDir":
		g.emitSysutilCreateDir()
	case "DeleteFile":
		g.emitSysutilDeleteFile()
	case "AppendFile":
		g.emitSysutilAppendFile()
	case "CopyFile":
		g.emitSysutilCopyFile()
	case "GetWorkingDir":
		g.emitSysutilGetWorkingDir()
	case "SetWorkingDir":
		g.emitSysutilSetWorkingDir()
	case "GetTempDir":
		g.emitSysutilGetTempDir()
	case "GetFileSize":
		g.emitSysutilGetFileSize()
	}
}

// ---- ReadFile: ptr @__kylix_sysutil_ReadFile(ptr %path) ----
//
//	fopen(path, "r") → fp; if fp null ret null
//	fseek(fp, 0, SEEK_END=2); size = ftell(fp); fseek(fp, 0, SEEK_SET=0)
//	buf = malloc(size+1); fread(buf, 1, size, fp); buf[size] = 0; fclose(fp)
//	ret buf
func (g *Generator) emitSysutilReadFile() {
	// v6.2.0: on Windows, fopen's "r" runs text-mode translation (\n→\r\n);
	// use binary mode so ReadFile returns the file's raw bytes (POSIX-like).
	modeR := "r"
	if g.targetOS == "windows" {
		modeR = "rb"
	}
	modeStr := g.addString(modeR)
	g.line("define ptr @__kylix_sysutil_ReadFile(ptr %path) {")
	g.line("entry:")
	modeRPtr := g.ptrTo(modeStr, len(modeR)+1)
	fp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @fopen(ptr %%path, ptr %s)", fp, modeRPtr))
	// null fp → return null
	nullBlk := g.label()
	okBlk := g.label()
	g.line(fmt.Sprintf("  %%null = icmp eq ptr %s, null", fp))
	g.line(fmt.Sprintf("  br i1 %%null, label %%%s, label %%%s", nullBlk, okBlk))
	g.line(nullBlk + ":")
	g.line("  ret ptr null")
	g.line(okBlk + ":")
	// size via fseek/ftell
	size := g.tmp()
	g.line(fmt.Sprintf("  call i32 @fseek(ptr %s, i64 0, i32 2)", fp)) // SEEK_END
	g.line(fmt.Sprintf("  %s = call i64 @ftell(ptr %s)", size, fp))
	g.line(fmt.Sprintf("  call i32 @fseek(ptr %s, i64 0, i32 0)", fp)) // SEEK_SET
	// malloc(size+1)
	plus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", plus1, size))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, plus1))
	// fread(buf, 1, size, fp)
	g.line(fmt.Sprintf("  call i64 @fread(ptr %s, i64 1, i64 %s, ptr %s)", buf, size, fp))
	// null-terminate: buf[size] = 0
	endPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", endPtr, buf, size))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", endPtr))
	g.line(fmt.Sprintf("  call i32 @fclose(ptr %s)", fp))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// ---- WriteFile: void @__kylix_sysutil_WriteFile(ptr %path, ptr %content) ----
//
//	fopen(path, "w") → fp; if null ret; fputs(content, fp); fclose(fp); ret
func (g *Generator) emitSysutilWriteFile() {
	// v6.2.0: Windows "w" writes \n as \r\n (text mode) — use "wb" so the
	// content bytes are written verbatim (POSIX-like).
	modeW := "w"
	if g.targetOS == "windows" {
		modeW = "wb"
	}
	modeStr := g.addString(modeW)
	g.line("define void @__kylix_sysutil_WriteFile(ptr %path, ptr %content) {")
	g.line("entry:")
	modeWPtr := g.ptrTo(modeStr, len(modeW)+1)
	fp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @fopen(ptr %%path, ptr %s)", fp, modeWPtr))
	// if fp null, return (silently — error path simplified)
	nullBlk := g.label()
	okBlk := g.label()
	g.line(fmt.Sprintf("  %%null = icmp eq ptr %s, null", fp))
	g.line(fmt.Sprintf("  br i1 %%null, label %%%s, label %%%s", nullBlk, okBlk))
	g.line(nullBlk + ":")
	g.line("  ret void")
	g.line(okBlk + ":")
	g.line(fmt.Sprintf("  call i32 @fputs(ptr %%content, ptr %s)", fp))
	g.line(fmt.Sprintf("  call i32 @fclose(ptr %s)", fp))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- FileExists: i1 @__kylix_sysutil_FileExists(ptr %path) ----
//
//	access(path, F_OK=0) == 0 → i1
func (g *Generator) emitSysutilFileExists() {
	// v6.2.0: Windows uses _access (POSIX access is not in UCRT).
	accessFunc := "@access"
	if g.targetOS == "windows" {
		accessFunc = "@_access"
		g.line("declare i32 @_access(ptr noundef, i32 noundef)")
	}
	g.line("define i1 @__kylix_sysutil_FileExists(ptr %path) {")
	g.line("entry:")
	ret := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 %s(ptr %%path, i32 0)", ret, accessFunc)) // F_OK=0
	g.line(fmt.Sprintf("  %%ok = icmp eq i32 %s, 0", ret))
	g.line("  ret i1 %ok")
	g.line("}")
	g.line("")
}

// ---- PathJoin: ptr @__kylix_sysutil_PathJoin_<n>(ptr %p0, ptr %p1, ...) ----
//
//	malloc a buffer, strcpy first arg, then for each remaining arg strcat a "/"
//	then strcat the arg. Returns the buffer.
func (g *Generator) emitSysutilPathJoin(n int) {
	if n < 1 {
		// No args — return empty string (malloc 1 byte, store 0).
		g.line(fmt.Sprintf("define ptr @__kylix_sysutil_PathJoin_%d() {", n))
		g.line("entry:")
		buf := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 1)", buf))
		g.line(fmt.Sprintf("  store i8 0, ptr %s", buf))
		g.line(fmt.Sprintf("  ret ptr %s", buf))
		g.line("}")
		g.line("")
		return
	}
	// Build param list: ptr %p0, ptr %p1, ...
	params := make([]string, n)
	for i := 0; i < n; i++ {
		params[i] = fmt.Sprintf("ptr %%p%d", i)
	}
	g.line(fmt.Sprintf("define ptr @__kylix_sysutil_PathJoin_%d(%s) {", n, strings.Join(params, ", ")))
	g.line("entry:")
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 4096)", buf))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %%p0)", buf))
	slash := g.addString("/")
	slashPtr := g.ptrTo(slash, 2)
	for i := 1; i < n; i++ {
		g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", buf, slashPtr))
		g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %%p%d)", buf, i))
	}
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// ---- PathBase: ptr @__kylix_sysutil_PathBase(ptr %path) ----
//
//	Find the last '/'. If found, return a malloc'd copy of the substring after
//	it; otherwise return a copy of the whole string. Uses an alloca slot for
//	the loop counter (SSA can't reassign a "variable" across loop iterations).
func (g *Generator) emitSysutilPathBase() {
	g.line("define ptr @__kylix_sysutil_PathBase(ptr %path) {")
	g.line("entry:")
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%path)", ln))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64", iSlot))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", ln, iSlot))
	loopCond := g.label()
	loopBody := g.label()
	foundBlk := g.label()
	notFoundBlk := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopCond))
	g.line(loopCond + ":")
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	zero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", zero, curI))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", zero, notFoundBlk, loopBody))
	g.line(loopBody + ":")
	// i-- ; charPtr = path + (i-1); c = load i8; if c == '/' → found
	dec := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 1", dec, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", dec, iSlot))
	charPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%path, i64 %s", charPtr, dec))
	ch := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", ch, charPtr))
	isSlash := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 47", isSlash, ch)) // '/' = 47
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isSlash, foundBlk, loopCond))
	// found: result starts at the char AFTER the last '/' (path[dec+1]); copy it.
	foundStart := g.tmp()
	afterSlash := g.tmp()
	g.line(foundBlk + ":")
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", afterSlash, dec))
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%path, i64 %s", foundStart, afterSlash))
	rlen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", rlen, foundStart))
	plus1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", plus1, rlen))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, plus1))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %s)", buf, foundStart))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	// not found: return a copy of the whole path.
	g.line(notFoundBlk + ":")
	plus1b := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", plus1b, ln))
	bufb := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", bufb, plus1b))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %%path)", bufb))
	g.line(fmt.Sprintf("  ret ptr %s", bufb))
	g.line("}")
	g.line("")
}

// ---- DirExists: i1 @__kylix_sysutil_DirExists(ptr %path) ----
// opendir(path) != null (POSIX). Windows has no opendir (v6.2.0 limitation).
func (g *Generator) emitSysutilDirExists() {
	g.line("define i1 @__kylix_sysutil_DirExists(ptr %path) {")
	g.line("entry:")
	d := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @opendir(ptr %%path)", d))
	ok := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", ok, d))
	notNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", notNull, d))
	closeLbl := g.label()
	skipLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", notNull, closeLbl, skipLbl))
	g.line(closeLbl + ":")
	g.line(fmt.Sprintf("  call i32 @closedir(ptr %s)", d))
	g.line(fmt.Sprintf("  br label %%%s", skipLbl))
	g.line(skipLbl + ":")
	g.line(fmt.Sprintf("  ret i1 %s", ok))
	g.line("}")
	g.line("")
}

// ---- CreateDir: void @__kylix_sysutil_CreateDir(ptr %path) ----
func (g *Generator) emitSysutilCreateDir() {
	g.line("define void @__kylix_sysutil_CreateDir(ptr %path) {")
	g.line("entry:")
	g.line("  call i32 @mkdir(ptr %path, i32 493) ; 0755 octal")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- DeleteFile: void @__kylix_sysutil_DeleteFile(ptr %path) ----
func (g *Generator) emitSysutilDeleteFile() {
	g.line("define void @__kylix_sysutil_DeleteFile(ptr %path) {")
	g.line("entry:")
	g.line("  call i32 @remove(ptr %path)")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- AppendFile: void @__kylix_sysutil_AppendFile(ptr %path, ptr %content) ----
func (g *Generator) emitSysutilAppendFile() {
	modeA := "a"
	if g.targetOS == "windows" {
		modeA = "ab"
	}
	modeStr := g.addString(modeA)
	g.line("define void @__kylix_sysutil_AppendFile(ptr %path, ptr %content) {")
	g.line("entry:")
	fp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @fopen(ptr %%path, ptr %s)", fp, g.ptrTo(modeStr, len(modeA)+1)))
	nullBlk := g.label()
	okBlk := g.label()
	g.line(fmt.Sprintf("  %%null = icmp eq ptr %s, null", fp))
	g.line(fmt.Sprintf("  br i1 %%null, label %%%s, label %%%s", nullBlk, okBlk))
	g.line(nullBlk + ":")
	g.line("  ret void")
	g.line(okBlk + ":")
	g.line(fmt.Sprintf("  call i32 @fputs(ptr %%content, ptr %s)", fp))
	g.line(fmt.Sprintf("  call i32 @fclose(ptr %s)", fp))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- CopyFile: void @__kylix_sysutil_CopyFile(ptr %src, ptr %dst) ----
// Opens src "rb" + dst "wb" and copies in 4KB chunks (binary-safe both platforms).
func (g *Generator) emitSysutilCopyFile() {
	g.line("define void @__kylix_sysutil_CopyFile(ptr %src, ptr %dst) {")
	g.line("entry:")
	rStr := g.addString("rb")
	wStr := g.addString("wb")
	in := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @fopen(ptr %%src, ptr %s)", in, g.ptrTo(rStr, 3)))
	nullIn := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", nullIn, in))
	doneLbl := g.label()
	openLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nullIn, doneLbl, openLbl))
	g.line(openLbl + ":")
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @fopen(ptr %%dst, ptr %s)", out, g.ptrTo(wStr, 3)))
	nullOut := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", nullOut, out))
	nullOutLbl := g.label()
	loopLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", nullOut, nullOutLbl, loopLbl))
	g.line(nullOutLbl + ":")
	g.line(fmt.Sprintf("  call i32 @fclose(ptr %s)", in))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	// copy loop
	g.line(loopLbl + ":")
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 4096)", buf))
	loopHdr := g.label()
	writeLbl := g.label()
	loopEnd := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopHdr))
	g.line(loopHdr + ":")
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @fread(ptr %s, i64 1, i64 4096, ptr %s)", n, buf, in))
	done := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", done, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, loopEnd, writeLbl))
	g.line(writeLbl + ":")
	g.line(fmt.Sprintf("  call i64 @fwrite(ptr %s, i64 1, i64 %s, ptr %s)", buf, n, out))
	g.line(fmt.Sprintf("  br label %%%s", loopHdr))
	g.line(loopEnd + ":")
	g.line(fmt.Sprintf("  call void @free(ptr %s)", buf))
	g.line(fmt.Sprintf("  call i32 @fclose(ptr %s)", in))
	g.line(fmt.Sprintf("  call i32 @fclose(ptr %s)", out))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(doneLbl + ":")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- GetWorkingDir: ptr @__kylix_sysutil_GetWorkingDir() ----
func (g *Generator) emitSysutilGetWorkingDir() {
	g.line("define ptr @__kylix_sysutil_GetWorkingDir() {")
	g.line("entry:")
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 1024)", buf))
	g.line(fmt.Sprintf("  call ptr @getcwd(ptr %s, i64 1024)", buf))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// ---- SetWorkingDir: void @__kylix_sysutil_SetWorkingDir(ptr %path) ----
func (g *Generator) emitSysutilSetWorkingDir() {
	g.line("define void @__kylix_sysutil_SetWorkingDir(ptr %path) {")
	g.line("entry:")
	g.line("  call i32 @chdir(ptr %path)")
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- GetTempDir: ptr @__kylix_sysutil_GetTempDir() ----
// getenv("TMPDIR"); fallback "/tmp" (TMP/TEMP on Windows is a v6.2.0 limitation).
func (g *Generator) emitSysutilGetTempDir() {
	g.line("define ptr @__kylix_sysutil_GetTempDir() {")
	g.line("entry:")
	tmpEnv := g.addString("TMPDIR")
	env := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @getenv(ptr %s)", env, tmpEnv))
	isNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", isNull, env))
	tmpStr := g.addString("/tmp")
	tmpPtr := g.ptrTo(tmpStr, 5)
	ret := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", ret, isNull, tmpPtr, env))
	g.line(fmt.Sprintf("  ret ptr %s", ret))
	g.line("}")
	g.line("")
}

// ---- GetFileSize: i64 @__kylix_sysutil_GetFileSize(ptr %path) ----
func (g *Generator) emitSysutilGetFileSize() {
	modeR := "rb"
	modeStr := g.addString(modeR)
	g.line("define i64 @__kylix_sysutil_GetFileSize(ptr %path) {")
	g.line("entry:")
	fp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @fopen(ptr %%path, ptr %s)", fp, g.ptrTo(modeStr, 3)))
	nullBlk := g.label()
	okBlk := g.label()
	g.line(fmt.Sprintf("  %%null = icmp eq ptr %s, null", fp))
	g.line(fmt.Sprintf("  br i1 %%null, label %%%s, label %%%s", nullBlk, okBlk))
	g.line(nullBlk + ":")
	g.line("  ret i64 -1")
	g.line(okBlk + ":")
	g.line(fmt.Sprintf("  call i32 @fseek(ptr %s, i64 0, i32 2)", fp))
	size := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @ftell(ptr %s)", size, fp))
	g.line(fmt.Sprintf("  call i32 @fclose(ptr %s)", fp))
	g.line(fmt.Sprintf("  ret i64 %s", size))
	g.line("}")
	g.line("")
}
