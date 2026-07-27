package main

import (
	"fmt"
)

// Kylix runtime exception base type
type Exception struct {
	Message string
}

func (e *Exception) Error() string { return e.Message }

//line example27_try_except.klx:4
func SafeDivide(a float64, b float64) float64 {
var result float64
//line example27_try_except.klx:6
if (b == 0)	 {
//line example27_try_except.klx:8
panic(&Exception{"Division by zero"})
}	
//line example27_try_except.klx:10
result = (a / b)
return result
}

func main() {
//line example27_try_except.klx:15
	func() {
		defer func() {
			if r := recover(); r != nil {
//line example27_try_except.klx:22
fmt.Println("An error occurred")
			}
		}()
//line example27_try_except.klx:17
result := SafeDivide(10, 2)
//line example27_try_except.klx:18
fmt.Println("10 / 2 = ", result)
	}()
//line example27_try_except.klx:27
	func() {
		defer func() {
			if r := recover(); r != nil {
//line example27_try_except.klx:34
fmt.Println("Cannot divide by zero!")
			}
		}()
//line example27_try_except.klx:29
result := SafeDivide(10, 0)
//line example27_try_except.klx:30
fmt.Println("Result: ", result)
	}()
}
