package moonraker

import (
	"image/color"

	"github.com/mazznoer/csscolorparser"

	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/moondeck/icon"
	"maze.io/moondeck/util"
)

type ColorConfig string

func (c ColorConfig) RGBA() color.RGBA {
	p, err := csscolorparser.Parse(string(c))
	if err != nil {
		panic(err)
	}
	r, g, b, a := p.RGBA255()
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func (c ColorConfig) RGBADefault(other color.RGBA) color.RGBA {
	if c == "" {
		return other
	}
	return c.RGBA()
}

type IconConfig struct {
	Name       string       `hcl:"name,label"`
	Color      *ColorConfig `hcl:"color"`
	Background *ColorConfig `hcl:"color"`
	At         util.Point   `hcl:"at"`
}

func (c *IconConfig) Widgets(_ *moondeck.App) ([]WidgetConfig, error) {
	var fg, bg color.RGBA
	if c.Color == nil {
		fg = icon.White
	} else {
		fg = c.Color.RGBA()
	}
	if c.Background == nil {
		bg = icon.Transparent
	} else {
		bg = c.Background.RGBA()
	}
	return []WidgetConfig{
		{
			Widget: moondeck.NewIconColorWidget(c.Name, fg, bg),
			Layer:  moondeck.Foreground,
			Point:  c.At,
		},
	}, nil
}
