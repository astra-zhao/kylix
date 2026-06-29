// websocket.go — Minimal RFC 6455 WebSocket client and server (pure stdlib).
//
// Supports: handshake (client + server), text frames, ping/pong, close.
// Does NOT support: binary frames, per-message-deflate, frame fragmentation,
// streaming reads of messages larger than MaxFrameBytes (default 1 MiB).
//
// Client example:
//
//	ws, _ := WsDial("localhost:8080", "/chat")
//	WsSend(ws, "hello")
//	msg, _ := WsRecv(ws)
//	WsClose(ws)
package stdlib

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// wsGUID is the RFC 6455 handshake magic string.
const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// MaxWsFrameBytes caps a single received message size to bound memory use.
const MaxWsFrameBytes = 1 << 20 // 1 MiB

// TWsConn is a WebSocket connection (client or server side).
type TWsConn struct {
	conn      net.Conn
	br        *bufio.Reader
	isServer  bool // server-side writes must NOT mask frames
	writeBuf  []byte
}

// wsOpcode per RFC 6455 section 5.2.
const (
	wsOpCont  = 0x0
	wsOpText  = 0x1
	wsOpClose = 0x8
	wsOpPing  = 0x9
	wsOpPong  = 0xA
)

// WsDial connects to a WebSocket server at addr/path and performs the
// handshake. addr is "host:port"; path must start with "/".
func WsDial(addr, path string) (*TWsConn, error) {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("WsDial: %w", err)
	}

	// Build client handshake request with a random Sec-WebSocket-Key.
	key, err := wsRandomKey()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("WsDial (key): %w", err)
	}
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\n"+
		"Connection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n",
		path, addr, key)
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("WsDial (write): %w", err)
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, &http.Request{Method: "GET"})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("WsDial (read response): %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		return nil, fmt.Errorf("WsDial: expected 101, got %d", resp.StatusCode)
	}
	// Verify Sec-WebSocket-Accept.
	want := wsAcceptKey(key)
	if resp.Header.Get("Sec-WebSocket-Accept") != want {
		conn.Close()
		return nil, fmt.Errorf("WsDial: bad Sec-WebSocket-Accept")
	}
	return &TWsConn{conn: conn, br: br, isServer: false, writeBuf: make([]byte, 0, 256)}, nil
}

// WsAccept wraps an already-accepted TCP connection (from net.TcpAccept) and
// performs the server-side WebSocket handshake. Use this after upgrading an
// HTTP request, or on a dedicated WS port.
func WsAccept(tcp *TTcpConn) (*TWsConn, error) {
	if tcp == nil || tcp.conn == nil {
		return nil, fmt.Errorf("WsAccept: nil connection")
	}
	conn := tcp.conn
	br := bufio.NewReader(conn)
	req, err := http.ReadRequest(br)
	if err != nil {
		return nil, fmt.Errorf("WsAccept (read request): %w", err)
	}
	if !strings.EqualFold(req.Header.Get("Upgrade"), "websocket") {
		return nil, fmt.Errorf("WsAccept: not a websocket upgrade")
	}
	key := req.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, fmt.Errorf("WsAccept: missing Sec-WebSocket-Key")
	}
	accept := wsAcceptKey(key)
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := conn.Write([]byte(resp)); err != nil {
		return nil, fmt.Errorf("WsAccept (write): %w", err)
	}
	return &TWsConn{conn: conn, br: br, isServer: true, writeBuf: make([]byte, 0, 256)}, nil
}

// WsSend writes a text frame containing msg.
func WsSend(ws *TWsConn, msg string) error {
	if ws == nil {
		return fmt.Errorf("WsSend: nil connection")
	}
	return ws.writeFrame(wsOpText, []byte(msg))
}

// WsRecv reads one message frame and returns its text payload. Ping frames
// are answered with pong automatically; close frames return an error.
func WsRecv(ws *TWsConn) (string, error) {
	if ws == nil {
		return "", fmt.Errorf("WsRecv: nil connection")
	}
	for {
		payload, opcode, err := ws.readFrame()
		if err != nil {
			return "", err
		}
		switch opcode {
		case wsOpText, wsOpCont:
			return string(payload), nil
		case wsOpPing:
			// Echo back as pong.
			if err := ws.writeFrame(wsOpPong, payload); err != nil {
				return "", err
			}
		case wsOpPong:
			// Ignore unsolicited pong.
		case wsOpClose:
			return "", fmt.Errorf("WsRecv: connection closed by peer")
		}
	}
}

// WsClose sends a close frame and closes the underlying TCP connection.
func WsClose(ws *TWsConn) error {
	if ws == nil {
		return nil
	}
	_ = ws.writeFrame(wsOpClose, nil)
	return ws.conn.Close()
}

// wsAcceptKey computes the Sec-WebSocket-Accept value per RFC 6455 §4.2.2:
// base64(sha1(key + GUID)).
func wsAcceptKey(key string) string {
	h := sha1.Sum([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(h[:])
}

// wsRandomKey returns 16 random bytes base64-encoded for Sec-WebSocket-Key.
func wsRandomKey() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// writeFrame encodes and writes a single WebSocket frame. Client-side frames
// must be masked (RFC 6455 §5.3); server-side frames are not masked.
func (ws *TWsConn) writeFrame(opcode byte, data []byte) error {
	var hdr [14]byte
	n := 0
	hdr[n] = 0x80 | opcode // FIN=1 + opcode
	n++
	maskBit := byte(0)
	if !ws.isServer {
		maskBit = 0x80
	}
	payloadLen := len(data)
	switch {
	case payloadLen < 126:
		hdr[n] = maskBit | byte(payloadLen)
		n++
	case payloadLen < 65536:
		hdr[n] = maskBit | 126
		n++
		binary.BigEndian.PutUint16(hdr[n:], uint16(payloadLen))
		n += 2
	default:
		hdr[n] = maskBit | 127
		n++
		binary.BigEndian.PutUint64(hdr[n:], uint64(payloadLen))
		n += 8
	}

	var mask [4]byte
	if !ws.isServer {
		if _, err := io.ReadFull(rand.Reader, mask[:]); err != nil {
			return fmt.Errorf("writeFrame (mask): %w", err)
		}
		copy(hdr[n:n+4], mask[:])
		n += 4
	}

	if _, err := ws.conn.Write(hdr[:n]); err != nil {
		return err
	}
	if payloadLen > 0 {
		out := data
		if !ws.isServer {
			out = make([]byte, payloadLen)
			for i, b := range data {
				out[i] = b ^ mask[i%4]
			}
		}
		if _, err := ws.conn.Write(out); err != nil {
			return err
		}
	}
	return nil
}

// readFrame reads one frame and returns its (payload, opcode, error).
// Fragmentation is not supported; each message must fit in one frame.
func (ws *TWsConn) readFrame() ([]byte, byte, error) {
	hdr := make([]byte, 2)
	if _, err := io.ReadFull(ws.br, hdr); err != nil {
		return nil, 0, fmt.Errorf("readFrame (header): %w", err)
	}
	opcode := hdr[0] & 0x0F
	masked := hdr[1]&0x80 != 0
	length := int(hdr[1] & 0x7F)
	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(ws.br, ext[:]); err != nil {
			return nil, 0, fmt.Errorf("readFrame (len16): %w", err)
		}
		length = int(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(ws.br, ext[:]); err != nil {
			return nil, 0, fmt.Errorf("readFrame (len64): %w", err)
		}
		length = int(binary.BigEndian.Uint64(ext[:]))
	}
	if length > MaxWsFrameBytes {
		return nil, 0, fmt.Errorf("readFrame: payload too large (%d > %d)", length, MaxWsFrameBytes)
	}

	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(ws.br, mask[:]); err != nil {
			return nil, 0, fmt.Errorf("readFrame (mask): %w", err)
		}
	}
	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(ws.br, payload); err != nil {
			return nil, 0, fmt.Errorf("readFrame (payload): %w", err)
		}
		if masked {
			for i, b := range payload {
				payload[i] = b ^ mask[i%4]
			}
		}
	}
	return payload, opcode, nil
}
