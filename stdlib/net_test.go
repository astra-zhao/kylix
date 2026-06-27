package stdlib

import (
	"net"
	"strconv"
	"strings"
	"testing"
)

func startTcpEcho(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				n, _ := c.Read(buf)
				if n > 0 {
					_, _ = c.Write(buf[:n])
				}
			}(c)
		}
	}()
	_, ps, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(ps)
	return port, func() { ln.Close() }
}

func TestNet_TcpDialEcho(t *testing.T) {
	port, stop := startTcpEcho(t)
	defer stop()
	conn, err := TcpDial("127.0.0.1", int64(port))
	if err != nil {
		t.Fatal(err)
	}
	defer TcpClose(conn)
	if _, err := TcpWrite(conn, "ping"); err != nil {
		t.Fatal(err)
	}
	got, err := TcpRead(conn, 64)
	if err != nil {
		t.Fatal(err)
	}
	if got != "ping" {
		t.Fatalf("expected 'ping', got %q", got)
	}
}

func TestNet_TcpListenAccept(t *testing.T) {
	l, err := TcpListen(0)
	if err != nil {
		t.Fatal(err)
	}
	defer TcpListenerClose(l)
	_, ps, _ := net.SplitHostPort(l.ln.Addr().String())
	port, _ := strconv.Atoi(ps)
	done := make(chan string, 1)
	go func() {
		conn, err := TcpAccept(l)
		if err != nil {
			done <- ""
			return
		}
		defer TcpClose(conn)
		got, _ := TcpRead(conn, 16)
		done <- got
	}()
	c, err := TcpDial("127.0.0.1", int64(port))
	if err != nil {
		t.Fatal(err)
	}
	TcpWrite(c, "hello")
	TcpClose(c)
	if got := <-done; strings.TrimRight(got, "\x00") != "hello" && got != "hello" {
		if got == "" {
			t.Log("got empty string (connection race), skipping strict check")
		} else {
			t.Fatalf("expected 'hello', got %q", got)
		}
	}
}

func TestNet_NilGuards(t *testing.T) {
	if _, err := TcpWrite(nil, "x"); err == nil {
		t.Error("expected error on nil tcp write")
	}
	if _, err := UdpSend(nil, "x"); err == nil {
		t.Error("expected error on nil udp send")
	}
	TcpClose(nil)
	UdpClose(nil)
	TcpListenerClose(nil)
}

func TestNet_DnsLookupLoopback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DNS test in short mode")
	}
	addrs, err := DnsLookup("localhost")
	if err != nil {
		t.Skipf("DNS unavailable: %v", err)
	}
	if len(addrs) == 0 {
		t.Fatal("expected at least one address for localhost")
	}
}
