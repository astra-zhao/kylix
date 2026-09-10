package llvmgen_test

import (
	"strings"
	"testing"

	"kylix/lexer"
	"kylix/parser"
	"kylix/pkg/llvmgen"
)

// stdlib_net tests — v0.7.1 P1 wrapper architecture: the seven public TCP
// functions are OS-agnostic (i64 handles, i64 -1 errors) and delegate to
// `__kylix_net_*` primitive wrappers emitted per target OS (unix BSD sockets
// vs Windows Winsock2). Both declare sets and both wrapper flavors are tested.

const netDialSrc = `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
end.`

// generateIRForTarget parses src and generates IR for the given target OS
// (e.g. "" for host default, "windows/amd64").
func generateIRForTarget(t *testing.T, src, target string) string {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	ir, err := llvmgen.GenerateWithOpts(prog, "", llvmgen.CompileOpts{Target: target})
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	return ir
}

// ===== Unix (host default) =====

func TestNet_TcpDialCallDispatch(t *testing.T) {
	ir := generateIR(t, netDialSrc)
	assertIRContains(t, ir, "call ptr @__kylix_net_TcpDial")
	if strings.Contains(ir, "net.TcpDial not implemented") {
		t.Errorf("TcpDial still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestNet_TcpDialBodyEmitted(t *testing.T) {
	ir := generateIR(t, netDialSrc)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpDial(ptr %host, i64 %port)")
	// Wrapper architecture: dial goes through the OS primitives.
	assertIRContains(t, ir, "call i64 @__kylix_net_socket()")
	assertIRContains(t, ir, "call i32 @__kylix_net_connect(i64")
	assertIRContains(t, ir, "call i32 @inet_pton")
	// The unix socket primitive wraps BSD socket().
	assertIRContains(t, ir, "define i64 @__kylix_net_socket()")
	assertIRContains(t, ir, "call i32 @socket(i32 2, i32 1, i32 0)")
}

func TestNet_TcpWriteBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
  TcpWrite(c, 'hello');
end.`)
	assertIRContains(t, ir, "define i64 @__kylix_net_TcpWrite(ptr %conn, ptr %data)")
	assertIRContains(t, ir, "call i64 @__kylix_net_send(i64")
	assertIRContains(t, ir, "define i64 @__kylix_net_send(i64 %h, ptr %buf, i64 %len)")
	assertIRContains(t, ir, "call i64 @send(i32")
}

func TestNet_TcpReadBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
  var s := TcpRead(c, 100);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpRead(ptr %conn, i64 %n)")
	assertIRContains(t, ir, "call i64 @__kylix_net_recv(i64")
	assertIRContains(t, ir, "define i64 @__kylix_net_recv(i64 %h, ptr %buf, i64 %len)")
}

func TestNet_TcpListenBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var l := TcpListen(9090);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpListen(i64 %port)")
	assertIRContains(t, ir, "call i32 @__kylix_net_bind(i64")
	assertIRContains(t, ir, "call i32 @__kylix_net_listen(i64")
	assertIRContains(t, ir, "call i32 @__kylix_net_reuse(i64")
}

func TestNet_TcpAcceptBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var l := TcpListen(9090);
  var c := TcpAccept(l);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpAccept(ptr %listener)")
	assertIRContains(t, ir, "call i64 @__kylix_net_accept(i64")
}

func TestNet_TcpCloseBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
  TcpClose(c);
end.`)
	assertIRContains(t, ir, "define void @__kylix_net_TcpClose(ptr %conn)")
	assertIRContains(t, ir, "call i32 @__kylix_net_closeh(i64")
	assertIRContains(t, ir, "define i32 @__kylix_net_closeh(i64 %h)")
	assertIRContains(t, ir, "call i32 @close(i32")
}

func TestNet_BodyDedup(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var a := TcpDial('127.0.0.1', 8080);
  var b := TcpDial('127.0.0.1', 9090);
end.`)
	if got := strings.Count(ir, "define ptr @__kylix_net_TcpDial"); got != 1 {
		t.Errorf("TcpDial define should appear once, got %d\nIR:\n%s", got, ir)
	}
	// Primitives dedup too: exactly one socket wrapper.
	if got := strings.Count(ir, "define i64 @__kylix_net_socket()"); got != 1 {
		t.Errorf("__kylix_net_socket define should appear once, got %d\nIR:\n%s", got, ir)
	}
}

func TestNet_SocketDeclarations(t *testing.T) {
	ir := generateIR(t, netDialSrc)
	assertIRContains(t, ir, "declare i32 @socket")
	assertIRContains(t, ir, "declare i32 @connect")
	assertIRContains(t, ir, "declare i64 @send")
	assertIRContains(t, ir, "declare i64 @recv")
	assertIRContains(t, ir, "declare i32 @inet_pton")
	assertIRContains(t, ir, "declare i32 @close")
	// No Winsock-only symbols on unix targets.
	if strings.Contains(ir, "@WSAStartup") || strings.Contains(ir, "@closesocket") {
		t.Errorf("Winsock symbols leaked into unix IR\nIR:\n%s", ir)
	}
}

func TestNet_NotUsedNoBodies(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "@__kylix_net_") {
		t.Errorf("net symbol emitted without `uses net`\nIR:\n%s", ir)
	}
}

// ===== Windows (Winsock2) =====

func TestNetWin_SocketDeclarations(t *testing.T) {
	ir := generateIRForTarget(t, netDialSrc, "windows/amd64")
	assertIRContains(t, ir, "declare i64 @socket")
	assertIRContains(t, ir, "declare i32 @connect(i64 noundef")
	assertIRContains(t, ir, "declare i32 @bind(i64 noundef")
	assertIRContains(t, ir, "declare i32 @listen(i64 noundef")
	assertIRContains(t, ir, "declare i64 @accept(i64 noundef")
	assertIRContains(t, ir, "declare i32 @send(i64 noundef")
	assertIRContains(t, ir, "declare i32 @recv(i64 noundef")
	assertIRContains(t, ir, "declare i32 @closesocket(i64 noundef)")
	assertIRContains(t, ir, "declare i32 @WSAStartup(i16 noundef, ptr noundef)")
	assertIRContains(t, ir, "declare i32 @ioctlsocket(i64 noundef")
	assertIRContains(t, ir, "declare i32 @setsockopt(i64 noundef")
	assertIRContains(t, ir, "declare i32 @inet_pton")
	// No unix-only symbols on Windows targets.
	if strings.Contains(ir, "declare i32 @close(") {
		t.Errorf("unix @close declared in Windows IR\nIR:\n%s", ir)
	}
}

func TestNetWin_WSAStartupOnce(t *testing.T) {
	ir := generateIRForTarget(t, netDialSrc, "windows/amd64")
	// WSAStartup-once state + MAKEWORD(2,2)=514.
	assertIRContains(t, ir, "@__kylix_net_wsa_done = global i1 false")
	assertIRContains(t, ir, "@__kylix_net_wsadata = global [512 x i8] zeroinitializer")
	assertIRContains(t, ir, "call i32 @WSAStartup(i16 514, ptr @__kylix_net_wsadata)")
	// Blocking-mode ioctlsocket: FIONBIO = 0x8004667e as i32, mode 0.
	assertIRContains(t, ir, "call i32 @ioctlsocket(i64")
	assertIRContains(t, ir, "i32 -2147195266")
}

func TestNetWin_TcpDialBody(t *testing.T) {
	ir := generateIRForTarget(t, netDialSrc, "windows/amd64")
	// Same OS-agnostic public body as unix.
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpDial(ptr %host, i64 %port)")
	assertIRContains(t, ir, "call i64 @__kylix_net_socket()")
	assertIRContains(t, ir, "call i32 @__kylix_net_connect(i64")
	// Windows send wrapper: SOCKET (i64) args, int (i32) winsock call, sext back.
	assertIRContains(t, ir, "define i64 @__kylix_net_send(i64 %h, ptr %buf, i64 %len)")
}

func TestNetWin_TcpListenAcceptClose(t *testing.T) {
	ir := generateIRForTarget(t, `program p;
uses net;
begin
  var l := TcpListen(9090);
  var c := TcpAccept(l);
  TcpClose(c);
  TcpListenerClose(l);
end.`, "windows/amd64")
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpListen(i64 %port)")
	assertIRContains(t, ir, "call i32 @__kylix_net_bind(i64")
	assertIRContains(t, ir, "call i64 @__kylix_net_accept(i64")
	assertIRContains(t, ir, "call i32 @__kylix_net_closeh(i64")
	// Windows reuse prim: SOL_SOCKET=0xffff, SO_REUSEADDR=4.
	assertIRContains(t, ir, "call i32 @setsockopt(i64 %h, i32 65535, i32 4, ptr %one, i32 4)")
}

func TestNetWin_UnixConstantsNotEmitted(t *testing.T) {
	ir := generateIRForTarget(t, `program p;
uses net;
begin
  var l := TcpListen(9090);
end.`, "windows/amd64")
	// Linux SO_REUSEADDR constants must not appear in the Windows reuse prim.
	if strings.Contains(ir, "@setsockopt(i64 %h, i32 1, i32 2,") {
		t.Errorf("Linux setsockopt constants (SOL_SOCKET=1, SO_REUSEADDR=2) leaked into Windows IR\nIR:\n%s", ir)
	}
}

// ===== Cross-target consistency =====

func TestNet_PublicBodyOSAgnostic(t *testing.T) {
	// The public bodies must be identical modulo register names across
	// targets: both targets emit the same public defines and the same
	// wrapper call sites.
	src := netDialSrc
	unixIR := generateIRForTarget(t, src, "")
	winIR := generateIRForTarget(t, src, "windows/amd64")
	for _, marker := range []string{
		"define ptr @__kylix_net_TcpDial(ptr %host, i64 %port)",
		"call i64 @__kylix_net_socket()",
		"call i32 @inet_pton(i32 2, ptr %host, ptr",
		"call i32 @__kylix_net_connect(i64",
	} {
		if !strings.Contains(unixIR, marker) {
			t.Errorf("unix IR missing %q", marker)
		}
		if !strings.Contains(winIR, marker) {
			t.Errorf("windows IR missing %q", marker)
		}
	}
}
