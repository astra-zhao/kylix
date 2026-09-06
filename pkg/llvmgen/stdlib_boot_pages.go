package llvmgen

import (
	"fmt"

	"kylix/ast"
)

// stdlib_boot_pages.go — v0.7.0 P2 page-rendering runtime for the LLVM boot
// server. Complements stdlib_boot.go / stdlib_boot_http.go:
//
//   - emitBootResponseMethodCall: TResponse fluent methods (Html/WithCookie/
//     WithHeader/Send/StatusCode) on the 40-byte response handle.
//   - emitBootStaticBody + emitBootServeStaticBody: BootStatic(dir) registers
//     the static root; BootRun's route-miss path serves GET /static/* files
//     with MIME types (mirrors Go Router.serveStatic).
//   - emitBootFormGetBody: req.Form(name) — application/x-www-form-urlencoded
//     body lookup with %XX and '+' URL decoding.
//   - emitBootCookieGetBody: req.Cookie(name) — Cookie header pair lookup.
//
// All helpers follow the v0.6.9 bootstrap rules: allocas live in the entry
// block only, every SSA value goes through g.tmp(), and null results are
// normalized by the caller.

// bootStaticDirGlobal is the module global BootStatic(dir) writes and
// serve_static reads. Declared once per module.
const bootStaticDirGlobal = "@__kylix_boot_static_dir"

// Error-page globals (v0.7.0 P3). BootNotFoundPage/BootErrorPage store the
// custom HTML into these; BootRun's 404/500 paths fall back to the built-in
// responses when they are null (the default).
const (
	boot404PageGlobal = "@__kylix_boot_404_page"
	boot500PageGlobal = "@__kylix_boot_500_page"
)

// Response content-type constants (mirrors pkg/boot/types.go).
const (
	bootCTHtml = "text/html; charset=utf-8"
	bootCTText = "text/plain; charset=utf-8"
)

// emitBootStaticBody — void @__kylix_boot_BootStatic(ptr %dir): store the
// static root into the module global (null = static serving disabled, the
// default; Go mirrors this with Router.StaticDir == "").
func (g *Generator) emitBootStaticBody() {
	g.bootDeclareStaticDirGlobal()
	g.line("define void @__kylix_boot_BootStatic(ptr %dir) {")
	g.line("entry:")
	g.line(fmt.Sprintf("  store ptr %%dir, ptr %s", bootStaticDirGlobal))
	g.line("  ret void")
	g.line("}")
}

// bootDeclareStaticDirGlobal emits the `@__kylix_boot_static_dir` global
// declaration once per module. BootRun's route-miss path always references
// it (serve_static reads the dir even when BootStatic was never called —
// null means disabled), so emitBootServeStaticBody calls this too.
func (g *Generator) bootDeclareStaticDirGlobal() {
	if g.bootStaticDirDeclared {
		return
	}
	g.bootStaticDirDeclared = true
	g.line(fmt.Sprintf("%s = global ptr null", bootStaticDirGlobal))
}

// emitBootErrorPageBody — void @__kylix_boot_BootNotFoundPage(ptr %html) /
// void @__kylix_boot_BootErrorPage(ptr %html): store the custom page into
// the module global (v0.7.0 P3). BootRun reads these on the 404/500 paths;
// null = built-in response, the default.
func (g *Generator) emitBootErrorPageBody(fn string, global string) {
	g.bootDeclareErrorPageGlobals()
	g.line(fmt.Sprintf("define void @__kylix_boot_%s(ptr %%html) {", fn))
	g.line("entry:")
	g.line(fmt.Sprintf("  store ptr %%html, ptr %s", global))
	g.line("  ret void")
	g.line("}")
}

// bootDeclareErrorPageGlobals emits the 404/500 page globals once per module.
// BootRun always references them (same pattern as the static dir), so the
// declaration is not gated on the setters having been called.
func (g *Generator) bootDeclareErrorPageGlobals() {
	if g.bootErrorPagesDeclared {
		return
	}
	g.bootErrorPagesDeclared = true
	g.line(fmt.Sprintf("%s = global ptr null", boot404PageGlobal))
	g.line(fmt.Sprintf("%s = global ptr null", boot500PageGlobal))
}

// --- small IR helpers shared by the defines below ---

func (g *Generator) bootGepI8(base, idx string) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", r, base, idx))
	return r
}

func (g *Generator) bootLoadI8(p string) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", r, p))
	return r
}

func (g *Generator) bootLoadPtr(p string) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, p))
	return r
}

func (g *Generator) bootStrlen(s string) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", r, s))
	return r
}

// bootStoreI8AtOffAlloca stores an i8 value (register) at buf + *(i64*)offAlloca.
func (g *Generator) bootStoreI8AtOffAlloca(val, buf, offAlloca string) {
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", out, offAlloca))
	p := g.bootGepI8(buf, out)
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", val, p))
}

// emitBootHexvalBody — i64 @__kylix_boot_hexval(i8 %c): hex digit value
// ('0'-'9' → 0-9, 'a'-'f' → 10-15, 'A'-'F' → 10-15, else 0).
func (g *Generator) emitBootHexvalBody() {
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	c("define i64 @__kylix_boot_hexval(i8 %%c) {")
	c("entry:")
	c("  %%v = zext i8 %%c to i64")
	is9a := g.tmp()
	c("  %s = icmp uge i64 %%v, 48", is9a)
	is9b := g.tmp()
	c("  %s = icmp ule i64 %%v, 57", is9b)
	is9 := g.tmp()
	c("  %s = and i1 %s, %s", is9, is9a, is9b)
	r9 := g.tmp()
	c("  %s = sub i64 %%v, 48", r9)
	isaLo := g.tmp()
	c("  %s = icmp uge i64 %%v, 97", isaLo)
	isaHi := g.tmp()
	c("  %s = icmp ule i64 %%v, 102", isaHi)
	isa := g.tmp()
	c("  %s = and i1 %s, %s", isa, isaLo, isaHi)
	ra := g.tmp()
	c("  %s = sub i64 %%v, 87", ra)
	isALo := g.tmp()
	c("  %s = icmp uge i64 %%v, 65", isALo)
	isAHi := g.tmp()
	c("  %s = icmp ule i64 %%v, 70", isAHi)
	isA := g.tmp()
	c("  %s = and i1 %s, %s", isA, isALo, isAHi)
	rA := g.tmp()
	c("  %s = sub i64 %%v, 55", rA)
	s1 := g.tmp()
	c("  %s = select i1 %s, i64 %s, i64 0", s1, is9, r9)
	s2 := g.tmp()
	c("  %s = select i1 %s, i64 %s, i64 %s", s2, isa, ra, s1)
	s3 := g.tmp()
	c("  %s = select i1 %s, i64 %s, i64 %s", s3, isA, rA, s2)
	c("  ret i64 %s", s3)
	c("}")
}

// emitBootResponseMethodCall lowers `resp.<method>(args)` on a TResponse
// handle ({i64 status@0, ptr body@8, ptr ctype@16, ptr cookie@24,
// ptr xhdrs@32}, 40 bytes — see emitBootResponseBody). All methods mutate the
// handle in place and return it (fluent API, matching pkg/boot/types.go).
// v0.7.0 P2.
func (g *Generator) emitBootResponseMethodCall(handle, method string, args []ast.Expression) (string, string, error) {
	// Evaluate + coerce all args up front (Variant boxes are unboxed to ptr).
	argRegs := make([]string, 0, len(args))
	for _, a := range args {
		r, rt, err := g.emitExpr(a)
		if err != nil {
			return "", "", err
		}
		if rt == variantT {
			r, _ = g.coerceValue(r, rt, "ptr")
		}
		argRegs = append(argRegs, r)
	}

	field := func(off int64) string {
		p := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", p, handle, off))
		return p
	}
	storePtr := func(v string, off int64) {
		g.line(fmt.Sprintf("  store ptr %s, ptr %s", v, field(off)))
	}

	switch method {
	case "Html":
		// resp.Html(body): replace body + force text/html.
		if len(argRegs) >= 1 {
			storePtr(argRegs[0], 8)
		}
		storePtr(g.ptrTo(g.addString(bootCTHtml), len(bootCTHtml)), 16)
		return handle, "ptr", nil

	case "Send", "SetText":
		// resp.Send(body): replace the body, default the type to text/plain.
		if len(argRegs) >= 1 {
			storePtr(argRegs[0], 8)
		}
		ct := g.bootLoadPtr(field(16))
		ctNull := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", ctNull, ct))
		fallback := g.ptrTo(g.addString(bootCTText), len(bootCTText))
		ctSel := g.tmp()
		g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", ctSel, ctNull, fallback, ct))
		storePtr(ctSel, 16)
		return handle, "ptr", nil

	case "StatusCode":
		if len(argRegs) >= 1 {
			g.line(fmt.Sprintf("  store i64 %s, ptr %s", argRegs[0], field(0)))
		}
		return handle, "ptr", nil

	case "WithCookie":
		// resp.WithCookie(name, value): cookie slot = "name=value; Path=/".
		if len(argRegs) < 2 {
			return "", "", fmt.Errorf("TResponse.WithCookie expects 2 arguments, got %d", len(argRegs))
		}
		nameLen := g.bootStrlen(argRegs[0])
		valLen := g.bootStrlen(argRegs[1])
		cap1 := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %s", cap1, nameLen, valLen))
		cap2 := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, 24", cap2, cap1))
		buf := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, cap2))
		fmtStr := g.addString("%s=%s; Path=/")
		fmtPtr := g.ptrTo(fmtStr, len("%s=%s; Path=/"))
		g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 %s, ptr %s, ptr %s, ptr %s)",
			buf, cap2, fmtPtr, argRegs[0], argRegs[1]))
		storePtr(buf, 24)
		return handle, "ptr", nil

	case "WithHeader":
		// resp.WithHeader(k, v): append "k: v\r\n" to the xhdrs string
		// (allocated 1024 on first use; the Go side keeps a real map).
		if len(argRegs) < 2 {
			return "", "", fmt.Errorf("TResponse.WithHeader expects 2 arguments, got %d", len(argRegs))
		}
		xh := g.bootLoadPtr(field(32))
		xhNull := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", xhNull, xh))
		allocLbl := g.label()
		joinLbl := g.label()
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", xhNull, allocLbl, joinLbl))
		g.line(fmt.Sprintf("%s:", allocLbl))
		fresh := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 1024)", fresh))
		g.line(fmt.Sprintf("  store i8 0, ptr %s", fresh))
		storePtr(fresh, 32)
		g.line(fmt.Sprintf("  br label %%%s", joinLbl))
		g.line(fmt.Sprintf("%s:", joinLbl))
		xh2 := g.bootLoadPtr(field(32))
		kLen := g.bootStrlen(argRegs[0])
		vLen := g.bootStrlen(argRegs[1])
		tot := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, %s", tot, kLen, vLen))
		cap := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, 8", cap, tot))
		entry := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", entry, cap))
		fmtStr := g.addString("%s: %s\r\n")
		fmtPtr := g.ptrTo(fmtStr, 10)
		g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 %s, ptr %s, ptr %s, ptr %s)",
			entry, cap, fmtPtr, argRegs[0], argRegs[1]))
		g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", xh2, entry))
		return handle, "ptr", nil

	case "Redirect":
		// resp.Redirect(url): status = 302 and append "Location: url\r\n"
		// to xhdrs (v0.7.0 P3; mirrors Go Response.Redirect).
		if len(argRegs) < 1 {
			return "", "", fmt.Errorf("TResponse.Redirect expects 1 argument, got %d", len(argRegs))
		}
		g.line(fmt.Sprintf("  store i64 302, ptr %s", field(0)))
		xh := g.bootLoadPtr(field(32))
		xhNull := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", xhNull, xh))
		allocLbl := g.label()
		joinLbl := g.label()
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", xhNull, allocLbl, joinLbl))
		g.line(fmt.Sprintf("%s:", allocLbl))
		fresh := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 1024)", fresh))
		g.line(fmt.Sprintf("  store i8 0, ptr %s", fresh))
		storePtr(fresh, 32)
		g.line(fmt.Sprintf("  br label %%%s", joinLbl))
		g.line(fmt.Sprintf("%s:", joinLbl))
		xh2 := g.bootLoadPtr(field(32))
		uLen := g.bootStrlen(argRegs[0])
		cap := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 %s, 16", cap, uLen))
		entry := g.tmp()
		g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", entry, cap))
		fmtStr := g.addString("Location: %s\r\n")
		fmtPtr := g.ptrTo(fmtStr, 14)
		g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 %s, ptr %s, ptr %s)",
			entry, cap, fmtPtr, argRegs[0]))
		g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", xh2, entry))
		return handle, "ptr", nil

	default:
		// Unknown method — null fallback keeps the IR legal.
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; TResponse.%s stub", r, method))
		return r, "ptr", nil
	}
}

// emitBootServeStaticBody — i1 @__kylix_boot_serve_static(ptr %conn,
// ptr %method, ptr %path): attempt to serve path under the static root.
// Returns 1 when a file was served (response fully sent), 0 when the caller
// must fall through to 404. Mirrors Go Router.serveStatic: GET only,
// /static/ prefix only, no directory listing.
func (g *Generator) emitBootServeStaticBody() {
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	g.bootDeclareStaticDirGlobal()
	c("define i1 @__kylix_boot_serve_static(ptr %%conn, ptr %%method, ptr %%path) {")
	c("entry:")
	dir := g.bootLoadPtr(bootStaticDirGlobal)
	dirNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", dirNull, dir)
	no1 := g.label()
	chkGet := g.label()
	c("  br i1 %s, label %%%s, label %%%s", dirNull, no1, chkGet)
	c("%s:", no1)
	c("  ret i1 0")
	c("%s:", chkGet)
	mCmp := g.tmp()
	c("  %s = call i32 @strcmp(ptr %%method, ptr %s)", mCmp, g.ptrTo(g.addString("GET"), 4))
	isGet := g.tmp()
	c("  %s = icmp eq i32 %s, 0", isGet, mCmp)
	no2 := g.label()
	chkPfx := g.label()
	c("  br i1 %s, label %%%s, label %%%s", isGet, chkPfx, no2)
	c("%s:", no2)
	c("  ret i1 0")
	c("%s:", chkPfx)
	pfxCmp := g.tmp()
	c("  %s = call i32 @strncmp(ptr %%path, ptr %s, i64 8)", pfxCmp, g.ptrTo(g.addString("/static/"), 9))
	hasPfx := g.tmp()
	c("  %s = icmp eq i32 %s, 0", hasPfx, pfxCmp)
	no3 := g.label()
	openLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", hasPfx, openLbl, no3)
	c("%s:", no3)
	c("  ret i1 0")

	// full = dir + "/" + rel ("//" when dir ends in / — harmless for open()).
	c("%s:", openLbl)
	rel := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %%path, i64 8", rel)
	dirLen := g.bootStrlen(dir)
	relLen := g.bootStrlen(rel)
	cap1 := g.tmp()
	c("  %s = add i64 %s, %s", cap1, dirLen, relLen)
	cap2 := g.tmp()
	c("  %s = add i64 %s, 2", cap2, cap1)
	full := g.tmp()
	c("  %s = call ptr @malloc(i64 %s)", full, cap2)
	c("  call ptr @strcpy(ptr %s, ptr %s)", full, dir)
	g.bootStrcat(full, g.ptrTo(g.addString("/"), 2))
	g.bootStrcat(full, rel)
	fp := g.tmp()
	c("  %s = call ptr @fopen(ptr %s, ptr %s)", fp, full, g.ptrTo(g.addString("rb"), 3))
	fpNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", fpNull, fp)
	no4 := g.label()
	readLbl := g.label()
	c("  br i1 %s, label %%%s, label %%%s", fpNull, no4, readLbl)
	c("%s:", no4)
	c("  call void @free(ptr %s)", full)
	c("  ret i1 0")

	c("%s:", readLbl)
	c("  call i32 @fseek(ptr %s, i64 0, i32 2)", fp) // SEEK_END
	size := g.tmp()
	c("  %s = call i64 @ftell(ptr %s)", size, fp)
	c("  call i32 @fseek(ptr %s, i64 0, i32 0)", fp) // SEEK_SET
	bufCap := g.tmp()
	c("  %s = add i64 %s, 1", bufCap, size)
	buf := g.tmp()
	c("  %s = call ptr @malloc(i64 %s)", buf, bufCap)
	total := g.tmp()
	c("  %s = call i64 @fread(ptr %s, i64 1, i64 %s, ptr %s)", total, buf, size, fp)
	c("  call i32 @fclose(ptr %s)", fp)
	end := g.bootGepI8(buf, total)
	c("  store i8 0, ptr %s", end)

	// MIME by extension ("" when no extension → application/octet-stream).
	ext := g.tmp()
	c("  %s = call ptr @strrchr(ptr %s, i32 46)", ext, rel)
	extNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", extNull, ext)
	extSel := g.tmp()
	c("  %s = select i1 %s, ptr %s, ptr %s", extSel, extNull,
		g.ptrTo(g.addString(""), 1), ext)
	mime := g.emitBootMimeSelect(extSel)

	hdr := g.tmp()
	c("  %s = call ptr @malloc(i64 256)", hdr)
	fmtStr := g.addString("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %lld\r\n\r\n")
	fmtPtr := g.ptrTo(fmtStr, 61)
	c("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 256, ptr %s, ptr %s, i64 %s)",
		hdr, fmtPtr, mime, total)
	fd := g.bootConnFd("%conn")
	hdrLen := g.bootStrlen(hdr)
	c("  call i64 @send(i32 %s, ptr %s, i64 %s, i32 0)", fd, hdr, hdrLen)
	c("  call i64 @send(i32 %s, ptr %s, i64 %s, i32 0)", fd, buf, total)
	c("  call void @free(ptr %s)", hdr)
	c("  call void @free(ptr %s)", buf)
	c("  call void @free(ptr %s)", full)
	c("  ret i1 1")
	c("}")
}

// emitBootMimeSelect builds a select chain mapping an extension pointer to a
// Content-Type constant (unknown → application/octet-stream). Mirrors Go
// mimeFor in pkg/boot/router.go.
func (g *Generator) emitBootMimeSelect(ext string) string {
	mime := g.ptrTo(g.addString("application/octet-stream"), 24)
	add := func(extension, ctype string) {
		cmp := g.tmp()
		g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, ext,
			g.ptrTo(g.addString(extension), len(extension)+1)))
		eq := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
		ctypePtr := g.ptrTo(g.addString(ctype), len(ctype))
		sel := g.tmp()
		g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", sel, eq, ctypePtr, mime))
		mime = sel
	}
	add(".html", bootCTHtml)
	add(".htm", bootCTHtml)
	add(".css", "text/css; charset=utf-8")
	add(".js", "application/javascript; charset=utf-8")
	add(".json", "application/json")
	add(".png", "image/png")
	add(".jpg", "image/jpeg")
	add(".jpeg", "image/jpeg")
	add(".svg", "image/svg+xml")
	add(".ico", "image/x-icon")
	add(".txt", bootCTText)
	return mime
}

// emitBootFormGetBody — ptr @__kylix_boot_form_get(ptr %body, ptr %name):
// scan an application/x-www-form-urlencoded body ("a=1&b=2") for name and
// return the URL-decoded value (malloc'd), or null on miss. '+' decodes to
// space, %XX to the raw byte. Mirrors Go Request.Form.
//
// Segment walk (indices into %body, NUL-terminated):
//
//	seg: pos = segment start; *pos == 0 → miss
//	scankey: k from pos; '=' → haveeq (key [pos,kk), value [kk+1, terminator));
//	         '&' or 0 → noeq (skip segment)
//	advance: pos = vv + (vc == '&' ? 1 : 0)
func (g *Generator) emitBootFormGetBody() {
	// hexval helper is only used by the %XX decode path.
	g.emitBootHexvalBody()
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	c("define ptr @__kylix_boot_form_get(ptr %%body, ptr %%name) {")
	c("entry:")
	nameLen := g.bootStrlen("%name")
	allocas := []struct{ name, ty string }{
		{"%res", "ptr"}, {"%pos", "i64"}, {"%k", "i64"},
		{"%v", "i64"}, {"%j", "i64"}, {"%out", "i64"},
	}
	for _, a := range allocas {
		c("  %s = alloca %s, align 8", a.name, a.ty)
	}
	c("  store ptr null, ptr %%res")
	c("  store i64 0, ptr %%pos")
	c("  br label %%seg")

	c("seg:")
	p0 := g.tmp()
	c("  %s = load i64, ptr %%pos", p0)
	c0 := g.bootLoadI8(g.bootGepI8("%body", p0))
	segDone := g.tmp()
	c("  %s = icmp eq i8 %s, 0", segDone, c0)
	c("  br i1 %s, label %%miss, label %%scankey", segDone)

	c("scankey:")
	c("  store i64 %s, ptr %%k", p0)
	c("  br label %%keyloop")
	c("keyloop:")
	kk := g.tmp()
	c("  %s = load i64, ptr %%k", kk)
	kc := g.bootLoadI8(g.bootGepI8("%body", kk))
	isEq := g.tmp()
	c("  %s = icmp eq i8 %s, 61", isEq, kc) // '='
	c("  br i1 %s, label %%haveeq, label %%keynoteq", isEq)
	c("keynoteq:")
	isAmp := g.tmp()
	c("  %s = icmp eq i8 %s, 38", isAmp, kc) // '&'
	isEnd := g.tmp()
	c("  %s = icmp eq i8 %s, 0", isEnd, kc)
	noVal := g.tmp()
	c("  %s = or i1 %s, %s", noVal, isAmp, isEnd)
	c("  br i1 %s, label %%noeq, label %%keyadv", noVal)
	c("keyadv:")
	kNext := g.tmp()
	c("  %s = add i64 %s, 1", kNext, kk)
	c("  store i64 %s, ptr %%k", kNext)
	c("  br label %%keyloop")

	// noeq: segment has no '=' — skip; next = kk + (kc == '&' ? 1 : 0).
	c("noeq:")
	kk1 := g.tmp()
	c("  %s = add i64 %s, 1", kk1, kk)
	nx := g.tmp()
	c("  %s = select i1 %s, i64 %s, i64 %s", nx, isAmp, kk1, kk)
	c("  store i64 %s, ptr %%pos", nx)
	c("  br label %%seg")

	// haveeq: key is [pos, kk), value starts at kk+1.
	c("haveeq:")
	vs := g.tmp()
	c("  %s = add i64 %s, 1", vs, kk)
	c("  store i64 %s, ptr %%v", vs)
	c("  br label %%valloop")
	c("valloop:")
	vv := g.tmp()
	c("  %s = load i64, ptr %%v", vv)
	vc := g.bootLoadI8(g.bootGepI8("%body", vv))
	vAmp := g.tmp()
	c("  %s = icmp eq i8 %s, 38", vAmp, vc)
	vEnd0 := g.tmp()
	c("  %s = icmp eq i8 %s, 0", vEnd0, vc)
	vStop := g.tmp()
	c("  %s = or i1 %s, %s", vStop, vAmp, vEnd0)
	c("  br i1 %s, label %%valend, label %%valadv", vStop)
	c("valadv:")
	vNext := g.tmp()
	c("  %s = add i64 %s, 1", vNext, vv)
	c("  store i64 %s, ptr %%v", vNext)
	c("  br label %%valloop")

	// valend: match key [pos, kk) against name; decode value [vs, vv).
	c("valend:")
	keyLen := g.tmp()
	c("  %s = sub i64 %s, %s", keyLen, kk, p0)
	lenEq := g.tmp()
	c("  %s = icmp eq i64 %s, %s", lenEq, keyLen, nameLen)
	c("  br i1 %s, label %%cmpname, label %%advance", lenEq)
	c("cmpname:")
	keyP := g.bootGepI8("%body", p0)
	cmp := g.tmp()
	c("  %s = call i32 @strncmp(ptr %s, ptr %%name, i64 %s)", cmp, keyP, keyLen)
	keyEq := g.tmp()
	c("  %s = icmp eq i32 %s, 0", keyEq, cmp)
	c("  br i1 %s, label %%decode, label %%advance", keyEq)

	// advance: pos = vv + (vc == '&' ? 1 : 0).
	c("advance:")
	vv1 := g.tmp()
	c("  %s = add i64 %s, 1", vv1, vv)
	npos := g.tmp()
	c("  %s = select i1 %s, i64 %s, i64 %s", npos, vAmp, vv1, vv)
	c("  store i64 %s, ptr %%pos", npos)
	c("  br label %%seg")

	// decode: URL-decode [vs, vv) into malloc(vLen+1).
	c("decode:")
	vLen := g.tmp()
	c("  %s = sub i64 %s, %s", vLen, vv, vs)
	bufCap := g.tmp()
	c("  %s = add i64 %s, 1", bufCap, vLen)
	buf := g.tmp()
	c("  %s = call ptr @malloc(i64 %s)", buf, bufCap)
	c("  store i64 0, ptr %%j")
	c("  store i64 0, ptr %%out")
	c("  br label %%dloop")

	c("dloop:")
	jj := g.tmp()
	c("  %s = load i64, ptr %%j", jj)
	dDone := g.tmp()
	c("  %s = icmp sge i64 %s, %s", dDone, jj, vLen)
	c("  br i1 %s, label %%dfin, label %%dbody", dDone)
	c("dbody:")
	sum := g.tmp()
	c("  %s = add i64 %s, %s", sum, vs, jj)
	bch := g.bootLoadI8(g.bootGepI8("%body", sum))
	isPlus := g.tmp()
	c("  %s = icmp eq i8 %s, 43", isPlus, bch) // '+'
	c("  br i1 %s, label %%dplus, label %%dnotplus", isPlus)

	// dplus: '+' → ' '.
	c("dplus:")
	g.bootStoreI8AtOffAlloca("32", buf, "%out")
	jAdv1 := g.tmp()
	c("  %s = add i64 %s, 1", jAdv1, jj)
	c("  store i64 %s, ptr %%j", jAdv1)
	c("  br label %%dout")

	c("dnotplus:")
	isPct := g.tmp()
	c("  %s = icmp eq i8 %s, 37", isPct, bch) // '%'
	j2 := g.tmp()
	c("  %s = add i64 %s, 2", j2, jj)
	guard := g.tmp()
	c("  %s = icmp slt i64 %s, %s", guard, j2, vLen)
	isPct2 := g.tmp()
	c("  %s = and i1 %s, %s", isPct2, isPct, guard)
	c("  br i1 %s, label %%dpct, label %%dcopy", isPct2)

	// dpct: "%XY" → hexval(X)*16 + hexval(Y).
	c("dpct:")
	sum1 := g.tmp()
	c("  %s = add i64 %s, 1", sum1, sum)
	h1 := g.bootLoadI8(g.bootGepI8("%body", sum1))
	sum2 := g.tmp()
	c("  %s = add i64 %s, 2", sum2, sum)
	h2 := g.bootLoadI8(g.bootGepI8("%body", sum2))
	d1 := g.tmp()
	c("  %s = call i64 @__kylix_boot_hexval(i8 %s)", d1, h1)
	d2 := g.tmp()
	c("  %s = call i64 @__kylix_boot_hexval(i8 %s)", d2, h2)
	d1s := g.tmp()
	c("  %s = shl i64 %s, 4", d1s, d1)
	byteVal := g.tmp()
	c("  %s = add i64 %s, %s", byteVal, d1s, d2)
	bv8 := g.tmp()
	c("  %s = trunc i64 %s to i8", bv8, byteVal)
	g.bootStoreI8AtOffAlloca(bv8, buf, "%out")
	jAdv3 := g.tmp()
	c("  %s = add i64 %s, 3", jAdv3, jj)
	c("  store i64 %s, ptr %%j", jAdv3)
	c("  br label %%dout")

	// dcopy: literal byte.
	c("dcopy:")
	g.bootStoreI8AtOffAlloca(bch, buf, "%out")
	jAdv2 := g.tmp()
	c("  %s = add i64 %s, 1", jAdv2, jj)
	c("  store i64 %s, ptr %%j", jAdv2)
	c("  br label %%dout")

	// dout: out++; loop.
	c("dout:")
	outV := g.tmp()
	c("  %s = load i64, ptr %%out", outV)
	outV2 := g.tmp()
	c("  %s = add i64 %s, 1", outV2, outV)
	c("  store i64 %s, ptr %%out", outV2)
	c("  br label %%dloop")

	// dfin: NUL-terminate + store result.
	c("dfin:")
	g.bootStoreI8AtOffAlloca("0", buf, "%out")
	c("  store ptr %s, ptr %%res", buf)
	c("  br label %%done")
	c("miss:")
	c("  br label %%done")
	c("done:")
	r := g.bootLoadPtr("%res")
	c("  ret ptr %s", r)
	c("}")
}

// emitBootCookieGetBody — ptr @__kylix_boot_cookie_get(ptr %headers, ptr %name):
// find "Cookie:" in the raw header block, scan ';'-separated pairs and return
// the value of name (malloc'd copy), or null on miss. Works on pointers (the
// block is a strstr hit, not an index base).
func (g *Generator) emitBootCookieGetBody() {
	c := func(format string, args ...interface{}) { g.line(fmt.Sprintf(format, args...)) }
	c("define ptr @__kylix_boot_cookie_get(ptr %%headers, ptr %%name) {")
	c("entry:")
	nameLen := g.bootStrlen("%name")
	allocas := []struct{ name, ty string }{
		{"%res", "ptr"}, {"%v", "ptr"}, {"%k", "ptr"}, {"%pairStart", "ptr"},
	}
	for _, a := range allocas {
		c("  %s = alloca %s, align 8", a.name, a.ty)
	}
	c("  store ptr null, ptr %%res")
	cst := func(s string) string { return g.ptrTo(g.addString(s), len(s)+1) }
	line := g.tmp()
	c("  %s = call ptr @strstr(ptr %%headers, ptr %s)", line, cst("Cookie:"))
	lineNull := g.tmp()
	c("  %s = icmp eq ptr %s, null", lineNull, line)
	c("  br i1 %s, label %%miss, label %%found", lineNull)

	// found: skip "Cookie:" + leading spaces.
	c("found:")
	v0 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 7", v0, line)
	c("  store ptr %s, ptr %%v", v0)
	c("  br label %%skipsp")
	c("skipsp:")
	sv := g.bootLoadPtr("%v")
	sc := g.bootLoadI8(sv)
	isSp := g.tmp()
	c("  %s = icmp eq i8 %s, 32", isSp, sc) // ' '
	c("  br i1 %s, label %%skipadv, label %%pairstart", isSp)
	c("skipadv:")
	sv1 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", sv1, sv)
	c("  store ptr %s, ptr %%v", sv1)
	c("  br label %%skipsp")

	c("pairstart:")
	vCur := g.bootLoadPtr("%v")
	c("  store ptr %s, ptr %%pairStart", vCur)
	c("  br label %%pairloop")
	c("pairloop:")
	ps := g.bootLoadPtr("%pairStart")
	pc := g.bootLoadI8(ps)
	pcCr := g.tmp()
	c("  %s = icmp eq i8 %s, 13", pcCr, pc) // '\r'
	pcNul := g.tmp()
	c("  %s = icmp eq i8 %s, 0", pcNul, pc)
	pcStop := g.tmp()
	c("  %s = or i1 %s, %s", pcStop, pcCr, pcNul)
	c("  br i1 %s, label %%miss, label %%scankey", pcStop)

	// scankey: kk from ps; '=' → haveeq, ';' → next pair, else advance.
	c("scankey:")
	c("  store ptr %s, ptr %%k", ps)
	c("  br label %%keyloop")
	c("keyloop:")
	kk := g.bootLoadPtr("%k")
	kc := g.bootLoadI8(kk)
	kEq := g.tmp()
	c("  %s = icmp eq i8 %s, 61", kEq, kc) // '='
	c("  br i1 %s, label %%haveeq, label %%keynoteq", kEq)
	c("keynoteq:")
	kSemi := g.tmp()
	c("  %s = icmp eq i8 %s, 59", kSemi, kc) // ';'
	kCr := g.tmp()
	c("  %s = icmp eq i8 %s, 13", kCr, kc)
	kNul := g.tmp()
	c("  %s = icmp eq i8 %s, 0", kNul, kc)
	kStop := g.tmp()
	c("  %s = or i1 %s, %s", kStop, kCr, kNul)
	c("  br i1 %s, label %%miss, label %%keysemi", kStop)
	c("keysemi:")
	c("  br i1 %s, label %%nextpair, label %%keyadv", kSemi)
	c("keyadv:")
	kk1 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", kk1, kk)
	c("  store ptr %s, ptr %%k", kk1)
	c("  br label %%keyloop")

	// haveeq: key is [ps, kk); match name, else skip the value.
	c("haveeq:")
	kkI := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", kkI, kk)
	psI := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", psI, ps)
	keyLen := g.tmp()
	c("  %s = sub i64 %s, %s", keyLen, kkI, psI)
	lenEq := g.tmp()
	c("  %s = icmp eq i64 %s, %s", lenEq, keyLen, nameLen)
	c("  br i1 %s, label %%cmpname, label %%skipval", lenEq)
	c("cmpname:")
	cmp := g.tmp()
	c("  %s = call i32 @strncmp(ptr %s, ptr %%name, i64 %s)", cmp, ps, keyLen)
	keyEq := g.tmp()
	c("  %s = icmp eq i32 %s, 0", keyEq, cmp)
	c("  br i1 %s, label %%valscan, label %%skipval", keyEq)

	// skipval: advance past this pair's value to its ';' terminator.
	c("skipval:")
	kv1 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", kv1, kk)
	c("  store ptr %s, ptr %%v", kv1)
	c("  br label %%skiploop")
	c("skiploop:")
	sv3 := g.bootLoadPtr("%v")
	sc3 := g.bootLoadI8(sv3)
	sSemi := g.tmp()
	c("  %s = icmp eq i8 %s, 59", sSemi, sc3) // ';'
	c("  br i1 %s, label %%skipsemi, label %%skipstop", sSemi)
	c("skipstop:")
	sCr := g.tmp()
	c("  %s = icmp eq i8 %s, 13", sCr, sc3)
	sNul := g.tmp()
	c("  %s = icmp eq i8 %s, 0", sNul, sc3)
	sDone := g.tmp()
	c("  %s = or i1 %s, %s", sDone, sCr, sNul)
	c("  br i1 %s, label %%miss, label %%skipadv2", sDone)
	c("skipadv2:")
	sv4 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", sv4, sv3)
	c("  store ptr %s, ptr %%v", sv4)
	c("  br label %%skiploop")

	// valscan: value is [kk+1, ';' | '\r' | 0).
	c("valscan:")
	kv2 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", kv2, kk)
	c("  store ptr %s, ptr %%v", kv2)
	c("  br label %%valloop")
	c("valloop:")
	vv := g.bootLoadPtr("%v")
	vc := g.bootLoadI8(vv)
	vSemi := g.tmp()
	c("  %s = icmp eq i8 %s, 59", vSemi, vc)
	vCr := g.tmp()
	c("  %s = icmp eq i8 %s, 13", vCr, vc)
	vNul := g.tmp()
	c("  %s = icmp eq i8 %s, 0", vNul, vc)
	vA := g.tmp()
	c("  %s = or i1 %s, %s", vA, vSemi, vCr)
	vStop := g.tmp()
	c("  %s = or i1 %s, %s", vStop, vA, vNul)
	c("  br i1 %s, label %%valend, label %%valadv", vStop)
	c("valadv:")
	vv1 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", vv1, vv)
	c("  store ptr %s, ptr %%v", vv1)
	c("  br label %%valloop")

	// valend: copy value into a malloc'd NUL-terminated buffer.
	c("valend:")
	vvI := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", vvI, vv)
	kkI2 := g.tmp()
	c("  %s = ptrtoint ptr %s to i64", kkI2, kk)
	kStart := g.tmp()
	c("  %s = add i64 %s, 1", kStart, kkI2)
	vLen := g.tmp()
	c("  %s = sub i64 %s, %s", vLen, vvI, kStart)
	bufCap := g.tmp()
	c("  %s = add i64 %s, 1", bufCap, vLen)
	buf := g.tmp()
	c("  %s = call ptr @malloc(i64 %s)", buf, bufCap)
	kv3 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", kv3, kk)
	c("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", buf, kv3, vLen)
	bufEnd := g.bootGepI8(buf, vLen)
	c("  store i8 0, ptr %s", bufEnd)
	c("  store ptr %s, ptr %%res", buf)
	c("  br label %%done")

	// nextpair: after a non-matching key's ';' — next pair starts at kk+1.
	c("nextpair:")
	np1 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", np1, kk)
	c("  store ptr %s, ptr %%pairStart", np1)
	c("  br label %%pairloop")
	// skipsemi: after a skipped value's ';' — next pair starts at sv3+1.
	c("skipsemi:")
	ss1 := g.tmp()
	c("  %s = getelementptr inbounds i8, ptr %s, i64 1", ss1, sv3)
	c("  store ptr %s, ptr %%pairStart", ss1)
	c("  br label %%pairloop")
	c("miss:")
	c("  br label %%done")
	c("done:")
	r := g.bootLoadPtr("%res")
	c("  ret ptr %s", r)
	c("}")
}
