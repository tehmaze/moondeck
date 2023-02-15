package icon

import (
	"embed"
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"golang.org/x/image/draw"

	_ "image/gif"  // GIF codec
	_ "image/jpeg" // JPEG codec
	_ "image/png"  // PNG codec

	_ "golang.org/x/image/bmp" // BMP codec
)

// Path is the default locations to look for icon files
// UNIX:    ~/.local/share/moondeck
// macOS:   ~/Library/Application Support/moondeck
// Windows: %LOCALAPPDATA%\moondeck
var Path = appendToPaths(xdg.DataDirs, "moondeck")

func appendToPaths(paths []string, suffix string) []string {
	for i, path := range paths {
		paths[i] = filepath.Join(path, suffix)
	}
	return paths
}

//go:embed data/*.png
var content embed.FS

// Common colors
var (
	Transparent = color.RGBA{}
	Black       = color.RGBA{A: 0xff}
	Blue        = color.RGBA{B: 0xff, A: 0xff}
	Green       = color.RGBA{G: 0xff, A: 0xff}
	Cyan        = color.RGBA{G: 0xff, B: 0xff, A: 0xff}
	Red         = color.RGBA{R: 0xff, A: 0xff}
	Magenta     = color.RGBA{R: 0xff, B: 0xff, A: 0xff}
	Yellow      = color.RGBA{R: 0xff, G: 0xff, A: 0xff}
	White       = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
)

func Brand(name string) (*image.RGBA, error) {
	return Icon(name + ".brand")
}

func Icon(name string) (*image.RGBA, error) {
	var (
		f   io.ReadCloser
		err error
	)
	for _, path := range Path {
		//log.Println("try", path)
		if f, err = os.Open(filepath.Join(path, name+".png")); err == nil {
			break
		} else if os.IsNotExist(err) {
			f = nil
			continue
		}
		return nil, err
	}
	if f == nil {
		if f, err = content.Open("data/" + name + ".png"); err != nil {
			//log.Println("tried internal", name, err)
			return nil, err
		}
	}

	//log.Printf("decoding image from %#+v", f)
	i, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return toRGBA(i), f.Close()
}

func Must(name string) *image.RGBA {
	i, err := Icon(name)
	if err != nil {
		panic(err)
	}
	return i
}

func Colorize(i *image.RGBA, fg, bg color.RGBA) *image.RGBA {
	var (
		r = i.Bounds()
		o = image.NewRGBA(r)
	)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if i.RGBAAt(x, y).R >= 0x80 {
				o.SetRGBA(x, y, fg)
			} else {
				o.SetRGBA(x, y, bg)
			}
		}
	}

	return o
}

func toRGBA(i image.Image) *image.RGBA {
	switch i := i.(type) {
	case *image.RGBA:
		return i
	}
	o := image.NewRGBA(i.Bounds())
	draw.Copy(o, image.Point{}, i, i.Bounds(), draw.Src, nil)
	return o
}
