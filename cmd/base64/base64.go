package main

import (
	"log"

	"github.com/cbluth/go/pkg/cmd/base64"
)

func main() {
	err := base64.ExecuteCommand()
	if err != nil {
		log.Fatalln(err)
	}
}
