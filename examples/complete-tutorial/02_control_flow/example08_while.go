package main

import (
	"fmt"
)

var i int64
var sum int64
var countdown int64
func main() {
//line example08_while.klx:10
i = 1
//line example08_while.klx:11
sum = 0
//line example08_while.klx:12
for (i <= 10)	 {
//line example08_while.klx:14
sum = (sum + i)
//line example08_while.klx:15
i = (i + 1)
	}
//line example08_while.klx:17
fmt.Println("Sum of 1-10: ", sum)
//line example08_while.klx:20
countdown = 5
//line example08_while.klx:21
fmt.Println("Countdown:")
//line example08_while.klx:22
for (countdown > 0)	 {
//line example08_while.klx:24
fmt.Println(countdown)
//line example08_while.klx:25
countdown = (countdown - 1)
	}
//line example08_while.klx:27
fmt.Println("Blast off!")
//line example08_while.klx:30
i = 100
//line example08_while.klx:31
for (i < 10)	 {
//line example08_while.klx:33
fmt.Println("This will not print")
//line example08_while.klx:34
i = (i + 1)
	}
//line example08_while.klx:36
fmt.Println("Loop skipped")
}
