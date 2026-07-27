package main

import (
	"fmt"
)

var a, b int64
var x, y float64
var flag1, flag2 bool
func main() {
//line example05_operators.klx:10
a = 10
//line example05_operators.klx:11
b = 3
//line example05_operators.klx:13
fmt.Println("=== Arithmetic ===")
//line example05_operators.klx:14
fmt.Println("a + b = ", (a + b))
//line example05_operators.klx:15
fmt.Println("a - b = ", (a - b))
//line example05_operators.klx:16
fmt.Println("a * b = ", (a * b))
//line example05_operators.klx:17
fmt.Println("a / b = ", (a / b))
//line example05_operators.klx:18
fmt.Println("a mod b = ", (a % b))
//line example05_operators.klx:19
fmt.Println("a div b = ", (a / b))
//line example05_operators.klx:22
fmt.Println("")
//line example05_operators.klx:23
fmt.Println("=== Comparison ===")
//line example05_operators.klx:24
fmt.Println("a > b: ", (a > b))
//line example05_operators.klx:25
fmt.Println("a < b: ", (a < b))
//line example05_operators.klx:26
fmt.Println("a >= b: ", (a >= b))
//line example05_operators.klx:27
fmt.Println("a <= b: ", (a <= b))
//line example05_operators.klx:28
fmt.Println("a = b: ", (a == b))
//line example05_operators.klx:29
fmt.Println("a <> b: ", (a != b))
//line example05_operators.klx:32
fmt.Println("")
//line example05_operators.klx:33
fmt.Println("=== Logical ===")
//line example05_operators.klx:34
flag1 = true
//line example05_operators.klx:35
flag2 = false
//line example05_operators.klx:37
fmt.Println("flag1 and flag2: ", (flag1 && flag2))
//line example05_operators.klx:38
fmt.Println("flag1 or flag2: ", (flag1 || flag2))
//line example05_operators.klx:39
fmt.Println("not flag1: ", (!flag1))
//line example05_operators.klx:40
fmt.Println("not flag2: ", (!flag2))
}
