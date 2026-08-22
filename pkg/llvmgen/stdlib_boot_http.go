package llvmgen

import (
	"fmt"

	"kylix/ast"
)

// stdlib_boot_http.go — v6.6.0: KylixBoot HTTP server for the LLVM backend.
//
// @__kylix_boot_BootRun(port) listens on a BSD socket (reusing the net module's
// TcpListen/TcpAccept/TcpClose), reads each request's HTTP headers, parses the
// request line, matches it against the route table
// (@__kylix_boot_routes, populated by @__kylix_boot_Boot<M>), dispatches to
// the generated @__kylix_boot_handler_<i> wrapper with a TRequest handle, and
// writes back an HTTP/1.1 response.
//
// TRequest handle layout (48 bytes):
//
//	0  ptr method    8  ptr path   16 ptr headers   24 ptr body
//	32 ptr params    40 i64 nparams     ; params = {ptr key, ptr value} pairs
//
// Helpers:
//
//	read_headers(conn) -> ptr                     recv until \r\n\r\n (4096 max)
//	parse_request(buf, methodSlot, pathSlot)      extract "METHOD PATH"
//	route_lookup(method, path, req) -> handler    match + fill req, or null
//	path_match(pattern, path, req) -> i1          segment match + :param fill

// bootRequestSize is the TRequest handle size in bytes.
const bootRequestSize = 48

// bootIntToStr converts an i64 register to a NUL-terminated decimal string.
func (g *Generator) bootIntToStr(v string) string {
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [24 x i8], align 1", buf))
	bufPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [24 x i8], ptr %s, i64 0, i64 0", bufPtr, buf))
	fmtStr := g.addString("%lld")
	fmtPtr := g.ptrTo(fmtStr, 5)
	g.line(fmt.Sprintf("  call i32 (ptr, i64, ptr, ...) @snprintf(ptr %s, i64 24, ptr %s, i64 %s)", bufPtr, fmtPtr, v))
	return bufPtr
}

// bootStrcat emits strcat(dst, src).
func (g *Generator) bootStrcat(dst, src string) {
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", dst, src))
}

// emitBootRunBody — the server main loop. Blocks forever on accept; every
// connection is handled serially (single-threaded, like the rest of the LLVM
// runtime).
func (g *Generator) emitBootRunBody() {
	// Reuse the net module's BSD-socket wrappers.
	g.enqueueStdlib("net", "TcpListen", "TcpListen", 0)
	g.enqueueStdlib("net", "TcpAccept", "TcpAccept", 0)
	g.enqueueStdlib("net", "TcpClose", "TcpClose", 0)
	// HTTP helpers.
	g.enqueueStdlib("boot", "readheaders", "readheaders", 0)
	g.enqueueStdlib("boot", "parsereq", "parsereq", 0)
	g.enqueueStdlib("boot", "routelookup", "routelookup", 0)

	g.line("define i64 @__kylix_boot_BootRun(i64 %port) {")
	g.line("entry:")
	listener := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpListen(i64 %%port)", listener))
	loopLbl := g.label()
	closeLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	conn := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpAccept(ptr %s)", conn, listener))
	headers := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_read_headers(ptr %s)", headers, conn))
	// method/path buffers
	methodBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 32)", methodBuf))
	pathBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 512)", pathBuf))
	g.line(fmt.Sprintf("  call void @__kylix_boot_parse_request(ptr %s, ptr %s, ptr %s)", headers, methodBuf, pathBuf))
	// req handle + route lookup
	req := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %d)", req, bootRequestSize))
	handler := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_boot_route_lookup(ptr %s, ptr %s, ptr %s)", handler, methodBuf, pathBuf, req))
	handlerNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", handlerNull, handler))
	doServeLbl := g.label()
	do404Lbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", handlerNull, do404Lbl, doServeLbl))

	// ---- serve: fill headers/body, dispatch, build+send response.
	g.line(fmt.Sprintf("%s:", doServeLbl))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", headers, g.bootReqField(req, 16)))
	// Body: read Content-Length bytes after the header block (best-effort;
	// the header buffer holds whatever followed the request line).
	g.line(fmt.Sprintf("  store ptr null, ptr %s", g.bootReqField(req, 24)))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr %s(ptr %s)", res, handler, req))
	status := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", status, res))
	bodyField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", bodyField, res))
	body := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", body, bodyField))
	respBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 2048)", respBuf))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", respBuf))
	g.line(fmt.Sprintf("  call ptr @strcpy(ptr %s, ptr %s)", respBuf, g.ptrTo(g.addString("HTTP/1.1 "), 10)))
	g.bootStrcat(respBuf, g.bootIntToStr(status))
	g.bootStrcat(respBuf, g.ptrTo(g.addString(" "), 2))
	g.bootStrcat(respBuf, g.bootReasonPhrase(status))
	g.bootStrcat(respBuf, g.ptrTo(g.addString("\r\nContent-Length: "), 19))
	bodyLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", bodyLen, body))
	g.bootStrcat(respBuf, g.bootIntToStr(bodyLen))
	g.bootStrcat(respBuf, g.ptrTo(g.addString("\r\n\r\n"), 5))
	g.bootStrcat(respBuf, body)
	g.emitBootSend(conn, respBuf)
	g.line(fmt.Sprintf("  br label %%%s", closeLbl))

	// ---- 404
	g.line(fmt.Sprintf("%s:", do404Lbl))
	g.line(fmt.Sprintf("  call i32 @send(i32 %s, ptr %s, i64 %d, i32 0)",
		g.bootConnFd(conn), g.ptrTo(g.addString("HTTP/1.1 404 Not Found\r\nContent-Length: 9\r\n\r\nNot Found"), 52), 52))
	g.line(fmt.Sprintf("  br label %%%s", closeLbl))

	// ---- close + loop
	g.line(fmt.Sprintf("%s:", closeLbl))
	g.line(fmt.Sprintf("  call void @__kylix_net_TcpClose(ptr %s)", conn))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line("}")
	g.line("")
}

// bootConnFd loads the fd from a net TTcpConn handle (malloc'd {i32 fd}).
func (g *Generator) bootConnFd(conn string) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load i32, ptr %s", r, conn))
	return r
}

// emitBootSend sends respBuf over conn via send(fd, buf, strlen, 0).
func (g *Generator) emitBootSend(conn, buf string) {
	fd := g.bootConnFd(conn)
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", ln, buf))
	g.line(fmt.Sprintf("  call i64 @send(i32 %s, ptr %s, i64 %s, i32 0)", fd, buf, ln))
}

// bootReqField returns the address of TRequest handle field at byte offset.
func (g *Generator) bootReqField(req string, off int) string {
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", p, req, off))
	return p
}

// bootReasonPhrase maps a status code to its HTTP reason phrase (200 → "OK",
// 404 → "Not Found", everything else → "OK").
func (g *Generator) bootReasonPhrase(status string) string {
	is404 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 404", is404, status))
	okPtr := g.ptrTo(g.addString("OK"), 3)
	nfPtr := g.ptrTo(g.addString("Not Found"), 10)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", r, is404, nfPtr, okPtr))
	return r
}

// emitBootReadHeadersBody — ptr @__kylix_boot_read_headers(ptr %conn).
// Reads bytes one at a time until \r\n\r\n (or 4095 bytes), NUL-terminated.
func (g *Generator) emitBootReadHeadersBody() {
	g.line("define ptr @__kylix_boot_read_headers(ptr %conn) {")
	g.line("entry:")
	fd := g.bootConnFd("%conn")
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 4096)", buf))
	posSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", posSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", posSlot))
	loopLbl := g.label()
	gotLbl := g.label()
	crlfLbl := g.label()
	exitLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	pos := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pos, posSlot))
	full := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 4095", full, pos))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", full, exitLbl, gotLbl))
	g.line(fmt.Sprintf("%s:", gotLbl))
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
	rc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @recv(i32 %s, ptr %s, i64 1, i32 0)", rc, fd, dst))
	rcPos := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i64 %s, 0", rcPos, rc))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", rcPos, crlfLbl, exitLbl))
	// \r\n\r\n check at pos2-4..pos2-1
	g.line(fmt.Sprintf("%s:", crlfLbl))
	pos2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", pos2, pos))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", pos2, posSlot))
	ge4 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 4", ge4, pos2))
	checkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ge4, checkLbl, loopLbl))
	g.line(fmt.Sprintf("%s:", checkLbl))
	off4 := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 4", off4, pos2))
	b1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", b1, buf, off4))
	c1 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c1, b1))
	isCR1 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 13", isCR1, c1))
	g.line(fmt.Sprintf("  br i1 %s, label %%nl1, label %%%s", isCR1, loopLbl))
	g.line("nl1:")
	off3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", off3, off4))
	b2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", b2, buf, off3))
	c2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c2, b2))
	isLF1 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 10", isLF1, c2))
	g.line(fmt.Sprintf("  br i1 %s, label %%cr2, label %%%s", isLF1, loopLbl))
	g.line("cr2:")
	off2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", off2, off3))
	b3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", b3, buf, off2))
	c3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c3, b3))
	isCR2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 13", isCR2, c3))
	g.line(fmt.Sprintf("  br i1 %s, label %%nl2, label %%%s", isCR2, loopLbl))
	g.line("nl2:")
	off1 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", off1, off2))
	b4 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", b4, buf, off1))
	c4 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c4, b4))
	isLF2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 10", isLF2, c4))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isLF2, exitLbl, loopLbl))
	g.line(fmt.Sprintf("%s:", exitLbl))
	finalPos := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", finalPos, posSlot))
	termPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, buf, finalPos))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// emitBootParseRequestBody — void @__kylix_boot_parse_request(ptr %buf,
// ptr %methodSlot, ptr %pathSlot). Copies "METHOD" into methodSlot and the
// request-target into pathSlot (up to the next space or EOL).
func (g *Generator) emitBootParseRequestBody() {
	g.line("define void @__kylix_boot_parse_request(ptr %buf, ptr %methodSlot, ptr %pathSlot) {")
	g.line("entry:")
	sp1 := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %%buf, i32 32)", sp1))
	// method length = sp1 - buf
	sp1Addr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", sp1Addr, sp1))
	bufAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %%buf to i64", bufAddr))
	mLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", mLen, sp1Addr, bufAddr))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %%methodSlot, ptr %%buf, i64 %s)", mLen))
	mTerm := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%methodSlot, i64 %s", mTerm, mLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", mTerm))
	// path = sp1+1 .. next space
	sp1p1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", sp1p1, sp1))
	sp2 := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 32)", sp2, sp1p1))
	sp2Addr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", sp2Addr, sp2))
	sp1p1Addr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", sp1p1Addr, sp1p1))
	pLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", pLen, sp2Addr, sp1p1Addr))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %%pathSlot, ptr %s, i64 %s)", sp1p1, pLen))
	pTerm := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%pathSlot, i64 %s", pTerm, pLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", pTerm))
	// Strip the query string (?k=v...) — route matching is path-only.
	qMark := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %%pathSlot, i32 63)", qMark))
	qNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", qNull, qMark))
	pDoneLbl := g.label()
	pStripLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", qNull, pDoneLbl, pStripLbl))
	g.line(fmt.Sprintf("%s:", pStripLbl))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", qMark))
	g.line(fmt.Sprintf("  br label %%%s", pDoneLbl))
	g.line(fmt.Sprintf("%s:", pDoneLbl))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitBootRouteLookupBody — ptr @__kylix_boot_route_lookup(ptr %method,
// ptr %path, ptr %req). Fills method/path/params on the request handle and
// returns the matching handler (or null). Linear scan of the route table.
func (g *Generator) emitBootRouteLookupBody() {
	g.enqueueStdlib("boot", "pathmatch", "pathmatch", 0)
	g.line("define ptr @__kylix_boot_route_lookup(ptr %method, ptr %path, ptr %req) {")
	g.line("entry:")
	// req->method/path, params buffer, nparams=0
	g.line(fmt.Sprintf("  store ptr %%method, ptr %s", g.bootReqField("%req", 0)))
	g.line(fmt.Sprintf("  store ptr %%path, ptr %s", g.bootReqField("%req", 8)))
	params := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 128)", params))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", params, g.bootReqField("%req", 32)))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", g.bootReqField("%req", 40)))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr @__kylix_boot_nroutes", n))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	loopLbl := g.label()
	nextLbl := g.label()
	notfoundLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	done := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", done, curI, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", done, notfoundLbl, nextLbl))
	g.line(fmt.Sprintf("%s:", nextLbl))
	slot := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [64 x { ptr, ptr, ptr }], ptr @__kylix_boot_routes, i64 0, i64 %s", slot, curI))
	rmethod := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 0", rmethod, slot))
	rmv := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", rmv, rmethod))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %%method, ptr %s)", cmp, rmv))
	neq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 0", neq, cmp))
	matchLbl := g.label()
	noMatchLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", neq, noMatchLbl, matchLbl))
	g.line(fmt.Sprintf("%s:", matchLbl))
	rpath := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 1", rpath, slot))
	rpv := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", rpv, rpath))
	matched := g.tmp()
	g.line(fmt.Sprintf("  %s = call i1 @__kylix_boot_path_match(ptr %s, ptr %%path, ptr %%req)", matched, rpv))
	matchedLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", matched, matchedLbl, noMatchLbl))
	g.line(fmt.Sprintf("%s:", matchedLbl))
	rhandler := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds { ptr, ptr, ptr }, ptr %s, i32 0, i32 2", rhandler, slot))
	rhv := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", rhv, rhandler))
	g.line(fmt.Sprintf("  ret ptr %s", rhv))
	g.line(fmt.Sprintf("%s:", noMatchLbl))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", notfoundLbl))
	g.line("  ret ptr null")
	g.line("}")
	g.line("")
}

// emitBootPathMatchBody — i1 @__kylix_boot_path_match(ptr %pattern, ptr %path,
// ptr %req). Segments by '/'; a pattern segment starting with ':' captures the
// path segment into the request params (max 8 params). Returns true on a full
// match.
func (g *Generator) emitBootPathMatchBody() {
	g.line("define i1 @__kylix_boot_path_match(ptr %pattern, ptr %path, ptr %req) {")
	g.line("entry:")
	// Both pattern and path start with '/' (Pascal route paths); skip it.
	pIdx := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", pIdx))
	g.line(fmt.Sprintf("  store i64 1, ptr %s", pIdx))
	xIdx := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", xIdx))
	g.line(fmt.Sprintf("  store i64 1, ptr %s", xIdx))
	loopLbl := g.label()
	falseLbl := g.label()
	doneLbl := g.label()
	advLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	pi := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pi, pIdx))
	xi := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", xi, xIdx))
	pCur := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%pattern, i64 %s", pCur, pi))
	xCur := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%path, i64 %s", xCur, xi))
	pSlash := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 47)", pSlash, pCur))
	xSlash := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 47)", xSlash, xCur))
	// segment length: pLen = pSlash ? pSlash-pCur : strlen(pCur)
	pSlashNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", pSlashNull, pSlash))
	pCurAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", pCurAddr, pCur))
	pSlashAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", pSlashAddr, pSlash))
	pSeg := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", pSeg, pSlashAddr, pCurAddr))
	pRest := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", pRest, pCur))
	pLen := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", pLen, pSlashNull, pRest, pSeg))
	xSlashNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", xSlashNull, xSlash))
	xCurAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", xCurAddr, xCur))
	xSlashAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", xSlashAddr, xSlash))
	xSeg := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", xSeg, xSlashAddr, xCurAddr))
	xRest := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", xRest, xCur))
	xLen := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", xLen, xSlashNull, xRest, xSeg))
	// both consumed → done
	pZero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", pZero, pLen))
	xZero := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", xZero, xLen))
	bothEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = and i1 %s, %s", bothEnd, pZero, xZero))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk1", bothEnd, doneLbl))
	g.line("chk1:")
	oneEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i1 %s, %s", oneEnd, pZero, xZero))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%chk_param", oneEnd, falseLbl))
	g.line("chk_param:")
	firstC := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", firstC, pCur))
	isColon := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 58", isColon, firstC))
	paramLbl := g.label()
	litLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isColon, paramLbl, litLbl))
	// literal segment: lengths equal && strncmp equal
	g.line(fmt.Sprintf("%s:", litLbl))
	lenEq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", lenEq, pLen, xLen))
	lenEqLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lenEq, lenEqLbl, falseLbl))
	g.line(fmt.Sprintf("%s:", lenEqLbl))
	sc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strncmp(ptr %s, ptr %s, i64 %s)", sc, pCur, xCur, pLen))
	scEq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", scEq, sc))
	litOkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", scEq, litOkLbl, falseLbl))
	g.line(fmt.Sprintf("%s:", litOkLbl))
	g.line(fmt.Sprintf("  br label %%%s", advLbl))
	// param segment: value must be non-empty; capture key=pCur+1, value=xCur
	g.line(fmt.Sprintf("%s:", paramLbl))
	g.line(fmt.Sprintf("  br i1 %s, label %%param_ok, label %%%s", g.icmpSgt64(xLen, 0), falseLbl))
	g.line("param_ok:")
	npPtr := g.bootReqField("%req", 40)
	np := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", np, npPtr))
	cap8 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 8", cap8, np))
	capLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%param_store", cap8, capLbl))
	g.line("param_store:")
	paramsPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", paramsPtr, g.bootReqField("%req", 32)))
	entryOff := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 16", entryOff, np))
	entry := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", entry, paramsPtr, entryOff))
	keyLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, 1", keyLen, pLen))
	keyBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 64)", keyBuf))
	g.needMemcpy = true
	keySrc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", keySrc, pCur))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", keyBuf, keySrc, keyLen))
	kTerm := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", kTerm, keyBuf, keyLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", kTerm))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", keyBuf, entry))
	valBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 256)", valBuf))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", valBuf, xCur, xLen))
	vTerm := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", vTerm, valBuf, xLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", vTerm))
	valField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", valField, entry))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", valBuf, valField))
	npNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", npNext, np))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", npNext, npPtr))
	g.line(fmt.Sprintf("  br label %%%s", capLbl))
	g.line(fmt.Sprintf("%s:", capLbl))
	g.line(fmt.Sprintf("  br label %%%s", advLbl))
	// advance: pIdx += pLen + (pSlash?1:0); xIdx += xLen + (xSlash?1:0)
	g.line(fmt.Sprintf("%s:", advLbl))
	pInc := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", pInc, pLen))
	g.line(fmt.Sprintf("  br i1 %s, label %%p_no, label %%p_yes", pSlashNull))
	g.line("p_yes:")
	pYes := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", pYes, pi, pInc))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", pYes, pIdx))
	g.line(fmt.Sprintf("  br label %%p_after"))
	g.line("p_no:")
	pNo := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", pNo, pi, pLen))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", pNo, pIdx))
	g.line(fmt.Sprintf("  br label %%p_after"))
	g.line("p_after:")
	xInc := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", xInc, xLen))
	g.line(fmt.Sprintf("  br i1 %s, label %%x_no, label %%x_yes", xSlashNull))
	g.line("x_yes:")
	xYes := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", xYes, xi, xInc))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", xYes, xIdx))
	g.line(fmt.Sprintf("  br label %%x_after"))
	g.line("x_no:")
	xNo := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", xNo, xi, xLen))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", xNo, xIdx))
	g.line(fmt.Sprintf("  br label %%x_after"))
	g.line("x_after:")
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", falseLbl))
	g.line("  ret i1 false")
	g.line(fmt.Sprintf("%s:", doneLbl))
	g.line("  ret i1 true")
	g.line("}")
	g.line("")
}

func (g *Generator) icmpSgt64(a string, v int64) string {
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sgt i64 %s, %d", r, a, v))
	return r
}

// emitBootRequestMethodCall lowers req.Param/Query/Header/Body on a TRequest
// handle (v6.6.0). The lookup IR is emitted INLINE at the call site (it runs
// inside the caller's function), so results go through an alloca slot and are
// returned as a register — never `ret` (that would return from the caller).
func (g *Generator) emitBootRequestMethodCall(req, method string, args []ast.Expression) (string, string, error) {
	switch method {
	case "Param":
		return g.emitBootReqParam(req, args)
	case "Header":
		return g.emitBootReqHeader(req, args)
	case "Query":
		return g.emitBootReqQuery(req, args)
	case "Body", "GetBody":
		return g.emitBootReqBody(req, args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = inttoptr i64 0 to ptr ; TRequest.%s stub", r, method))
		return r, "ptr", nil
	}
}

// emitBootReqParam — req.Param(name): scan the {key,value} params array.
func (g *Generator) emitBootReqParam(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TRequest.Param expects 1 argument, got %d", len(args))
	}
	nameReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", resSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", g.ptrTo(g.addString(""), 1), resSlot))
	np := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", np, g.bootReqField(req, 40)))
	params := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", params, g.bootReqField(req, 32)))
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	loopLbl := g.label()
	foundLbl := g.label()
	nextLbl := g.label()
	doneLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	curI := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", curI, iSlot))
	dv := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %s", dv, curI, np))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", dv, doneLbl, foundLbl))
	g.line(fmt.Sprintf("%s:", foundLbl))
	off := g.tmp()
	g.line(fmt.Sprintf("  %s = mul i64 %s, 16", off, curI))
	entry := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", entry, params, off))
	key := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", key, entry))
	cmp := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strcmp(ptr %s, ptr %s)", cmp, key, nameReg))
	eq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", eq, cmp))
	g.line(fmt.Sprintf("  br i1 %s, label %%hit, label %%%s", eq, nextLbl))
	g.line("hit:")
	valField := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", valField, entry))
	val := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", val, valField))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", val, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", nextLbl))
	iNext := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iNext, curI))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iNext, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", res, resSlot))
	return res, "ptr", nil
}

// emitBootReqHeader — req.Header(name): find "Name:" in the header block and
// return the trimmed value (up to \r), else "".
func (g *Generator) emitBootReqHeader(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TRequest.Header expects 1 argument, got %d", len(args))
	}
	nameReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", resSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", g.ptrTo(g.addString(""), 1), resSlot))
	headers := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", headers, g.bootReqField(req, 16)))
	p := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", p, headers, nameReg))
	pNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", pNull, p))
	doneLbl := g.label()
	haveLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", pNull, doneLbl, haveLbl))
	g.line(fmt.Sprintf("%s:", haveLbl))
	colon := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 58)", colon, p))
	// valStart lives in a slot so the trim loop can advance it.
	valSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", valSlot))
	val0 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", val0, colon))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", val0, valSlot))
	trimLbl := g.label()
	trimDoneLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", trimLbl))
	g.line(fmt.Sprintf("%s:", trimLbl))
	valStart := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", valStart, valSlot))
	sc := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", sc, valStart))
	isSp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 32", isSp, sc))
	isTb := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i8 %s, 9", isTb, sc))
	isWs := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", isWs, isSp, isTb))
	g.line(fmt.Sprintf("  br i1 %s, label %%sp, label %%%s", isWs, trimDoneLbl))
	g.line("sp:")
	vs := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", vs, valStart))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", vs, valSlot))
	g.line(fmt.Sprintf("  br label %%%s", trimLbl))
	g.line(fmt.Sprintf("%s:", trimDoneLbl))
	cr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 13)", cr, valStart))
	crAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", crAddr, cr))
	vsAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", vsAddr, valStart))
	vLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", vLen, crAddr, vsAddr))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 128)", buf))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", buf, valStart, vLen))
	term := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", term, buf, vLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", term))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", buf, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", res, resSlot))
	return res, "ptr", nil
}

// emitBootReqQuery — req.Query(name): parse `?k=v&...` from the request path.
func (g *Generator) emitBootReqQuery(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("TRequest.Query expects 1 argument, got %d", len(args))
	}
	nameReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", resSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", g.ptrTo(g.addString(""), 1), resSlot))
	// The request path stored on the handle is query-stripped (route matching
	// is path-only), so parse the query from the raw header block instead —
	// its first line is "METHOD /path?query HTTP/1.1".
	headers := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", headers, g.bootReqField(req, 16)))
	q := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 63)", q, headers))
	qNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", qNull, q))
	doneLbl := g.label()
	scanLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", qNull, doneLbl, scanLbl))
	g.line(fmt.Sprintf("%s:", scanLbl))
	cur := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", cur, q))
	curSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", curSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", cur, curSlot))
	loopLbl := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", loopLbl))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", c, curSlot))
	amp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 38)", amp, c))
	eqSign := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 61)", eqSign, c))
	eqNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", eqNull, eqSign))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%have_eq", eqNull, doneLbl))
	g.line("have_eq:")
	eqAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", eqAddr, eqSign))
	cAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", cAddr, c))
	keyLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", keyLen, eqAddr, cAddr))
	nameLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", nameLen, nameReg))
	lenEq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, %s", lenEq, keyLen, nameLen))
	lenEqLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%key_mismatch", lenEq, lenEqLbl))
	g.line(fmt.Sprintf("%s:", lenEqLbl))
	sc := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @strncmp(ptr %s, ptr %s, i64 %s)", sc, c, nameReg, keyLen))
	scEq := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", scEq, sc))
	g.line(fmt.Sprintf("  br i1 %s, label %%hit, label %%%s", scEq, keyMismatchLbl()))
	g.line("hit:")
	valStart := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", valStart, eqSign))
	ampNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", ampNull, amp))
	ampAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", ampAddr, amp))
	vsAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", vsAddr, valStart))
	ampLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", ampLen, ampAddr, vsAddr))
	vsLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", vsLen, valStart))
	// The value ends at the next '&', or — inside the raw header block — at
	// the next space (start of " HTTP/1.1") or CR (end of request line).
	spQ := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 32)", spQ, valStart))
	spNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", spNull, spQ))
	spAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", spAddr, spQ))
	spLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", spLen, spAddr, vsAddr))
	crQ := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 13)", crQ, valStart))
	crNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", crNull, crQ))
	crAddr := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", crAddr, crQ))
	crLen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", crLen, crAddr, vsAddr))
	// vLen = ampNull ? (spNull ? (crNull ? strlen : crLen) : spLen) : ampLen
	selCR := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", selCR, crNull, vsLen, crLen))
	selSP := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", selSP, spNull, selCR, spLen))
	vLen := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 %s, i64 %s", vLen, ampNull, selSP, ampLen))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 128)", buf))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", buf, valStart, vLen))
	term := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", term, buf, vLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", term))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", buf, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", doneLbl))
	g.line(keyMismatchLbl() + ":")
	ampNull2 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", ampNull2, amp))
	advLbl := g.label()
	// No more '&' → no further pairs; stop scanning.
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ampNull2, doneLbl, advLbl))
	g.line(fmt.Sprintf("%s:", advLbl))
	adv := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", adv, amp))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", adv, curSlot))
	g.line(fmt.Sprintf("  br label %%%s", loopLbl))
	g.line(fmt.Sprintf("%s:", doneLbl))
	res := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", res, resSlot))
	return res, "ptr", nil
}

func keyMismatchLbl() string { return "key_mismatch" }

// emitBootReqBody — req.Body: the raw request body pointer (may be null).
func (g *Generator) emitBootReqBody(req string, args []ast.Expression) (string, string, error) {
	if len(args) != 0 {
		return "", "", fmt.Errorf("TRequest.Body expects 0 arguments, got %d", len(args))
	}
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, g.bootReqField(req, 24)))
	return r, "ptr", nil
}
