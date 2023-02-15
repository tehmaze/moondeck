package main

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"maze.io/moondeck/moondeck"
)

var listCommand = &cli.Command{
	Name:    "list",
	Usage:   "List connected compatible devices",
	Aliases: []string{"l", "discover"},
	Action:  runList,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
		},
	},
}

func runList(ctx *cli.Context) error {
	devices, err := moondeck.Discover()
	if err != nil {
		return err
	}

	fmt.Printf("discovered %d compatible devices:\n", len(devices))
	for _, d := range devices {
		if ctx.Bool("verbose") {
			if err = d.Open(); err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println(d.Name())
		fmt.Println("\tSerial: ", d.SerialNumber())
		if ctx.Bool("verbose") {
			fmt.Println("\tVersion:", d.Version())
			fmt.Println("\tPath:   ", d.Path())
		}
		if ctx.Bool("verbose") {
			if err = d.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}

	return nil
}
