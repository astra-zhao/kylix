package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_websocket.go — RFC 6455 WebSocket client + server for the LLVM backend.
//
// Mirrors stdlib/websocket.go (Go): handshake (client WsDial + server WsAccept),
// text frames, ping/pong auto-answer, close. A TWsConn handle is a 16-byte
// block: {i64 fd, i64 isServer}. Frame IO is exact-length recv/send on the fd
// (net's TcpRead/TcpWrite are strlen/one-shot and unsuitable for binary frames).
//
// Internal helpers (emitted once when needWebsocketHelpers is set):
//   @__kylix_ws_recvn(i32 fd, ptr buf, i64 n)   → i64 bytes read
//   @__kylix_ws_sendall(i32 fd, ptr buf, i64 n) → i64 bytes sent
//   @__kylix_ws_buildframe(i64 opcode, i64 isServer, ptr payload, i64 len, ptr out) → i64 frame length
//   @__kylix_ws_sha1 / @__kylix_ws_b64 / @__kylix_ws_rand  (stdlib_websocket_crypto.go)
//
// WsDial/WsAccept (the handshake) live in stdlib_websocket_handshake.go.

func (g *Generator) emitWebsocketCall(funcName string, args []ast.Expression) (string, string, error) {
	g.needWebsocketHelpers = true
	switch funcName {
	case "WsDial":
		g.enqueueStdlib("net", "TcpDial", "TcpDial", 0)
	}
	switch funcName {
	case "WsDial":
		return g.emitWsDialCall(args)
	case "WsAccept":
		return g.emitWsAcceptCall(args)
	case "WsSend":
		return g.emitWsSendCall(args)
	case "WsRecv":
		return g.emitWsRecvCall(args)
	case "WsClose":
		return g.emitWsCloseCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; websocket.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

func (g *Generator) emitWebsocketBody(funcName string) {
	switch funcName {
	case "WsDial":
		g.emitWsDialBody()
	case "WsAccept":
		g.emitWsAcceptBody()
	case "WsSend":
		g.emitWsSendBody()
	case "WsRecv":
		g.emitWsRecvBody()
	case "WsClose":
		g.emitWsCloseBody()
	}
}

// wsLoadFd emits a load of the fd field (offset 0) from a TWsConn handle.
func (g *Generator) wsLoadFd(wsReg string) string {
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", fd, wsReg))
	return fd
}

// wsLoadIsServer emits a load of the isServer field (offset 8).
func (g *Generator) wsLoadIsServer(wsReg string) string {
	loc := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 8", loc, wsReg))
	sv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", sv, loc))
	return sv
}

// ---- call sites: enqueue + module-level call ----

func (g *Generator) emitWsDialCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("websocket.WsDial expects 2 arguments, got %d", len(args))
	}
	addrReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	pathReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("websocket", "WsDial", "WsDial", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_websocket_WsDial(ptr %s, ptr %s)", r, addrReg, pathReg))
	return r, "ptr", nil
}

func (g *Generator) emitWsAcceptCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("websocket.WsAccept expects 1 argument, got %d", len(args))
	}
	tcpReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("websocket", "WsAccept", "WsAccept", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_websocket_WsAccept(ptr %s)", r, tcpReg))
	return r, "ptr", nil
}

func (g *Generator) emitWsSendCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("websocket.WsSend expects 2 arguments, got %d", len(args))
	}
	wsReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	msgReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("websocket", "WsSend", "WsSend", 0)
	g.line(fmt.Sprintf("  call void @__kylix_websocket_WsSend(ptr %s, ptr %s)", wsReg, msgReg))
	return "0", "void", nil
}

func (g *Generator) emitWsRecvCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("websocket.WsRecv expects 1 argument, got %d", len(args))
	}
	wsReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("websocket", "WsRecv", "WsRecv", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_websocket_WsRecv(ptr %s)", r, wsReg))
	return r, "ptr", nil
}

func (g *Generator) emitWsCloseCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("websocket.WsClose expects 1 argument, got %d", len(args))
	}
	wsReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("websocket", "WsClose", "WsClose", 0)
	g.line(fmt.Sprintf("  call void @__kylix_websocket_WsClose(ptr %s)", wsReg))
	return "0", "void", nil
}

// ---- exact-length IO helpers ----

// emitWsRecvnBody: i64 @__kylix_ws_recvn(i32 %fd, ptr %buf, i64 %n)
// recv loop until n bytes read or EOF. Returns bytes actually read.
func (g *Generator) emitWsRecvnBody() {
	g.line("define i64 @__kylix_ws_recvn(i32 %fd, ptr %buf, i64 %n) {")
	g.line("entry:")
	tSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", tSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", tSlot))
	loop := g.label()
	body := g.label()
	ok := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", loop))
	tv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", tv, tSlot))
	te := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %%n", te, tv))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", te, done, body))
	g.line(fmt.Sprintf("%s:", body))
	tv2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", tv2, tSlot))
	dst := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%buf, i64 %s", dst, tv2))
	want := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %%n, %s", want, tv2))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @recv(i32 %%fd, ptr %s, i64 %s, i32 0)", r, dst, want))
	le0 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sle i64 %s, 0", le0, r))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", le0, done, ok))
	g.line(fmt.Sprintf("%s:", ok))
	tv3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", tv3, tSlot))
	tv4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", tv4, tv3, r))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", tv4, tSlot))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", done))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", out, tSlot))
	g.line(fmt.Sprintf("  ret i64 %s", out))
	g.line("}")
	g.line("")
}

// emitWsSendallBody: i64 @__kylix_ws_sendall(i32 %fd, ptr %buf, i64 %n)
func (g *Generator) emitWsSendallBody() {
	g.line("define i64 @__kylix_ws_sendall(i32 %fd, ptr %buf, i64 %n) {")
	g.line("entry:")
	tSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", tSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", tSlot))
	loop := g.label()
	body := g.label()
	ok := g.label()
	done := g.label()
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", loop))
	tv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", tv, tSlot))
	te := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sge i64 %s, %%n", te, tv))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", te, done, body))
	g.line(fmt.Sprintf("%s:", body))
	tv2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", tv2, tSlot))
	src := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%buf, i64 %s", src, tv2))
	want := g.tmp()
	g.line(fmt.Sprintf("  %s = sub i64 %%n, %s", want, tv2))
	s := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @send(i32 %%fd, ptr %s, i64 %s, i32 0)", s, src, want))
	le0 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sle i64 %s, 0", le0, s))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", le0, done, ok))
	g.line(fmt.Sprintf("%s:", ok))
	tv3 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", tv3, tSlot))
	tv4 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", tv4, tv3, s))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", tv4, tSlot))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", done))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", out, tSlot))
	g.line(fmt.Sprintf("  ret i64 %s", out))
	g.line("}")
	g.line("")
}

// emitWsBuildframeBody: i64 @__kylix_ws_buildframe(i64 %opcode, i64 %isServer,
// ptr %payload, i64 %len, ptr %out) — writes a single WebSocket frame into
// %out (caller-sized ≥ len+15) and returns its length. Client frames get a
// random 4-byte mask and masked payload (RFC 6455 §5.3); server frames don't.
func (g *Generator) emitWsBuildframeBody() {
	g.line("define i64 @__kylix_ws_buildframe(i64 %opcode, i64 %isServer, ptr %payload, i64 %len, ptr %out) {")
	g.line("entry:")
	isSrv := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %%isServer, 0", isSrv))
	maskBit := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 0, i64 128", maskBit, isSrv))
	// out[0] = 0x80 | opcode
	b0 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 128, %%opcode", b0))
	b0t := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", b0t, b0))
	g.line(fmt.Sprintf("  store i8 %s, ptr %%out", b0t))
	hdrLenSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", hdrLenSlot))
	maskStart := g.label()
	lt126 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %%len, 126", lt126))
	c1 := g.label()
	c2 := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lt126, c1, c2))
	// len < 126: b1 = maskBit|len, hdrLen=2
	g.line(fmt.Sprintf("%s:", c1))
	b1 := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %%len", b1, maskBit))
	b1t := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", b1t, b1))
	o1 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 1", o1))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", b1t, o1))
	g.line(fmt.Sprintf("  store i64 2, ptr %s", hdrLenSlot))
	g.line(fmt.Sprintf("  br label %%%s", maskStart))
	// len 126..65535: b1 = maskBit|126 + 2-byte BE
	c2a := g.label()
	c2b := g.label()
	g.line(fmt.Sprintf("%s:", c2))
	lt65536 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ult i64 %%len, 65536", lt65536))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lt65536, c2a, c2b))
	g.line(fmt.Sprintf("%s:", c2a))
	b1a := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, 126", b1a, maskBit))
	b1at := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", b1at, b1a))
	o1a := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 1", o1a))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", b1at, o1a))
	hi := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %%len, 8", hi))
	hiv := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", hiv, hi))
	hivt := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", hivt, hiv))
	o2 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 2", o2))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", hivt, o2))
	lo := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %%len, 255", lo))
	lot := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", lot, lo))
	o3 := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 3", o3))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", lot, o3))
	g.line(fmt.Sprintf("  store i64 4, ptr %s", hdrLenSlot))
	g.line(fmt.Sprintf("  br label %%%s", maskStart))
	// len >= 65536: b1 = maskBit|127 + 8-byte BE
	g.line(fmt.Sprintf("%s:", c2b))
	b1b := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, 127", b1b, maskBit))
	b1bt := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", b1bt, b1b))
	o1b := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 1", o1b))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", b1bt, o1b))
	for j := 0; j < 8; j++ {
		sh := 8 * (7 - j)
		by := g.tmp()
		g.line(fmt.Sprintf("  %s = lshr i64 %%len, %d", by, sh))
		bv := g.tmp()
		g.line(fmt.Sprintf("  %s = and i64 %s, 255", bv, by))
		bt := g.tmp()
		g.line(fmt.Sprintf("  %s = trunc i64 %s to i8", bt, bv))
		op := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 %d", op, 2+j))
		g.line(fmt.Sprintf("  store i8 %s, ptr %s", bt, op))
	}
	g.line(fmt.Sprintf("  store i64 10, ptr %s", hdrLenSlot))
	g.line(fmt.Sprintf("  br label %%%s", maskStart))
	// mask + payload
	g.line(fmt.Sprintf("%s:", maskStart))
	hdrLen := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", hdrLen, hdrLenSlot))
	maskLen := g.tmp()
	g.line(fmt.Sprintf("  %s = select i1 %s, i64 0, i64 4", maskLen, isSrv))
	srvLbl := g.label()
	cliLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isSrv, srvLbl, cliLbl))
	// server: memcpy(out + hdrLen, payload, len)
	g.line(fmt.Sprintf("%s:", srvLbl))
	dstS := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 %s", dstS, hdrLen))
	g.needMemcpy = true
	g.line(fmt.Sprintf("  call ptr @memcpy(ptr %s, ptr %%payload, i64 %%len)", dstS))
	frameLenS := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", frameLenS, hdrLen, "%len"))
	g.line(fmt.Sprintf("  ret i64 %s", frameLenS))
	// client: 4 random mask bytes then XOR-copy payload
	g.line(fmt.Sprintf("%s:", cliLbl))
	maskPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 %s", maskPtr, hdrLen))
	g.line(fmt.Sprintf("  call void @__kylix_ws_rand(ptr %s, i64 4)", maskPtr))
	// XOR loop: out[hdrLen+4+i] = payload[i] ^ mask[i%4]
	iSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", iSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", iSlot))
	xloop := g.label()
	xbody := g.label()
	xdone := g.label()
	g.line(fmt.Sprintf("  br label %%%s", xloop))
	g.line(fmt.Sprintf("%s:", xloop))
	iv := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", iv, iSlot))
	ie := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %%len", ie, iv))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ie, xbody, xdone))
	g.line(fmt.Sprintf("%s:", xbody))
	iv2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", iv2, iSlot))
	// mask[i%4]
	mi := g.tmp()
	g.line(fmt.Sprintf("  %s = urem i64 %s, 4", mi, iv2))
	mp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", mp, maskPtr, mi))
	mb := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", mb, mp))
	// payload[i]
	pp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%payload, i64 %s", pp, iv2))
	pb := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", pb, pp))
	xb := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i8 %s, %s", xb, pb, mb))
	// out[hdrLen+4+i]
	oi := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 4", oi, hdrLen))
	oi2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", oi2, oi, iv2))
	op := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %%out, i64 %s", op, oi2))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", xb, op))
	iv3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", iv3, iv2))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", iv3, iSlot))
	g.line(fmt.Sprintf("  br label %%%s", xloop))
	g.line(fmt.Sprintf("%s:", xdone))
	frameLenC := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, %s", frameLenC, hdrLen, "%len"))
	frameLenC2 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 4", frameLenC2, frameLenC))
	g.line(fmt.Sprintf("  ret i64 %s", frameLenC2))
	g.line("}")
	g.line("")
}

// ---- WsSend / WsClose / WsRecv bodies ----

func (g *Generator) emitWsSendBody() {
	g.line("define void @__kylix_websocket_WsSend(ptr %ws, ptr %msg) {")
	g.line("entry:")
	fd := g.wsLoadFd("%ws")
	isServer := g.wsLoadIsServer("%ws")
	msgLen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%msg)", msgLen))
	frameSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 15", frameSize, msgLen))
	frame := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", frame, frameSize))
	flen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_ws_buildframe(i64 1, i64 %s, ptr %%msg, i64 %s, ptr %s)", flen, isServer, msgLen, frame))
	fdi := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fdi, fd))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_sendall(i32 %s, ptr %s, i64 %s)", fdi, frame, flen))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", frame))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

func (g *Generator) emitWsCloseBody() {
	g.line("define void @__kylix_websocket_WsClose(ptr %ws) {")
	g.line("entry:")
	fd := g.wsLoadFd("%ws")
	isServer := g.wsLoadIsServer("%ws")
	frame := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", frame))
	flen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_ws_buildframe(i64 8, i64 %s, ptr null, i64 0, ptr %s)", flen, isServer, frame))
	fdi := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fdi, fd))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_sendall(i32 %s, ptr %s, i64 %s)", fdi, frame, flen))
	g.line(fmt.Sprintf("  call i32 @close(i32 %s)", fdi))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", frame))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// emitWsRecvBody: ptr @__kylix_websocket_WsRecv(ptr %ws) — reads frames until a
// text/continuation payload, auto-answering ping with pong. Returns a
// NUL-terminated malloc'd string ("" on close/error).
func (g *Generator) emitWsRecvBody() {
	g.line("define ptr @__kylix_websocket_WsRecv(ptr %ws) {")
	g.line("entry:")
	fd := g.wsLoadFd("%ws")
	isServer := g.wsLoadIsServer("%ws")
	fdi := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fdi, fd))
	empty := g.addString("")
	emptyPtr := g.ptrTo(empty, 1)
	resSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca ptr, align 8", resSlot))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", emptyPtr, resSlot))

	loop := g.label()
	parse := g.label()
	extNot126 := g.label()
	ext2 := g.label()
	ext8 := g.label()
	maskRead := g.label()
	readMask := g.label()
	payloadRead := g.label()
	unmask := g.label()
	unmaskInit := g.label()
	unmaskCheck := g.label()
	unmaskWork := g.label()
	unmaskDone := g.label()
	notText := g.label()
	textDone := g.label()
	notPing := g.label()
	pingLbl := g.label()
	done := g.label()

	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", loop))
	hdr := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 2)", hdr))
	nh := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_ws_recvn(i32 %s, ptr %s, i64 2)", nh, fdi, hdr))
	lth := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, 2", lth, nh))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", lth, done, parse))
	g.line(fmt.Sprintf("%s:", parse))
	b0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", b0, hdr))
	b0i := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", b0i, b0))
	opcode := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 15", opcode, b0i))
	h1p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", h1p, hdr))
	b1 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", b1, h1p))
	b1i := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", b1i, b1))
	masked := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 128", masked, b1i))
	len7 := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 127", len7, b1i))
	lenSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", lenSlot))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", len7, lenSlot))
	eq126 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 126", eq126, len7))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eq126, ext2, extNot126))
	g.line(fmt.Sprintf("%s:", extNot126))
	eq127 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 127", eq127, len7))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", eq127, ext8, maskRead))
	g.line(fmt.Sprintf("%s:", ext2))
	ext2b := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 2)", ext2b))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_recvn(i32 %s, ptr %s, i64 2)", fdi, ext2b))
	e0 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", e0, ext2b))
	e0i := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", e0i, e0))
	e1p := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 1", e1p, ext2b))
	e1 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", e1, e1p))
	e1i := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i8 %s to i64", e1i, e1))
	be16a := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 8", be16a, e0i))
	be16b := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", be16b, be16a, e1i))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", be16b, lenSlot))
	g.line(fmt.Sprintf("  br label %%%s", maskRead))
	g.line(fmt.Sprintf("%s:", ext8))
	ext8b := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", ext8b))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_recvn(i32 %s, ptr %s, i64 8)", fdi, ext8b))
	acc := ""
	for j := 0; j < 8; j++ {
		ep := g.tmp()
		g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %d", ep, ext8b, j))
		eb := g.tmp()
		g.line(fmt.Sprintf("  %s = load i8, ptr %s", eb, ep))
		ei := g.tmp()
		g.line(fmt.Sprintf("  %s = zext i8 %s to i64", ei, eb))
		if j == 0 {
			acc = ei
		} else {
			sl := g.tmp()
			g.line(fmt.Sprintf("  %s = shl i64 %s, %d", sl, ei, 8*(7-j)))
			nw := g.tmp()
			g.line(fmt.Sprintf("  %s = or i64 %s, %s", nw, acc, sl))
			acc = nw
		}
	}
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", acc, lenSlot))
	g.line(fmt.Sprintf("  br label %%%s", maskRead))
	g.line(fmt.Sprintf("%s:", maskRead))
	mask := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 4)", mask))
	maskedBool := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i64 %s, 0", maskedBool, masked))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", maskedBool, readMask, payloadRead))
	g.line(fmt.Sprintf("%s:", readMask))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_recvn(i32 %s, ptr %s, i64 4)", fdi, mask))
	g.line(fmt.Sprintf("  br label %%%s", payloadRead))
	g.line(fmt.Sprintf("%s:", payloadRead))
	len64 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", len64, lenSlot))
	plSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", plSize, len64))
	payload := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", payload, plSize))
	npl := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_ws_recvn(i32 %s, ptr %s, i64 %s)", npl, fdi, payload, len64))
	ltpl := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", ltpl, npl, len64))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", ltpl, done, unmask))
	// unmask: XOR the payload if the frame was masked (client→server).
	g.line(fmt.Sprintf("%s:", unmask))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", maskedBool, unmaskInit, unmaskDone))
	g.line(fmt.Sprintf("%s:", unmaskInit))
	uiSlot := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i64, align 8", uiSlot))
	g.line(fmt.Sprintf("  store i64 0, ptr %s", uiSlot))
	g.line(fmt.Sprintf("  br label %%%s", unmaskCheck))
	g.line(fmt.Sprintf("%s:", unmaskCheck))
	ui := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", ui, uiSlot))
	uie := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp slt i64 %s, %s", uie, ui, len64))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", uie, unmaskWork, unmaskDone))
	g.line(fmt.Sprintf("%s:", unmaskWork))
	ui2 := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %s", ui2, uiSlot))
	mi := g.tmp()
	g.line(fmt.Sprintf("  %s = urem i64 %s, 4", mi, ui2))
	mp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", mp, mask, mi))
	mb := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", mb, mp))
	pp := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", pp, payload, ui2))
	pb := g.tmp()
	g.line(fmt.Sprintf("  %s = load i8, ptr %s", pb, pp))
	xb := g.tmp()
	g.line(fmt.Sprintf("  %s = xor i8 %s, %s", xb, pb, mb))
	g.line(fmt.Sprintf("  store i8 %s, ptr %s", xb, pp))
	ui3 := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %s, 1", ui3, ui2))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", ui3, uiSlot))
	g.line(fmt.Sprintf("  br label %%%s", unmaskCheck))
	// NUL-terminate + dispatch by opcode.
	g.line(fmt.Sprintf("%s:", unmaskDone))
	termP := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termP, payload, len64))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termP))
	isText1 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 1", isText1, opcode))
	isText0 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 0", isText0, opcode))
	isText := g.tmp()
	g.line(fmt.Sprintf("  %s = or i1 %s, %s", isText, isText1, isText0))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isText, textDone, notText))
	g.line(fmt.Sprintf("%s:", textDone))
	g.line(fmt.Sprintf("  store ptr %s, ptr %s", payload, resSlot))
	g.line(fmt.Sprintf("  br label %%%s", done))
	g.line(fmt.Sprintf("%s:", notText))
	isPing := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 9", isPing, opcode))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isPing, pingLbl, notPing))
	g.line(fmt.Sprintf("%s:", pingLbl))
	// answer ping with pong carrying the same payload
	pf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 16)", pf))
	pflen := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_ws_buildframe(i64 10, i64 %s, ptr %s, i64 %s, ptr %s)", pflen, isServer, payload, len64, pf))
	g.line(fmt.Sprintf("  call i64 @__kylix_ws_sendall(i32 %s, ptr %s, i64 %s)", fdi, pf, pflen))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", pf))
	g.line(fmt.Sprintf("  br label %%%s", loop))
	g.line(fmt.Sprintf("%s:", notPing))
	// opcode 10 (pong) → keep reading; anything else (close/unknown) → done.
	isPong := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, 10", isPong, opcode))
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", isPong, loop, done))
	g.line(fmt.Sprintf("%s:", done))
	out := g.tmp()
	g.line(fmt.Sprintf("  %s = load ptr, ptr %s", out, resSlot))
	g.line(fmt.Sprintf("  ret ptr %s", out))
	g.line("}")
	g.line("")
}
