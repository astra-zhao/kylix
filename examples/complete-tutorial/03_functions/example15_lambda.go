package main

import (
	"fmt"
)

func main() {
//line example15_lambda.klx:14
greet := func(name string) 	{
//line example15_lambda.klx:16
fmt.Println((("Hello, " + name) + "!"))
}
//line example15_lambda.klx:19
greet("Alice")
//line example15_lambda.klx:20
greet("Bob")
//line example15_lambda.klx:21
greet("Kylix")
//line example15_lambda.klx:23
printLine := func(label string, value int64) 	{
//line example15_lambda.klx:25
fmt.Println(((label + ": ") + fmt.Sprintf("%d", value)))
}
//line example15_lambda.klx:28
printLine("Answer", 42)
//line example15_lambda.klx:29
printLine("Year", 2024)
//line example15_lambda.klx:30
printLine("Items", 7)
}
