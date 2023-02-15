package main

import (
	"flag"
	"fmt"
	"log"

	"maze.io/moondeck/moonraker"
)

func main() {
	configFile := flag.String("config", "moondeck.hcl", "Configuration file")
	flag.Parse()

	c, err := moonraker.Load(*configFile)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("config:\n%#+v\n", c)
}
