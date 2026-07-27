package main

import (
	"kylix/stdlib"
	"fmt"
)

func main() {
//line example55_websocket.klx:6
ln := func() *stdlib.TTcpListener { _v, _ := stdlib.TcpListen(18080); return _v }()
//line example55_websocket.klx:18
stdlib.TcpListenerClose(ln)
//line example55_websocket.klx:19
fmt.Println("websocket demo OK")
}
