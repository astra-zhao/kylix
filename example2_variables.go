package main

import (
	"fmt"
)

var name string
var age int64
var score float64
var passed bool
func main() {
//line /tmp/example2_variables.klx:8
name = "Alice"
//line /tmp/example2_variables.klx:9
age = 25
//line /tmp/example2_variables.klx:10
score = 89.500000
//line /tmp/example2_variables.klx:11
passed = (score >= 60.000000)
//line /tmp/example2_variables.klx:13
fmt.Println("Name: ", name)
//line /tmp/example2_variables.klx:14
fmt.Println("Age: ", fmt.Sprintf("%d", age))
//line /tmp/example2_variables.klx:15
fmt.Println("Score: ", RealToStr(score))
//line /tmp/example2_variables.klx:16
if passed	 {
//line /tmp/example2_variables.klx:17
fmt.Println("Status: Passed")
}	 else {
//line /tmp/example2_variables.klx:19
fmt.Println("Status: Failed")
	}
}
