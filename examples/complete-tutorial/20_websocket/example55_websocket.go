package main

import (
	"kylix/stdlib"
	"fmt"
)

func main() {
//line example55_websocket.klx:9
ln := func() *stdlib.TTcpListener { _v, _ := stdlib.TcpListen(18080); return _v }()
//line example55_websocket.klx:10
cli := func() *stdlib.TWsConn { _v, _ := stdlib.WsDialConnect("127.0.0.1:18080", "/echo"); return _v }()
//line example55_websocket.klx:11
if (cli == nil)	 {
//line example55_websocket.klx:13
fmt.Println("connect failed")
//line example55_websocket.klx:14
return
}	
//line example55_websocket.klx:16
tcp := func() *stdlib.TTcpConn { _v, _ := stdlib.TcpAccept(ln); return _v }()
//line example55_websocket.klx:17
srv := func() *stdlib.TWsConn { _v, _ := stdlib.WsAccept(tcp); return _v }()
//line example55_websocket.klx:18
if (!stdlib.WsDialFinish(cli))	 {
//line example55_websocket.klx:20
fmt.Println("handshake failed")
//line example55_websocket.klx:21
return
}	
//line example55_websocket.klx:23
stdlib.WsSend(srv, "hello from server")
//line example55_websocket.klx:24
fmt.Println(("client got: " + func() string { _v, _ := stdlib.WsRecv(cli); return _v }()))
//line example55_websocket.klx:25
stdlib.WsClose(cli)
//line example55_websocket.klx:26
stdlib.WsClose(srv)
//line example55_websocket.klx:27
stdlib.TcpListenerClose(ln)
//line example55_websocket.klx:28
fmt.Println("websocket demo OK")
}
