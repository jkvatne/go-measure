package main

import (
	"flag"
	"time"

	"github.com/jkvatne/go-measure/plot"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/jkvatne/go-measure/ad2"
	"github.com/jkvatne/go-measure/alog"
	"github.com/jkvatne/go-measure/instr"
	"github.com/jkvatne/go-measure/tps2000"
)

type event int

const (
	evNone event = iota
	evFetchData
	evRedraw
	evDone
)

var events chan event

func handleEvents(f *plot.Frame, window fyne.Window) {
	time.Sleep(time.Second)
	for {
		select {
		case e := <-events:
			if e == evFetchData {

			}
			if e == evDone {
				return
			}
			if e == evRedraw || e == evFetchData {
				f.Redraw(window)
			}
		}
	}
}

// FetchCurve will read points from scope
func FetchCurve(scope instr.Scope) [][]float64 {
	data, err := scope.Curve([]instr.Chan{instr.Ch1, instr.Ch2}, 2500)
	if err != nil {
		alog.Error("Error fetching curve, %s", err)
	}
	return data
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
		if err != nil {
			alog.Error("No tps2000 scope found")
		}
	} else {
		digilentAd2, err = ad2.New("")
		err = SetupAd2(digilentAd2, 1000)
		if err != nil {
			alog.Error("Error setting up AD2")
		}
		scope = instr.Scope(digilentAd2)
	}

	time.Sleep(500 * time.Millisecond)
	data := FetchCurve(scope)

	m := glfw.GetMonitors()[0].GetVideoMode()
	alog.Info("Monitor W=%d, H=%d\n", m.Width, m.Height)

	// Top header label
	top := widget.NewLabelWithStyle("Oscilloscope", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	// Setup form contents
	f := plot.NewFrame()
	window.SetContent(fyne.NewContainerWithLayout(layout.NewBorderLayout(top, nil, nil, nil), top, fyne.NewContainerWithLayout(f, f.Canvas)))
	// Center On screen only needed if not maximized
	window.CenterOnScreen()
	// Maximize to fill all of screen
	window.Maximize()
	f.SetData(data)
	go handleEvents(f, window)
	events <- evFetchData
	window.ShowAndRun()
}
