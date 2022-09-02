package main

import (
	"log"

	"github.com/cbluth/go/pkg/cmd"
)

func main() {
	err := cmd.Base64()
	if err != nil {
		log.Fatalln(err)
	}
}
