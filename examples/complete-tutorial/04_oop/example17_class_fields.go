package main

import (
	"fmt"
)

type TPerson struct {
Name string
Age int64
Email string
}

func main() {
//line example17_class_fields.klx:19
p := &TPerson{}
//line example17_class_fields.klx:20
p.Name = "Alice"
//line example17_class_fields.klx:21
p.Age = 30
//line example17_class_fields.klx:22
p.Email = "alice@example.com"
//line example17_class_fields.klx:24
fmt.Println(("Name: " + p.Name))
//line example17_class_fields.klx:25
fmt.Println(("Age: " + fmt.Sprintf("%d", p.Age)))
//line example17_class_fields.klx:26
fmt.Println(("Email: " + p.Email))
//line example17_class_fields.klx:29
p2 := &TPerson{}
//line example17_class_fields.klx:30
p2.Name = "Bob"
//line example17_class_fields.klx:31
p2.Age = 25
//line example17_class_fields.klx:32
p2.Email = "bob@example.com"
//line example17_class_fields.klx:34
fmt.Println("--- Second person ---")
//line example17_class_fields.klx:35
fmt.Println(("Name: " + p2.Name))
//line example17_class_fields.klx:36
fmt.Println(("Age: " + fmt.Sprintf("%d", p2.Age)))
}
