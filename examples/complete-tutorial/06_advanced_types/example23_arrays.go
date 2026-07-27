package main

import (
	"fmt"
)

var numbers [((4 - 0) + 1)]int64
var names [((2 - 0) + 1)]string
var i int64
func main() {
//line example23_arrays.klx:10
numbers[0] = 10
//line example23_arrays.klx:11
numbers[1] = 20
//line example23_arrays.klx:12
numbers[2] = 30
//line example23_arrays.klx:13
numbers[3] = 40
//line example23_arrays.klx:14
numbers[4] = 50
//line example23_arrays.klx:17
fmt.Println("Numbers array:")
//line example23_arrays.klx:18
for i = 0; i <= 4; i++ {
//line example23_arrays.klx:20
fmt.Println("numbers[", i, "] = ", numbers[i])
	}
//line example23_arrays.klx:24
names[0] = "Alice"
//line example23_arrays.klx:25
names[1] = "Bob"
//line example23_arrays.klx:26
names[2] = "Charlie"
//line example23_arrays.klx:28
fmt.Println("Names:")
//line example23_arrays.klx:29
for i = 0; i <= 2; i++ {
//line example23_arrays.klx:31
fmt.Println((i + 1), ". ", names[i])
	}
}
