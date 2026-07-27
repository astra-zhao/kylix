package main

import (
	"fmt"
)

type TShape struct {
Color string
}

func (self *TShape) Describe() {
//line example19_inheritance.klx:17
fmt.Println(("Shape, color: " + self.Color))
}

type TRectangle struct {
	TShape
Width int64
Height int64
}

func (self *TRectangle) Area() int64 {
var result int64
//line example19_inheritance.klx:28
result = (self.Width * self.Height)
	return result
}

func (self *TRectangle) Perimeter() int64 {
var result int64
//line example19_inheritance.klx:33
result = (2 * (self.Width + self.Height))
	return result
}

type TSquare struct {
	TRectangle
}

func (self *TSquare) SetSide(n int64) {
//line example19_inheritance.klx:41
self.Width = n
//line example19_inheritance.klx:42
self.Height = n
}

func main() {
//line example19_inheritance.klx:47
rect := &TRectangle{}
//line example19_inheritance.klx:48
rect.Color = "blue"
//line example19_inheritance.klx:49
rect.Width = 8
//line example19_inheritance.klx:50
rect.Height = 5
//line example19_inheritance.klx:51
rect.Describe()
//line example19_inheritance.klx:52
fmt.Println(("Area: " + fmt.Sprintf("%d", rect.Area())))
//line example19_inheritance.klx:53
fmt.Println(("Perimeter: " + fmt.Sprintf("%d", rect.Perimeter())))
//line example19_inheritance.klx:55
fmt.Println("---")
//line example19_inheritance.klx:57
sq := &TSquare{}
//line example19_inheritance.klx:58
sq.Color = "red"
//line example19_inheritance.klx:59
sq.SetSide(6)
//line example19_inheritance.klx:60
sq.Describe()
//line example19_inheritance.klx:61
fmt.Println(("Area: " + fmt.Sprintf("%d", sq.Area())))
}
