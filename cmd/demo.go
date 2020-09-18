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
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/colornames"
)

var scopeImg *image.RGBA
var data [][]float64

type event int

type aScopeFrame struct {
	window fyne.Window
	canvas fyne.CanvasObject
}

const (
	updateEvent event = iota
	evRefresh
	doneEevnt
)

var events chan event

func handleEvents(w fyne.Window, f *aScopeFrame) {
	time.Sleep(time.Second)
	for {
		select {
		case e := <-events:
			if e == doneEevnt {
				return
			}
			update(e, f)
		}
	}
}

func update(e event, f *aScopeFrame) {
	time.Sleep(time.Millisecond * 500)
	size := f.canvas.Size()
	//size := f.canvas.Size()
	if e == evRefresh {
		scopeImg = image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
		draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Blue), image.Pt(0, 0), draw.Src)
	}
	t1 := 0.0
	t2 := 1.0e-3
	if data != nil {
		t1 = data[0][0]
		t2 = data[0][len(data[0])-1]
	}
	Grid(scopeImg, t1, t2, data[len(data)-2], data[len(data)-1])
	plot(scopeImg, data)
	f.canvas = canvas.NewRasterFromImage(scopeImg)
	//f.refresh()
}

// FetchCurve will read points from scope
func FetchCurve(scope instr.Scope) {
	fmt.Printf("Fetch curve\n")
	var err error
	data, err = scope.Curve([]instr.Chan{instr.Ch1, instr.Ch2}, 2500)
	if err != nil {
		alog.Error("Error fetching curve, %s", err)
	}
}

func getCurve(scope instr.Scope) {
	time.Sleep(500 * time.Millisecond)
	FetchCurve(scope)
	events <- evRefresh
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

func (f *aScopeFrame) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	f.canvas.Resize(size)
	f.redrawScope()
	f.refresh()
}

func (f *aScopeFrame) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(1000, 700)
}

func (f *aScopeFrame) refresh() {
	f.window.Canvas().Refresh(f.canvas)
}

var n int

func (f *aScopeFrame) redrawScope() {
	n++
	scopeImg = image.NewRGBA(image.Rect(0, 0, f.canvas.Size().Width, f.canvas.Size().Height))
	t1 := 0.0
	t2 := 1.0e-3
	if data != nil {
		t1 = data[0][0]
		t2 = data[0][len(data[0])-1]
	}
	Grid(scopeImg, t1, t2, data[len(data)-2], data[len(data)-1])
	plot(scopeImg, data)
	Label(scopeImg, 100, 100, fmt.Sprintf("n=%d", n), colornames.Orange, Regular10)
	f.canvas = canvas.NewRasterFromImage(scopeImg)
}

func main() {
	a := app.NewWithID("io.fyne.demo")
	window := a.NewWindow("Scope")
	window.SetPadded(false)

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

	m := glfw.GetMonitors()[0].GetVideoMode()
	fmt.Printf("Monitor W=%d, H=%d\n", m.Width, m.Height)

	b := image.Rect(0, 0, 1000, 700)
	scopeImg = image.NewRGBA(b)
	draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Red), image.Pt(0, 0), draw.Src)
	//fyneImg = canvas.NewImageFromImage(scopeImg)

	time.Sleep(500 * time.Millisecond)
	FetchCurve(scope)
	t1 := 0.0
	t2 := 1.0e-3
	if data != nil {
		t1 = data[0][0]
		t2 = data[0][len(data[0])-1]
	}
	Grid(scopeImg, t1, t2, data[len(data)-2], data[len(data)-1])
	plot(scopeImg, data)
	Label(scopeImg, 100, 100, fmt.Sprintf("Initial"), colornames.Orange, Regular12)
	scopeFrame := &aScopeFrame{window: window}
	scopeFrame.canvas = canvas.NewRasterFromImage(scopeImg)
	// Top header label
	//top := widget.NewLabelWithStyle("Oscilloscope", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	// Setup form contents
	//window.SetContent(fyne.NewContainerWithLayout(layout.NewBorderLayout(	top,nil,nil,nil),top, fyne.NewContainerWithLayout(scopeFrame, scopeFrame.canvas)))
	window.SetContent(fyne.NewContainerWithLayout(scopeFrame, scopeFrame.canvas))

	//w.Maximize()
	window.Resize(fyne.Size{1000, 750})
	window.CenterOnScreen()
	//go handleEvents(window, scopeFrame)
	//go getCurve(scope)
	window.ShowAndRun()
}
