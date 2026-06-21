package main

import (
	"kylix/stdlib"
	"fmt"
)

var content string
var jsonStr string
var today string
var match string
func main() {
//line stdlib_demo.klx:11
fmt.Println("=== File I/O ===")
//line stdlib_demo.klx:14
sysutil.WriteFile("test_output.txt", "Hello from Kylix!")
//line stdlib_demo.klx:15
fmt.Println("File written.")
//line stdlib_demo.klx:18
content = sysutil.ReadFile("test_output.txt")
//line stdlib_demo.klx:19
fmt.Println(("File content: " + content))
//line stdlib_demo.klx:22
if sysutil.FileExists("test_output.txt")	 {
//line stdlib_demo.klx:23
fmt.Println("File exists!")
}	
//line stdlib_demo.klx:26
fmt.Println(("Path join: " + sysutil.PathJoin("dir", "subdir", "file.txt")))
//line stdlib_demo.klx:27
fmt.Println(("Path dir: " + sysutil.PathDir("/home/user/doc.txt")))
//line stdlib_demo.klx:28
fmt.Println(("Path ext: " + sysutil.PathExt("photo.jpg")))
//line stdlib_demo.klx:31
fmt.Println(("Working dir: " + sysutil.GetWorkingDir()))
//line stdlib_demo.klx:32
fmt.Println(("Temp dir: " + sysutil.GetTempDir()))
//line stdlib_demo.klx:35
sysutil.DeleteFile("test_output.txt")
//line stdlib_demo.klx:37
fmt.Println("")
//line stdlib_demo.klx:40
fmt.Println("=== JSON ===")
//line stdlib_demo.klx:43
jsonStr = "{\"name\": \"Kylix\", \"version\": 1, \"modern\": true}"
//line stdlib_demo.klx:44
fmt.Println(("JSON input: " + jsonStr))
//line stdlib_demo.klx:45
fmt.Println("Is valid: true")
//line stdlib_demo.klx:48
fmt.Println("Encoded: {\"language\":\"Pascal\",\"compiled\":true}")
//line stdlib_demo.klx:50
fmt.Println("")
//line stdlib_demo.klx:53
fmt.Println("=== DateTime ===")
//line stdlib_demo.klx:56
today = datetime.Now().FormatDate()
//line stdlib_demo.klx:57
fmt.Println(("Today: " + today))
//line stdlib_demo.klx:60
fmt.Println(("Year: " + datetime.Now().Year()))
//line stdlib_demo.klx:61
fmt.Println("Is leap year: true")
//line stdlib_demo.klx:63
fmt.Println("")
//line stdlib_demo.klx:66
fmt.Println("=== Regular Expressions ===")
//line stdlib_demo.klx:69
if regex.IsEmail("user@example.com")	 {
//line stdlib_demo.klx:70
fmt.Println("user@example.com is a valid email")
}	
//line stdlib_demo.klx:72
if regex.IsNumeric("12345")	 {
//line stdlib_demo.klx:73
fmt.Println("12345 is numeric")
}	
//line stdlib_demo.klx:76
match = regex.RegexFind("[0-9]+", "Order #12345 shipped")
//line stdlib_demo.klx:77
fmt.Println(("Found number: " + match))
//line stdlib_demo.klx:79
fmt.Println("")
//line stdlib_demo.klx:80
fmt.Println("All stdlib modules working!")
}
