package main

import (
	"fmt"
)

var quotient, remainder int64
var minVal, maxVal int64
var q2, r2 int64
//line example16_multireturn.klx:4
func DivMod(dividend int64, divisor int64) (int64, int64) {
//line example16_multireturn.klx:6
return (dividend / divisor), (dividend % divisor)
}

//line example16_multireturn.klx:10
func MinMax(a int64, b int64, c int64) (int64, int64) {
var min int64
var max int64
//line example16_multireturn.klx:14
min = a
//line example16_multireturn.klx:15
max = a
//line example16_multireturn.klx:17
if (b < min)	 {
//line example16_multireturn.klx:17
min = b
}	
//line example16_multireturn.klx:18
if (c < min)	 {
//line example16_multireturn.klx:18
min = c
}	
//line example16_multireturn.klx:20
if (b > max)	 {
//line example16_multireturn.klx:20
max = b
}	
//line example16_multireturn.klx:21
if (c > max)	 {
//line example16_multireturn.klx:21
max = c
}	
//line example16_multireturn.klx:23
return min, max
}

func main() {
//line example16_multireturn.klx:33
quotient, remainder := DivMod(17, 5)
//line example16_multireturn.klx:34
fmt.Println("17 div 5 = ", quotient)
//line example16_multireturn.klx:35
fmt.Println("17 mod 5 = ", remainder)
//line example16_multireturn.klx:37
minVal, maxVal := MinMax(5, 12, 3)
//line example16_multireturn.klx:38
fmt.Println("Min of 5,12,3: ", minVal)
//line example16_multireturn.klx:39
fmt.Println("Max of 5,12,3: ", maxVal)
//line example16_multireturn.klx:42
q2, r2 := DivMod(100, 7)
//line example16_multireturn.klx:43
fmt.Println("100 div 7 = ", q2, ", remainder = ", r2)
}
