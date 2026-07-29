package main

import (
	"fmt"
	"kylix/stdlib"
)

var m map[string]interface{} = map[string]interface{}{}
func main() {
//line example57_variant_map.klx:19
m = func() map[string]interface{} { _v, _ := stdlib.JsonDecodeMap("{\"pi\":3.14,\"name\":\"Kylix\",\"flag\":true,\"count\":5}"); return _v }()
//line example57_variant_map.klx:25
if (m["pi"] == 3.140000)	 {
//line example57_variant_map.klx:25
fmt.Println("pi match")
}	
//line example57_variant_map.klx:26
fmt.Println(m["name"])
//line example57_variant_map.klx:27
if (m["flag"] == true)	 {
//line example57_variant_map.klx:27
fmt.Println("flag match")
}	
//line example57_variant_map.klx:28
fmt.Println(m["count"])
//line example57_variant_map.klx:29
if (m["count"] == 5.000000)	 {
//line example57_variant_map.klx:29
fmt.Println("count match")
}	
//line example57_variant_map.klx:32
fmt.Println(m["name"])
}
