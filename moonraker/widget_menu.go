package moonraker

import (
	"fmt"
	"image/color"

	log "github.com/sirupsen/logrus"

	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/util"
)

type AppMenuConfig struct {
	Name string     `hcl:"name,label"`
	At   util.Point `hcl:"at"`
}

func (c *AppMenuConfig) Widgets(menus []*MenuConfig) func(_ *moondeck.App) ([]WidgetConfig, error) {
	return func(_ *moondeck.App) ([]WidgetConfig, error) {
		for _, menu := range menus {
			if menu.Name == c.Name {
				return []WidgetConfig{{
					Widget: NewMenu(menu.Icon, menu.Item, menu.ForegroundColor.RGBA(), menu.BackgroundColor.RGBA()),
					Layer:  moondeck.Foreground,
					Point:  c.At,
				}}, nil
			}
		}
		return nil, fmt.Errorf("moonraker: menu %q not defined", c.Name)
	}
}

type Menu struct {
	moondeck.BaseWidget
	menu      *moondeck.ImageWidget
	open      *moondeck.ImageWidget
	item      []*moondeck.ImageWidget
	isVisible bool
}

func NewMenu(iconName string, item []MenuItemConfig, fg, bg color.RGBA) *Menu {
	w := &Menu{
		menu: moondeck.NewIconWidget("bars"),
		open: moondeck.NewIconWidget("bars-staggered"),
	}

	w.open.OnPress = func(_ moondeck.Button, app *moondeck.App) {
		app.CloseOverlay()
	}

	for _, i := range item {
		//j := moondeck.NewIconWidget(i.Icon)
		j := moondeck.NewIconColorWidget(i.Icon, fg, bg)
		j.OnPress = w.jump(i.App)
		w.item = append(w.item, j)
	}

	return w
}

func (w *Menu) jump(name string) func(moondeck.Button, *moondeck.App) {
	return func(b moondeck.Button, app *moondeck.App) {
		app.Dashboard.Start(name)
	}
}

func (w *Menu) Pressed(b moondeck.Button, app *moondeck.App) {
	log.WithField("visible", w.isVisible).Debug("menu pressed")

	app.CloseOverlay()

	pos := b.Pos()
	app.AddWidget(w.open, moondeck.Overlay, pos.X, pos.Y)

	size := app.Dashboard.Size()
	for x, i := pos.X+1, 0; x < size.W-1 && i < len(w.item); x, i = x+1, i+1 {
		app.AddWidget(w.item[i], moondeck.Overlay, x, pos.Y)
	}
}

func (w *Menu) Released(b moondeck.Button, app *moondeck.App) {}

func (w *Menu) Render(b moondeck.Button) error {
	return w.menu.Render(b)
}

func (w *Menu) Closed(app *moondeck.App) {
	w.isVisible = false
}

var _ moondeck.Widget = (*Menu)(nil)
