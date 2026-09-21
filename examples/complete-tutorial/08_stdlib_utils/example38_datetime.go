package main

import (
	"kylix/stdlib"
	"fmt"
)

func main() {
//line example38_datetime.klx:12
now := stdlib.Now()
//line example38_datetime.klx:13
fmt.Println(("Current year: " + fmt.Sprintf("%d", now.Year())))
//line example38_datetime.klx:14
fmt.Println(("Current month: " + fmt.Sprintf("%d", now.Month())))
//line example38_datetime.klx:15
fmt.Println(("Current day: " + fmt.Sprintf("%d", now.Day())))
//line example38_datetime.klx:17
custom := stdlib.MakeDate(2024, 12, 25)
//line example38_datetime.klx:18
fmt.Println(("Christmas 2024: " + custom.FormatDate()))
//line example38_datetime.klx:20
future := custom.AddDays(7)
//line example38_datetime.klx:21
fmt.Println(("One week later: " + future.FormatDate()))
//line example38_datetime.klx:23
past := custom.AddDays((-7))
//line example38_datetime.klx:24
fmt.Println(("One week before: " + past.FormatDate()))
}
