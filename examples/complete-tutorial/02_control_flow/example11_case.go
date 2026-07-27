package main

import (
	"fmt"
)

var day int64
var grade string
var month int64
func main() {
//line example11_case.klx:10
day = 3
//line example11_case.klx:11
fmt.Println("Day ", day, " is: ")
//line example11_case.klx:12
switch day	 {
case 1		:
//line example11_case.klx:13
fmt.Println("Monday")
case 2		:
//line example11_case.klx:14
fmt.Println("Tuesday")
case 3		:
//line example11_case.klx:15
fmt.Println("Wednesday")
case 4		:
//line example11_case.klx:16
fmt.Println("Thursday")
case 5		:
//line example11_case.klx:17
fmt.Println("Friday")
case 6		:
//line example11_case.klx:18
fmt.Println("Saturday")
case 7		:
//line example11_case.klx:19
fmt.Println("Sunday")
	}
//line example11_case.klx:23
fmt.Println("Day type: ")
//line example11_case.klx:24
switch day	 {
case 1, 2, 3, 4, 5		:
//line example11_case.klx:25
fmt.Println("Weekday")
case 6, 7		:
//line example11_case.klx:26
fmt.Println("Weekend")
	}
//line example11_case.klx:30
month = 7
//line example11_case.klx:31
fmt.Println("Season for month ", month, ": ")
//line example11_case.klx:32
switch month	 {
case 12, 1, 2		:
//line example11_case.klx:33
fmt.Println("Winter")
case 3, 4, 5		:
//line example11_case.klx:34
fmt.Println("Spring")
case 6, 7, 8		:
//line example11_case.klx:35
fmt.Println("Summer")
case 9, 10, 11		:
//line example11_case.klx:36
fmt.Println("Fall")
	}
}
