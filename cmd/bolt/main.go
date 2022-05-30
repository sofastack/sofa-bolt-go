package main

import (
	"log"

	"github.com/sofastack/sofa-bolt-go/sofabolt/cmd"
)

func main() {
	if err := cmd.RootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
