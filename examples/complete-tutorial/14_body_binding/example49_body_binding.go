package main

import (
	"strings"
	"regexp"
	"kylix/stdlib"
	"fmt"
)

type TCreateUser struct {
Email string
Password string
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
	return errors
}

func (self *TCreateUser) IsValid() bool {
	return len(self.Validate()) == 0
}

func (self *TCreateUser) ToRow() map[string]interface{} {
	return map[string]interface{}{
		"Email": self.Email,
		"Password": self.Password,
	}
}

func (self *TCreateUser) FromRow(row map[string]interface{}) {
	if v, ok := row["Email"].(string); ok {
		self.Email = v
	}
	if v, ok := row["Password"].(string); ok {
		self.Password = v
	}
}

type TUserController struct {
}

func (self *TUserController) CreateUser(req *stdlib.BootRequest) *stdlib.BootResponse {
var result *stdlib.BootResponse
//line example49_body_binding.klx:22
result = stdlib.BootText(201, "user created")
	return result
}

func main() {
	__kylix_ctrl_TUserController := &TUserController{}
	stdlib.BootPOST("/api/users", func(req *stdlib.BootRequest) *stdlib.BootResponse {
		var __body TCreateUser
		if err := stdlib.BootReadJSON(req, &__body); err != nil {
			return stdlib.BootJSON(400, map[string]string{"error": "invalid JSON"})
		}
		if !__body.IsValid() {
			return stdlib.BootJSON(400, __body.Validate())
		}
		return __kylix_ctrl_TUserController.CreateUser(req)
	})
//line example49_body_binding.klx:27
fmt.Println("Body binding annotations OK")
}
