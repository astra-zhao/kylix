package main

import (
	"fmt"
)

func main() {
//line example25_string_ops.klx:10
s := "Hello, Kylix!"
//line example25_string_ops.klx:13
fmt.Println(("String: " + s))
//line example25_string_ops.klx:14
fmt.Println(("Length: " + fmt.Sprintf("%d", int64(len(s)))))
//line example25_string_ops.klx:17
fmt.Println(("First 5 chars: " + s[0:5]))
//line example25_string_ops.klx:18
fmt.Println(("From index 7: " + s[7:int64(len(s))]))
//line example25_string_ops.klx:21
a := "foo"
//line example25_string_ops.klx:22
b := "bar"
//line example25_string_ops.klx:23
fmt.Println((("Concat: " + a) + b))
//line example25_string_ops.klx:26
if (a == "foo")	 {
//line example25_string_ops.klx:27
fmt.Println("a equals foo")
}	
//line example25_string_ops.klx:28
if (a != b)	 {
//line example25_string_ops.klx:29
fmt.Println("a and b differ")
}	
//line example25_string_ops.klx:32
n := 2024
//line example25_string_ops.klx:33
fmt.Println(("Year: " + fmt.Sprintf("%d", n)))
//line example25_string_ops.klx:36
empty := ""
//line example25_string_ops.klx:37
if (int64(len(empty)) == 0)	 {
//line example25_string_ops.klx:38
fmt.Println("empty string")
}	
}
