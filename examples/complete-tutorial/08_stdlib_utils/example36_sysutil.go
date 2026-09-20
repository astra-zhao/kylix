package main

import (
	"kylix/stdlib"
	"fmt"
)

var fname string
var content string
func main() {
//line example36_sysutil.klx:14
fname = "/tmp/kylix_example_sysutil.txt"
//line example36_sysutil.klx:16
func() { stdlib.WriteFile(fname, "Hello from Kylix sysutil!") }()
//line example36_sysutil.klx:17
fmt.Println(("File written to: " + fname))
//line example36_sysutil.klx:19
if stdlib.FileExists(fname)	 {
//line example36_sysutil.klx:20
fmt.Println("File exists: confirmed")
}	
//line example36_sysutil.klx:22
content = func() string { _v, _ := stdlib.ReadFile(fname); return _v }()
//line example36_sysutil.klx:23
fmt.Println(("Read back: " + content))
//line example36_sysutil.klx:25
fmt.Println(("PathJoin example: " + stdlib.PathJoin("/usr", "local", "bin")))
//line example36_sysutil.klx:26
fmt.Println(("PathBase example: " + stdlib.PathBase("/path/to/file.txt")))
}
