package main

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/moonraker"
)

var runFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "config",
		Usage:       "Path to configuration file",
		DefaultText: filepath.Join(xdg.ConfigHome, "moondeck", "moondeck.hcl"),
	},
	&cli.StringFlag{
		Name:  "device-name",
		Usage: "Use device with provided name",
	},
	&cli.StringFlag{
		Name:  "device-path",
		Usage: "Use device with provided path",
	},
	&cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug output",
	},
	&cli.BoolFlag{
		Name:  "trace",
		Usage: "Enable trace output",
	},
}

func run(ctx *cli.Context) error {
	if ctx.Bool("trace") {
		log.SetLevel(log.TraceLevel)
	} else if ctx.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	log.
		WithFields(log.Fields{
			"args": os.Args,
			"flag": ctx.FlagNames(),
		}).
		Trace("starting")

	config, err := moonraker.Load(ctx.String("config"))
	if err != nil {
		return err
	}

	deck, err := openDeck(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = deck.Close() }()

	if err = deck.Reset(); err != nil {
		return err
	}

	/*
		dashboard := moondeck.NewDashboard(deck)
		for _, appConfig := range config.App {
			log.Println("moondeck: loading app", appConfig.Name)
			app, err := appConfig.App(dashboard, config)
			if err != nil {
				return err
			}
			dashboard.Add(app)
		}

		if !dashboard.Start("home") {
			return errors.New("No \"home\" app configured")
		}
	*/

	dashboard, err := moonraker.NewDashboard(config, moondeck.NewDashboard(deck))
	if err != nil {
		return err
	}
	return dashboard.Run()
}

func openDeck(ctx *cli.Context) (moondeck.Deck, error) {
	var (
		path   = ctx.String("device-path")
		name   = ctx.String("device-name")
		filter func(moondeck.Deck) bool
	)
	switch {
	case path != "":
		filter = func(deck moondeck.Deck) bool {
			return deck.Path() == path
		}
	case name != "":
		filter = func(deck moondeck.Deck) bool {
			return deck.Name() == name
		}
	default:
		return moondeck.Open()
	}

	devices, err := moondeck.Discover()
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		if filter(device) {
			if err = device.Open(); err != nil {
				return nil, err
			}
			return device, nil
		}
	}
	return nil, moondeck.ErrNoDevices
}
