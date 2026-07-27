package main

import (
	"fmt"
)

type TPoint struct {
X float64
Y float64
}

type TPerson struct {
Name string
Age int64
Email string
}

var point TPoint
var person TPerson
func main() {
//line example22_records.klx:22
point.X = 10.500000
//line example22_records.klx:23
point.Y = 20.300000
//line example22_records.klx:25
fmt.Println("Point: (", point.X, ", ", point.Y, ")")
//line example22_records.klx:28
person.Name = "Alice"
//line example22_records.klx:29
person.Age = 25
//line example22_records.klx:30
person.Email = "alice@example.com"
//line example22_records.klx:32
fmt.Println("Person: ", person.Name, ", Age: ", person.Age)
//line example22_records.klx:33
fmt.Println("Email: ", person.Email)
}
