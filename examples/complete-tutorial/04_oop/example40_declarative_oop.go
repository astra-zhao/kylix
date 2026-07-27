package main

import (
	"fmt"
)

type TAnimal struct {
Name string
Sound string
}

func (self *TAnimal) Speak() {
//line example40_declarative_oop.klx:20
fmt.Println(((self.Name + " says ") + self.Sound))
}

type TDog struct {
	TAnimal
Breed string
}

func (self *TDog) Describe() {
//line example40_declarative_oop.klx:30
fmt.Println((((("Dog: " + self.Name) + " (") + self.Breed) + ")"))
}

var cat *TAnimal
var dog *TDog
func main() {
//line example40_declarative_oop.klx:42
cat = &TAnimal{}
//line example40_declarative_oop.klx:43
cat.Name = "Whiskers"
//line example40_declarative_oop.klx:44
cat.Sound = "Meow"
//line example40_declarative_oop.klx:45
cat.Speak()
//line example40_declarative_oop.klx:47
dog = &TDog{}
//line example40_declarative_oop.klx:48
dog.Name = "Rex"
//line example40_declarative_oop.klx:49
dog.Sound = "Woof"
//line example40_declarative_oop.klx:50
dog.Breed = "Labrador"
//line example40_declarative_oop.klx:51
dog.Speak()
//line example40_declarative_oop.klx:52
dog.Describe()
}
