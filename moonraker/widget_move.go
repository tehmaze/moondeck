package moonraker

import (
	"log"
	"strconv"

	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/moondeck/font"
	"maze.io/moondeck/moondeck/icon"
	"maze.io/moondeck/util"
)

type MoveConfig struct {
	At util.Point `hcl:"at"`
}

func (c *MoveConfig) Widgets(app *moondeck.App) ([]WidgetConfig, error) {
	m := NewMove(c.At)
	return []WidgetConfig{
		{Widget: m.cross[0], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(0, 0))},
		{Widget: m.cross[1], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(1, 0))},
		{Widget: m.cross[2], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(2, 0))},
		{Widget: m.cross[3], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(3, 0))},
		{Widget: m.cross[4], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(0, 1))},
		{Widget: m.cross[5], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(1, 1))},
		{Widget: m.cross[6], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(2, 1))},
		{Widget: m.cross[7], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(3, 1))},
		{Widget: m.cross[8], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(0, 2))},
		{Widget: m.cross[9], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(1, 2))},
		{Widget: m.cross[10], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(2, 2))},
		{Widget: m.cross[11], Layer: moondeck.Foreground, Point: c.At.Add(util.Pt(3, 2))},
	}, nil
}

var moveStepSizes = []float64{0.1, 1, 10, 25, 100, 250}

type Move struct {
	at        util.Point
	cross     [12]moondeck.Widget
	step      float64
	stepIndex int
}

func NewMove(at util.Point) *Move {
	w := &Move{
		at:        at,
		step:      moveStepSizes[1],
		stepIndex: 1,
	}
	w.cross = [12]moondeck.Widget{
		moondeck.NewIconWidget("arrow-up-left"),           // 00: move to back left
		moondeck.NewIconWidget("arrow-up"),                // 01: move back
		moondeck.NewIconWidget("arrow-up-right"),          // 02: move to back right
		moondeck.NewIconWidget("arrow-up-from-bracket"),   // 03: move bed up
		moondeck.NewIconWidget("arrow-left"),              // 04: move to left
		moondeck.NewIconWidget("crosshairs"),              // 05: home
		moondeck.NewIconWidget("arrow-right"),             // 06: move to right
		nil,                                               // 07: tbd
		moondeck.NewIconWidget("arrow-down-left"),         // 08: move to front left
		moondeck.NewIconWidget("arrow-down"),              // 09: move to front
		moondeck.NewIconWidget("arrow-down-right"),        // 10: move to front right
		moondeck.NewIconWidget("arrow-down-from-bracket"), // 11: move bed down
	}
	w.cross[0].(*moondeck.ImageWidget).OnPress = w.move(-1, -1, +0)
	w.cross[1].(*moondeck.ImageWidget).OnPress = w.move(+0, -1, +0)
	w.cross[2].(*moondeck.ImageWidget).OnPress = w.move(+1, -1, +0)
	w.cross[3].(*moondeck.ImageWidget).OnPress = w.move(+0, +0, +1)

	w.cross[4].(*moondeck.ImageWidget).OnPress = w.move(-1, +0, +0)
	w.cross[5].(*moondeck.ImageWidget).OnPress = w.home
	w.cross[6].(*moondeck.ImageWidget).OnPress = w.move(+1, +0, +0)
	w.cross[7] = &moondeck.TextWidget{
		Text:     strconv.FormatFloat(w.step, 'f', 0, 64),
		Font:     font.RobotoBold,
		FontSize: 16,
	}
	w.cross[7].(*moondeck.TextWidget).OnPress = w.stepChange

	w.cross[8].(*moondeck.ImageWidget).OnPress = w.move(-1, +1, +0)
	w.cross[9].(*moondeck.ImageWidget).OnPress = w.move(+0, +1, +0)
	w.cross[10].(*moondeck.ImageWidget).OnPress = w.move(+1, +1, +0)
	w.cross[11].(*moondeck.ImageWidget).OnPress = w.move(+0, +0, -1)
	return w
}

func (w *Move) move(x, y, z float64) func(moondeck.Button, *moondeck.App) {
	return func(b moondeck.Button, app *moondeck.App) {
		x *= w.step
		y *= w.step
		z *= w.step
		log.Printf("moonraker: move toolhead %f, %f, %f", x, y, z)
	}
}

func (w *Move) home(b moondeck.Button, app *moondeck.App) {
	//log.Println("moonraker: move home")
	var (
		home  = moondeck.NewIconColorWidget("crosshairs", icon.Yellow, icon.Transparent)
		homeA = &moondeck.TextWidget{Font: font.RobotoBold, FontSize: 16, TextColor: icon.Yellow, Text: "All"}
		homeX = &moondeck.TextWidget{Font: font.RobotoBold, FontSize: 16, TextColor: icon.Yellow, Text: "X"}
		homeY = &moondeck.TextWidget{Font: font.RobotoBold, FontSize: 16, TextColor: icon.Yellow, Text: "Y"}
		homeZ = &moondeck.TextWidget{Font: font.RobotoBold, FontSize: 16, TextColor: icon.Yellow, Text: "Z"}
	)
	home.OnPress = func(_ moondeck.Button, _ *moondeck.App) {
		app.CloseOverlay()
		home = nil
		homeA = nil
		homeX = nil
		homeZ = nil
	}
	homeA.OnPress = w.homeAxis("")
	homeX.OnPress = w.homeAxis("X")
	homeY.OnPress = w.homeAxis("Y")
	homeZ.OnPress = w.homeAxis("Z")
	app.AddWidget(home, moondeck.Overlay, w.at.X+1, w.at.Y+1)
	app.AddWidget(homeA, moondeck.Overlay, w.at.X+0, w.at.Y)
	app.AddWidget(homeX, moondeck.Overlay, w.at.X+1, w.at.Y)
	app.AddWidget(homeY, moondeck.Overlay, w.at.X+2, w.at.Y)
	app.AddWidget(homeZ, moondeck.Overlay, w.at.X+3, w.at.Y)
}

func (w *Move) homeAxis(axis string) func(moondeck.Button, *moondeck.App) {
	return func(b moondeck.Button, app *moondeck.App) {
		app.CloseOverlay()
		log.Println("moonraker: home", axis)
	}
}

func (w *Move) stepChange(b moondeck.Button, app *moondeck.App) {
	w.stepIndex++
	if w.stepIndex >= len(moveStepSizes) {
		w.stepIndex = 0
	}
	w.step = moveStepSizes[w.stepIndex]

	var precision int
	if w.step < 1 {
		precision = 1
	}
	w.cross[7].(*moondeck.TextWidget).Text = strconv.FormatFloat(w.step, 'f', precision, 64)

	w.cross[7].(*moondeck.TextWidget).Dirty()
}
