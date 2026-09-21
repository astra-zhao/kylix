package main

import (
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

func main() {
//line example45_validation_annotations.klx:17
user := &TCreateUser{}
//line example45_validation_annotations.klx:18
user.Email = "bad-email"
//line example45_validation_annotations.klx:19
user.Password = "short"
//line example45_validation_annotations.klx:20
user.Age = 16
//line example45_validation_annotations.klx:22
if (!user.IsValid())	 {
//line example45_validation_annotations.klx:23
fmt.Println("Validation failed")
}	 else {
//line example45_validation_annotations.klx:25
fmt.Println("Validation passed")
	}
}
