package main

import (
	"kylix/stdlib"
	"fmt"
)

type TAdminController struct {
}

func (self *TAdminController) Dashboard(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example46_security_annotations.klx:11
result = stdlib.BootText(200, "admin dashboard")
	return result
}

func (self *TAdminController) ListUsers(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example46_security_annotations.klx:18
result = stdlib.BootText(200, "admin users")
	return result
}

func main() {
	__kylix_ctrl_TAdminController := &TAdminController{}
	stdlib.BootGET("/admin/dashboard", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }
		return __kylix_ctrl_TAdminController.Dashboard(req)
	})
	stdlib.BootGET("/admin/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }
		if __r := stdlib.BootEnforceRole(req, "admin"); __r != nil { return __r }
		return __kylix_ctrl_TAdminController.ListUsers(req)
	})
//line example46_security_annotations.klx:23
fmt.Println("KylixBoot security annotations OK")
}
