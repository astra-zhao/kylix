package main

import (
	"fmt"
)

type TCounter struct {
Value int64
}

func (self *TCounter) Increment() {
//line example18_class_methods.klx:16
self.Value = (self.Value + 1)
}

func (self *TCounter) Add(n int64) {
//line example18_class_methods.klx:21
self.Value = (self.Value + n)
}

func (self *TCounter) Reset() {
//line example18_class_methods.klx:26
self.Value = 0
}

func (self *TCounter) Get() int64 {
var result int64
//line example18_class_methods.klx:31
result = self.Value
	return result
}

func main() {
//line example18_class_methods.klx:36
c := &TCounter{}
//line example18_class_methods.klx:37
c.Value = 0
//line example18_class_methods.klx:39
c.Increment()
//line example18_class_methods.klx:40
c.Increment()
//line example18_class_methods.klx:41
c.Increment()
//line example18_class_methods.klx:42
fmt.Println(("After 3 Increment: " + fmt.Sprintf("%d", c.Get())))
//line example18_class_methods.klx:44
c.Add(10)
//line example18_class_methods.klx:45
fmt.Println(("After Add(10): " + fmt.Sprintf("%d", c.Get())))
//line example18_class_methods.klx:47
c.Reset()
//line example18_class_methods.klx:48
fmt.Println(("After Reset: " + fmt.Sprintf("%d", c.Get())))
}
