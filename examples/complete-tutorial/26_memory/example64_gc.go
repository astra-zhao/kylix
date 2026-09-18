package main

import (
	"fmt"
)

type TAccum struct {
Total int64
}

func (self *TAccum) Add(v int64) {
//line example64_gc.klx:22
self.Total = (self.Total + v)
}

var i int64
var j int64
var t string
var acc *TAccum
var grand int64
func main() {
//line example64_gc.klx:33
grand = 0
//line example64_gc.klx:34
for i = 1; i <= 2500; i++ {
//line example64_gc.klx:36
for j = 1; j <= 40; j++ {
//line example64_gc.klx:38
t = (((("item-" + fmt.Sprintf("%d", i)) + "-") + fmt.Sprintf("%d", j)) + "-payloadpadding-x")
//line example64_gc.klx:39
t = (t + t)
//line example64_gc.klx:40
t = (t + t)
//line example64_gc.klx:41
t = (t + t)
//line example64_gc.klx:42
grand = (grand + int64(len(t)))
		}
//line example64_gc.klx:44
acc = &TAccum{}
//line example64_gc.klx:45
acc.Add(grand)
//line example64_gc.klx:46
if (i == 625)		 {
//line example64_gc.klx:48
fmt.Println("checkpoint ", i, " grand=", grand)
}		
//line example64_gc.klx:50
if (i == 1250)		 {
//line example64_gc.klx:52
fmt.Println("checkpoint ", i, " grand=", grand)
}		
//line example64_gc.klx:54
if (i == 1875)		 {
//line example64_gc.klx:56
fmt.Println("checkpoint ", i, " grand=", grand)
}		
	}
//line example64_gc.klx:59
fmt.Println("done: grand=", grand, " sample=", int64(len(t)), " total=", acc.Total)
}
