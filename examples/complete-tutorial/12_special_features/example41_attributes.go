package main

import (
	"fmt"
	"kylix/stdlib"
)

type TUserRepository struct {
}

func (self *TUserRepository) Name() string {
var result string
//line example41_attributes.klx:19
result = "UserRepo"
	return result
}

type TUserController struct {
Repo *TUserRepository
}

func (self *TUserController) ListUsers(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example41_attributes.klx:33
result = stdlib.BootText(200, ("GET /api/users → " + self.Repo.Name()))
	return result
}

func (self *TUserController) CreateUser(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example41_attributes.klx:39
result = stdlib.BootText(201, "POST /api/users → created")
	return result
}

func (self *TUserController) GetUser(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example41_attributes.klx:45
result = stdlib.BootText(200, "GET /api/users/:id")
	return result
}

type TAppConfig struct {
AppName string
Port int64
}

//line example41_attributes.klx:61
func HealthCheck() string {
var result string
//line example41_attributes.klx:63
result = "OK"
return result
}

func main() {
	__kylix_svc_TUserRepository := &TUserRepository{}
	stdlib.BootRegisterInstance("TUserRepository", __kylix_svc_TUserRepository)
	stdlib.BootRegisterInstance("UserRepository", __kylix_svc_TUserRepository)
	__kylix_svc_TAppConfig := &TAppConfig{}
	stdlib.BootRegisterInstance("TAppConfig", __kylix_svc_TAppConfig)
	stdlib.BootRegisterInstance("AppConfig", __kylix_svc_TAppConfig)
	__kylix_ctrl_TUserController := &TUserController{}
	__kylix_ctrl_TUserController.Repo = __kylix_svc_TUserRepository
	stdlib.BootGET("/api/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TUserController.ListUsers(req)
	})
	stdlib.BootPOST("/api/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TUserController.CreateUser(req)
	})
	stdlib.BootGET("/api/users/:id", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TUserController.GetUser(req)
	})
//line example41_attributes.klx:67
repo := &TUserRepository{}
//line example41_attributes.klx:68
ctrl := &TUserController{}
//line example41_attributes.klx:69
ctrl.Repo = repo
//line example41_attributes.klx:70
fmt.Println(ctrl.ListUsers(nil).Body)
//line example41_attributes.klx:71
fmt.Println(ctrl.CreateUser(nil).Body)
//line example41_attributes.klx:72
fmt.Println(ctrl.GetUser(nil).Body)
//line example41_attributes.klx:74
cfg := &TAppConfig{}
//line example41_attributes.klx:75
cfg.AppName = "KylixDemo"
//line example41_attributes.klx:76
cfg.Port = 9090
//line example41_attributes.klx:77
fmt.Println(("App: " + cfg.AppName))
//line example41_attributes.klx:78
fmt.Println(("Port: " + fmt.Sprintf("%d", cfg.Port)))
//line example41_attributes.klx:80
fmt.Println(HealthCheck())
}
