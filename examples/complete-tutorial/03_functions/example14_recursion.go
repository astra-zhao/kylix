package main

import (
	"fmt"
)

//line example14_recursion.klx:4
func Factorial(n int64) int64 {
var result int64
//line example14_recursion.klx:6
if (n <= 1)	 {
//line example14_recursion.klx:7
result = 1
}	 else {
//line example14_recursion.klx:9
result = (n * Factorial((n - 1)))
	}
return result
}

//line example14_recursion.klx:13
func Fibonacci(n int64) int64 {
var result int64
//line example14_recursion.klx:15
if (n <= 1)	 {
//line example14_recursion.klx:16
result = n
}	 else {
//line example14_recursion.klx:18
result = (Fibonacci((n - 1)) + Fibonacci((n - 2)))
	}
return result
}

//line example14_recursion.klx:22
func Power(base int64, exp int64) int64 {
var result int64
//line example14_recursion.klx:24
if (exp == 0)	 {
//line example14_recursion.klx:25
result = 1
}	 else {
//line example14_recursion.klx:27
result = (base * Power(base, (exp - 1)))
	}
return result
}

func main() {
//line example14_recursion.klx:31
fmt.Println("Factorial of 5: ", Factorial(5))
//line example14_recursion.klx:32
fmt.Println("Factorial of 10: ", Factorial(10))
//line example14_recursion.klx:34
fmt.Println("Fibonacci sequence:")
//line example14_recursion.klx:35
fmt.Println("F(0) = ", Fibonacci(0))
//line example14_recursion.klx:36
fmt.Println("F(1) = ", Fibonacci(1))
//line example14_recursion.klx:37
fmt.Println("F(5) = ", Fibonacci(5))
//line example14_recursion.klx:38
fmt.Println("F(10) = ", Fibonacci(10))
//line example14_recursion.klx:40
fmt.Println("2^3 = ", Power(2, 3))
//line example14_recursion.klx:41
fmt.Println("5^4 = ", Power(5, 4))
}
