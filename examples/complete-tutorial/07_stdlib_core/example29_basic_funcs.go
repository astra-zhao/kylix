package main

import (
	"fmt"
)

//line example29_basic_funcs.klx:4
func Max(a int64, b int64) int64 {
var result int64
//line example29_basic_funcs.klx:6
if (a > b)	 {
//line example29_basic_funcs.klx:7
result = a
}	 else {
//line example29_basic_funcs.klx:9
result = b
	}
return result
}

//line example29_basic_funcs.klx:12
func Min(a int64, b int64) int64 {
var result int64
//line example29_basic_funcs.klx:14
if (a < b)	 {
//line example29_basic_funcs.klx:15
result = a
}	 else {
//line example29_basic_funcs.klx:17
result = b
	}
return result
}

//line example29_basic_funcs.klx:20
func Abs(n int64) int64 {
var result int64
//line example29_basic_funcs.klx:22
if (n < 0)	 {
//line example29_basic_funcs.klx:23
result = (-n)
}	 else {
//line example29_basic_funcs.klx:25
result = n
	}
return result
}

func main() {
//line example29_basic_funcs.klx:29
fmt.Println("Max(10, 20): ", Max(10, 20))
//line example29_basic_funcs.klx:30
fmt.Println("Min(10, 20): ", Min(10, 20))
//line example29_basic_funcs.klx:31
fmt.Println("Abs(-15): ", Abs((-15)))
//line example29_basic_funcs.klx:32
fmt.Println("Abs(25): ", Abs(25))
}
