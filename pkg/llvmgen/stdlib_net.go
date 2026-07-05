package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_net.go — LLVM IR implementation for the `net` stdlib module (TCP).
//
// TTcpConn / TTcpListener are heap-allocated 8-byte cells holding a single
// i32 file descriptor (the OS socket fd). Mirrors the Go-backend stdlib/net.go
// surface for the TCP subset that example55 (websocket) and example54 (http)
// depend on:
//
//   TcpDial(host, port)   -> ptr (TTcpConn)   socket()+connect()
//   TcpWrite(c, data)     -> i64 (bytes written) send()
//   TcpRead(c, n)         -> ptr (String)     recv()
//   TcpClose(c)           -> void             close()
//   TcpListen(port)       -> ptr (TTcpListener) socket()+bind()+listen()
//   TcpAccept(l)          -> ptr (TTcpConn)   accept()
//   TcpListenerClose(l)   -> void             close()
//
// UDP / DNS lookup are not implemented here (no example currently uses them
// on the LLVM path) and fall through to the not-implemented stub.
//
// Address handling: TcpDial uses inet_pton(AF_INET=2) so the host must be a
// dotted-quad IPv4 literal (e.g. "127.0.0.1"). Hostname resolution
// (gethostbyname) is deferred — keeps the first cut small and matches the
// tutorial examples which all use loopback.

// emitNetCall dispatches a `net.Func(args)` / bare `Func(args)` call.
func (g *Generator) emitNetCall(funcName string, args []ast.Expression) (string, string, error) {
	switch funcName {
	case "TcpDial":
		return g.emitNetTcpDialCall(args)
	case "TcpWrite":
		return g.emitNetTcpWriteCall(args)
	case "TcpRead":
		return g.emitNetTcpReadCall(args)
	case "TcpClose":
		return g.emitNetTcpCloseCall(args)
	case "TcpListen":
		return g.emitNetTcpListenCall(args)
	case "TcpAccept":
		return g.emitNetTcpAcceptCall(args)
	case "TcpListenerClose":
		return g.emitNetTcpListenerCloseCall(args)
	default:
		r := g.tmp()
		g.line(fmt.Sprintf("  %s = add i64 0, 0 ; net.%s not implemented", r, funcName))
		return r, "i64", nil
	}
}

// emitNetBody dispatches the deferred body emitter.
func (g *Generator) emitNetBody(funcName string) {
	switch funcName {
	case "TcpDial":
		g.emitNetTcpDialBody()
	case "TcpWrite":
		g.emitNetTcpWriteBody()
	case "TcpRead":
		g.emitNetTcpReadBody()
	case "TcpClose":
		g.emitNetTcpCloseBody()
	case "TcpListen":
		g.emitNetTcpListenBody()
	case "TcpAccept":
		g.emitNetTcpAcceptBody()
	case "TcpListenerClose":
		g.emitNetTcpListenerCloseBody()
	}
}

// netFDTypeName is the Kylix type name recorded for conn/listener locals so
// that method-style dispatch (c.Method()) can recognize them. We don't add
// real methods here (TcpWrite etc. are free functions, not methods), but
// recording the type keeps localTypes consistent for any future method work.
const netConnTypeName = "TTcpConn"
const netListenerTypeName = "TTcpListener"

// ---- TcpDial: ptr @__kylix_net_TcpDial(ptr %host, i64 %port) ----
//
//	fd = socket(AF_INET=2, SOCK_STREAM=1, 0)
//	addr = { sa_family=2 (i16), sin_port=htons(port) (i16), sin_addr (i32), zero (i64) }
//	inet_pton(AF_INET, host, &addr.sin_addr)
//	connect(fd, &addr, 16)
//	malloc(8), store fd, return ptr  (fd==-1 → return null)
func (g *Generator) emitNetTcpDialCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("net.TcpDial expects 2 arguments, got %d", len(args))
	}
	hostReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	portReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpDial", "TcpDial", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpDial(ptr %s, i64 %s)", r, hostReg, portReg))
	return r, netConnTypeName, nil
}

func (g *Generator) emitNetTcpDialBody() {
	g.line("define ptr @__kylix_net_TcpDial(ptr %host, i64 %port) {")
	g.line("entry:")
	// fd = socket(2, 1, 0) -> i32
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @socket(i32 2, i32 1, i32 0)", fd))
	// if fd == -1 → ret null
	bad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, -1", bad, fd))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	// sockaddr_in: 16 bytes. alloca, zero it.
	addr := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [16 x i8], align 4", addr))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 16, i1 false)", addr))
	// sa_family = AF_INET (2) at offset 0 (i16)
	famPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 0", famPtr, addr))
	g.line(fmt.Sprintf("  store i16 2, ptr %s", famPtr))
	// sin_port = htons(port) at offset 2 (i16). htons on a constant-foldable
	// runtime value: shift bytes. We compute ((port & 0xff) << 8) | (port >> 8).
	lo := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %%port, 255", lo))
	loSh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 8", loSh, lo))
	hi := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %%port, 8", hi))
	portNet := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", portNet, loSh, hi))
	portNet16 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i16", portNet16, portNet))
	portPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 2", portPtr, addr))
	g.line(fmt.Sprintf("  store i16 %s, ptr %s", portNet16, portPtr))
	// sin_addr at offset 4: inet_pton(AF_INET, host, &addr+4)
	inAddrPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 4", inAddrPtr, addr))
	inetRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @inet_pton(i32 2, ptr %%host, ptr %s)", inetRes, inAddrPtr))
	// inet_pton returns 1 on success; !=1 → fail
	inetOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 1", inetOk, inetRes))
	inetFailLbl := g.label()
	inetOkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", inetOk, inetFailLbl, inetOkLbl))
	g.line(fmt.Sprintf("%s:", inetFailLbl))
	// close fd before returning null
	g.line(fmt.Sprintf("  call i32 @close(i32 %s)", fd))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", inetOkLbl))
	// connect(fd, &addr, 16)
	connRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @connect(i32 %s, ptr %s, i32 16)", connRes, fd, addr))
	connOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", connOk, connRes))
	connFailLbl := g.label()
	connOkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", connOk, connOkLbl, connFailLbl))
	g.line(fmt.Sprintf("%s:", connFailLbl))
	g.line(fmt.Sprintf("  call i32 @close(i32 %s)", fd))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", connOkLbl))
	// malloc(8), store fd (as i32), return ptr
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", inst))
	fdExt := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", fdExt, fd))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", fdExt, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// ---- TcpWrite: i64 @__kylix_net_TcpWrite(ptr %conn, ptr %data) ----
//
//	fd = load i64 from conn (truncate to i32)
//	n = send(fd, data, strlen(data), 0)
//	ret n
func (g *Generator) emitNetTcpWriteCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("net.TcpWrite expects 2 arguments, got %d", len(args))
	}
	connReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	dataReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpWrite", "TcpWrite", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_net_TcpWrite(ptr %s, ptr %s)", r, connReg, dataReg))
	return r, "i64", nil
}

func (g *Generator) emitNetTcpWriteBody() {
	g.line("define i64 @__kylix_net_TcpWrite(ptr %conn, ptr %data) {")
	g.line("entry:")
	fdVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%conn", fdVal))
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fd, fdVal))
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%data)", ln))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @send(i32 %s, ptr %%data, i64 %s, i32 0)", n, fd, ln))
	g.line(fmt.Sprintf("  ret i64 %s", n))
	g.line("}")
	g.line("")
}

// ---- TcpRead: ptr @__kylix_net_TcpRead(ptr %conn, i64 %n) ----
//
//	buf = malloc(n+1)
//	r = recv(fd, buf, n, 0)
//	if r <= 0 → free buf, ret null  (best-effort: caller treats null as EOF/error)
//	buf[r] = 0; ret buf
func (g *Generator) emitNetTcpReadCall(args []ast.Expression) (string, string, error) {
	if len(args) != 2 {
		return "", "", fmt.Errorf("net.TcpRead expects 2 arguments, got %d", len(args))
	}
	connReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	nReg, _, err := g.emitExpr(args[1])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpRead", "TcpRead", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpRead(ptr %s, i64 %s)", r, connReg, nReg))
	return r, "ptr", nil
}

func (g *Generator) emitNetTcpReadBody() {
	g.line("define ptr @__kylix_net_TcpRead(ptr %conn, i64 %n) {")
	g.line("entry:")
	fdVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%conn", fdVal))
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fd, fdVal))
	// bufSize = n + 1
	bufSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %%n, 1", bufSize))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, bufSize))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @recv(i32 %s, ptr %s, i64 %%n, i32 0)", r, fd, buf))
	// if r <= 0 → free(buf); ret null
	le0 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sle i64 %s, 0", le0, r))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", le0, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", buf))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	// buf[r] = 0
	termPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds i8, ptr %s, i64 %s", termPtr, buf, r))
	g.line(fmt.Sprintf("  store i8 0, ptr %s", termPtr))
	g.line(fmt.Sprintf("  ret ptr %s", buf))
	g.line("}")
	g.line("")
}

// ---- TcpClose: void @__kylix_net_TcpClose(ptr %conn) ----
func (g *Generator) emitNetTcpCloseCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("net.TcpClose expects 1 argument, got %d", len(args))
	}
	connReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpClose", "TcpClose", 0)
	g.line(fmt.Sprintf("  call void @__kylix_net_TcpClose(ptr %s)", connReg))
	return "0", "void", nil
}

func (g *Generator) emitNetTcpCloseBody() {
	g.line("define void @__kylix_net_TcpClose(ptr %conn) {")
	g.line("entry:")
	fdVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%conn", fdVal))
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fd, fdVal))
	g.line(fmt.Sprintf("  call i32 @close(i32 %s)", fd))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- TcpListen: ptr @__kylix_net_TcpListen(i64 %port) ----
//
//	fd = socket(2,1,0)
//	addr: family=2, port=htons(port), addr=INADDR_ANY(0)
//	bind(fd, &addr, 16); listen(fd, 128); malloc+store fd → ptr
func (g *Generator) emitNetTcpListenCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("net.TcpListen expects 1 argument, got %d", len(args))
	}
	portReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpListen", "TcpListen", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpListen(i64 %s)", r, portReg))
	return r, netListenerTypeName, nil
}

func (g *Generator) emitNetTcpListenBody() {
	g.line("define ptr @__kylix_net_TcpListen(i64 %port) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @socket(i32 2, i32 1, i32 0)", fd))
	bad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, -1", bad, fd))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	// SO_REUSEADDR = 1 to avoid "address in use" on quick restarts.
	one := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca i32, align 4", one))
	g.line(fmt.Sprintf("  store i32 1, ptr %s", one))
	g.line(fmt.Sprintf("  call i32 @setsockopt(i32 %s, i32 1, i32 2, ptr %s, i32 4)", fd, one)) // SOL_SOCKET=1, SO_REUSEADDR=2
	// build sockaddr_in (bind to INADDR_ANY=0)
	addr := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [16 x i8], align 4", addr))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 16, i1 false)", addr))
	famPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 0", famPtr, addr))
	g.line(fmt.Sprintf("  store i16 2, ptr %s", famPtr))
	// htons(port)
	lo := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %%port, 255", lo))
	loSh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 8", loSh, lo))
	hi := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %%port, 8", hi))
	portNet := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", portNet, loSh, hi))
	portNet16 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i16", portNet16, portNet))
	portPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 2", portPtr, addr))
	g.line(fmt.Sprintf("  store i16 %s, ptr %s", portNet16, portPtr))
	// sin_addr already 0 (INADDR_ANY) from memset
	bindRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @bind(i32 %s, ptr %s, i32 16)", bindRes, fd, addr))
	bindOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", bindOk, bindRes))
	bindFailLbl := g.label()
	bindOkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bindOk, bindOkLbl, bindFailLbl))
	g.line(fmt.Sprintf("%s:", bindFailLbl))
	g.line(fmt.Sprintf("  call i32 @close(i32 %s)", fd))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", bindOkLbl))
	g.line(fmt.Sprintf("  call i32 @listen(i32 %s, i32 128)", fd))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", inst))
	fdExt := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", fdExt, fd))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", fdExt, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// ---- TcpAccept: ptr @__kylix_net_TcpAccept(ptr %listener) ----
//
//	cfd = accept(lfd, null, null)
//	malloc+store cfd → ptr (null on -1)
func (g *Generator) emitNetTcpAcceptCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("net.TcpAccept expects 1 argument, got %d", len(args))
	}
	lReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpAccept", "TcpAccept", 0)
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpAccept(ptr %s)", r, lReg))
	return r, netConnTypeName, nil
}

func (g *Generator) emitNetTcpAcceptBody() {
	g.line("define ptr @__kylix_net_TcpAccept(ptr %listener) {")
	g.line("entry:")
	lfdVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%listener", lfdVal))
	lfd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", lfd, lfdVal))
	cfd := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @accept(i32 %s, ptr null, ptr null)", cfd, lfd))
	bad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, -1", bad, cfd))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", inst))
	fdExt := g.tmp()
	g.line(fmt.Sprintf("  %s = zext i32 %s to i64", fdExt, cfd))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", fdExt, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// ---- TcpListenerClose: void @__kylix_net_TcpListenerClose(ptr %listener) ----
func (g *Generator) emitNetTcpListenerCloseCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("net.TcpListenerClose expects 1 argument, got %d", len(args))
	}
	lReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpListenerClose", "TcpListenerClose", 0)
	g.line(fmt.Sprintf("  call void @__kylix_net_TcpListenerClose(ptr %s)", lReg))
	return "0", "void", nil
}

func (g *Generator) emitNetTcpListenerCloseBody() {
	g.line("define void @__kylix_net_TcpListenerClose(ptr %listener) {")
	g.line("entry:")
	fdVal := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%listener", fdVal))
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i32", fd, fdVal))
	g.line(fmt.Sprintf("  call i32 @close(i32 %s)", fd))
	g.line("  ret void")
	g.line("}")
	g.line("")
}
