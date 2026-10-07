package main

import (
	"fmt"
	"kylix/stdlib"
)

var json string
var obj map[string]interface{} = map[string]interface{}{}
func main() {
//line example37_jsonutil.klx:15
json = "{\"name\":\"Kylix\",\"version\":3,\"active\":true}"
//line example37_jsonutil.klx:17
if stdlib.JsonIsValid(json)	 {
//line example37_jsonutil.klx:18
fmt.Println("JSON is valid")
}	
//line example37_jsonutil.klx:20
obj = func() map[string]interface{} { _v, _ := stdlib.JsonDecodeMap(json); return _v }()
//line example37_jsonutil.klx:21
fmt.Println(("name: " + stdlib.JsonGetString(obj, "name")))
//line example37_jsonutil.klx:22
fmt.Println(("version: " + fmt.Sprintf("%d", stdlib.JsonGetInt(obj, "version"))))
//line example37_jsonutil.klx:24
if stdlib.JsonGetBool(obj, "active")	 {
//line example37_jsonutil.klx:25
fmt.Println("active: true")
}	
//line example37_jsonutil.klx:27
if (!stdlib.JsonIsValid("bad json {"))	 {
//line example37_jsonutil.klx:28
fmt.Println("Invalid JSON detected correctly")
}	
}
