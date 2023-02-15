package moonraker

import (
	"fmt"
	"image"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/bamiaux/rez"
	"golang.org/x/image/draw"

	"maze.io/moondeck/gfx/mjpeg"
	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/util"
)

var DefaultCameraFPS = 10.0

type CameraConfig struct {
	URL  string     `hcl:"url"`
	FPS  float64    `hcl:"fps,optional"`
	At   util.Point `hcl:"at"`
	Size util.Size  `hcl:"size,optional"`
}

func (c *CameraConfig) Widgets(app *moondeck.App) ([]WidgetConfig, error) {
	if _, err := url.Parse(c.URL); err != nil {
		return nil, fmt.Errorf("moonraker: invalid camera URL %q: %w", c.URL, err)
	}
	if c.FPS <= 0 {
		c.FPS = DefaultCameraFPS
	}
	if c.Size.W == 0 {
		c.Size.W = app.Dashboard.Size().W - c.At.X
	}
	if c.Size.H == 0 {
		c.Size.H = app.Dashboard.Size().H - c.At.Y
	}

	var (
		wcs  []WidgetConfig
		grid = make([]*moondeck.ImageWidget, c.Size.Area()-1)
		tile = image.NewRGBA(image.Rectangle{Max: app.Dashboard.ButtonSize().ImagePoint()})
	)
	for i := range grid {
		grid[i] = moondeck.NewImageWidget(tile)

		x := ((i + 1) % c.Size.W) + c.At.X
		y := ((i + 1) / c.Size.W) + c.At.Y
		wcs = append(wcs, WidgetConfig{
			Widget: grid[i],
			Layer:  moondeck.Foreground,
			Point:  util.Pt(x, y),
		})
	}

	return append([]WidgetConfig{{
		Widget: NewCamera(app, c.URL, c.FPS, grid, c.Size),
		Layer:  moondeck.Foreground,
		Point:  c.At,
	}}, wcs...), nil
}

type Camera struct {
	start        chan struct{}
	frame        chan image.Image
	url          string
	size         util.Size
	deck         moondeck.Deck
	grid         []*moondeck.ImageWidget
	layout       util.Size
	decoderMutex sync.RWMutex
	decoder      *mjpeg.Decoder
	isClean      bool
	isVisible    bool
}

func NewCamera(app *moondeck.App, url string, fps float64, grid []*moondeck.ImageWidget, layout util.Size) *Camera {
	w := &Camera{
		start:  make(chan struct{}, 1),
		frame:  make(chan image.Image, 1),
		url:    url,
		size:   app.Dashboard.ButtonSize(),
		deck:   app.Dashboard.Deck,
		layout: layout,
	}

	tile := image.NewRGBA(image.Rectangle{Max: w.size.ImagePoint()})
	w.grid = append(w.grid, moondeck.NewImageWidget(tile))
	w.grid = append(w.grid, grid...)
	log.Printf("moonraker: camera of %s with %d grid tiles", w.size, len(w.grid))

	go w.stream()
	go w.render()

	app.OnStart(w.onAppStart)
	app.OnClose(w.onAppClose)

	return w
}

func (w *Camera) stream() {
connecting:
	for {
		<-w.start

		log.Println("moonraker: starting camera stream to", w.url)

		var err error
		w.decoderMutex.Lock()
		w.decoder, err = mjpeg.NewDecoderFromURL(w.url)
		w.decoderMutex.Unlock()
		if err != nil {
			log.Println("moonraker: error connecting to MJPEG stream", w.url+":", err)
			time.Sleep(1)
			if w.isVisible {
				w.start <- struct{}{}
			}
			continue connecting
		}

		for {
			w.decoderMutex.RLock()
			i, err := w.decoder.Next()
			w.decoderMutex.RUnlock()
			if err != nil {
				log.Println("moonraker: error decoding MJPEG frame:", err)
				if w.isVisible {
					w.start <- struct{}{}
				}
				continue connecting
			}

			select {
			case w.frame <- i:
				//log.Println("moonraker: camera received frame of", i.Bounds().Max)
			default:
				log.Println("moonraker: camera frame dropped")
			}
		}
	}
}

func (w *Camera) render() {
	var (
		resized   *image.RGBA
		cfg       *rez.ConverterConfig
		converter rez.Converter
		err       error
	)
	for {
		frame := toRGBA(<-w.frame)
		//log.Printf("moonraker: camera frame %T", frame)
		if resized == nil {
			size := image.Rectangle{Max: w.size.Mul(w.layout).ImagePoint()}
			resized = image.NewRGBA(size)
			if cfg, err = rez.PrepareConversion(resized, frame); err != nil {
				panic(err)
			}
			//if converter, err = rez.NewConverter(cfg, rez.NewLanczosFilter(3)); err != nil {
			if converter, err = rez.NewConverter(cfg, rez.NewBilinearFilter()); err != nil {
				panic(err)
			}
		}
		if converter.Convert(resized, frame) == nil {
			for i, t := range w.grid {
				//log.Printf("moonraker: camera grid %d is %+v", i, t)
				var (
					x  = (i % w.layout.W) * w.size.W
					y  = (i / w.layout.W) * w.size.H
					sp = image.Pt(x, y)
				)
				draw.Draw(t.Image, t.Image.Bounds(), resized, sp, draw.Src)
				t.Dirty()
			}
		}
	}
}

func (w *Camera) Dirty()        { w.isClean = false }
func (w *Camera) IsDirty() bool { return !w.isClean }

func (w *Camera) onAppClose(_ *moondeck.App) {
	w.isVisible = false
	w.decoderMutex.Lock()
	if w.decoder != nil {
		_ = w.decoder.Close()
		w.decoder = nil
	}
	w.decoderMutex.Unlock()
}

func (w *Camera) onAppStart(_ *moondeck.App) {
	if !w.isVisible {
		log.Println("moonraker: camera visible (triggered by render)")
		w.isVisible = true
		select {
		case w.start <- struct{}{}:
		default:
		}
	}
}

func (Camera) Pressed(_ moondeck.Button, _ *moondeck.App)  {}
func (Camera) Released(_ moondeck.Button, _ *moondeck.App) {}

func (w *Camera) Render(b moondeck.Button) error {
	return w.grid[0].Render(b)
}

func toRGBA(i image.Image) *image.RGBA {
	if i, ok := i.(*image.RGBA); ok {
		return i
	}
	o := image.NewRGBA(i.Bounds())
	draw.Copy(o, image.Point{}, i, i.Bounds(), draw.Src, nil)
	return o
}

var _ moondeck.Widget = (*Camera)(nil)
