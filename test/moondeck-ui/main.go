package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math/rand"
	"os"
	"time"

	"maze.io/moondeck/moondeck"
	"maze.io/moondeck/moonraker"
)

func main() {
	bg, err := loadBackground("klipper-logo-5x3.png")
	if err != nil {
		log.Fatalln(err)
	}

	deck, err := moondeck.Open()
	if err != nil {
		log.Fatalln(err)
	}
	if err = deck.Reset(); err != nil {
		log.Fatalln(err)
	}

	dash := moondeck.NewDashboard(deck)
	app := moondeck.NewApp(dash)
	app.SetBackgroundImage(bg)

	/*
		app.AddWidget(moondeck.NewImageWidget(bg), moondeck.Foreground, 1, 1)
		{
			w := &moondeck.TextWidget{Text: "23.5C"}
			w.OnPress = randomOnTextPress
			app.AddWidget(w, moondeck.Foreground, 0, 0)
		}
		{
			w := &moondeck.TextWidget{Text: "104.2C"}
			w.OnPress = randomOnTextPress
			app.AddWidget(w, moondeck.Foreground, 0, 1)
		}
		app.AddWidget(randomFloatWidget(), moondeck.Foreground, 3, 2)
		app.AddWidget(moondeck.NewIconWidget("arrow-up.solid", icon.White, icon.Transparent), moondeck.Foreground, 2, 2)
	*/

	{
		w := moonraker.NewTemperature("test")
		w.Value = 42
		app.AddWidget(w, moondeck.Foreground, 0, 1)
	}
	{
		w := moonraker.NewTemperature("test2")
		w.Value = 23
		app.AddWidget(w, moondeck.Foreground, 1, 1)
	}

	//d.Add(moondeck.NewRange(23, 0, 100), 0, 0)
	//d.Add(moondeck.NewRange(42, 23, 100), 2, 1)
	//moondeck.NewRange(23, 0, 100).Connect(d, 0, 0)
	//moondeck.NewRange(42, 23, 100).Connect(d, 2, 1)
	dash.SetActive(app)
	if err := dash.Run(); err != nil {
		log.Fatalln(err)
	}
}

func loadBackground(name string) (image.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return png.Decode(f)
}

func randomFloatWidget() moondeck.Widget {
	w := &moondeck.FloatWidget{
		Precision: 2,
	}
	go func() {
		for range time.Tick(time.Second) {
			w.Value = rand.Float64() * 999
			w.Dirty()
		}
	}()
	return w
}

func randomOnTextPress(w moondeck.Widget, b moondeck.Button, app *moondeck.App) {
	w.(*moondeck.TextWidget).Text = fmt.Sprintf("%.1fC", rand.Float64()*250)
	w.Dirty()
}

/*
func temperatureGaugeWidget() moondeck.Widget {
	i := image.NewRGBA(image.Rect(0, 0, 96, 96))
	w := moondeck.NewImageWidget(i)
}
*/
