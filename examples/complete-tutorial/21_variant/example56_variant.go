package main

import (
	"kylix/stdlib"
	"fmt"
)

var arr []interface{}
var v interface{}
var i int64
func main() {
//line example56_variant.klx:22
v = 10.0
//line example56_variant.klx:23
if (v == 10.0)	 {
//line example56_variant.klx:23
fmt.Println("scalar match")
}	
//line example56_variant.klx:24
fmt.Println(v)
//line example56_variant.klx:27
arr = stdlib.JsonGetArray(func() map[string]interface{} { _v, _ := stdlib.JsonDecodeMap("{\"nums\":[10,20,30]}"); return _v }(), "nums")
//line example56_variant.klx:28
if (arr[0] == 10.0)	 {
//line example56_variant.klx:28
fmt.Println("arr[0] is 10")
}	
//line example56_variant.klx:29
fmt.Println(arr[1])
//line example56_variant.klx:30
fmt.Println(int64(len(arr)))
//line example56_variant.klx:33
arr = stdlib.JsonGetArray(func() map[string]interface{} { _v, _ := stdlib.JsonDecodeMap("{\"fruits\":[\"apple\",\"banana\",\"cherry\"]}"); return _v }(), "fruits")
//line example56_variant.klx:34
fmt.Println(arr[1])
//line example56_variant.klx:35
if (arr[2] == "cherry")	 {
//line example56_variant.klx:35
fmt.Println("cherry match")
}	
//line example56_variant.klx:38
arr = stdlib.JsonGetArray(func() map[string]interface{} { _v, _ := stdlib.JsonDecodeMap("{\"nums\":[1,2,3]}"); return _v }(), "nums")
//line example56_variant.klx:39
for i = 0; i <= (int64(len(arr)) - 1); i++ {
//line example56_variant.klx:40
fmt.Println(arr[i])
	}
}
