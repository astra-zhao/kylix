package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_net tests — verify the IR generation for the TCP subset of the net
// stdlib module lowers to libc socket-backed defines (not stubs).

func TestNet_TcpDialCallDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_net_TcpDial")
	if strings.Contains(ir, "net.TcpDial not implemented") {
		t.Errorf("TcpDial still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestNet_TcpDialBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpDial(ptr %host, i64 %port)")
	assertIRContains(t, ir, "call i32 @socket(i32 2, i32 1, i32 0)")
	assertIRContains(t, ir, "call i32 @inet_pton")
	assertIRContains(t, ir, "call i32 @connect")
}

func TestNet_TcpWriteBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
  TcpWrite(c, 'hello');
end.`)
	assertIRContains(t, ir, "define i64 @__kylix_net_TcpWrite(ptr %conn, ptr %data)")
	assertIRContains(t, ir, "call i64 @send")
}

func TestNet_TcpReadBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
  var s := TcpRead(c, 100);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpRead(ptr %conn, i64 %n)")
	assertIRContains(t, ir, "call i64 @recv")
}

func TestNet_TcpListenBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var l := TcpListen(9090);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpListen(i64 %port)")
	assertIRContains(t, ir, "call i32 @bind")
	assertIRContains(t, ir, "call i32 @listen")
}

func TestNet_TcpAcceptBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var l := TcpListen(9090);
  var c := TcpAccept(l);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_net_TcpAccept(ptr %listener)")
	assertIRContains(t, ir, "call i32 @accept")
}

func TestNet_TcpCloseBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
  TcpClose(c);
end.`)
	assertIRContains(t, ir, "define void @__kylix_net_TcpClose(ptr %conn)")
	assertIRContains(t, ir, "call i32 @close")
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
}

func TestNet_SocketDeclarations(t *testing.T) {
	ir := generateIR(t, `program p;
uses net;
begin
  var c := TcpDial('127.0.0.1', 8080);
end.`)
	assertIRContains(t, ir, "declare i32 @socket")
	assertIRContains(t, ir, "declare i32 @connect")
	assertIRContains(t, ir, "declare i64 @send")
	assertIRContains(t, ir, "declare i64 @recv")
	assertIRContains(t, ir, "declare i32 @inet_pton")
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
