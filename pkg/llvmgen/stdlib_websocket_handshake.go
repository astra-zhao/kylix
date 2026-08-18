package llvmgen

import "fmt"

// stdlib_websocket_handshake.go — RFC 6455 handshake bodies for the LLVM
// backend: client (WsDial) and server (WsAccept). Both share the
// @__kylix_ws_readheaders helper (read an HTTP/1.1 header block until
// CRLFCRLF) and compute Sec-WebSocket-Accept = b64(sha1(key + GUID)).

// emitWsReadHeadersBody: ptr @__kylix_ws_readheaders(i32 %fd) — reads bytes
// until "\r\n\r\n" or a 4096-byte cap. Returns a NUL-terminated malloc'd buf.
func (g *Generator) emitWsReadHeadersBody() {
	g.line("define ptr @__kylix_ws_readheaders(i32 %fd) {")
	g.line("entry:")
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 4096)", buf))
	posSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", posSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", posSlot))
	crlf := g.addString("\r\n\r\n")
	crlfPtr := g.ptrTo(crlf, len("\r\n\r\n")+1)
	loop := g.label()
	body := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", loop))
	byteBuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 1)", byteBuf))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_ws_recvn(i32 %%fd, ptr %s, i64 1)", n, byteBuf))
	le0 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sle i64 %s, 0", le0, n))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", le0, done, body))
	g.line(fmt.Sprintf("%s:", body))
	pos := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", pos, posSlot))
	b := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", b, byteBuf))
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst, buf, pos))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", b, dst))
	pos2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", pos2, pos))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", pos2, posSlot))
	dst2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", dst2, buf, pos2))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", dst2))
	found := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", found, buf, crlfPtr))
	got := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", got, found))
	full := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, 4094", full, pos2))
	stop := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", stop, got, full))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", stop, done, loop))
	g.line(fmt.Sprintf("%s:", done))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// wsAcceptKey emits IR computing b64(sha1(key+GUID)) for a key string and
// returns the register holding the expected Sec-WebSocket-Accept string.
func (g *Generator) wsAcceptKey(keyReg string) string {
	keyLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", keyLen, keyReg))
	kpLen := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 36", kpLen, keyLen))
	kpSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", kpSize, kpLen))
	kp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", kp, kpSize))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", kp, keyReg, keyLen))
	guid := g.addString(wsWsGUID)
	guidPtr := g.ptrTo(guid, len(wsWsGUID)+1)
	kp2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", kp2, kp, keyLen))
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 36)", kp2, guidPtr))
	sha1buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 20)", sha1buf))
	// OpenSSL SHA1 (avoid the hand-written IR SHA-1's correctness bugs).
	g.needLibcrypto = true
	g.line(fmt.Sprintf("  call ptr @SHA1(ptr %s, i64 %s, ptr %s)", kp, kpLen, sha1buf))
	expect := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_ws_b64(ptr %s, i64 20)", expect, sha1buf))
	return expect
}

// wsHeaderField emits IR extracting the value of a request/response header
// field (e.g. "Sec-WebSocket-Key:") from a header buffer, NUL-terminated at
// the line end. Returns (valueReg, linePtrReg) — valueReg is a malloc'd copy
// with the CR/LF trimmed.
func (g *Generator) wsHeaderField(headerReg, field string) (string, string) {
	fs := g.addString(field)
	fsPtr := g.ptrTo(fs, len(field)+1)
	line := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", line, headerReg, fsPtr))
	val := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", val, line, len(field)))
	// skip leading spaces/tabs after the colon ("Field: value")
	trimSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", trimSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", val, trimSlot))
	trimLoop := g.label()
	trimNext := g.label()
	trimDone := g.label()
	g.line(fmt.Sprintf("  br label %%%s", trimLoop))
	g.line(fmt.Sprintf("%s:", trimLoop))
	cur := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", cur, trimSlot))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", c, cur))
	ci := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i32", ci, c))
	isSp := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 32", isSp, ci))
	isTab := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 9", isTab, ci))
	isWs := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", isWs, isSp, isTab))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isWs, trimNext, trimDone))
	g.line(fmt.Sprintf("%s:", trimNext))
	nxt := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", nxt, cur))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", nxt, trimSlot))
	g.line(fmt.Sprintf("  br label %%%s", trimLoop))
	g.line(fmt.Sprintf("%s:", trimDone))
	val2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", val2, trimSlot))
	// find line end: strchr(val, 13) ('\r') else strchr(val, 10) ('\n')
	cr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 13)", cr, val2))
	crNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", crNull, cr))
	nl := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 10)", nl, val2))
	lineEnd := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, ptr %s, ptr %s", lineEnd, crNull, nl, cr))
	// value length = lineEnd - val2
	lineEndI := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", lineEndI, lineEnd))
	valI := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", valI, val2))
	vlen := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", vlen, lineEndI, valI))
	// malloc + memcpy + NUL
	vsize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", vsize, vlen))
	vbuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", vbuf, vsize))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", vbuf, val2, vlen))
	vterm := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", vterm, vbuf, vlen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", vterm))
	return vbuf, line
}

// emitWsDialBody: ptr @__kylix_websocket_WsDial(ptr %addr, ptr %path)
// addr = "host:port". Client handshake: TCP connect → GET upgrade request →
// verify 101 + Sec-WebSocket-Accept. Returns a TWsConn handle or null.
//
// The request is built with a strcat chain rather than snprintf: LLVM -O0
// mis-handles the parameter spill for a varargs call in this large body (the
// path parameter read back as garbage), and pinning the parameters into
// explicit allocas sidesteps the same spill issue.
func (g *Generator) emitWsDialBody() {
	g.line("define ptr @__kylix_websocket_WsDial(ptr %addr, ptr %path) {")
	g.line("entry:")
	// pin params into explicit slots (avoid LLVM -O0 param-spill corruption)
	addrSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", addrSlot))
	g.line(fmt.Sprintf("  store ptr %%addr, ptr %s", addrSlot))
	pathSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", pathSlot))
	g.line(fmt.Sprintf("  store ptr %%path, ptr %s", pathSlot))
	addrLoad := func() string {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, addrSlot))
		return r
	}
	pathLoad := func() string {
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = load ptr, ptr %s", r, pathSlot))
		return r
	}
	// ---- parse host:port
	addr := addrLoad()
	colon := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strchr(ptr %s, i32 58)", colon, addr)) // ':'
	hostLen := g.tmp()
	colonI := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", colonI, colon))
	addrI := g.tmp()
	g.line(fmt.Sprintf("  %s = ptrtoint ptr %s to i64", addrI, addr))
	g.line(fmt.Sprintf("  %s = sub i64 %s, %s", hostLen, colonI, addrI))
	hostSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", hostSize, hostLen))
	host := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", host, hostSize))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %s, i64 %s)", host, addr, hostLen))
	hterm := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", hterm, host, hostLen))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", hterm))
	portStr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", portStr, colon))
	port := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @atoll(ptr %s)", port, portStr))
	// TEMP debug
	// ---- TCP connect
	conn := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpDial(ptr %s, i64 %s)", conn, host, port))
	failLbl := g.label()
	proceedLbl := g.label()
	verifyAccept := g.label()
	okLbl := g.label()
	connNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", connNull, conn))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", connNull, failLbl, proceedLbl))
	g.line(fmt.Sprintf("%s:", proceedLbl))
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", fd, conn))
	fdi := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fdi, fd))
	// ---- key = b64(16 random bytes)
	keybuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", keybuf))
	g.line(fmt.Sprintf("  call void @__kylix_ws_rand(ptr %s, i64 16)", keybuf))
	key := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_ws_b64(ptr %s, i64 16)", key, keybuf))
	// ---- build request with a strcat chain
	req := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 1024)", req))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", req))
	cat := func(s string) {
		sc := g.addString(s)
		scPtr := g.ptrTo(sc, len(s)+1)
		g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", req, scPtr))
	}
	cat("GET ")
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", req, pathLoad()))
	cat(" HTTP/1.1\r\nHost: ")
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", req, addrLoad()))
	cat("\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: ")
	g.line(fmt.Sprintf("  call ptr @strcat(ptr %s, ptr %s)", req, key))
	cat("\r\nSec-WebSocket-Version: 13\r\n\r\n")
	reqLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %s)", reqLen, req))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_sendall(i32 %s, ptr %s, i64 %s)", fdi, req, reqLen))
	// ---- read response headers
	resp := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_ws_readheaders(i32 %s)", resp, fdi))
	// ---- verify 101 + Sec-WebSocket-Accept
	st101 := g.addString(" 101 ")
	st101Ptr := g.ptrTo(st101, len(" 101 ")+1)
	stFound := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", stFound, resp, st101Ptr))
	stNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", stNull, stFound))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", stNull, failLbl, verifyAccept))
	g.line(fmt.Sprintf("%s:", verifyAccept))
	acceptVal, _ := g.wsHeaderField(resp, "Sec-WebSocket-Accept:")
	expect := g.wsAcceptKey(key)
	acChk := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @strstr(ptr %s, ptr %s)", acChk, acceptVal, expect))
	acOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne ptr %s, null", acOk, acChk))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", acOk, okLbl, failLbl))
	// ---- build handle {fd, isServer=0}
	g.line(fmt.Sprintf("%s:", okLbl))
	ws := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", ws))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", fd, ws))
	wsSv := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", wsSv, ws))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", wsSv))
	g.line(fmt.Sprintf("  ret ptr %s", ws))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line("}")
	g.line("")
}

// emitWsAcceptBody: ptr @__kylix_websocket_WsAccept(ptr %tcp) — a net TTcpConn
// handle (whose fd field is at offset 0). Server handshake: read the upgrade
// request, compute Sec-WebSocket-Accept from the client key, reply 101.
func (g *Generator) emitWsAcceptBody() {
	g.line("define ptr @__kylix_websocket_WsAccept(ptr %tcp) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%tcp", fd))
	fdi := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fdi, fd))
	// read request headers
	req := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_ws_readheaders(i32 %s)", req, fdi))
	// extract Sec-WebSocket-Key
	failLbl := g.label()
	okLbl := g.label()
	key, keyLine := g.wsHeaderField(req, "Sec-WebSocket-Key:")
	keyNull := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq ptr %s, null", keyNull, keyLine))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", keyNull, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", okLbl))
	expect := g.wsAcceptKey(key)
	// reply 101
	fmtStr := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n"
	fmtReg := g.addString(fmtStr)
	fmtPtr := g.ptrTo(fmtReg, len(fmtStr)+1)
	rbuf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 256)", rbuf))
	rl := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @snprintf(ptr %s, i64 256, ptr %s, ptr %s)", rl, rbuf, fmtPtr, expect))
	rl64 := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", rl64, rl))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_sendall(i32 %s, ptr %s, i64 %s)", fdi, rbuf, rl64))
	// build handle {fd, isServer=1}
	ws := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", ws))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", fd, ws))
	wsSv := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", wsSv, ws))
	g.line(fmt.Sprintf("  store i64 1, ptr %s", wsSv))
	g.line(fmt.Sprintf("  ret ptr %s", ws))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line("}")
	g.line("")
}
