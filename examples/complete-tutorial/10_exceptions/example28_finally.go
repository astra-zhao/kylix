package main

import (
	"fmt"
)

// Kylix runtime exception base type
type Exception struct {
	Message string
}

func (e *Exception) Error() string { return e.Message }

//line example28_finally.klx:3
func ProcessFile(filename string) {
//line example28_finally.klx:5
fmt.Println("Opening file: ", filename)
//line example28_finally.klx:7
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic(r)
			}
		}()
//line example28_finally.klx:9
fmt.Println("Processing data...")
//line example28_finally.klx:11
fmt.Println("Data processed successfully")
	}()
	// finally block
//line example28_finally.klx:15
fmt.Println("Closing file: ", filename)
}

func main() {
//line example28_finally.klx:21
ProcessFile("data.txt")
//line example28_finally.klx:23
fmt.Println("---")
//line example28_finally.klx:26
	func() {
		defer func() {
			if r := recover(); r != nil {
//line example28_finally.klx:34
fmt.Println("Error occurred")
			}
		}()
//line example28_finally.klx:28
fmt.Println("Starting operation...")
//line example28_finally.klx:29
x := (10 / 2)
//line example28_finally.klx:30
fmt.Println("Result: ", x)
	}()
	// finally block
//line example28_finally.klx:38
fmt.Println("Cleanup complete")
}
