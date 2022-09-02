package main

import (
	"log"

	"github.com/cbluth/go/pkg/cmd"
)

func main() {
	err := cmd.Cat()
	if err != nil {
		log.Fatalln(err)
	}
}
