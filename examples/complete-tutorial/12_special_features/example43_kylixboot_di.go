package main

import (
	"kylix/stdlib"
	"fmt"
)

type TUserService struct {
}

func (self *TUserService) Greeting() string {
var result string
//line example43_kylixboot_di.klx:9
result = "Hello from service"
	return result
}

type TUserController struct {
UserService *TUserService
}

func (self *TUserController) Hello(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example43_kylixboot_di.klx:22
result = stdlib.BootText(200, self.UserService.Greeting())
	return result
}

func main() {
	__kylix_svc_TUserService := &TUserService{}
	stdlib.BootRegisterInstance("TUserService", __kylix_svc_TUserService)
	stdlib.BootRegisterInstance("UserService", __kylix_svc_TUserService)
	__kylix_ctrl_TUserController := &TUserController{}
	__kylix_ctrl_TUserController.UserService = __kylix_svc_TUserService
	stdlib.BootGET("/di/hello", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TUserController.Hello(req)
	})
//line example43_kylixboot_di.klx:27
fmt.Println("KylixBoot DI auto-wire OK")
}
