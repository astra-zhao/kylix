package main

import (
	"fmt"
	"errors"
)

var name string
var err error
var e2 error
//line example58_error_type.klx:8
func FindUser(id int64) (string, error) {
//line example58_error_type.klx:10
if (id <= 0)	 {
//line example58_error_type.klx:11
return "", errors.New("invalid id")
}	 else {
//line example58_error_type.klx:13
return "user", nil
	}
}

//line example58_error_type.klx:17
func CheckAge(age int64) error {
var result error
//line example58_error_type.klx:19
if (age < 0)	 {
//line example58_error_type.klx:20
result = errors.New("age cannot be negative")
}	 else {
//line example58_error_type.klx:22
result = nil
	}
return result
}

func main() {
//line example58_error_type.klx:31
name, err = FindUser((-1))
//line example58_error_type.klx:32
if (err != nil)	 {
//line example58_error_type.klx:33
fmt.Println("Error: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example58_error_type.klx:35
fmt.Println("Name: ", name)
	}
//line example58_error_type.klx:37
name, err = FindUser(5)
//line example58_error_type.klx:38
if (err != nil)	 {
//line example58_error_type.klx:39
fmt.Println("Error: ", func() string { if e := err; e != nil { return e.Error() }; return "" }())
}	 else {
//line example58_error_type.klx:41
fmt.Println("Name: ", name)
	}
//line example58_error_type.klx:43
e2 = CheckAge((-3))
//line example58_error_type.klx:44
if (e2 != nil)	 {
//line example58_error_type.klx:45
fmt.Println("CheckAge: ", func() string { if e := e2; e != nil { return e.Error() }; return "" }())
}	 else {
//line example58_error_type.klx:47
fmt.Println("CheckAge: ok")
	}
//line example58_error_type.klx:49
e2 = CheckAge(30)
//line example58_error_type.klx:50
if (e2 != nil)	 {
//line example58_error_type.klx:51
fmt.Println("CheckAge: ", func() string { if e := e2; e != nil { return e.Error() }; return "" }())
}	 else {
//line example58_error_type.klx:53
fmt.Println("CheckAge: ok")
	}
}
