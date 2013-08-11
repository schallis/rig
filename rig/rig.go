package main

import (
	"flag"
	"log"
	"os"
)

var (
	defaultProto string = "http"
	defaultAddr  string = "0.0.0.0:9696"
)

func main() {
	flag.Parse()
	cli := NewCli(defaultProto, defaultAddr)

	if err := cli.ParseCommand(flag.Args()...); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
