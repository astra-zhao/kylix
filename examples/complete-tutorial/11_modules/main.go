package main

import (
	"fmt"
)

//line math_helper.klx:11
func Square(x int64) int64 {
var result int64
//line math_helper.klx:13
result = (x * x)
return result
}

//line math_helper.klx:16
func Cube(x int64) int64 {
var result int64
//line math_helper.klx:18
result = ((x * x) * x)
return result
}

//line math_helper.klx:21
func IsEven(n int64) bool {
var result bool
//line math_helper.klx:23
result = ((n % 2) == 0)
return result
}

func main() {
//line example33_use_module.klx:5
fmt.Println("Square of 5: ", Square(5))
//line example33_use_module.klx:6
fmt.Println("Cube of 3: ", Cube(3))
//line example33_use_module.klx:8
fmt.Println("Is 4 even? ", IsEven(4))
//line example33_use_module.klx:9
fmt.Println("Is 7 even? ", IsEven(7))
//line example33_use_module.klx:11
fmt.Println("Square of 10: ", Square(10))
//line example33_use_module.klx:12
fmt.Println("Cube of 4: ", Cube(4))
}
