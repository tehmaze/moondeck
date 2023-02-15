package moondeck

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	builtin "maze.io/moondeck/moondeck/font"
	"maze.io/moondeck/moondeck/icon"
	"maze.io/moondeck/util"
)

// ButtonTrigger is used for the ButtonTriggerSchedule.
type ButtonTrigger struct {
	After   time.Duration
	Trigger time.Duration
}

// ButtonTriggerSchedule is used to determine how often a button is retriggered
// when the user keeps the button pressed down.
var ButtonTriggerSchedule = []ButtonTrigger{
	{1000 * time.Millisecond, time.Second / 2},
	{2500 * time.Millisecond, time.Second / 3},
	{5000 * time.Millisecond, time.Second / 4},
}

// Button is a single button on a Deck.
type Button interface {
	// Deck the button is on.
	Deck() Deck

	// Index of the button.
	Index() int

	// Size of the button in pixels.
	Size() util.Size

	// Pos is the button position.
	Pos() util.Point

	// SetColor sets the button to a uniform color.
	SetColor(color.Color) error

	// SetImage sets the button to an image.
	SetImage(image.Image) error
}

// ButtonEvent is fired when a key is pressed or released.
type ButtonEvent struct {
	Button

	// Pressed is true if the button is pressed, false when it's released.
	Pressed bool

	// Duration indicates how long the button is pressed on.
	Duration time.Duration
}

// ButtonAt is a helper that resolves the Button at position (x, y).
func ButtonAt(deck Deck, x, y int) (Button, bool) {
	if deck == nil {
		return nil, false
	}
	var (
		s = deck.Size()
		i = y*s.W + s.H
	)
	return deck.Button(i)
}

// Text is a helper to render text onto a button, the text is center aligned.
type Text struct {
	Font     *truetype.Font
	FontSize float64
}

// Defaults for Text.
var (
	DefaultTextBackground = color.Transparent
	DefaultTextForeground = color.White
	DefaultTextSize       = 16.0
	DefaultText           = &Text{
		Font:     builtin.Roboto,
		FontSize: DefaultTextSize,
	}
)

// Render text onto a Button.
func (t *Text) Render(b Button, text string, fg, bg color.Color) error {
	var (
		s = b.Size()
		i = image.NewRGBA(image.Rectangle{Max: s.ImagePoint()})
		f = t.Font
		//bg = t.Background
		//fg = t.Foreground
	)
	if f == nil {
		f = builtin.Roboto
	}
	if bg == nil {
		bg = DefaultTextBackground
	}
	if fg == nil {
		fg = icon.White
	}
	draw.Draw(i, i.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(t.Font)

	fontSize := t.FontSize
	if fontSize == 0 {
		fontSize = DefaultTextSize
	}

	ctx.SetFontSize(fontSize)

	fontFace := truetype.NewFace(f, &truetype.Options{
		Size: t.FontSize,
		DPI:  72,
	})
	fontDraw := &font.Drawer{
		Dst:  i,
		Src:  image.NewUniform(fg),
		Face: fontFace,
	}
	textBounds, _ := fontDraw.BoundString(text)
	xPosition := (fixed.I(s.W) - fontDraw.MeasureString(text)) / 2
	textHeight := textBounds.Max.Y - textBounds.Min.Y
	yPosition := fixed.I((s.H)-textHeight.Ceil())/2 + fixed.I(textHeight.Ceil())
	fontDraw.Dot = fixed.Point26_6{
		X: xPosition,
		Y: yPosition,
	}
	fontDraw.DrawString(text)

	return b.SetImage(i)
}

// ButtonText is a helper that uses the Text defaults.
func ButtonText(b Button, text string) error {
	return DefaultText.Render(b, text, icon.White, icon.Transparent)
}

func ButtonTextColor(b Button, text string, fg, bg color.Color) error {
	return DefaultText.Render(b, text, fg, bg)
}
