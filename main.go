package main

import (
	"log"

	"github.com/crossjoin-io/crossjoin/cmd"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cmd.Execute()
}
