// net.go — Kylix stdlib net module: TCP/UDP clients + DNS lookup.
package stdlib

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type TTcpConn struct{ conn net.Conn }
type TTcpListener struct{ ln net.Listener }
type TUdpConn struct{ conn *net.UDPConn }

func TcpDial(host string, port int64) (*TTcpConn, error) {
	c, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.FormatInt(port, 10)), 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &TTcpConn{conn: c}, nil
}

func TcpWrite(c *TTcpConn, data string) (int64, error) {
	if c == nil || c.conn == nil {
		return 0, fmt.Errorf("net: nil tcp connection")
	}
	n, err := c.conn.Write([]byte(data))
	return int64(n), err
}

func TcpRead(c *TTcpConn, n int64) (string, error) {
	if c == nil || c.conn == nil {
		return "", fmt.Errorf("net: nil tcp connection")
	}
	buf := make([]byte, n)
	r, err := c.conn.Read(buf)
	if r > 0 {
		return string(buf[:r]), err
	}
	return "", err
}

func TcpClose(c *TTcpConn) {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func TcpListen(port int64) (*TTcpListener, error) {
	ln, err := net.Listen("tcp", ":"+strconv.FormatInt(port, 10))
	if err != nil {
		return nil, err
	}
	return &TTcpListener{ln: ln}, nil
}

func TcpAccept(l *TTcpListener) (*TTcpConn, error) {
	if l == nil || l.ln == nil {
		return nil, fmt.Errorf("net: nil tcp listener")
	}
	c, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}
	return &TTcpConn{conn: c}, nil
}

func TcpListenerClose(l *TTcpListener) {
	if l != nil && l.ln != nil {
		_ = l.ln.Close()
	}
}

func UdpDial(host string, port int64) (*TUdpConn, error) {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.FormatInt(port, 10)))
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	return &TUdpConn{conn: c}, nil
}

func UdpSend(c *TUdpConn, data string) (int64, error) {
	if c == nil || c.conn == nil {
		return 0, fmt.Errorf("net: nil udp connection")
	}
	n, err := c.conn.Write([]byte(data))
	return int64(n), err
}

func UdpRecv(c *TUdpConn, n int64) (string, error) {
	if c == nil || c.conn == nil {
		return "", fmt.Errorf("net: nil udp connection")
	}
	buf := make([]byte, n)
	r, err := c.conn.Read(buf)
	if r > 0 {
		return string(buf[:r]), err
	}
	return "", err
}

func UdpClose(c *TUdpConn) {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func DnsLookup(host string) ([]string, error) {
	return net.LookupHost(host)
}

func DnsLookupCNAME(host string) (string, error) {
	return net.LookupCNAME(host)
}
