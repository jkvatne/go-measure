package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/image/colornames"

	"github.com/goki/freetype/truetype"

	"golang.org/x/image/font/gofont/goregular"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

func Ticks(img draw.Image, p1, p3 image.Point) {
	p2 := image.Point{X: p1.X, Y: p3.Y}
	p4 := image.Point{X: p3.X, Y: p1.Y}
	Rect(scopeImg, p1, p3, colornames.Gray, 1)
	vTicks(img, p1, p2, 10, 16)
	vTicks(img, p1, p2, 20, 8)
	vTicks(img, p1, p2, 100, 4)
	vTicks(img, p4, p3, 10, -16)
	vTicks(img, p4, p3, 20, -8)
	vTicks(img, p4, p3, 100, -4)
	hTicks(img, p2, p3, 10, -16)
	hTicks(img, p2, p3, 20, -8)
	hTicks(img, p2, p3, 100, -4)
	hTicks(img, p1, p4, 10, 16)
	hTicks(img, p1, p4, 20, 8)
	hTicks(img, p1, p4, 100, 4)
	h := p3.Y - p1.Y
	for i := 0; i < 10; i++ {
		hDot(img, p1.X, p3.X, p1.Y+i*h/10, colornames.Gray)
	}
	w := p3.X - p1.X
	for i := 0; i < 10; i++ {
		vDot(img, p1.X+i*w/10, p1.Y, p3.Y, colornames.Gray)
	}
}

func hDot(img draw.Image, x1, x2, y int, c color.Color) {
	for x := x1; x < x2; x += 4 {
		img.Set(x, y, c)
	}
}

func vDot(img draw.Image, x, y1, y2 int, c color.Color) {
	for y := y1; y < y2; y += 4 {
		img.Set(x, y, c)
	}
}

func vTicks(img draw.Image, p1, p2 image.Point, n, dx int) {
	h := p2.Y - p1.Y
	for i := 0; i < n; i++ {
		y := p1.Y + i*h/n
		Line(img, image.Point{X: p1.X, Y: y}, image.Point{X: p1.X + dx, Y: y}, colornames.Gray, 1)
	}
}

func hNum(img draw.Image, p1, p2 image.Point, t1, t2 float64) {
	tmax := math.Max(t1, t2)
	unit := "s"
	if tmax < 1e-6 {
		unit = "nS"
		t1 = t1 * 1e9
		t2 = t2 * 1e9
	} else if tmax < 1e-3 {
		unit = "uS"
		t1 = t1 * 1e6
		t2 = t2 * 1e6
	} else if tmax < 1.0 {
		unit = "mS"
		t1 = t1 * 1e3
		t2 = t2 * 1e3
	} else {
		unit = "S"
		t1 = t1
		t2 = t2
	}
	dp := 0
	v := math.Max(t1, t2)
	if v >= 100.0 {
		dp = 0
	} else if v >= 10.0 {
		dp = 1
	} else if v >= 1.0 {
		dp = 2
	} else {
		dp = 3
	}
	w := p2.X - p1.X
	for i := 0; i <= 10; i++ {
		x := p1.X + i*w/10
		val := t1 + float64(i)*(t2-t1)/10
		s := fmt.Sprintf("%0.*f", dp, val) + unit
		Label(img, x, p1.Y+10, s, colornames.White, Regular10)
	}
}

func vNum(img draw.Image, p1, p2 image.Point, topVal, btmVal float64) {
	h := p2.Y - p1.Y
	dp := 1
	v := math.Max(math.Abs(topVal), math.Abs(btmVal))
	if v >= 100.0 {
		dp = 0
	} else if v >= 10.0 {
		dp = 1
	} else if v >= 1.0 {
		dp = 2
	} else {
		dp = 3
	}
	for i := 0; i <= 10; i++ {
		y := p1.Y + i*h/10
		val := topVal - float64(i)*(topVal-btmVal)/10
		s := fmt.Sprintf("%0.*f", dp, val)
		dx := Measure(s, Regular10) + 1
		Label(img, p1.X-dx, y+h10/2, s, colornames.White, Regular10)
	}
}

func hTicks(img draw.Image, p1, p2 image.Point, n, dy int) {
	w := p2.X - p1.X
	for i := 0; i < n; i++ {
		x := p1.X + i*w/n
		Line(img, image.Point{X: x, Y: p1.Y}, image.Point{X: x, Y: p2.Y + dy}, colornames.Gray, 1)
	}
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

func Label(img draw.Image, x, y int, label string, c color.RGBA, face font.Face) {
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
