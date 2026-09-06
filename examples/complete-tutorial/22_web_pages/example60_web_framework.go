package main

import (
	"kylix/stdlib"
	"fmt"
)

type TWebController struct {
}

func (self *TWebController) NewPage(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example60_web_framework.klx:24
result = stdlib.BootHTML(200, "you made it")
	return result
}

func (self *TWebController) OldPage(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example60_web_framework.klx:30
result = stdlib.BootHTML(200, "old page")
//line example60_web_framework.klx:31
result = result.Redirect("/api/new")
	return result
}

func (self *TWebController) BoomPage(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example60_web_framework.klx:37
result = stdlib.BootHTML(200, "never")
//line example60_web_framework.klx:38
panic(&Exception{"handler blew up"})
	return result
}

// Kylix runtime exception base type
type Exception struct {
	Message string
}

func (e *Exception) Error() string { return e.Message }

func main() {
	__kylix_ctrl_TWebController := &TWebController{}
	stdlib.BootGET("/api/new", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TWebController.NewPage(req)
	})
	stdlib.BootGET("/api/old", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TWebController.OldPage(req)
	})
	stdlib.BootGET("/api/boom", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		return __kylix_ctrl_TWebController.BoomPage(req)
	})
//line example60_web_framework.klx:43
stdlib.BootNotFoundPage("<html><body><h1>Custom 404</h1></body></html>")
//line example60_web_framework.klx:44
stdlib.BootErrorPage("<html><body><h1>Custom 500</h1></body></html>")
//line example60_web_framework.klx:45
fmt.Println("web framework demo on :8077")
//line example60_web_framework.klx:46
stdlib.BootRun(8077)
}
