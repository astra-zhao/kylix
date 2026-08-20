package llvmgen_test

import (
	"strings"
	"testing"
)

// stdlib_websocket tests — verify the RFC 6455 client+server bodies lower to
// the expected IR: OpenSSL SHA-1 for the handshake accept, strcat request
// building, exact-length recv/send helpers, and frame building.

func TestWs_DialDispatch(t *testing.T) {
	ir := generateIR(t, `program p;
uses websocket;
begin
  var ws := WsDial('127.0.0.1:8080', '/echo');
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_websocket_WsDial")
	if strings.Contains(ir, "websocket.WsDial not implemented") {
		t.Errorf("WsDial still routed to not-implemented stub\nIR:\n%s", ir)
	}
}

func TestWs_DialBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses websocket;
begin
  var ws := WsDial('127.0.0.1:8080', '/echo');
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_websocket_WsDial(ptr %addr, ptr %path)")
	// TCP connect + client handshake: key → strcat request → send → read headers.
	assertIRContains(t, ir, "call ptr @__kylix_net_TcpDial")
	assertIRContains(t, ir, "call ptr @__kylix_ws_b64")
	assertIRContains(t, ir, "call void @__kylix_ws_sha1")
	assertIRContains(t, ir, "call ptr @strcat")
	assertIRContains(t, ir, "call ptr @__kylix_ws_readheaders")
	assertIRContains(t, ir, "call ptr @strstr")
}

func TestWs_AcceptBodyEmitted(t *testing.T) {
	ir := generateIR(t, `program p;
uses net, websocket;
begin
  var ln := TcpListen(8080);
  var tcp := TcpAccept(ln);
  var ws := WsAccept(tcp);
end.`)
	assertIRContains(t, ir, "define ptr @__kylix_websocket_WsAccept(ptr %tcp)")
	assertIRContains(t, ir, "call ptr @__kylix_ws_readheaders")
	assertIRContains(t, ir, "call void @__kylix_ws_sha1")
	assertIRContains(t, ir, "call i64 @__kylix_ws_sendall")
}

func TestWs_SendRecvCloseBodies(t *testing.T) {
	ir := generateIR(t, `program p;
uses websocket;
begin
  var ws := WsDial('127.0.0.1:8080', '/echo');
  WsSend(ws, 'hi');
  var m := WsRecv(ws);
  WsClose(ws);
end.`)
	assertIRContains(t, ir, "define void @__kylix_websocket_WsSend(ptr %ws, ptr %msg)")
	assertIRContains(t, ir, "define ptr @__kylix_websocket_WsRecv(ptr %ws)")
	assertIRContains(t, ir, "define void @__kylix_websocket_WsClose(ptr %ws)")
	// frame + exact-length IO helpers
	assertIRContains(t, ir, "define i64 @__kylix_ws_buildframe")
	assertIRContains(t, ir, "define i64 @__kylix_ws_recvn")
	assertIRContains(t, ir, "define i64 @__kylix_ws_sendall")
	// WsRecv parses frames (opcode/mask/length) and answers ping with pong.
	assertIRContains(t, ir, "call i64 @__kylix_ws_buildframe(i64 10")
}

func TestWs_DialConnectFinishDispatch(t *testing.T) {
	// v6.5.0: two-phase client handshake.
	ir := generateIR(t, `program p;
uses websocket;
begin
  var c := WsDialConnect('127.0.0.1:8080', '/echo');
  var ok := WsDialFinish(c);
end.`)
	assertIRContains(t, ir, "call ptr @__kylix_websocket_WsDialConnect")
	assertIRContains(t, ir, "call i1 @__kylix_websocket_WsDialFinish")
	// WsDial composes the two phases.
	assertIRContains(t, ir, "call ptr @__kylix_websocket_WsDialConnect")
}

func TestWs_NotUsedNoSymbols(t *testing.T) {
	ir := generateIR(t, `program p;
begin
  WriteLn('hi');
end.`)
	if strings.Contains(ir, "@__kylix_websocket_") {
		t.Errorf("websocket symbols emitted without `uses websocket`\nIR:\n%s", ir)
	}
}