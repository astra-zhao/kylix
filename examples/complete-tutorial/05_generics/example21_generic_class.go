package main

import (
	"fmt"
)

type TStack[T interface{}] struct {
Items [((99 - 0) + 1)]T
Count int64
}

func (self *TStack[T]) Create() {
//line example21_generic_class.klx:12
self.Count = 0
}

func (self *TStack[T]) Push(item T) {
//line example21_generic_class.klx:17
if (self.Count < 100)	 {
//line example21_generic_class.klx:19
self.Items[self.Count] = item
//line example21_generic_class.klx:20
self.Count = (self.Count + 1)
}	
}

func (self *TStack[T]) Pop() T {
var result T
//line example21_generic_class.klx:26
if (self.Count > 0)	 {
//line example21_generic_class.klx:28
self.Count = (self.Count - 1)
//line example21_generic_class.klx:29
result = self.Items[self.Count]
}	
	return result
}

func (self *TStack[T]) GetCount() int64 {
var result int64
//line example21_generic_class.klx:35
result = self.Count
	return result
}

func main() {
//line example21_generic_class.klx:40
intStack := &TStack[int64]{}
//line example21_generic_class.klx:41
intStack.Push(10)
//line example21_generic_class.klx:42
intStack.Push(20)
//line example21_generic_class.klx:43
intStack.Push(30)
//line example21_generic_class.klx:45
fmt.Println("Stack count: ", intStack.GetCount())
//line example21_generic_class.klx:46
fmt.Println("Pop: ", intStack.Pop())
//line example21_generic_class.klx:47
fmt.Println("Pop: ", intStack.Pop())
//line example21_generic_class.klx:48
fmt.Println("Stack count: ", intStack.GetCount())
//line example21_generic_class.klx:50
strStack := &TStack[string]{}
//line example21_generic_class.klx:51
strStack.Push("Hello")
//line example21_generic_class.klx:52
strStack.Push("World")
//line example21_generic_class.klx:53
fmt.Println("Pop: ", strStack.Pop())
}
