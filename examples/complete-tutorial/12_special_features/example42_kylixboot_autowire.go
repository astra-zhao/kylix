package main

import (
	"fmt"
	"kylix/stdlib"
)

type THelloController struct {
}

func (self *THelloController) Hello(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example42_kylixboot_autowire.klx:10
result = stdlib.BootText(200, "Hello from auto-wire")
	return result
}

func main() {
	__kylix_ctrl_THelloController := &THelloController{}
	stdlib.BootGET("/api/hello", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_THelloController.Hello(req)
	})
//line example42_kylixboot_autowire.klx:15
fmt.Println("KylixBoot auto-wire OK")
}
