package main

import (
	"fmt"
)

var square = func(x int64) {
return (x * x)
}
var add = func(a int64, b int64) {
return (a + b)
}
var numbers = []interface{}{1, 2, 3, 4, 5}
var name = "World"
var greeting = "Hello, ${name}!"
//line modern.klx:8
func Apply(fn function, value int64) int64 {
var result int64
//line modern.klx:10
result = fn(value)
return result
}

//line modern.klx:14
func FetchData(url string) <-chan string {
ch := make(chan string	, 1)
	go func() {
var result string
//line modern.klx:17
result = ("Data from " + url)
		ch <- result
	}()
	return ch
}

//line modern.klx:21
func Describe(value int64) string {
var result string
//line modern.klx:23
switch _v := value	 {
case _v == 0		:
//line modern.klx:24
"zero"
case _v == 1		:
//line modern.klx:25
"one"
case _v == 2 || _v == 3		:
//line modern.klx:26
"small prime or two"
case (value < 0)		:
//line modern.klx:27
"negative"
case (value > 100)		:
//line modern.klx:28
"large"
		default:
//line modern.klx:29
"other"
	}
return result
}

func main() {
//line modern.klx:35
for _, num := range numbers	 {
//line modern.klx:37
fmt.Println("Number: ", num)
	}
//line modern.klx:45
fmt.Println("Square of 5: ", square(5))
//line modern.klx:46
fmt.Println("3 + 4 = ", add(3, 4))
//line modern.klx:47
fmt.Println("Applied: ", Apply(square, 7))
//line modern.klx:49
data := <-FetchData("http://example.com")
//line modern.klx:50
fmt.Println(data)
//line modern.klx:52
fmt.Println("Describe 5: ", Describe(5))
//line modern.klx:53
fmt.Println(greeting)
}
