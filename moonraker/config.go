package moonraker

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hclsimple"

	"maze.io/moondeck/gfx/background"
	"maze.io/moondeck/gfx/icon"
	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/util"

	_ "image/gif"  // GIF codec
	_ "image/jpeg" // JPEG codec
	_ "image/png"  // PNG codec

	_ "golang.org/x/image/bmp" // BMP codec
)

const (
	DefaultKlipperAPI = "127.0.0.1:7125"
)

type Config struct {
	Klipper    *KlipperConfig  `hcl:"klipper,block"`
	Loading    string          `hcl:"loading,optional"`
	Connecting string          `hcl:"connecting,optional"`
	Menu       []*MenuConfig   `hcl:"menu,block"`
	App        []*AppConfig    `hcl:"app,block"`
	Preset     []*PresetConfig `hcl:"preset,optional"`
}

type KlipperConfig struct {
	API string `hcl:"api"`
}

type PresetConfig struct {
	Name     string  `hcl:"name" cty:"name"`
	Extruder float64 `hcl:"extruder" cty:"extruder"`
	Bed      float64 `hcl:"bed" cty:"bed"`
}

type MenuConfig struct {
	Name            string           `hcl:"name,label"`
	Icon            string           `hcl:"icon"`
	Item            []MenuItemConfig `hcl:"item"`
	ForegroundColor ColorConfig      `hcl:"fg,optional"`
	BackgroundColor ColorConfig      `hcl:"bg,optional"`
}

type MenuItemConfig struct {
	App  string `hcl:"app" cty:"app"`
	Icon string `hcl:"icon" cty:"icon"`
}

type AppConfig struct {
	Name       string           `hcl:"name,label"`
	Background string           `hcl:"background,optional"`
	Menu       *AppMenuConfig   `hcl:"menu,block"`
	Icon       []*IconConfig    `hcl:"icon,block"`
	Temp       []*TempConfig    `hcl:"temp,block"`
	Camera     []*CameraConfig  `hcl:"camera,block"`
	GCode      []*GCodeConfig   `hcl:"gcode,block"`
	Emergency  *EmergencyConfig `hcl:"emergency,block"`
	Move       *MoveConfig      `hcl:"move,block"`
}

func (c *AppConfig) App(dash *moondeck.Dashboard, config *Config) (*moondeck.App, error) {
	app := moondeck.NewApp(c.Name, dash)

	if c.Background != "" {
		/*
			f, err := os.Open(c.Background)
			if err != nil {
				return nil, fmt.Errorf("moonraker: could not open background image %q for icon %q", c.Background, c.Name)
			}
			defer func() { _ = f.Close() }()

			i, _, err := image.Decode(f)
			if err != nil {
				return nil, fmt.Errorf("moonraker: could not load background image %q for icon %q", c.Background, c.Name)
			}
		*/
		i, err := background.Logo(c.Background)
		if err != nil {
			return nil, fmt.Errorf("could not load background image %q for app %q", c.Background, c.Name)
		}

		app.SetBackgroundImage(i)
	}

	var wcs []widgetConfigurator
	for _, ci := range c.Icon {
		wcs = append(wcs, ci)
	}
	for _, ct := range c.Temp {
		log.Println("moonraker: config temp", ct.Heater, "at", ct.At)
		wcs = append(wcs, ct)
	}
	for _, cc := range c.Camera {
		wcs = append(wcs, cc)
	}
	for _, cg := range c.GCode {
		wcs = append(wcs, cg)
	}
	if c.Menu != nil {
		wcs = append(wcs, widgetConfiguratorFunc(c.Menu.Widgets(config.Menu)))
	}
	if c.Move != nil {
		wcs = append(wcs, widgetConfiguratorFunc(c.Move.Widgets))
	}

	for _, wc := range wcs {
		ws, err := wc.Widgets(app)
		if err != nil {
			return nil, fmt.Errorf("moonraker: error configuring widgets %T: %w", wc, err)
		}
		for _, w := range ws {
			app.AddWidget(w.Widget, w.Layer, w.X, w.Y)
		}
	}

	if c.Emergency != nil {
		w := moondeck.NewIconColorWidget("emergency", icon.Red, icon.Transparent)
		w.OnPress = func(_ moondeck.Button, _ *moondeck.App) {
			panic("emergency")
		}
		w.OnRemove = func(_ *moondeck.App) {
			app.AddWidget(w, moondeck.Overlay, c.Emergency.At.X, c.Emergency.At.Y)
		}
		app.AddWidget(w, moondeck.Overlay, c.Emergency.At.X, c.Emergency.At.Y)
	}

	return app, nil
}

type EmergencyConfig struct {
	At      util.Point `hcl:"at"`
	Confirm bool       `hcl:"confirm"`
}

type widgetConfigurator interface {
	Widgets(*moondeck.App) ([]WidgetConfig, error)
}

type widgetConfiguratorFunc func(*moondeck.App) ([]WidgetConfig, error)

func (f widgetConfiguratorFunc) Widgets(app *moondeck.App) ([]WidgetConfig, error) {
	return f(app)
}

type WidgetConfig struct {
	moondeck.Widget
	Layer int
	util.Point
}

// Load a Moondeck UI configuration file.
func Load(name string) (*Config, error) {
	var config Config
	if err := hclsimple.DecodeFile(name, nil, &config); err != nil {
		return nil, err
	}

	if len(config.App) == 0 {
		return nil, fmt.Errorf("moonraker: %s has no apps configured", name)
	}

	var homeFound bool
	for _, app := range config.App {
		if homeFound = app.Name == "home"; homeFound {
			break
		}
	}
	if !homeFound {
		return nil, fmt.Errorf("moonraker: %s has no \"home\" app configured", name)
	}

	if config.Klipper.API == "" {
		log.Println("moonraker: no klipper.api address configured, using", DefaultKlipperAPI)
		config.Klipper.API = DefaultKlipperAPI
	}

	return &config, nil
}
