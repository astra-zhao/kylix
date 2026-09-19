package main

import (
	"kylix/stdlib"
	"fmt"
)

func main() {
//line example39_regex.klx:10
if stdlib.IsEmail("user@example.com")	 {
//line example39_regex.klx:11
fmt.Println("Valid email")
}	
//line example39_regex.klx:13
if (!stdlib.IsEmail("bad-email"))	 {
//line example39_regex.klx:14
fmt.Println("Invalid email detected")
}	
//line example39_regex.klx:16
if stdlib.IsURL("https://kylix.top")	 {
//line example39_regex.klx:17
fmt.Println("Valid URL")
}	
//line example39_regex.klx:19
if stdlib.IsNumeric("12345")	 {
//line example39_regex.klx:20
fmt.Println("Valid numeric string")
}	
//line example39_regex.klx:22
if (!stdlib.IsNumeric("abc123"))	 {
//line example39_regex.klx:23
fmt.Println("Non-numeric string detected")
}	
//line example39_regex.klx:25
if stdlib.IsAlpha("hello")	 {
//line example39_regex.klx:26
fmt.Println("Valid alphabetic string")
}	
}
