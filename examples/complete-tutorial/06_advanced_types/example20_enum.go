package main

import (
	"fmt"
)

const (
	North TDirection = iota
	South
	East
	West
)

type TDirection int

const (
	Red TColor = iota
	Green
	Blue
	Yellow
)

type TColor int

const (
	Pending TStatus = iota
	Active
	Inactive
	Deleted
)

type TStatus int

var dir TDirection
var color TColor
var status TStatus
func main() {
//line example20_enum.klx:19
dir = North
//line example20_enum.klx:20
switch dir	 {
case North		:
//line example20_enum.klx:21
fmt.Println("Direction: North")
case South		:
//line example20_enum.klx:22
fmt.Println("Direction: South")
case East		:
//line example20_enum.klx:23
fmt.Println("Direction: East")
case West		:
//line example20_enum.klx:24
fmt.Println("Direction: West")
	}
//line example20_enum.klx:27
color = Green
//line example20_enum.klx:28
if (color == Green)	 {
//line example20_enum.klx:29
fmt.Println("Color is green")
}	
//line example20_enum.klx:31
status = Active
//line example20_enum.klx:32
switch status	 {
case Pending		:
//line example20_enum.klx:33
fmt.Println("Status: Pending")
case Active		:
//line example20_enum.klx:34
fmt.Println("Status: Active")
case Inactive		:
//line example20_enum.klx:35
fmt.Println("Status: Inactive")
case Deleted		:
//line example20_enum.klx:36
fmt.Println("Status: Deleted")
	}
//line example20_enum.klx:39
dir = West
//line example20_enum.klx:40
if (dir != North)	 {
//line example20_enum.klx:41
fmt.Println("Not going North")
}	
}
