package icon

import (
	"embed"
	"image"
	"image/color"

	"github.com/adrg/xdg"
	"maze.io/moondeck/gfx"
	"maze.io/moondeck/util"
)

//go:embed image/*
var content embed.FS

// Path is the default locations to look for icon files
// UNIX:    ~/.local/share/moondeck/icon
// macOS:   ~/Library/Application Support/moondeck/icon
// Windows: %LOCALAPPDATA%\moondeck\icon
var Path = util.AppendToPaths(xdg.DataDirs, "moondeck", "icon")

// Common colors
var (
	Transparent = color.RGBA{}
	Black       = color.RGBA{A: 0xff}
	Blue        = color.RGBA{B: 0xff, A: 0xff}
	Green       = color.RGBA{G: 0xff, A: 0xff}
	Cyan        = color.RGBA{G: 0xff, B: 0xff, A: 0xff}
	Red         = color.RGBA{R: 0xff, A: 0xff}
	Magenta     = color.RGBA{R: 0xff, B: 0xff, A: 0xff}
	Yellow      = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
	White       = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
)

func Brand(name string) (*image.RGBA, error) {
	return Icon(name + ".brand")
}

func Icon(name string) (*image.RGBA, error) {
	here := []interface{}{
		util.Jail{Opener: content, Prefix: []string{"image"}},
	}
	for _, path := range Path {
		here = append(here, path)
	}

	i, err := gfx.LoadImage(name, here...)
	if err != nil {
		return nil, err
	}

	return gfx.ToRGBA(i), nil
}

func Must(name string) *image.RGBA {
	i, err := Icon(name)
	if err != nil {
		panic(err)
	}
	return i
}
