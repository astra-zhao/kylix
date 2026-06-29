package stdlib

import (
	"net"
	"strings"
	"testing"
	"time"
)

// startWsServer launches a TCP server that performs the WS handshake and
// echoes any received text back to the client. Returns the listener and port.
func startWsServer(t *testing.T) (net.Listener, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			// Wrap raw net.Conn into our TTcpConn so WsAccept can consume it.
			tcp := &TTcpConn{conn: conn}
			ws, err := WsAccept(tcp)
			if err != nil {
				conn.Close()
				continue
			}
			// Echo loop: read a message, send it back, until closed.
			for {
				msg, err := WsRecv(ws)
				if err != nil {
					WsClose(ws)
					break
				}
				if err := WsSend(ws, "echo:"+msg); err != nil {
					WsClose(ws)
					break
				}
			}
		}
	}()
	return ln, ln.Addr().(*net.TCPAddr).Port
}

func TestWs_DialHandshake(t *testing.T) {
	ln, port := startWsServer(t)
	defer ln.Close()

	ws, err := WsDial("127.0.0.1:"+itoa(port), "/chat")
	if err != nil {
		t.Fatalf("WsDial failed: %v", err)
	}
	defer WsClose(ws)
}

func TestWs_EchoRoundTrip(t *testing.T) {
	ln, port := startWsServer(t)
	defer ln.Close()

	ws, err := WsDial("127.0.0.1:"+itoa(port), "/chat")
	if err != nil {
		t.Fatalf("WsDial failed: %v", err)
	}
	defer WsClose(ws)

	if err := WsSend(ws, "hello"); err != nil {
		t.Fatalf("WsSend failed: %v", err)
	}
	msg, err := WsRecv(ws)
	if err != nil {
		t.Fatalf("WsRecv failed: %v", err)
	}
	if msg != "echo:hello" {
		t.Errorf("got %q, want echo:hello", msg)
	}
}

func TestWs_LargeMessage(t *testing.T) {
	ln, port := startWsServer(t)
	defer ln.Close()

	ws, err := WsDial("127.0.0.1:"+itoa(port), "/chat")
	if err != nil {
		t.Fatalf("WsDial failed: %v", err)
	}
	defer WsClose(ws)

	// 5000 bytes triggers the 16-bit length encoding path (>125, <65536).
	big := strings.Repeat("x", 5000)
	if err := WsSend(ws, big); err != nil {
		t.Fatalf("WsSend large failed: %v", err)
	}
	msg, err := WsRecv(ws)
	if err != nil {
		t.Fatalf("WsRecv large failed: %v", err)
	}
	want := "echo:" + big
	if msg != want {
		t.Errorf("large msg len = %d, want %d", len(msg), len(want))
	}
}

func TestWs_PingAutoPong(t *testing.T) {
	ln, port := startWsServer(t)
	defer ln.Close()

	ws, err := WsDial("127.0.0.1:"+itoa(port), "/chat")
	if err != nil {
		t.Fatalf("WsDial failed: %v", err)
	}
	defer WsClose(ws)

	// Force-send a ping frame directly; server's WsRecv answers with pong.
	if err := ws.writeFrame(wsOpPing, []byte("pingdata")); err != nil {
		t.Fatalf("writeFrame ping failed: %v", err)
	}
	// Now send a real text message; the pong is consumed silently inside
	// WsRecv and we should still get the echo.
	if err := WsSend(ws, "after-ping"); err != nil {
		t.Fatalf("WsSend failed: %v", err)
	}
	msg, err := WsRecv(ws)
	if err != nil {
		t.Fatalf("WsRecv failed: %v", err)
	}
	if msg != "echo:after-ping" {
		t.Errorf("got %q, want echo:after-ping", msg)
	}
}

func TestWs_CloseReturnsError(t *testing.T) {
	ln, port := startWsServer(t)
	defer ln.Close()

	ws, err := WsDial("127.0.0.1:"+itoa(port), "/chat")
	if err != nil {
		t.Fatalf("WsDial failed: %v", err)
	}

	// Close from client; server echo loop reads the close frame, returns err,
	// and closes its side. A subsequent WsRecv on a fresh client that the
	// server closed should error.
	WsClose(ws)

	// Dial again and have the server close immediately by sending close.
	ws2, err := WsDial("127.0.0.1:"+itoa(port), "/chat")
	if err != nil {
		t.Fatalf("WsDial2 failed: %v", err)
	}
	// Send a close frame; server will close its conn; next recv errors.
	if err := ws2.writeFrame(wsOpClose, nil); err != nil {
		t.Fatalf("writeFrame close failed: %v", err)
	}
	// Give the server a moment to process and close.
	time.Sleep(50 * time.Millisecond)
	_, err = WsRecv(ws2)
	if err == nil {
		t.Error("expected error after peer close, got nil")
	}
	ws2.conn.Close()
}

func TestWs_NilGuards(t *testing.T) {
	if err := WsSend(nil, "x"); err == nil {
		t.Error("WsSend(nil) should error")
	}
	if _, err := WsRecv(nil); err == nil {
		t.Error("WsRecv(nil) should error")
	}
	if err := WsClose(nil); err != nil {
		t.Errorf("WsClose(nil) should be nil, got %v", err)
	}
	if _, err := WsAccept(nil); err == nil {
		t.Error("WsAccept(nil) should error")
	}
}

// itoa avoids pulling strconv just for an int→string in tests.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
