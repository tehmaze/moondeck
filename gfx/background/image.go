package background

import (
	"embed"
	"image"

	"github.com/adrg/xdg"
	"maze.io/moondeck/gfx"
	"maze.io/moondeck/util"
)

//go:embed image/*
var content embed.FS

// Path is the default locations to look for icon files
// UNIX:    ~/.local/share/moondeck/logo
// macOS:   ~/Library/Application Support/moondeck/logo
// Windows: %LOCALAPPDATA%\moondeck\logo
var Path = util.AppendToPaths(xdg.DataDirs, "moondeck", "logo")

// Logo image by name.
func Logo(name string) (*image.RGBA, error) {
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
