package util

import (
	"fmt"
	"image"
)

type Point struct {
	X int `cty:"x" hcl:"x,optional"`
	Y int `cty:"y" hcl:"y,optional"`
}

func Pt(x, y int) Point { return Point{x, y} }

func (p Point) Add(o Point) Point       { return Point{X: p.X + o.X, Y: p.Y + o.Y} }
func (p Point) Sub(o Point) Point       { return Point{X: p.X - o.X, Y: p.Y - o.Y} }
func (p Point) Div(o Point) Point       { return Point{X: p.X / o.X, Y: p.Y / o.Y} }
func (p Point) Mul(o Point) Point       { return Point{X: p.X * o.X, Y: p.Y * o.Y} }
func (p Point) Size() Size              { return Size{W: p.X, H: p.Y} }
func (p Point) ImagePoint() image.Point { return image.Point{p.X, p.Y} }
func (p Point) String() string          { return fmt.Sprintf("(%d,%d)", p.X, p.Y) }

type Size struct {
	W int `cty:"w" hcl:"w,optional"`
	H int `cty:"h" hcl:"h,optional"`
}

func Sz(w, h int) Size { return Size{w, h} }

func (s Size) Area() int {
	return s.W * s.H
}

func (s Size) Add(o Size) Size         { return Size{W: s.W + o.W, H: s.H + o.H} }
func (s Size) Sub(o Size) Size         { return Size{W: s.W - o.W, H: s.H - o.H} }
func (s Size) Mul(o Size) Size         { return Size{W: s.W * o.W, H: s.H * o.H} }
func (s Size) Div(o Size) Size         { return Size{W: s.W / o.W, H: s.H / o.H} }
func (s Size) Point() Point            { return Point{s.W, s.H} }
func (s Size) ImagePoint() image.Point { return image.Point{s.W, s.H} }
func (s Size) String() string          { return fmt.Sprintf("%dx%d", s.W, s.H) }
