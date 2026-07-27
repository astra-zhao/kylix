package main

import (
	"fmt"
)

//line example13_functions.klx:4
func Add(a int64, b int64) int64 {
var result int64
//line example13_functions.klx:6
result = (a + b)
return result
}

//line example13_functions.klx:10
func Average(x float64, y float64, z float64) float64 {
var result float64
//line example13_functions.klx:12
result = (((x + y) + z) / 3)
return result
}

//line example13_functions.klx:16
func GetPI() float64 {
var result float64
//line example13_functions.klx:18
result = 3.141593
return result
}

//line example13_functions.klx:22
func Greet(name string) {
//line example13_functions.klx:24
fmt.Println("Hello, ", name, "!")
}

//line example13_functions.klx:28
func PrintInfo(name string, age int64) {
//line example13_functions.klx:30
fmt.Println("Name: ", name)
//line example13_functions.klx:31
fmt.Println("Age: ", age)
//line example13_functions.klx:32
fmt.Println("---")
}

func main() {
//line example13_functions.klx:36
fmt.Println("5 + 3 = ", Add(5, 3))
//line example13_functions.klx:37
fmt.Println("Average of 10, 20, 30 = ", Average(10, 20, 30))
//line example13_functions.klx:38
fmt.Println("PI = ", GetPI())
//line example13_functions.klx:40
Greet("Alice")
//line example13_functions.klx:41
Greet("Bob")
//line example13_functions.klx:43
PrintInfo("Charlie", 25)
//line example13_functions.klx:44
PrintInfo("Diana", 30)
}
