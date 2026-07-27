package main

import (
	"fmt"
)

var x int64
var age int64
var score int64
func main() {
//line example07_if_else.klx:10
x = 10
//line example07_if_else.klx:11
if (x > 5)	 {
//line example07_if_else.klx:13
fmt.Println("x is greater than 5")
}	
//line example07_if_else.klx:17
age = 18
//line example07_if_else.klx:18
if (age >= 18)	 {
//line example07_if_else.klx:20
fmt.Println("You are an adult")
}	 else {
//line example07_if_else.klx:24
fmt.Println("You are a minor")
	}
//line example07_if_else.klx:28
score = 85
//line example07_if_else.klx:29
if (score >= 90)	 {
//line example07_if_else.klx:31
fmt.Println("Grade: A")
}	 else {
//line example07_if_else.klx:33
if (score >= 80)		 {
//line example07_if_else.klx:35
fmt.Println("Grade: B")
}		 else {
//line example07_if_else.klx:37
if (score >= 70)			 {
//line example07_if_else.klx:39
fmt.Println("Grade: C")
}			 else {
//line example07_if_else.klx:43
fmt.Println("Grade: F")
			}
		}
	}
}
