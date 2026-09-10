package llvmgen

import (
	"fmt"
	"kylix/ast"
)

// stdlib_net.go — LLVM IR implementation for the `net` stdlib module (TCP).
//
// v0.7.1 P1: target-independent wrapper architecture. The seven public
// functions below are OS-agnostic: they operate on an i64 socket handle
// (stored in the 8-byte heap cell that backs TTcpConn/TTcpListener) and
// delegate every OS touchpoint to the `__kylix_net_*` primitive wrappers,
// which are emitted once per target OS (unix BSD sockets vs Windows Winsock2):
//
//	__kylix_net_socket()                      -> i64 handle (-1 on failure; on
//	                                             Windows runs WSAStartup-once and
//	                                             forces blocking mode)
//	__kylix_net_connect(h, addr, len)         -> i32 (0 = ok)
//	__kylix_net_bind(h, addr, len)            -> i32
//	__kylix_net_listen(h, backlog)            -> i32
//	__kylix_net_accept(h)                     -> i64 handle (-1 on failure)
//	__kylix_net_send(h, buf, len)             -> i64 bytes (-1 on failure)
//	__kylix_net_recv(h, buf, len)             -> i64 bytes (-1 on failure)
//	__kylix_net_closeh(h)                     -> i32
//	__kylix_net_reuse(h)                      -> i32 (SO_REUSEADDR)
//
// Signatures are unified (handles/errors as i64 -1) so the public bodies never
// branch on the target. Unix maps handles to i32 fds (sext/trunc); Windows
// keeps SOCKET (UINT_PTR) as i64 natively. The corresponding declares in
// codegen.go switch on g.targetOS to match.
//
// TTcpConn / TTcpListener remain heap-allocated 8-byte cells holding the
// handle. Mirrors the Go-backend stdlib/net.go surface for the TCP subset that
// example55 (websocket) and example54 (http) depend on:
//
//   TcpDial(host, port)   -> ptr (TTcpConn)     socket()+connect()
//   TcpWrite(c, data)     -> i64 (bytes written) send()
//   TcpRead(c, n)         -> ptr (String)      recv()
//   TcpClose(c)           -> void              close
//   TcpListen(port)       -> ptr (TTcpListener) socket()+bind()+listen()
//   TcpAccept(l)          -> ptr (TTcpConn)    accept()
//   TcpListenerClose(l)   -> void              close
//
// UDP / DNS lookup are not implemented here (no example currently uses them
// on the LLVM path) and fall through to the not-implemented stub.
//
// Address handling: TcpDial uses inet_pton(AF_INET=2) — identical signature on
// both platforms (ws2_32 provides it) — so the host must be a dotted-quad IPv4
// literal (e.g. "127.0.0.1"). sockaddr_in layout is also identical (family
// i16 @0, port i16 @2, addr i32 @4). Hostname resolution is deferred.

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
	// OS primitives (emitted once, deduped by enqueueStdlib)
	case "__net_socket":
		g.emitNetSocketPrim()
	case "__net_connect":
		g.emitNetConnectPrim()
	case "__net_bind":
		g.emitNetBindPrim()
	case "__net_listen":
		g.emitNetListenPrim()
	case "__net_accept":
		g.emitNetAcceptPrim()
	case "__net_send":
		g.emitNetSendPrim()
	case "__net_recv":
		g.emitNetRecvPrim()
	case "__net_closeh":
		g.emitNetClosePrim()
	case "__net_reuse":
		g.emitNetReusePrim()
	}
}

// enqueueNetPrims queues the OS primitive wrappers for deferred emission.
// Deduped by enqueueStdlib, so calling it from every public Call emitter is
// cheap and guarantees the wrappers exist whenever any net function does.
func (g *Generator) enqueueNetPrims() {
	g.enqueueStdlib("net", "__net_socket", "__net_socket", 0)
	g.enqueueStdlib("net", "__net_connect", "__net_connect", 0)
	g.enqueueStdlib("net", "__net_bind", "__net_bind", 0)
	g.enqueueStdlib("net", "__net_listen", "__net_listen", 0)
	g.enqueueStdlib("net", "__net_accept", "__net_accept", 0)
	g.enqueueStdlib("net", "__net_send", "__net_send", 0)
	g.enqueueStdlib("net", "__net_recv", "__net_recv", 0)
	g.enqueueStdlib("net", "__net_closeh", "__net_closeh", 0)
	g.enqueueStdlib("net", "__net_reuse", "__net_reuse", 0)
}

// enqueueNetPublic queues one public net function plus the OS primitives it
// needs, for cross-module references (the boot HTTP server and websocket
// handshake emit calls to the public functions directly, with no user-visible
// `net` call site that would otherwise trigger emission). Deduped.
func (g *Generator) enqueueNetPublic(name string) {
	g.enqueueStdlib("net", name, name, 0)
	g.enqueueNetPrims()
}

// netFDTypeName is the Kylix type name recorded for conn/listener locals so
// that method-style dispatch (c.Method()) can recognize them. We don't add
// real methods here (TcpWrite etc. are free functions, not methods), but
// recording the type keeps localTypes consistent for any future method work.
const netConnTypeName = "TTcpConn"
const netListenerTypeName = "TTcpListener"

// emitNetHtons emits the byte-swap for sin_port: ((port & 0xff) << 8) |
// (port >> 8), truncated to i16. Identical on both targets (big-endian wire
// order).
func (g *Generator) emitNetHtons(portReg string) string {
	lo := g.tmp()
	g.line(fmt.Sprintf("  %s = and i64 %s, 255", lo, portReg))
	loSh := g.tmp()
	g.line(fmt.Sprintf("  %s = shl i64 %s, 8", loSh, lo))
	hi := g.tmp()
	g.line(fmt.Sprintf("  %s = lshr i64 %s, 8", hi, portReg))
	portNet := g.tmp()
	g.line(fmt.Sprintf("  %s = or i64 %s, %s", portNet, loSh, hi))
	portNet16 := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %s to i16", portNet16, portNet))
	return portNet16
}

// emitNetSockaddr allocates and fills a 16-byte sockaddr_in with AF_INET and
// htons(port); sin_addr stays 0 (INADDR_ANY) for the caller to overwrite
// (TcpDial does via inet_pton, TcpListen leaves it). Returns the alloca reg.
func (g *Generator) emitNetSockaddr(portReg string) string {
	addr := g.tmp()
	g.line(fmt.Sprintf("  %s = alloca [16 x i8], align 4", addr))
	g.line(fmt.Sprintf("  call void @llvm.memset.p0.i64(ptr %s, i8 0, i64 16, i1 false)", addr))
	famPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 0", famPtr, addr))
	g.line(fmt.Sprintf("  store i16 2, ptr %s", famPtr)) // AF_INET
	portNet16 := g.emitNetHtons(portReg)
	portPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 2", portPtr, addr))
	g.line(fmt.Sprintf("  store i16 %s, ptr %s", portNet16, portPtr))
	return addr
}

// ---- OS primitive wrappers (one define per target OS) ----

// emitNetSocketPrim: i64 @__kylix_net_socket()
//
//	Unix: fd = socket(AF_INET, SOCK_STREAM, 0), sext to i64 (-1 preserved).
//	Windows: WSAStartup-once (MAKEWORD(2,2)=514), then socket() (SOCKET is
//	UINT_PTR = i64, INVALID_SOCKET = -1 bit pattern), then ioctlsocket
//	(FIONBIO, mode=0) to force blocking mode.
func (g *Generator) emitNetSocketPrim() {
	if g.targetOS == "windows" {
		// WSAStartup state (module-level globals).
		g.line("@__kylix_net_wsa_done = global i1 false")
		g.line("@__kylix_net_wsadata = global [512 x i8] zeroinitializer")
		g.line("define i64 @__kylix_net_socket() {")
		g.line("entry:")
		done := g.tmp()
		g.line(fmt.Sprintf("  %s = load i1, ptr @__kylix_net_wsa_done", done))
		needInit := g.tmp()
		g.line(fmt.Sprintf("  %s = icmp eq i1 %s, false", needInit, done))
		initLbl := g.label()
		afterLbl := g.label()
		g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", needInit, initLbl, afterLbl))
		g.line(fmt.Sprintf("%s:", initLbl))
		g.line(fmt.Sprintf("  %%wsar = call i32 @WSAStartup(i16 514, ptr @__kylix_net_wsadata) ; MAKEWORD(2,2)"))
		g.line("  store i1 true, ptr @__kylix_net_wsa_done")
		g.line(fmt.Sprintf("  br label %%%s", afterLbl))
		g.line(fmt.Sprintf("%s:", afterLbl))
		h := g.tmp()
		g.line(fmt.Sprintf("  %s = call i64 @socket(i32 2, i32 1, i32 0)", h))
		// FIONBIO = 0x8004667e; mode 0 = blocking (explicit Winsock default).
		mode := g.tmp()
		g.line(fmt.Sprintf("  %s = alloca i32, align 4", mode))
		g.line(fmt.Sprintf("  store i32 0, ptr %s", mode))
		g.line(fmt.Sprintf("  %%ior = call i32 @ioctlsocket(i64 %s, i32 -2147195266, ptr %s)", h, mode))
		g.line(fmt.Sprintf("  ret i64 %s", h))
		g.line("}")
		g.line("")
		return
	}
	g.line("define i64 @__kylix_net_socket() {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @socket(i32 2, i32 1, i32 0)", fd))
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", h, fd))
	g.line(fmt.Sprintf("  ret i64 %s", h))
	g.line("}")
	g.line("")
}

func (g *Generator) emitNetConnectPrim() {
	if g.targetOS == "windows" {
		g.line("define i32 @__kylix_net_connect(i64 %h, ptr %addr, i32 %len) {")
		g.line("entry:")
		g.line("  %1 = call i32 @connect(i64 %h, ptr %addr, i32 %len)")
		g.line("  ret i32 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i32 @__kylix_net_connect(i64 %h, ptr %addr, i32 %len) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	g.line(fmt.Sprintf("  %%1 = call i32 @connect(i32 %s, ptr %%addr, i32 %%len)", fd))
	g.line("  ret i32 %1")
	g.line("}")
	g.line("")
}

func (g *Generator) emitNetBindPrim() {
	if g.targetOS == "windows" {
		g.line("define i32 @__kylix_net_bind(i64 %h, ptr %addr, i32 %len) {")
		g.line("entry:")
		g.line("  %1 = call i32 @bind(i64 %h, ptr %addr, i32 %len)")
		g.line("  ret i32 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i32 @__kylix_net_bind(i64 %h, ptr %addr, i32 %len) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	g.line(fmt.Sprintf("  %%1 = call i32 @bind(i32 %s, ptr %%addr, i32 %%len)", fd))
	g.line("  ret i32 %1")
	g.line("}")
	g.line("")
}

func (g *Generator) emitNetListenPrim() {
	if g.targetOS == "windows" {
		g.line("define i32 @__kylix_net_listen(i64 %h, i32 %backlog) {")
		g.line("entry:")
		g.line("  %1 = call i32 @listen(i64 %h, i32 %backlog)")
		g.line("  ret i32 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i32 @__kylix_net_listen(i64 %h, i32 %backlog) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	g.line(fmt.Sprintf("  %%1 = call i32 @listen(i32 %s, i32 %%backlog)", fd))
	g.line("  ret i32 %1")
	g.line("}")
	g.line("")
}

func (g *Generator) emitNetAcceptPrim() {
	if g.targetOS == "windows" {
		g.line("define i64 @__kylix_net_accept(i64 %h) {")
		g.line("entry:")
		g.line("  %1 = call i64 @accept(i64 %h, ptr null, ptr null)")
		g.line("  ret i64 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i64 @__kylix_net_accept(i64 %h) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	cfd := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @accept(i32 %s, ptr null, ptr null)", cfd, fd))
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = sext i32 %s to i64", h, cfd))
	g.line(fmt.Sprintf("  ret i64 %s", h))
	g.line("}")
	g.line("")
}

// send/recv: lengths unify as i64 in / i64 out. Unix send/recv take size_t
// (i64) and return ssize_t; Windows takes int (i32) and returns int, so the
// wrapper truncs/sexts (SOCKET_ERROR=-1 sexts back to -1).
func (g *Generator) emitNetSendPrim() {
	if g.targetOS == "windows" {
		g.line("define i64 @__kylix_net_send(i64 %h, ptr %buf, i64 %len) {")
		g.line("entry:")
		l := g.tmp()
		g.line(fmt.Sprintf("  %s = trunc i64 %%len to i32", l))
		g.line(fmt.Sprintf("  %%r = call i32 @send(i64 %%h, ptr %%buf, i32 %s, i32 0)", l))
		g.line("  %1 = sext i32 %r to i64")
		g.line("  ret i64 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i64 @__kylix_net_send(i64 %h, ptr %buf, i64 %len) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	g.line(fmt.Sprintf("  %%1 = call i64 @send(i32 %s, ptr %%buf, i64 %%len, i32 0)", fd))
	g.line("  ret i64 %1")
	g.line("}")
	g.line("")
}

func (g *Generator) emitNetRecvPrim() {
	if g.targetOS == "windows" {
		g.line("define i64 @__kylix_net_recv(i64 %h, ptr %buf, i64 %len) {")
		g.line("entry:")
		l := g.tmp()
		g.line(fmt.Sprintf("  %s = trunc i64 %%len to i32", l))
		g.line(fmt.Sprintf("  %%r = call i32 @recv(i64 %%h, ptr %%buf, i32 %s, i32 0)", l))
		g.line("  %1 = sext i32 %r to i64")
		g.line("  ret i64 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i64 @__kylix_net_recv(i64 %h, ptr %buf, i64 %len) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	g.line(fmt.Sprintf("  %%1 = call i64 @recv(i32 %s, ptr %%buf, i64 %%len, i32 0)", fd))
	g.line("  ret i64 %1")
	g.line("}")
	g.line("")
}

func (g *Generator) emitNetClosePrim() {
	if g.targetOS == "windows" {
		g.line("define i32 @__kylix_net_closeh(i64 %h) {")
		g.line("entry:")
		g.line("  %1 = call i32 @closesocket(i64 %h)")
		g.line("  ret i32 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i32 @__kylix_net_closeh(i64 %h) {")
	g.line("entry:")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	g.line(fmt.Sprintf("  %%1 = call i32 @close(i32 %s)", fd))
	g.line("  ret i32 %1")
	g.line("}")
	g.line("")
}

// SO_REUSEADDR — constants are NOT portable: Linux uses SOL_SOCKET=1 /
// SO_REUSEADDR=2, while BSD sockets (macOS, *BSD) and Winsock2 (Windows) both
// use SOL_SOCKET=0xffff (65535) / SO_REUSEADDR=4. Getting this wrong makes
// setsockopt silently target a bogus option level and the reuse never applies
// (surfaced as EADDRINUSE on quick restart after a clean exit leaves TIME_WAIT
// sockets behind; v0.7.1 P1).
func (g *Generator) emitNetReusePrim() {
	solSocket := "65535" // 0xffff — macOS/BSD/Windows
	soReuse := "4"
	if g.targetOS == "linux" {
		solSocket, soReuse = "1", "2"
	}
	if g.targetOS == "windows" {
		g.line("define i32 @__kylix_net_reuse(i64 %h) {")
		g.line("entry:")
		g.line("  %one = alloca i32, align 4")
		g.line("  store i32 1, ptr %one")
		g.line(fmt.Sprintf("  %%1 = call i32 @setsockopt(i64 %%h, i32 %s, i32 %s, ptr %%one, i32 4)", solSocket, soReuse))
		g.line("  ret i32 %1")
		g.line("}")
		g.line("")
		return
	}
	g.line("define i32 @__kylix_net_reuse(i64 %h) {")
	g.line("entry:")
	g.line("  %one = alloca i32, align 4")
	g.line("  store i32 1, ptr %one")
	fd := g.tmp()
	g.line(fmt.Sprintf("  %s = trunc i64 %%h to i32", fd))
	g.line(fmt.Sprintf("  %%1 = call i32 @setsockopt(i32 %s, i32 %s, i32 %s, ptr %%one, i32 4)", fd, solSocket, soReuse))
	g.line("  ret i32 %1")
	g.line("}")
	g.line("")
}

// ---- TcpDial: ptr @__kylix_net_TcpDial(ptr %host, i64 %port) ----
//
//	h = __kylix_net_socket()                (-1 → null)
//	addr = sockaddr_in(AF_INET, htons(port))
//	inet_pton(AF_INET, host, &addr.sin_addr) (!=1 → close, null)
//	__kylix_net_connect(h, addr, 16)        (!=0 → close, null)
//	malloc(8), store h, return ptr
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
	g.enqueueNetPrims()
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpDial(ptr %s, i64 %s)", r, hostReg, portReg))
	return r, netConnTypeName, nil
}

func (g *Generator) emitNetTcpDialBody() {
	g.line("define ptr @__kylix_net_TcpDial(ptr %host, i64 %port) {")
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_net_socket()", h))
	hBad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, -1", hBad, h))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hBad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	addr := g.emitNetSockaddr("%port")
	// sin_addr at offset 4: inet_pton(AF_INET, host, &addr+4)
	inAddrPtr := g.tmp()
	g.line(fmt.Sprintf("  %s = getelementptr inbounds [16 x i8], ptr %s, i64 0, i64 4", inAddrPtr, addr))
	inetRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @inet_pton(i32 2, ptr %%host, ptr %s)", inetRes, inAddrPtr))
	inetOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp ne i32 %s, 1", inetOk, inetRes))
	inetFailLbl := g.label()
	inetOkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", inetOk, inetFailLbl, inetOkLbl))
	g.line(fmt.Sprintf("%s:", inetFailLbl))
	g.line(fmt.Sprintf("  %%ign1 = call i32 @__kylix_net_closeh(i64 %s)", h))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", inetOkLbl))
	connRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @__kylix_net_connect(i64 %s, ptr %s, i32 16)", connRes, h, addr))
	connOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", connOk, connRes))
	connFailLbl := g.label()
	connOkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", connOk, connOkLbl, connFailLbl))
	g.line(fmt.Sprintf("%s:", connFailLbl))
	g.line(fmt.Sprintf("  %%ign2 = call i32 @__kylix_net_closeh(i64 %s)", h))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", connOkLbl))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", h, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// ---- TcpWrite: i64 @__kylix_net_TcpWrite(ptr %conn, ptr %data) ----
//
//	h = load i64 from conn
//	n = __kylix_net_send(h, data, strlen(data))
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
	g.enqueueNetPrims()
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_net_TcpWrite(ptr %s, ptr %s)", r, connReg, dataReg))
	return r, "i64", nil
}

func (g *Generator) emitNetTcpWriteBody() {
	g.line("define i64 @__kylix_net_TcpWrite(ptr %conn, ptr %data) {")
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%conn", h))
	ln := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @strlen(ptr %%data)", ln))
	n := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_net_send(i64 %s, ptr %%data, i64 %s)", n, h, ln))
	g.line(fmt.Sprintf("  ret i64 %s", n))
	g.line("}")
	g.line("")
}

// ---- TcpRead: ptr @__kylix_net_TcpRead(ptr %conn, i64 %n) ----
//
//	buf = malloc(n+1)
//	r = __kylix_net_recv(h, buf, n)
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
	g.enqueueNetPrims()
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpRead(ptr %s, i64 %s)", r, connReg, nReg))
	return r, "ptr", nil
}

func (g *Generator) emitNetTcpReadBody() {
	g.line("define ptr @__kylix_net_TcpRead(ptr %conn, i64 %n) {")
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%conn", h))
	bufSize := g.tmp()
	g.line(fmt.Sprintf("  %s = add i64 %%n, 1", bufSize))
	buf := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 %s)", buf, bufSize))
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_net_recv(i64 %s, ptr %s, i64 %%n)", r, h, buf))
	le0 := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp sle i64 %s, 0", le0, r))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", le0, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line(fmt.Sprintf("  call void @free(ptr %s)", buf))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
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
	g.enqueueNetPrims()
	g.line(fmt.Sprintf("  call void @__kylix_net_TcpClose(ptr %s)", connReg))
	return "0", "void", nil
}

func (g *Generator) emitNetTcpCloseBody() {
	g.line("define void @__kylix_net_TcpClose(ptr %conn) {")
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%conn", h))
	g.line(fmt.Sprintf("  %%ign = call i32 @__kylix_net_closeh(i64 %s)", h))
	g.line("  ret void")
	g.line("}")
	g.line("")
}

// ---- TcpListen: ptr @__kylix_net_TcpListen(i64 %port) ----
//
//	h = __kylix_net_socket()             (-1 → null)
//	__kylix_net_reuse(h)                  (SO_REUSEADDR before bind)
//	addr = sockaddr_in(AF_INET, htons(port)); sin_addr = INADDR_ANY
//	__kylix_net_bind(h, addr, 16)        (!=0 → close, null)
//	__kylix_net_listen(h, 128)
//	malloc(8), store h, return ptr
func (g *Generator) emitNetTcpListenCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("net.TcpListen expects 1 argument, got %d", len(args))
	}
	portReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpListen", "TcpListen", 0)
	g.enqueueNetPrims()
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpListen(i64 %s)", r, portReg))
	return r, netListenerTypeName, nil
}

func (g *Generator) emitNetTcpListenBody() {
	g.line("define ptr @__kylix_net_TcpListen(i64 %port) {")
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_net_socket()", h))
	hBad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, -1", hBad, h))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", hBad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	// SO_REUSEADDR to avoid "address in use" on quick restarts.
	g.line(fmt.Sprintf("  %%ignr = call i32 @__kylix_net_reuse(i64 %s)", h))
	addr := g.emitNetSockaddr("%port")
	bindRes := g.tmp()
	g.line(fmt.Sprintf("  %s = call i32 @__kylix_net_bind(i64 %s, ptr %s, i32 16)", bindRes, h, addr))
	bindOk := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i32 %s, 0", bindOk, bindRes))
	bindFailLbl := g.label()
	bindOkLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", bindOk, bindOkLbl, bindFailLbl))
	g.line(fmt.Sprintf("%s:", bindFailLbl))
	g.line(fmt.Sprintf("  %%ignb = call i32 @__kylix_net_closeh(i64 %s)", h))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", bindOkLbl))
	g.line(fmt.Sprintf("  %%ignl = call i32 @__kylix_net_listen(i64 %s, i32 128)", h))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", h, inst))
	g.line(fmt.Sprintf("  ret ptr %s", inst))
	g.line("}")
	g.line("")
}

// ---- TcpAccept: ptr @__kylix_net_TcpAccept(ptr %listener) ----
//
//	c = __kylix_net_accept(h)   (-1 → null)
//	malloc(8), store c, return ptr
func (g *Generator) emitNetTcpAcceptCall(args []ast.Expression) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("net.TcpAccept expects 1 argument, got %d", len(args))
	}
	lReg, _, err := g.emitExpr(args[0])
	if err != nil {
		return "", "", err
	}
	g.enqueueStdlib("net", "TcpAccept", "TcpAccept", 0)
	g.enqueueNetPrims()
	r := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @__kylix_net_TcpAccept(ptr %s)", r, lReg))
	return r, netConnTypeName, nil
}

func (g *Generator) emitNetTcpAcceptBody() {
	g.line("define ptr @__kylix_net_TcpAccept(ptr %listener) {")
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%listener", h))
	c := g.tmp()
	g.line(fmt.Sprintf("  %s = call i64 @__kylix_net_accept(i64 %s)", c, h))
	cBad := g.tmp()
	g.line(fmt.Sprintf("  %s = icmp eq i64 %s, -1", cBad, c))
	okLbl := g.label()
	failLbl := g.label()
	g.line(fmt.Sprintf("  br i1 %s, label %%%s, label %%%s", cBad, failLbl, okLbl))
	g.line(fmt.Sprintf("%s:", failLbl))
	g.line("  ret ptr null")
	g.line(fmt.Sprintf("%s:", okLbl))
	inst := g.tmp()
	g.line(fmt.Sprintf("  %s = call ptr @malloc(i64 8)", inst))
	g.line(fmt.Sprintf("  store i64 %s, ptr %s", c, inst))
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
	g.enqueueNetPrims()
	g.line(fmt.Sprintf("  call void @__kylix_net_TcpListenerClose(ptr %s)", lReg))
	return "0", "void", nil
}

func (g *Generator) emitNetTcpListenerCloseBody() {
	g.line("define void @__kylix_net_TcpListenerClose(ptr %listener) {")
	g.line("entry:")
	h := g.tmp()
	g.line(fmt.Sprintf("  %s = load i64, ptr %%listener", h))
	g.line(fmt.Sprintf("  %%ign = call i32 @__kylix_net_closeh(i64 %s)", h))
	g.line("  ret void")
	g.line("}")
	g.line("")
}
