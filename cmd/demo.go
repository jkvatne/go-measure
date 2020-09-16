package main

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/jkvatne/go-measure/instr"

	"github.com/jkvatne/go-measure/alog"
	"github.com/jkvatne/go-measure/tps2000"

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
var data [][]float64

type event int

const (
	updateEvent event = iota
	evRefresh
	doneEevnt
)

var events chan event

func handleEvents(w fyne.Window) {
	time.Sleep(time.Second)
	for {
		select {
		case e := <-events:
			if e == doneEevnt {
				return
			}
			update(e)
		}
	}
}

func update(e event) {
	time.Sleep(time.Millisecond * 500)
	size := fyneImg.Size()
	if e == evRefresh {
		scopeImg = image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
		draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
		fyneImg.Image = scopeImg
	}
	t1 := 0.0
	t2 := 1.0e-3
	voltTop := 10.0
	voltBtm := -10.0
	if data != nil {
		t1 = data[0][0]
		t2 = data[0][len(data[0])-1]
		voltTop = data[len(data)-2][0]
		voltBtm = data[len(data)-1][0]
	}
	Grid(scopeImg, t1, t2, voltTop, voltBtm)
	plot(scopeImg, data)
	fyneImg.Refresh()
}

func refresh() {
	events <- evRefresh
}

// FetchCurve will read points from scope
func FetchCurve(scope instr.Scope) {
	fmt.Printf("Fetch curve\n")
	var err error
	data, err = scope.Curve([]instr.Chan{instr.Ch1, instr.Ch2, instr.Ch3, instr.Ch4}, 2500)
	if err != nil {
		alog.Error("Error fetching curve, %s", err)
	}
	refresh()
}

func getCurve(scope instr.Scope) {
	time.Sleep(500 * time.Millisecond)
	FetchCurve(scope)
}

func main() {
	events = make(chan event)
	scope, err := tps2000.New("")
	if err != nil {
		alog.Error("No scope found")
	}

	sampleInterval, xPos := scope.GetTime()
	alog.Info("sampleInterval=%0.3e, xpos=%0.3e", sampleInterval, xPos)

	a := app.NewWithID("io.fyne.demo")
	w := a.NewWindow("Analog Discovery2 scope")

	m := glfw.GetMonitors()[0].GetVideoMode()
	fmt.Printf("Monitor W=%d, H=%d\n", m.Width, m.Height)

	b := image.Rect(0, 0, 1024, 712)
	scopeImg = image.NewRGBA(b)
	draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	fyneImg = canvas.NewImageFromImage(scopeImg)
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
	go handleEvents(w)
	go getCurve(scope)
	refresh()
	w.ShowAndRun()
}
