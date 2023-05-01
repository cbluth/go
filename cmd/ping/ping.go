package main

import (
	"log"

	"github.com/cbluth/go/pkg/cmd"
)

func main() {
	err := cmd.Ping()
	if err != nil {
		log.Fatalln(err)
	}
}
