package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/image/colornames"
)

// Grid will draw a oscilloscope grid with labels.
func Grid(img draw.Image, t1, t2 float64, voltTop, voltBtm float64) {
	// Fill black background
	draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	// The four corners of the grid
	tl := img.Bounds().Min.Add(image.Pt(30, 10))
	br := img.Bounds().Max.Add(image.Pt(-20, -20))
	bl := image.Pt(tl.X, br.Y)
	tr := image.Pt(br.X, tl.Y)
	// Voltage labels
	vNum(scopeImg, tl, bl, t1, -1)
	// Time labels
	hNum(scopeImg, bl, br, t1, t2)
	// Frame around grid
	Rect(scopeImg, tl, br, colornames.Gray, 1)
	// Vertical ticks
	vTicks(img, tl, bl, 10, 16)
	vTicks(img, tl, bl, 20, 8)
	vTicks(img, tl, bl, 100, 4)
	vTicks(img, tr, br, 10, -16)
	vTicks(img, tr, br, 20, -8)
	vTicks(img, tr, br, 100, -4)
	// Horizontal ticks
	hTicks(img, bl, br, 10, -16)
	hTicks(img, bl, br, 20, -8)
	hTicks(img, bl, br, 100, -4)
	hTicks(img, tl, tr, 10, 16)
	hTicks(img, tl, tr, 20, 8)
	hTicks(img, tl, tr, 100, 4)
	// grid lines horizontal
	h := br.Y - tl.Y
	for i := 0; i < 10; i++ {
		hDot(img, tl.X, br.X, tl.Y+i*h/10, colornames.Gray)
	}
	// grid lines verticaln
	w := br.X - tl.X
	for i := 0; i < 10; i++ {
		vDot(img, tl.X+i*w/10, tl.Y, br.Y, colornames.Gray)
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