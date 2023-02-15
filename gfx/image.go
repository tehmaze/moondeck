package gfx

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"golang.org/x/image/draw"
	"maze.io/moondeck/util"

	_ "image/gif"  // GIF codec
	_ "image/jpeg" // JPEG codec
	_ "image/png"  // PNG codec

	_ "golang.org/x/image/bmp" // BMP codec
)

var imageExts = []string{
	".png",
	".jpg",
	".jpeg",
	".gif",
	".bmp",
}

// LoadImage does a best guess on image names from the specified locations. The
// locations can be file paths, or instances implementing the util.Opener interface.
func LoadImage(name string, locations ...interface{}) (image.Image, error) {
	logger := log.WithField("name", name)
	for _, ext := range append([]string{""}, imageExts...) {
		for _, l := range locations {
			var (
				rc  io.ReadCloser
				i   image.Image
				err error
			)
			if o, ok := l.(util.Opener); ok {
				rc, err = o.Open(name + ext)
			} else if path, ok := l.(string); ok {
				rc, err = os.Open(filepath.Join(path, name+ext))
			} else {
				continue
			}

			if err != nil {
				logger.
					WithField("location", fmt.Sprintf("%T", l)).
					WithError(err).
					Trace("error loading")
			}

			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				return nil, err
			}

			defer func() { _ = rc.Close() }()
			if i, _, err = image.Decode(rc); err != nil {
				logger.WithError(err).Warn("error decoding")
			}
			return i, err
		}
	}
	return nil, &os.PathError{
		Path: name,
		Err:  os.ErrNotExist,
	}
}

func MustLoadImage(name string, locations ...interface{}) image.Image {
	i, err := LoadImage(name, locations...)
	if err != nil {
		panic(err)
	}
	return i
}

// Colorize does threshold-based colorization of an image.
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

func sinc(x float64) float64 {
	if x == 0 {
		return 1
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

// Lanczos interpolation kernel.
var Lanczos = &draw.Kernel{
	Support: 3.0,
	At: func(x float64) float64 {
		x = math.Abs(x)
		if x < 3.0 {
			return sinc(x) * sinc(x/3.0)
		}
		return 0
	},
}

// Resize an image using Lanczos filter for scaling up, and a Nearest Neighbor filter for scaling down.
func Resize(i *image.RGBA, size util.Size) *image.RGBA {
	var (
		o = image.NewRGBA(image.Rectangle{Max: size.ImagePoint()})
		k draw.Interpolator
	)
	if Area(i) > Area(o) {
		k = draw.NearestNeighbor
	} else {
		k = Lanczos
	}
	k.Scale(o, o.Bounds(), i, i.Bounds(), draw.Src, nil)
	return o
}

// ResizeTo resizes src to fit in dst.
func ResizeTo(dst, src *image.RGBA) {
	if src == nil || dst == nil {
		return
	}

	var k draw.Interpolator
	if Area(src) > Area(dst) {
		k = draw.NearestNeighbor
	} else {
		k = Lanczos
	}
	k.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
}

// Area of an image.
func Area(i image.Image) int {
	r := i.Bounds()
	return r.Dx() * r.Dy()
}

// ToRGBA converts the input image to RGBA format, if required.
func ToRGBA(i image.Image) *image.RGBA {
	if i, ok := i.(*image.RGBA); ok {
		return i
	}

	o := image.NewRGBA(i.Bounds())
	draw.Copy(o, image.Point{}, i, i.Bounds(), draw.Src, nil)
	return o
}

// ToRGBAColor converts the input color to RGBA format, if required.
func ToRGBAColor(c color.Color) color.RGBA {
	if c, ok := c.(color.RGBA); ok {
		return c
	}

	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}
