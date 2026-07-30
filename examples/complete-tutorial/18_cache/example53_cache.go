package main

import (
	"fmt"
	"kylix/stdlib"
)

func main() {
//line example53_cache.klx:6
c := stdlib.NewCache(4, 0)
//line example53_cache.klx:9
c.Put("name", "alice")
//line example53_cache.klx:10
c.Put("role", "admin")
//line example53_cache.klx:13
fmt.Println(("name: " + c.GetString("name")))
//line example53_cache.klx:14
fmt.Println(("role: " + c.GetString("role")))
//line example53_cache.klx:17
if c.Has("name")	 {
//line example53_cache.klx:18
fmt.Println("has name: yes")
}	
//line example53_cache.klx:19
fmt.Println(("missing: " + c.GetString("missing")))
//line example53_cache.klx:22
fmt.Println(c.Size())
//line example53_cache.klx:25
c.Delete("role")
//line example53_cache.klx:26
if (!c.Has("role"))	 {
//line example53_cache.klx:27
fmt.Println("role deleted")
}	
//line example53_cache.klx:30
c.Clear()
//line example53_cache.klx:31
fmt.Println(c.Size())
//line example53_cache.klx:33
fmt.Println("cache demo OK")
}
