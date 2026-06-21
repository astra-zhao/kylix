package main

import (
	"fmt"
)

type Animal struct {
Name string
Age int64
}

func (self *Animal) Create(name string, age int64) {
//line classes.klx:12
Name = name
//line classes.klx:13
Age = age
}

func (self *Animal) Speak() {
//line classes.klx:18
fmt.Println(Name, " makes a sound")
}

func (self *Animal) GetName() string {
return self.Name
}

func (self *Animal) GetAge() int64 {
return self.Age
}

type Dog struct {
	Animal
Breed string
}

func (self *Dog) Create(name string, age int64, breed string) {
//line classes.klx:33
self.Create(name, age)
//line classes.klx:34
Breed = breed
}

func (self *Dog) Speak() {
//line classes.klx:39
fmt.Println(Name, " barks!")
}

func (self *Dog) Fetch(item string) {
//line classes.klx:44
fmt.Println(Name, " fetches the ", item)
}

type IMovable interface {
Move(distance float64)
}

type Vehicle struct {
Speed float64
Position float64
}

func (self *Vehicle) Create(speed float64) {
//line classes.klx:62
Speed = speed
//line classes.klx:63
Position = 0
}

func (self *Vehicle) Move(distance float64) {
//line classes.klx:68
Position = (Position + distance)
//line classes.klx:69
fmt.Println("Vehicle moved to position: ", Position)
}

func (self *Vehicle) GetSpeed() float64 {
return self.Speed
}

func main() {
//line classes.klx:76
dog := &Dog{Breed: "Rex", 3, "German Shepherd"}
//line classes.klx:77
dog.Speak()
//line classes.klx:78
dog.Fetch("ball")
//line classes.klx:80
car := &Vehicle{Speed: 60.000000}
//line classes.klx:81
car.Move(100)
}
