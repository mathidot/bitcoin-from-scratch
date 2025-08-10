package main

import (
	"blockchain/cli"
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	cli := cli.CLI{}
	cli.Run()
}
