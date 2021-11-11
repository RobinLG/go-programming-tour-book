package main

import (
	"github.com/go-programming-tour-book/tour/1_2_tools/cmd"
	"log"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("cmd.Execute err: %v", err)
	}
}
