package main

import (
	"kylix/stdlib"
	"fmt"
)

func main() {
//line example54_http.klx:6
c := stdlib.NewHttpClient("https://httpbin.org", 5000)
//line example54_http.klx:9
c.SetHeader("X-Demo", "kylix")
//line example54_http.klx:10
c.SetHeader("Authorization", "Bearer token123")
//line example54_http.klx:32
fmt.Println(("BaseURL: " + c.BaseURL))
//line example54_http.klx:33
fmt.Println("http demo OK")
}
