package main

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/colornames"
)

var fyneImg *canvas.Image
var scopeImg *image.RGBA

func doResize() {
	//m := glfw.GetMonitors()[0].GetVideoMode()
	//scopeImg = image.NewRGBA(image.Rect(0, 0, 500, height-200))
	//draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	//fyneImg = canvas.NewImageFromImage(scopeImg)
}

/*
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
*/

func update(w fyne.Window) {
	n := 0
	size := fyne.Size{500, 500}
	for {
		time.Sleep(time.Second)
		//drawScope(size.Width, size.Height)

		n = n + 10
		size = fyneImg.Size()
		scopeImg = image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
		draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
		Rect(scopeImg, image.Pt(10, 10), image.Pt(size.Width-10, size.Height-10), colornames.Yellow, 1)
		Line(scopeImg, image.Pt(40, 140), image.Pt(250, 140+n), colornames.Blue, 2)
		fyneImg.Image = scopeImg
		fyneImg.Refresh()
		fmt.Printf("W=%d, H=%d\n", size.Width, size.Height)
	}

}

func minmax(x, y int) (int, int) {
	if x < y {
		return x, y
	}
	return y, x
}

func main() {
	a := app.NewWithID("io.fyne.demo")
	w := a.NewWindow("Analog Discovery2 scope")
	fmt.Printf("Done\n")

	m := glfw.GetMonitors()[0].GetVideoMode()
	fmt.Printf("W=%d, H=%d\n", m.Width, m.Height)

	b := image.Rect(0, 0, 1024, 712)
	scopeImg = image.NewRGBA(b)
	draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	fyneImg = canvas.NewImageFromImage(scopeImg)
	NewLabel(scopeImg, 50, 50, "Tekst", colornames.White)
	Label(scopeImg, 50, 100, "Regular", colornames.White, Regular10)
	Label(scopeImg, 50, 120, "Regular", colornames.White, Regular12)
	top := widget.NewLabelWithStyle("Oscilloscope", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	btm := widget.NewLabelWithStyle("Bottom", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	w.SetContent(fyne.NewContainerWithLayout(
		layout.NewBorderLayout(
			top,
			btm,
			nil,
			nil),
		top, btm, fyneImg),
	)
	//w.Maximize()
	w.Resize(fyne.Size{1000, 750})
	w.CenterOnScreen()
	go update(w)
	w.ShowAndRun()
	fmt.Printf("Done\n")
}
