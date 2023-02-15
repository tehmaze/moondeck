package font

import (
	"embed"
	"io"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

//go:embed *.ttf
var content embed.FS

var (
	Roboto     *truetype.Font
	RobotoBold *truetype.Font
)

func loadFont(name string) (*truetype.Font, error) {
	r, err := content.Open(name)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return freetype.ParseFont(b)
}

func mustLoadFont(name string) *truetype.Font {
	f, err := loadFont(name)
	if err != nil {
		panic(err)
	}
	return f
}

func init() {
	Roboto = mustLoadFont("Roboto-Regular.ttf")
	RobotoBold = mustLoadFont("Roboto-Bold.ttf")
}
