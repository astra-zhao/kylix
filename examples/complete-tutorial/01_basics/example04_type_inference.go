package main

import (
	"fmt"
)

func main() {
//line example04_type_inference.klx:6
count := 42
//line example04_type_inference.klx:7
message := "Hello"
//line example04_type_inference.klx:8
ratio := 3.141590
//line example04_type_inference.klx:9
active := true
//line example04_type_inference.klx:11
fmt.Println("Count: ", count)
//line example04_type_inference.klx:12
fmt.Println("Message: ", message)
//line example04_type_inference.klx:13
fmt.Println("Ratio: ", ratio)
//line example04_type_inference.klx:14
fmt.Println("Active: ", active)
//line example04_type_inference.klx:17
result := 10
//line example04_type_inference.klx:18
result = (result + 5)
//line example04_type_inference.klx:19
fmt.Println("Result: ", result)
//line example04_type_inference.klx:22
sum := ((10 + 20) + 30)
//line example04_type_inference.klx:23
product := (5 * 6)
//line example04_type_inference.klx:24
division := (100 / 4)
//line example04_type_inference.klx:26
fmt.Println("Sum: ", sum)
//line example04_type_inference.klx:27
fmt.Println("Product: ", product)
//line example04_type_inference.klx:28
fmt.Println("Division: ", division)
}
