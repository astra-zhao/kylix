package main

import (
	"fmt"
)

var i int64
var guess int64
func main() {
//line example10_repeat.klx:9
i = 1
//line example10_repeat.klx:10
	for {
//line example10_repeat.klx:11
fmt.Println("Iteration: ", i)
//line example10_repeat.klx:12
i = (i + 1)
if (i > 5)		 {
			break
		}
	}
//line example10_repeat.klx:17
guess = 100
//line example10_repeat.klx:18
fmt.Println("This will print once even though condition is true:")
//line example10_repeat.klx:19
	for {
//line example10_repeat.klx:20
fmt.Println("Guess value: ", guess)
//line example10_repeat.klx:21
guess = (guess + 1)
if (guess > 50)		 {
			break
		}
	}
//line example10_repeat.klx:25
i = 10
//line example10_repeat.klx:26
fmt.Println("Countdown with repeat:")
//line example10_repeat.klx:27
	for {
//line example10_repeat.klx:28
fmt.Println(i)
//line example10_repeat.klx:29
i = (i - 1)
if (i <= 0)		 {
			break
		}
	}
//line example10_repeat.klx:31
fmt.Println("Done!")
}
