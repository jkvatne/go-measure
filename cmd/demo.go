package main

import (
	"flag"
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
	"github.com/jkvatne/go-measure/ad2"
	"github.com/jkvatne/go-measure/alog"
	"github.com/jkvatne/go-measure/instr"
	"github.com/jkvatne/go-measure/tps2000"
	"golang.org/x/image/colornames"
)

var data [][]float64
var scopeFrame *aScopeFrame
var scopeImg *image.RGBA

type event int

type aScopeFrame struct {
	window fyne.Window
	canvas fyne.CanvasObject
}

const (
	updateEvent event = iota
	evRedraw
	evDone
)

var events chan event

func handleEvents(f *aScopeFrame) {
	time.Sleep(time.Second)
	for {
		select {
		case e := <-events:
			if e == evDone {
				return
			}
			if e == evRedraw && len(events) == 0 {
				f.redrawScope()
				f.window.Canvas().Refresh(f.canvas)
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
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
	if size.Width != f.canvas.Size().Width || size.Height != f.canvas.Size().Height {
		f.canvas.Resize(size)
		events <- evRedraw
	}
}

func (f *aScopeFrame) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(640, 480)
}

var n int

func (f *aScopeFrame) redrawScope() {
	n++
	scopeImg = image.NewRGBA(image.Rect(0, 0, f.canvas.Size().Width, f.canvas.Size().Height))
	plot(scopeImg, data)
	Label(scopeImg, 100, 100, fmt.Sprintf("n=%d", n), colornames.Orange, Regular10)

}

var useName string

func main() {
	flag.StringVar(&useName, "use", "ad2", "Use ad2 or tps2000 as digitizer")

	a := app.NewWithID("io.fyne.demo")
	window := a.NewWindow("Scope")
	window.SetPadded(false)

	events = make(chan event, 10)

	var scope instr.Scope
	var digilentAd2 *ad2.Ad2
	var err error
	if useName == "tps2000" {
		scope, err = tps2000.New("")
	} else {
		digilentAd2, err = ad2.New("")
		err = SetupAd2(digilentAd2, 1000)
		scope = instr.Scope(digilentAd2)
	}
	if err != nil {
		alog.Error("No %s scope found", useName)
	}
	if err != nil {
		alog.Error("Error setting up AD2")
	}

	m := glfw.GetMonitors()[0].GetVideoMode()
	fmt.Printf("Monitor W=%d, H=%d\n", m.Width, m.Height)

	scopeImg = image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(scopeImg, scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	time.Sleep(300 * time.Millisecond)
	FetchCurve(scope)
	scopeFrame = &aScopeFrame{window: window}
	scopeFrame.canvas = canvas.NewRaster(func(w, h int) image.Image { return scopeImg })
	// Top header label
	top := widget.NewLabelWithStyle("Oscilloscope", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	// Setup form contents
	window.SetContent(fyne.NewContainerWithLayout(layout.NewBorderLayout(top, nil, nil, nil), top, fyne.NewContainerWithLayout(scopeFrame, scopeFrame.canvas)))
	// Center On screen only needed if not maximized
	window.CenterOnScreen()
	// Maximize to fill all of screen
	window.Maximize()
	go handleEvents(scopeFrame)
	window.ShowAndRun()
}
