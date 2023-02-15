package moonraker

import (
	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/util"
)

type GCodeConfig struct {
	Name  string     `hcl:"name,label"`
	Icon  string     `hcl:"icon"`
	At    util.Point `hcl:"at"`
	GCode string     `hcl:"gcode"`
}

func (c *GCodeConfig) Widgets(_ *moondeck.App) ([]WidgetConfig, error) {
	w, err := moondeck.NewIconWidget(c.Icon)
	if err != nil {
		return nil, err
	}

	o := make(chan string, 8)
	go func(o <-chan string) {
		IsOnline(func(c *Client, _ error) {
			for gcode := range o {
				c.SendGCode(gcode)
			}
		})
	}(o)

	w.OnPress = func(b moondeck.Button, a *moondeck.App) {
		o <- c.GCode
	}

	return []WidgetConfig{{
		Widget: w,
		Layer:  moondeck.Foreground,
		Point:  c.At,
	}}, nil
}
