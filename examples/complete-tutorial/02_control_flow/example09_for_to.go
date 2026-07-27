package main

import (
	"fmt"
)

var i int64
var sum int64
var j int64
func main() {
//line example09_for_to.klx:10
fmt.Println("Counting up:")
//line example09_for_to.klx:11
for i = 1; i <= 5; i++ {
//line example09_for_to.klx:13
fmt.Println(i)
	}
//line example09_for_to.klx:17
fmt.Println("Counting down:")
//line example09_for_to.klx:18
for i = 5; i >= 1; i-- {
//line example09_for_to.klx:20
fmt.Println(i)
	}
//line example09_for_to.klx:24
sum = 0
//line example09_for_to.klx:25
for i = 1; i <= 100; i++ {
//line example09_for_to.klx:27
sum = (sum + i)
	}
//line example09_for_to.klx:29
fmt.Println("Sum of 1-100: ", sum)
//line example09_for_to.klx:32
fmt.Println("Multiplication table (3x3):")
//line example09_for_to.klx:33
for i = 1; i <= 3; i++ {
//line example09_for_to.klx:35
for j = 1; j <= 3; j++ {
//line example09_for_to.klx:37
fmt.Println(i, " x ", j, " = ", (i * j))
		}
	}
}
