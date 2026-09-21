package main

import (
	"fmt"
	"kylix/stdlib"
)

type TAuthController struct {
}

func (self *TAuthController) Login(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example50_jwt_auth.klx:11
token := func() string { _v, _ := stdlib.JwtSign("my-secret", "user1", 3600, nil); return _v }()
//line example50_jwt_auth.klx:12
result = stdlib.BootText(200, token)
	return result
}

func (self *TAuthController) Me(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example50_jwt_auth.klx:20
result = stdlib.BootText(200, "JWT demo OK")
	return result
}

func main() {
	__kylix_ctrl_TAuthController := &TAuthController{}
	stdlib.BootPOST("/api/login", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TAuthController.Login(req)
	})
	stdlib.BootGET("/api/me", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }
		return __kylix_ctrl_TAuthController.Me(req)
	})
//line example50_jwt_auth.klx:26
stdlib.BootRegisterJwtAuth("my-secret")
//line example50_jwt_auth.klx:27
fmt.Println("JWT demo OK")
}
