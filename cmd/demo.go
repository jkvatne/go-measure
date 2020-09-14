package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"

	"golang.org/x/image/colornames"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"fyne.io/fyne/app"
)

var pict *canvas.Image
var scope *fyne.Container

func doResize() {
	m := glfw.GetMonitors()[0].GetVideoMode()
	width := m.Width
	height := m.Height
	scopeImg := image.NewRGBA(image.Rect(0, 0, width-100, height-200))
	draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	pict = canvas.NewImageFromImage(scopeImg)

}

func main() {
	a := app.NewWithID("io.fyne.demo")
	w := a.NewWindow("Analog Discovery2 scope")
	fmt.Printf("Done\n")

	m := glfw.GetMonitors()[0].GetVideoMode()
	fmt.Printf("W=%d, H=%d\n", m.Width, m.Height)

	//scopeImg := image.NewRGBA(image.Rect(0, 0, 1920, 1024))
	//draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	//newLabel(scopeImg, 20, 40, "Osciloscope", colornames.Yellow)
	//pict = canvas.NewImageFromImage(scopeImg)
	top := widget.NewLabelWithStyle("Oscilloscope", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	btm := widget.NewLabelWithStyle("Bottom", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	scope = fyne.NewContainer()
	drawScope(500, 500)
	w.SetContent(fyne.NewContainerWithLayout(
		layout.NewBorderLayout(
			top,
			btm,
			nil,
			nil),
		top, btm, scope),
	)
	//w.Maximize()
	w.Resize(fyne.Size{500, 500})
	go update(w)
	w.ShowAndRun()
	fmt.Printf("Done\n")
}

func drawScope(x, y int) {
	p1 := fyne.Position{0, 0}
	p2 := fyne.Position{x, y}
	scope = fyne.NewContainer()
	scope.AddObject(&canvas.Line{Position1: p1, Position2: p2, StrokeColor: colornames.Black, StrokeWidth: 1})
	grid := &canvas.Rectangle{StrokeColor: colornames.Black, StrokeWidth: 1}
	grid.Move(fyne.Position{X: 50, Y: 50})
	grid.Resize(fyne.Size{Width: x - 100, Height: y - 100})
	scope.AddObject(grid)
}

func update(w fyne.Window) {
	n := 0
	size := fyne.Size{500, 500}
	for {
		time.Sleep(time.Second)
		/*
			top := widget.NewLabelWithStyle("Oscilloscope", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
			btm := widget.NewLabelWithStyle("Bottom", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			drawScope(size.Width, size.Height)
			w.SetContent(fyne.NewContainerWithLayout(
				layout.NewBorderLayout(
					top,
					btm,
					nil,
					nil),
				top, btm, scope),
			)
		*/
		drawScope(size.Width, size.Height)
		n = n + 10
		//Line(img, image.Pt(40, 140), image.Pt(250, 140+n), colornames.Blue, 2)
		scope.Refresh()
		// Window size: w, h := w.GetSize()
		size = scope.Size()
		fmt.Printf("W=%d, H=%d\n", size.Width, size.Height)
	}

}

func minmax(x, y int) (int, int) {
	if x < y {
		return x, y
	}
	return y, x
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

func newLabel(img *image.RGBA, x, y int, label string, c color.RGBA) {
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: inconsolata.Bold8x16,
		Dot:  point,
	}
	d.DrawString(label)
}
