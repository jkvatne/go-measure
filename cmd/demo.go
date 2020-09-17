package main

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/jkvatne/go-measure/ad2"

	"github.com/jkvatne/go-measure/instr"

	"github.com/jkvatne/go-measure/alog"

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
	//voltTop := 10.0
	//voltBtm := -10.0
	if data != nil {
		t1 = data[0][0]
		t2 = data[0][len(data[0])-1]
		//voltTop = data[len(data)-2][0]
		//voltBtm = data[len(data)-1][0]
	}
	Grid(scopeImg, t1, t2, data[len(data)-2], data[len(data)-1])
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
	data, err = scope.Curve([]instr.Chan{instr.Ch1, instr.Ch2}, 2500)
	if err != nil {
		alog.Error("Error fetching curve, %s", err)
	}
	refresh()
}

func getCurve(scope instr.Scope) {
	time.Sleep(500 * time.Millisecond)
	FetchCurve(scope)
}

// SetupAd2 will initialize Digilent Analog Discovery 2
func SetupAd2(a *ad2.Ad2, freq float64) error {
	err := a.SetOutput(0, 3.0, 0.0)
	err = a.SetOutput(1, -2.0, 0.0)
	err = a.SetAnalogOut(instr.Ch1, freq, 0.0, ad2.WfSine, 2.0, 0)
	err = a.SetAnalogOut(instr.Ch2, freq, 90.0, ad2.WfSine, 2.0, 0)
	a.StartAnalogOut(instr.TRIG)
	err = a.SetupChannel(instr.Ch1, 10.0, 1.0, instr.DC)
	err = a.SetupChannel(instr.Ch2, 10.0, -1.0, instr.DC)
	err = a.SetupTrigger(instr.Ch1, instr.DC, instr.Rising, 0.0, false, 0.0)
	sampleInterval := 1.0 / 2500.0 / freq
	err = a.SetupTime(sampleInterval, 0.0, instr.Sample)

	return err
}

func main() {
	events = make(chan event)
	//scope, err := tps2000.New("")
	scope, err := ad2.New("")
	if err != nil {
		alog.Error("No scope found")
	}
	err = SetupAd2(scope, 1000)
	if err != nil {
		alog.Error("Error setting up AD2")
	}
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
