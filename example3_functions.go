package main

import (
	"fmt"
)

var x, y, sum int64
var fact int64
//line /tmp/example3_functions.klx:3
func Add(a int64, b int64) int64 {
var result int64
//line /tmp/example3_functions.klx:5
result = (a + b)
return result
}

//line /tmp/example3_functions.klx:8
func Factorial(n int64) int64 {
var result int64
//line /tmp/example3_functions.klx:10
if (n <= 1)	 {
//line /tmp/example3_functions.klx:11
result = 1
}	 else {
//line /tmp/example3_functions.klx:13
result = (n * Factorial((n - 1)))
	}
return result
}

//line /tmp/example3_functions.klx:16
func Greet(name string) {
//line /tmp/example3_functions.klx:18
fmt.Println("Hello, ", name, "!")
}

func main() {
//line /tmp/example3_functions.klx:25
x = 10
//line /tmp/example3_functions.klx:26
y = 20
//line /tmp/example3_functions.klx:27
sum = Add(x, y)
//line /tmp/example3_functions.klx:28
fmt.Println(fmt.Sprintf("%d", x), " + ", fmt.Sprintf("%d", y), " = ", fmt.Sprintf("%d", sum))
//line /tmp/example3_functions.klx:30
fact = Factorial(5)
//line /tmp/example3_functions.klx:31
fmt.Println("5! = ", fmt.Sprintf("%d", fact))
//line /tmp/example3_functions.klx:33
Greet("Kylix User")
}
