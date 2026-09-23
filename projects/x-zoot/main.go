package main

import "fmt"

type operator struct {
	Name 		string
	ID			string
	Class 		string
	Rarity 		uint8
}

type TaskStatus int 

const (
	Pending TaskStatus = iota
	Running
	Completed
	Failed
)

func main() {
	fmt.Println("hello from projects/x-zoot")
}
