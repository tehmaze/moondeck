package sparkline

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"maze.io/moondeck/gfx/blend"
	"maze.io/moondeck/gfx/icon"
)

type Sparkline struct {
	size      int
	min       *float64
	values    []float64
	threshold []blend.Threshold
	bg        *image.Uniform
}

func New(size int, threshold []blend.Threshold, background color.RGBA) *Sparkline {
	return &Sparkline{
		size:      size,
		threshold: threshold,
		bg:        image.NewUniform(background),
	}
}

func (s *Sparkline) FixedMin(value float64) {
	s.min = &value
}

func (s *Sparkline) Push(value float64) {
	if s.min != nil && value < *s.min {
		value = *s.min
	}

	if len(s.values) < s.size {
		s.values = append(s.values, value)
	} else {
		copy(s.values, s.values[1:])
		s.values[len(s.values)-1] = value
	}
}

func (s *Sparkline) Draw(i *image.RGBA) {
	if len(s.values) == 0 {
		return
	}
	var (
		min = math.Inf(+1)
		max = math.Inf(-1)
	)
	if s.min != nil {
		min = *s.min
	}
	for _, v := range s.values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max < 1 {
		max = 1
	}
	if min >= max-1 {
		min = max - 1
	}

	max = math.Round(max/25)*25 + 25

	var (
		r  = i.Bounds()
		ys = float64(r.Dy()) / (max - min)
	)
	//log.Printf("sparkline: scale on y %.1f (%.1f - %.1f)", ys, min, max)
	draw.Draw(i, r, s.bg, image.Point{}, draw.Src)
	for x, l := 0, len(s.values); x < l && x < r.Dx(); x++ {
		c := blend.PickColor(s.threshold, uint64(s.values[x]))
		for y := int(ys * (max - s.values[x])); y < r.Dy(); y++ {
			i.Set(x, y, c)
		}
	}
	for x := 0; x < 5; x++ {
		for scale := 0.0; scale < max; scale += 25 {
			i.SetRGBA(x, int(ys*scale), icon.White)
		}
	}
}
