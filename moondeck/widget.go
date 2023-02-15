package moondeck

import (
	"image"
	"image/color"
	"image/draw"
	"strconv"

	"github.com/golang/freetype/truetype"

	"maze.io/moondeck/gfx"
	builtin "maze.io/moondeck/gfx/font"
	"maze.io/moondeck/gfx/icon"
)

type Widget interface {
	// Pressed the Widget Button.
	Pressed(Button, *App)

	// Released the Widget Button.
	Released(Button, *App)

	// Dirty marks this Widget as dirty.
	Dirty()

	// IsDirty checks if the Widget needs to be rendered.
	IsDirty() bool

	// Render the Widget onto the Button.
	Render(Button) error
}

type RemovableWidget interface {
	Removed(*App)
}

type BaseWidget struct {
	BeforeRender func(Widget, Button)
	OnPress      func(Button, *App)
	OnRelease    func(Button, *App)
	OnRemove     func(*App)
	isClean      bool
}

func (w *BaseWidget) Dirty()        { w.isClean = false }
func (w *BaseWidget) IsDirty() bool { return !w.isClean }

func (w *BaseWidget) Removed(app *App) {
	if w.OnRemove != nil {
		w.OnRemove(app)
	}
}

func (w *BaseWidget) Render(_ Button) error {
	w.isClean = true
	return nil
}

type ImageWidget struct {
	BaseWidget
	Image *image.RGBA
}

func NewImageWidget(i image.Image) *ImageWidget {
	w := &ImageWidget{}
	w.Update(i)
	return w
}

func NewIconWidget(name string) (*ImageWidget, error) {
	i, err := icon.Icon(name)
	if err != nil {
		return nil, err
	}

	w := new(ImageWidget)
	w.Update(i)

	return w, nil
}

func MustIconWidget(name string) *ImageWidget {
	w, err := NewIconWidget(name)
	if err != nil {
		panic(err)
	}
	return w
}

func NewIconColorWidget(name string, fg, bg color.RGBA) *ImageWidget {
	w := &ImageWidget{}
	w.Update(gfx.Colorize(icon.Must(name), fg, bg))
	return w
}

func (w *ImageWidget) Update(i image.Image) {
	w.Image = image.NewRGBA(i.Bounds())
	draw.Draw(w.Image, w.Image.Bounds(), i, image.Point{}, draw.Src)
}

func (w *ImageWidget) Render(b Button) error {
	if w.BeforeRender != nil {
		w.BeforeRender(w, b)
	}
	w.isClean = true
	return b.SetImage(w.Image)
}

func (w *ImageWidget) Pressed(b Button, app *App) {
	// log.Printf("moondeck: pressed %T %d", w, b.Index())
	if w.OnPress != nil {
		w.OnPress(b, app)
	}
}

func (w *ImageWidget) Released(b Button, app *App) {
	// log.Printf("moondeck: released %T %d", w, b.Index())
	if w.OnRelease != nil {
		w.OnRelease(b, app)
	}
}

type TextWidget struct {
	BaseWidget
	Text       string
	TextColor  color.Color
	Background color.Color
	Font       *truetype.Font
	FontSize   float64
}

func (w *TextWidget) Render(b Button) error {
	if w.BeforeRender != nil {
		w.BeforeRender(w, b)
	}
	w.isClean = true
	//return ButtonText(b, w.Text)
	t := Text{
		Font:     w.Font,
		FontSize: w.FontSize,
	}
	if t.Font == nil {
		t.Font = builtin.RobotoBold
	}
	if t.FontSize <= 0 {
		t.FontSize = DefaultTextSize
	}
	return t.Render(b, w.Text, w.TextColor, w.Background)
}

func (w *TextWidget) Pressed(b Button, app *App) {
	// log.Printf("moondeck: pressed %T %d", w, b.Index())
	if w.OnPress != nil {
		w.OnPress(b, app)
	}
}

func (w *TextWidget) Released(b Button, app *App) {
	//log.Printf("moondeck: released %T %d", w, b.Index())
	if w.OnRelease != nil {
		w.OnRelease(b, app)
	}
}

type FloatWidget struct {
	TextWidget
	Value     float64
	Precision int
	Prefix    string
	Suffix    string
}

func (w *FloatWidget) Render(b Button) error {
	w.Text = w.Prefix + strconv.FormatFloat(w.Value, 'f', w.Precision, 64) + w.Suffix
	return w.TextWidget.Render(b)
}

var (
	_ Widget = (*ImageWidget)(nil)
	_ Widget = (*TextWidget)(nil)
	_ Widget = (*FloatWidget)(nil)
)
