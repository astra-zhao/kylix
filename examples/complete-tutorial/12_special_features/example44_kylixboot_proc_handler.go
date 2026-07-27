package main

import (
	"kylix/stdlib"
	"fmt"
)

type TProcController struct {
}

func (self *TProcController) Hello(req *stdlib.BootRequest, res *stdlib.BootResponse) {
//line example44_kylixboot_proc_handler.klx:10
res.StatusCode(200)
//line example44_kylixboot_proc_handler.klx:11
res.Send("Hello from procedure handler")
}

func main() {
	__kylix_ctrl_TProcController := &TProcController{}
	stdlib.BootGET("/proc/hello", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		res := stdlib.BootText(200, "")
		__kylix_ctrl_TProcController.Hello(req, res)
		return res
	})
//line example44_kylixboot_proc_handler.klx:16
fmt.Println("KylixBoot procedure handler OK")
}
