package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ibigbug/GArrow/arrow"
)

const (
	// VERSION App version
	VERSION = "0.1.0"
)

func main() {
	var mode = flag.String("m", "client", "Run mode, can be client|server, default to beclient")
	var config = flag.String("c", "g-arrow.yaml", "Config path, default to be ./g-arrow.yaml")

	flag.Parse()

	c := arrow.NewConfig(*config)

	var s arrow.Runnable
	if *mode == "client" {
		s = arrow.NewClient(c)
	} else if *mode == "server" {
		s = arrow.NewServer(c)
	} else {
		fmt.Println("Unknow run mode")
		os.Exit(1)
	}
	log.Fatal(s.Run())
}
