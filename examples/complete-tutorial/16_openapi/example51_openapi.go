package main

import (
	"kylix/stdlib"
	"fmt"
	"strings"
	"regexp"
)

type TCreateUser struct {
Email string
Password string
Age int64
}

func (self *TCreateUser) Validate() map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(self.Email) == "" {
		errors["Email"] = "is required"
	}
	if self.Email != "" && !regexp.MustCompile("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$").MatchString(self.Email) {
		errors["Email"] = "must be a valid email address"
	}
	if strings.TrimSpace(self.Password) == "" {
		errors["Password"] = "is required"
	}
	if len(self.Password) < 8 {
		errors["Password"] = "must be at least 8 characters"
	}
	if self.Age < 18 {
		errors["Age"] = "must be at least 18"
	}
	return errors
}

func (self *TCreateUser) IsValid() bool {
	return len(self.Validate()) == 0
}

func (self *TCreateUser) ToRow() map[string]interface{} {
	return map[string]interface{}{
		"Email": self.Email,
		"Password": self.Password,
		"Age": self.Age,
	}
}

func (self *TCreateUser) FromRow(row map[string]interface{}) {
	if v, ok := row["Email"].(string); ok {
		self.Email = v
	}
	if v, ok := row["Password"].(string); ok {
		self.Password = v
	}
	if v, ok := row["Age"].(int64); ok {
		self.Age = v
	}
}

type TUserController struct {
}

func (self *TUserController) ListUsers(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example51_openapi.klx:24
result = stdlib.BootJSON(200, nil)
	return result
}

func (self *TUserController) CreateUser(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example51_openapi.klx:31
result = stdlib.BootText(200, "created")
	return result
}

func (self *TUserController) GetUser(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example51_openapi.klx:38
result = stdlib.BootJSON(200, nil)
	return result
}

func (self *TUserController) DeleteUser(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example51_openapi.klx:45
result = stdlib.BootJSON(200, nil)
	return result
}

func main() {
	__kylix_ctrl_TUserController := &TUserController{}
	stdlib.BootGET("/api/v1/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }
		return __kylix_ctrl_TUserController.ListUsers(req)
	})
	stdlib.BootPOST("/api/v1/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		var __body TCreateUser
		if err := stdlib.BootReadJSON(req, &__body); err != nil {
			return stdlib.BootJSON(400, map[string]string{"error": "invalid JSON"})
		}
		if !__body.IsValid() {
			return stdlib.BootJSON(400, __body.Validate())
		}
		return __kylix_ctrl_TUserController.CreateUser(req)
	})
	stdlib.BootGET("/api/v1/users/:id", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }
		return __kylix_ctrl_TUserController.GetUser(req)
	})
	stdlib.BootDELETE("/api/v1/users/:id", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		if __r := stdlib.BootEnforceAuth(req); __r != nil { return __r }
		if __r := stdlib.BootEnforceRole(req, "admin"); __r != nil { return __r }
		return __kylix_ctrl_TUserController.DeleteUser(req)
	})
//line example51_openapi.klx:50
fmt.Println("OpenAPI demo OK")
}
