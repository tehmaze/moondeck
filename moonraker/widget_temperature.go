package moonraker

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/image/draw"

	"maze.io/moondeck/gfx/blend"
	"maze.io/moondeck/gfx/font"
	"maze.io/moondeck/gfx/icon"
	"maze.io/moondeck/gfx/sparkline"
	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/util"
)

var (
	defaultTempForeground = icon.Yellow
	defaultTempBackground = icon.Transparent
)

type TempConfig struct {
	Heater     string      `hcl:"heater,label"`
	At         util.Point  `hcl:"at"`
	Graph      util.Size   `hcl:"graph,optional"`
	Foreground ColorConfig `hcl:"fg,optional"`
	Background ColorConfig `hcl:"bg,optional"`
}

func (c *TempConfig) Widgets(app *moondeck.App) (wcs []WidgetConfig, err error) {
	w := NewTemperature(c.Heater, app,
		c.Foreground.RGBADefault(defaultTempForeground),
		c.Background.RGBADefault(defaultTempBackground))
	wcs = append(wcs, WidgetConfig{w, moondeck.Foreground, c.At})
	if c.Graph.W > 0 && c.Graph.H > 0 {
		for _, wc := range w.Graph(72, c.Graph.W, c.Graph.H) {
			wcs = append(wcs, WidgetConfig{
				Widget: wc.Widget,
				Layer:  wc.Layer,
				Point:  c.At.Add(wc.Point).Add(util.Point{X: 1}),
			})
		}
	}
	return
}

type Temperature struct {
	Name      string
	Value     float64
	Target    float64
	Color     color.RGBA
	ColorEdit color.RGBA
	fg        color.RGBA
	graph     *TemperatureGraph
	isClean   bool
	isEditing bool
	up, dn    *moondeck.ImageWidget
	edit      *moondeck.FloatWidget
}

func NewTemperature(name string, app *moondeck.App, fg, bg color.RGBA) *Temperature {
	w := &Temperature{
		Name:      name,
		Color:     icon.White,
		ColorEdit: icon.Yellow,
		up:        moondeck.MustIconWidget("arrow-up"),
		dn:        moondeck.MustIconWidget("arrow-down"),
		edit: &moondeck.FloatWidget{
			TextWidget: moondeck.TextWidget{
				Font:       font.RobotoBold,
				FontSize:   16,
				TextColor:  fg,
				Background: bg,
			},
		},
		fg: icon.White,
	}
	w.up.OnPress = w.increaseTarget
	w.dn.OnPress = w.decreaseTarget
	w.edit.OnPress = w.setTarget
	app.OnClose(w.onAppClose)

	/*
		go func() {
			w.Update(rand.Float64() * 150)
			for range time.Tick(time.Second / 10) {
				w.Value += (rand.Float64() - 0.5) * 10
				if w.Value < 0 {
					w.Value = -w.Value
				}
				w.Update(w.Value)
			}
		}()
	*/

	var (
		online atomic.Bool
		once   sync.Once
	)

	IsOnline(func(c *Client, _ error) {
		online.Store(true)
		once.Do(func() {
			go w.updateTemperature(c, online)
		})
	})

	IsOnline(func(c *Client, _ error) {
		ts, err := c.TemperatureStores()
		if err != nil {
			log.WithError(err).Error("error obtaining temperature stores")
			return
		}
		if t, ok := ts[name]; ok {
			for _, v := range t.Temperatures {
				w.Update(v)
			}
		} else {
			log.
				WithError(err).
				WithField("heater", name).
				Error("error obtaining temperature stores for heater")
		}
	})

	return w
}

func (w *Temperature) Graph(pixels, width, height int) []WidgetConfig {
	if w.graph == nil {
		w.graph = NewTemperatureGraph(pixels, width, height)
		w.graph.Push(0)
	}
	return w.graph.widgetConfig
}

func (w *Temperature) Update(value float64) {
	w.Value = value
	w.isClean = false
	if w.graph != nil {
		w.graph.Push(value)
	}
}

func (w *Temperature) updateTemperature(c *Client, online atomic.Bool) {
	logger := log.WithField("heater", w.Name)
	for range time.Tick(time.Second) {
		if online.Load() && c != nil {
			switch w.Name {
			case "heater_bed":
				t, err := c.HeaterBedStatus()
				if err != nil {
					logger.
						WithError(err).
						Error("failed to retrieve heater status")
				} else {
					logger.
						WithFields(log.Fields{
							"temperature": t.Temperature,
							"target":      t.Target,
							"power":       t.Power,
						}).
						Debug("heater status")
					w.Update(t.Temperature)
					if !w.isEditing {
						w.Target = t.Target
					}
				}
			default:
				t, err := c.ExtruderStatus(0)
				if err != nil {
					logger.
						WithError(err).
						Error("failed to retrieve heater status")
				} else {
					w.Update(t.Temperature)
					if !w.isEditing {
						w.Target = t.Target
					}
				}
			}
		}
	}
}

func (w *Temperature) Dirty()        { w.isClean = false }
func (w *Temperature) IsDirty() bool { return !w.isClean }

func (w *Temperature) increaseTarget(b moondeck.Button, _ *moondeck.App) {
	w.Target += 1.0
	w.edit.Value = w.Target
	w.edit.Dirty()
}

func (w *Temperature) decreaseTarget(b moondeck.Button, _ *moondeck.App) {
	w.Value -= 1.0
	w.edit.Value = w.Target
	w.edit.Dirty()
}

func (w *Temperature) setTarget(b moondeck.Button, app *moondeck.App) {
	app.CloseOverlay()
	w.isEditing = false
	w.Dirty()
}

func (w *Temperature) Pressed(b moondeck.Button, app *moondeck.App) {
	log.Println("moonraker: pressed temperature", w.Name)
	if w.isEditing {
		w.isEditing = false
		w.fg = w.Color
		app.RemoveWidget(w.up)
		app.RemoveWidget(w.dn)
	} else {
		app.CloseOverlay()
		w.isEditing = true
		var (
			size = app.Dashboard.Size()
			pos  = b.Pos()
		)
		app.AddWidget(w.edit, moondeck.Overlay, pos.X, pos.Y)
		if size.H < 3 {
			switch pos.X {
			case 0:
				// two buttons right of temperature
				app.AddWidget(w.up, moondeck.Overlay, pos.X+1, pos.Y)
				app.AddWidget(w.dn, moondeck.Overlay, pos.X+2, pos.Y)
			case size.W - 1:
				// two buttons left of temperature
				app.AddWidget(w.up, moondeck.Overlay, pos.X-2, pos.Y)
				app.AddWidget(w.dn, moondeck.Overlay, pos.X-1, pos.Y)
			default:
				// two buttons next to temperature
				app.AddWidget(w.up, moondeck.Overlay, pos.X-1, pos.Y)
				app.AddWidget(w.dn, moondeck.Overlay, pos.X+1, pos.Y)
			}
		} else {
			switch pos.Y {
			case 0:
				// two buttons below temperature
				app.AddWidget(w.up, moondeck.Overlay, pos.X, pos.Y+1)
				app.AddWidget(w.dn, moondeck.Overlay, pos.X, pos.Y+2)
			case size.H - 1:
				// two buttons above temperature
				app.AddWidget(w.up, moondeck.Overlay, pos.X, pos.Y-2)
				app.AddWidget(w.dn, moondeck.Overlay, pos.X, pos.Y-1)
			default:
				// two buttons next to temperature
				app.AddWidget(w.up, moondeck.Overlay, pos.X, pos.Y-1)
				app.AddWidget(w.dn, moondeck.Overlay, pos.X, pos.Y+1)
			}
		}
	}
}

func (w *Temperature) Released(b moondeck.Button, app *moondeck.App) {}

func (w *Temperature) onAppClose(app *moondeck.App) {
	w.isEditing = false
	w.fg = w.Color
}

func (w *Temperature) Render(b moondeck.Button) error {
	t := &moondeck.Text{
		Font:     font.RobotoBold,
		FontSize: 16,
	}

	return t.Render(b, fmt.Sprintf("%.1fC", w.Value), w.fg, color.Transparent)
}

type TemperatureGraph struct {
	pixels       int
	w, h         int
	sparkline    *sparkline.Sparkline
	image        *image.RGBA
	widgetConfig []WidgetConfig
}

func NewTemperatureGraph(pixels, w, h int) *TemperatureGraph {
	g := &TemperatureGraph{
		pixels:    pixels,
		w:         w,
		h:         h,
		sparkline: sparkline.New(pixels*w, flames, icon.Transparent), // , white, flames),
		image:     image.NewRGBA(image.Rect(0, 0, w*pixels, h*pixels)),
	}
	g.sparkline.FixedMin(0)

	// Create widgets
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g.widgetConfig = append(g.widgetConfig, WidgetConfig{
				Widget: moondeck.NewImageWidget(image.NewRGBA(image.Rect(0, 0, pixels, pixels))),
				Layer:  moondeck.Foreground,
				Point:  util.Pt(x, y),
			})
		}
	}

	return g
}

func (g *TemperatureGraph) Push(value float64) {
	//log.Printf("moonraker: push %.1f to time series graph", value)
	//g.sparkline.Add(value)
	//g.sparkline.Draw(g.image)
	g.sparkline.Push(value)
	g.sparkline.Draw(g.image) // , icon.Red, icon.White)

	var j int
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			var (
				sp = image.Pt(x*g.pixels, y*g.pixels)
				iw = g.widgetConfig[j].Widget.(*moondeck.ImageWidget)
			)
			j++
			/*
				draw.Copy(
					iw.Image,      // dst
					image.Point{}, // dp
					g.image,       // src
					sr,            // sr
					draw.Src,      // op
					nil,           // options
				)
			*/
			draw.Draw(iw.Image, iw.Image.Bounds(), g.image, sp, draw.Src)
			iw.Dirty()
		}
	}

	//w := g.widgetConfig[0].Widget.(*moondeck.ImageWidget)
	//w.Update(g.image)
}

var (
	_ moondeck.Widget = (*Temperature)(nil)
)

var (
	white  = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	green  = color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff}
	orange = color.RGBA{R: 0xff, G: 0x7f, B: 0x00, A: 0xff}
	yellow = color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}
	red    = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}

	flames = []blend.Threshold{
		{Max: 500, Color: white},
		{Max: 250, Color: yellow},
		{Max: 100, Color: red},
		{Max: 60, Color: orange},
		{Max: 25, Color: green},
	}
	trafficLights = []blend.Threshold{
		{
			Max:   100,
			Color: white,
		},
		{
			Max:   50,
			Color: green,
		},
		{
			Max:   30,
			Color: yellow,
		},
		{
			Max:   20,
			Color: red,
		},
	}
)
