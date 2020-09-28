package main

import (
	"flag"
	"log"
	"os"
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

func handleEvents(f *plot.Frame, scope instr.Scope) {
	time.Sleep(time.Second)
	for {
		select {
		case e := <-events:
			if e == evFetchData {
				data, _ := scope.GetSamples()
				f.DataMutex.Lock()
				f.Data = data
				f.DataMutex.Unlock()
			}
			if e == evDone {
				return
			}
			if e == evRedraw || e == evFetchData {
				f.Refresh()
			}
		}
	}
}

func startPolling() {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			events <- evFetchData
		}
	}()
}

// SetupAd2 will initialize Digilent Analog Discovery 2
func SetupAd2(a *ad2.Ad2, freq float64) error {
	err := a.SetOutput(0, 3.0, 0.0)
	err = a.SetOutput(1, -2.0, 0.0)
	err = a.SetAnalogOut(instr.Ch1, freq, 0.0, ad2.WfSine, 2.0, 0)
	err = a.SetAnalogOut(instr.Ch2, freq, 90.0, ad2.WfSine, 2.0, 0)
	a.StartAnalogOut(instr.TRIG)
	err = a.SetupChannel(instr.Ch1, 10.0, 0.0, instr.DC)
	err = a.SetupChannel(instr.Ch2, 10.0, 0.0, instr.DC)
	err = a.SetupTrigger(instr.Ch2, instr.DC, instr.Rising, 1.0, false, 0.0)
	sampleInterval := 1.0 / 2500.0 / freq
	err = a.SetupTime(sampleInterval, 0.0, instr.Average, 2500)
	return err
}

var useName string

func init() {
	flag.StringVar(&useName, "use", "ad2", "Use ad2 or tps2000 as digitizer")
}

func main() {
	flag.Parse()
	alog.Setup(os.Stdout, alog.InfoLevel, log.Ltime|log.Ldate|log.Lmicroseconds)
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
			alog.Fatal("No tps2000 scope found, %s", err)
		}
		_ = scope.SetupTime(1e-3/250, 0.0, instr.MinMax, 2500)
		_ = scope.SetupChannel(instr.Ch1, 10.0, -4.0, instr.DC)
	} else {
		digilentAd2, err = ad2.New("")
		err = SetupAd2(digilentAd2, 1000)
		if err != nil {
			alog.Error("Error setting up AD2")
		}
		scope = instr.Scope(digilentAd2)
	}

	time.Sleep(100 * time.Millisecond)

	m := glfw.GetMonitors()[0].GetVideoMode()
	alog.Info("Monitor W=%d, H=%d\n", m.Width, m.Height)

	// Top header label
	top := widget.NewLabelWithStyle("Oscilloscope", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	// Setup form contents
	f := plot.NewFrame()
	window.SetContent(fyne.NewContainerWithLayout(layout.NewBorderLayout(top, nil, nil, nil), top, f))
	// Center On screen only needed if not maximized
	window.CenterOnScreen()
	// Maximize to fill all of screen
	window.Maximize()
	go handleEvents(f, scope)
	events <- evFetchData
	startPolling()
	window.ShowAndRun()
}
