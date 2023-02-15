package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "moondeck",
		Usage:  "Control Klipper using a compatible Deck",
		Action: run,
		Flags:  runFlags,
		Commands: []*cli.Command{
			listCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
