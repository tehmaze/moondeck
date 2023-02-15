package moondeck

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const DashboardFPS = 25.0

type Dashboard struct {
	Deck
	Apps     []*App
	active   *App
	previous *App
}

func NewDashboard(deck Deck) *Dashboard {
	return &Dashboard{
		Deck: deck,
	}
}

func (dash *Dashboard) Add(app *App) {
	dash.Apps = append(dash.Apps, app)
}

func (dash *Dashboard) Start(name string) bool {
	if dash.active != nil && dash.active.Name == name {
		// Already started
		return true
	}

	for _, app := range dash.Apps {
		if app.Name == name {
			// Close current App.
			if dash.active != nil {
				dash.active.Close()
			}

			dash.Reset()
			dash.previous = dash.active
			dash.active = app
			dash.active.Start()
			dash.active.render(true)
			return true
		}
	}

	return false
}

// Back starts the previous app.
func (dash *Dashboard) Back() bool {
	if dash.previous != nil {
		return dash.Start(dash.previous.Name)
	}
	return false
}

func (dash *Dashboard) Run() error {
	var (
		events = dash.ButtonEvents()
		render = time.NewTicker(time.Second / DashboardFPS)
	)
	defer render.Stop()
	for {
		select {
		case event := <-events:
			dash.handle(event.Button, event.Pressed)
		case <-render.C:
			if dash.active != nil {
				// log.Printf("moondeck: dashboard render app %q", dash.active.Name)
				if err := dash.active.Render(); err != nil {
					log.WithError(err).Error("render failed!")
				}
			}
		}
	}
}

func (dash *Dashboard) handle(b Button, pressed bool) {
	if dash.active != nil {
		if pressed {
			dash.active.ButtonPressed(b)
		} else {
			dash.active.ButtonReleased(b)
		}
	}
}
