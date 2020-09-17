package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/goki/freetype/truetype"

	"golang.org/x/image/font/gofont/goregular"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

func minmax(x, y int) (int, int) {
	if x < y {
		return x, y
	}
	return y, x
}

func Rect(img draw.Image, p0, p2 image.Point, c color.Color, w int) {
	p1 := image.Point{X: p0.X, Y: p2.Y}
	p3 := image.Point{X: p2.X, Y: p0.Y}
	Line(img, p0, p1, c, w)
	Line(img, p1, p2, c, w)
	Line(img, p2, p3, c, w)
	Line(img, p3, p0, c, w)
}

// Line draws a line
func Line(img draw.Image, a, b image.Point, c color.Color, width int) {
	minx, maxx := minmax(a.X, b.X)
	miny, maxy := minmax(a.Y, b.Y)

	dx := float64(b.X - a.X)
	dy := float64(b.Y - a.Y)

	if maxx-minx > maxy-miny {
		d := 1
		if a.X > b.X {
			d = -1
		}
		for x := 0; x != b.X-a.X+d; x += d {
			y := int(float64(x) * dy / dx)
			for i := 0; i < width; i++ {
				img.Set(a.X+x, a.Y+y+i, c)
			}
		}
	} else {
		d := 1
		if a.Y > b.Y {
			d = -1
		}
		for y := 0; y != b.Y-a.Y+d; y += d {
			x := int(float64(y) * dx / dy)
			for i := 0; i < width; i++ {
				img.Set(a.X+x+i, a.Y+y, c)
			}
		}
	}
}

func pt(x, y int) fixed.Point26_6 {
	return fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
}

func Measure(label string, face font.Face) int {
	d := &font.Drawer{
		Dst:  nil,
		Src:  nil,
		Face: face,
		Dot:  pt(0, 0),
	}
	return (int(d.MeasureString(label)) + 63) / 64
}

func Label(img draw.Image, x, y int, label string, c color.Color, face font.Face) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  pt(x, y),
	}
	d.MeasureString(label)
	d.DrawString(label)
}

func NewLabel(img *image.RGBA, x, y int, label string, c color.RGBA) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: inconsolata.Regular8x16,
		Dot:  pt(x, y),
	}
	d.DrawString(label)
}

var Regular12 font.Face
var Regular10 font.Face
var h10 int

func init() {
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	Regular10 = truetype.NewFace(f, &truetype.Options{Size: 10, DPI: 72})
	h10 = int(Regular10.Metrics().Ascent / 64)
	Regular12 = truetype.NewFace(f, &truetype.Options{Size: 12, DPI: 72})
}
