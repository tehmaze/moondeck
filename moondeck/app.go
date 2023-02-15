package moondeck

import (
	"fmt"
	"image"
	"image/draw"

	log "github.com/sirupsen/logrus"

	"maze.io/moondeck/util"
)

// Layer names.
const (
	Background = iota
	Foreground
	Overlay
)

// App is a collection of widgets.
type App struct {
	// Dashboard this App is connected to.
	Dashboard *Dashboard

	// Name of this app.
	Name string

	// IsActive if the App is currently active.
	IsActive bool

	// Layers are the background, foreground and overlay layers.
	Layers [3]Layer

	// size of the grid
	size util.Size

	// Callbacks
	onStart []func(*App)
	onClose []func(*App)
}

func NewApp(name string, dash *Dashboard) *App {
	return &App{
		Dashboard: dash,
		Name:      name,
		Layers:    NewLayers(dash.Buttons()),
		size:      dash.Size(),
	}
}

func (app *App) ButtonPressed(b Button) {
	if w := app.widgetAt(b.Index()); w != nil {
		w.Pressed(b, app)
	}
}

func (app *App) ButtonReleased(b Button) {
	if w := app.widgetAt(b.Index()); w != nil {
		w.Released(b, app)
	}
}

func (app *App) Start() {
	for _, f := range app.onStart {
		f(app)
	}
}

func (app *App) Close() {
	app.CloseOverlay()
	for _, f := range app.onClose {
		f(app)
	}
}

func (app *App) CloseOverlay() {
	for i, w := range app.Layers[Overlay] {
		if w != nil {
			if w, ok := w.(RemovableWidget); ok {
				w.Removed(app)
			}
			app.Layers[Overlay][i] = nil
			app.renderButton(i, true)
		}
	}
}

func (app *App) SetBackgroundImage(i image.Image) {
	var (
		size  = app.Dashboard.Size()
		r     = i.Bounds()
		dx    = r.Dx()
		dy    = r.Dy()
		stepx = dx / size.W
		stepy = dy / size.H
		tile  = image.NewRGBA(image.Rect(0, 0, stepx, stepy))
		j     int
	)
	dx = size.W * stepx
	dy = size.H * stepy
	for y := 0; y < dy; y += stepy {
		for x := 0; x < dx; x += stepx {
			draw.Draw(tile, tile.Bounds(), i, image.Pt(x, y), draw.Src)
			app.Layers[Background][j] = NewImageWidget(tile)
			j++
		}
	}
}

func (app *App) widgetAt(index int) Widget {
	for layer := Overlay; layer >= Background; layer-- {
		if w := app.Layers[layer][index]; w != nil {
			return w
		}
	}
	return nil
}

func (app *App) AddWidget(w Widget, layer, x, y int) bool {
	return app.AddWidgetAt(w, layer, y*app.size.W+x)
}

func (app *App) AddWidgetAt(w Widget, layer, index int) bool {
	logger := log.WithFields(log.Fields{
		"layer": layer,
		"index": index,
	})
	if layer < Background || layer > Overlay {
		logger.WithField("type", fmt.Sprintf("%T", w)).Error("widget not added: out of bounds")
		return false
	}
	if index < 0 || index > len(app.Layers[layer]) {
		logger.WithField("type", fmt.Sprintf("%T", w)).Error("widget not added: out of range")
		return false
	}
	logger.WithField("type", fmt.Sprintf("%T", w)).Debug("widget added")
	app.Layers[layer][index] = w
	app.RenderButton(index)
	return true
}

func (app *App) RemoveWidget(w Widget) bool {
	for layer := Overlay; layer >= Background; layer-- {
		for index, o := range app.Layers[layer] {
			if o == w {
				app.Layers[layer][index] = nil
				app.renderButton(index, true)
				return true
			}
		}
	}
	return false
}

func (app *App) RemoveWidgetAt(layer, index int) bool {
	if layer < Background || layer > Overlay {
		return false
	}
	if index < 0 || index > len(app.Layers[layer]) || app.Layers[layer][index] == nil {
		return false
	}
	app.Layers[layer][index] = nil
	app.renderButton(index, true)
	return true
}

func (app *App) Render() error {
	return app.render(false)
}

func (app *App) render(force bool) error {
	for i := 0; i < app.Dashboard.Buttons(); i++ {
		if err := app.renderButton(i, force); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) RenderButton(index int) error {
	return app.renderButton(index, true)
}

func (app *App) renderButton(index int, force bool) error {
	w := app.widgetAt(index)
	if w == nil {
		return nil
	}

	if !(force || w.IsDirty()) {
		return nil
	}

	b, ok := app.Dashboard.Button(index)
	if !ok {
		return nil
	}

	return w.Render(b)
}

func (app *App) OnStart(f func(*App)) {
	app.onStart = append(app.onStart, f)
}

func (app *App) OnClose(f func(*App)) {
	app.onClose = append(app.onClose, f)
}

type Layer []Widget

func NewLayers(buttons int) [3]Layer {
	return [3]Layer{
		make([]Widget, buttons),
		make([]Widget, buttons),
		make([]Widget, buttons),
	}
}
