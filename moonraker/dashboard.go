package moonraker

import (
	"image"
	"image/color"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/image/draw"

	"maze.io/moondeck/gfx/blend"
	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/moondeck/icon"
)

const (
	appNameConnecting = "moonraker-connecting"
	appNameLoading    = "moonraker-loading"
)

type Dashboard struct {
	*moondeck.Dashboard
	config              *Config
	connecting, loading *moondeck.App
}

func NewDashboard(config *Config, dash *moondeck.Dashboard) (*Dashboard, error) {
	d := &Dashboard{
		Dashboard:  dash,
		config:     config,
		connecting: moondeck.NewApp(appNameConnecting, dash),
		loading:    moondeck.NewApp(appNameLoading, dash),
	}

	dash.Add(d.connecting)
	dash.Add(d.loading)

	size := dash.Size()
	var (
		connectingWidget = pulsingIconWidget("network-wired")
		loadingWidget    = pulsingIconWidget("gears")
	)
	d.connecting.AddWidget(connectingWidget, moondeck.Background, size.W/2, size.H/2)
	d.loading.AddWidget(loadingWidget, moondeck.Background, size.W/2, size.H/2)

	IsOffline(func(c *Client, _ error) {
		d.Dashboard.Start(appNameConnecting)
	})
	IsOnline(func(c *Client, _ error) {
		if !d.Dashboard.Back() {
			d.Dashboard.Start("home")
		}
	})

	for _, appConfig := range config.App {
		log.WithField("app", appConfig.Name).Debug("loading app")
		app, err := appConfig.App(dash, config)
		if err != nil {
			return nil, err
		}
		dash.Add(app)
	}

	return d, nil
}

func (d *Dashboard) Run() error {
	_ = d.Dashboard.Start(appNameConnecting)

	_, err := New(d.config.Klipper.API)
	if err != nil {
		return err
	}

	return d.connecting.Dashboard.Run()
}

func pulsingIconWidget(name string) *moondeck.ImageWidget {
	w := moondeck.NewIconWidget(name)

	go func(w *moondeck.ImageWidget) {
		icons := make([]*image.RGBA, 0, 10)
		for p := float32(0.2); p <= 1.0; p += 0.2 {
			c := toRGBAColor(blend.Gradient(color.Black, color.White, p))
			icons = append(icons, icon.Colorize(w.Image, c, icon.Transparent))
		}
		for p := float32(1.0); p >= 0.2; p -= 0.2 {
			c := toRGBAColor(blend.Gradient(color.Black, color.White, p))
			icons = append(icons, icon.Colorize(w.Image, c, icon.Transparent))
		}

		j := -1
		t := time.NewTicker(time.Second / 5)
		for {
			<-t.C

			j++
			if j >= len(icons) {
				j = 0
			}

			draw.Copy(w.Image, image.Point{}, icons[j], icons[j].Bounds(), draw.Src, nil)
			w.Dirty()
		}
	}(w)

	return w
}

func toRGBAColor(c color.Color) color.RGBA {
	if c, ok := c.(color.RGBA); ok {
		return c
	}

	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}
