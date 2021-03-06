// Copyright 2020 Jan Kåre Vatne. All rights reserved.

package plot

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/image/colornames"
)

var chanColor = []color.Color{
	colornames.Yellow,
	colornames.Cyan,
	colornames.Magenta,
	colornames.Green,
	colornames.Red,
	colornames.Gray,
	colornames.White,
	colornames.Beige,
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
	for i := 1; i <= 10; i++ {
		x := p1.X + i*w/10
		val := t1 + float64(i)*(t2-t1)/10
		s := fmt.Sprintf("%0.*f", dp, val) + unit
		y := p1.Y + int(Regular10.Metrics().Ascent/54) // Divide by 58 instead of 64 to get some padding
		Label(img, x-8, y, s, colornames.White, Regular10)
	}
}

func vNum(img draw.Image, p1, p2 image.Point, topVal, btmVal float64, col color.Color) {
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
		Label(img, p1.X-dx, y+h10/2, s, col, Regular10)
	}
}

func hTicks(img draw.Image, p1, p2 image.Point, n, dy int) {
	w := p2.X - p1.X
	for i := 0; i < n; i++ {
		x := p1.X + i*w/n
		Line(img, image.Point{X: x, Y: p1.Y}, image.Point{X: x, Y: p2.Y + dy}, colornames.Gray, 1)
	}
}

func plot(img draw.Image, data [][]float64) {
	topMargin := h10 + 4
	leftMargin := 45
	rightMargin := 20
	// Fill black background
	draw.Draw(img, img.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	// Exit if no data - leave black screen
	if data == nil {
		Label(img, 150, h10+2, "No data", colornames.White, Regular10)
		return
	}

	tl := img.Bounds().Min.Add(image.Pt(leftMargin, topMargin))
	br := img.Bounds().Max.Add(image.Pt(-rightMargin, -topMargin))
	bl := image.Pt(tl.X, br.Y)
	tr := image.Pt(br.X, tl.Y)

	// Voltage labels
	for i := 0; i < len(data[len(data)-2]); i++ {
		t := tl.Add(image.Pt(0, i*h10))
		b := bl.Add(image.Pt(0, i*h10))
		vNum(img, t, b, data[len(data)-2][i], data[len(data)-1][i], chanColor[i])
	}
	// Time labels
	hNum(img, bl, br, data[0][0], data[0][len(data[0])-1])
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
	// Frame around grid
	Rect(img, tl, br, colornames.Gray, 1)
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
	chanNum := len(data) - 1
	// Check that y range is present
	scalingPresent := len(data[0]) >= 4 && len(data[len(data)-1]) <= 16 && len(data[len(data)-2]) <= 16
	if scalingPresent {
		chanNum = chanNum - 2
	}
	for ch := 0; ch < len(data[len(data)-2]); ch++ {
		col := chanColor[ch]
		voltTop := data[len(data)-2][ch]
		voltBtm := data[len(data)-1][ch]
		y0 := bl.Y + int(float64(tl.Y-bl.Y)*(data[ch+1][0]-voltBtm)/(voltTop-voltBtm))
		x0 := tl.X + int(float64(tr.X-tl.X)*data[0][0]/data[0][len(data[0])-1])
		for i := 0; i < len(data[0]); i++ {
			y1 := bl.Y + int(float64(tl.Y-bl.Y)*(data[ch+1][i]-voltBtm)/(voltTop-voltBtm))
			x1 := tl.X + int(float64(tr.X-tl.X)*data[0][i]/data[0][len(data[0])-1])
			p1 := image.Point{x0, y0}
			p2 := image.Point{x1, y1}
			Line(img, p1, p2, col, 1)
			x0 = x1
			y0 = y1
		}
	}
}
