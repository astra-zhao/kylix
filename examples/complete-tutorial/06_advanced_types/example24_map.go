package main

import (
	"fmt"
)

var scores map[string]int64 = map[string]int64{}
var ages map[string]int64 = map[string]int64{}
func main() {
//line example24_map.klx:9
scores["Alice"] = 95
//line example24_map.klx:10
scores["Bob"] = 87
//line example24_map.klx:11
scores["Charlie"] = 92
//line example24_map.klx:13
ages["Alice"] = 25
//line example24_map.klx:14
ages["Bob"] = 30
//line example24_map.klx:15
ages["Charlie"] = 28
//line example24_map.klx:18
fmt.Println("Scores:")
//line example24_map.klx:19
fmt.Println("Alice: ", scores["Alice"])
//line example24_map.klx:20
fmt.Println("Bob: ", scores["Bob"])
//line example24_map.klx:21
fmt.Println("Charlie: ", scores["Charlie"])
//line example24_map.klx:23
fmt.Println("Ages:")
//line example24_map.klx:24
fmt.Println("Alice: ", ages["Alice"])
//line example24_map.klx:25
fmt.Println("Bob: ", ages["Bob"])
//line example24_map.klx:26
fmt.Println("Charlie: ", ages["Charlie"])
}
