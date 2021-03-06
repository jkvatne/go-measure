package fluke

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jkvatne/go-measure/instr"
)

// Check if tti interface satisfies Psu interface
var _ instr.Dmm = &Fluke{}

// Fluke stores setup for a Fluke multimeter
type Fluke struct {
	instr.Connection
	setup   instr.Setup
	request string // The request string is set by calling Configure()
}

// New will return an instrument instance
func New(port string) (*Fluke, error) {
	dmm := &Fluke{}
	dmm.Port = port
	dmm.Timeout = 3000 * time.Millisecond
	dmm.Eol = instr.Lf
	err := dmm.Open(port)
	if err != nil {
		return nil, fmt.Errorf("error opening port, %s", err)
	}
	_ = dmm.Write("SYST:REM")
	_ = dmm.Write("*RST")
	time.Sleep(time.Millisecond * 50)
	dmm.Name, err = dmm.Ask("*IDN?")
	time.Sleep(time.Millisecond * 50)
	if err != nil {
		return nil, fmt.Errorf("no instrument found at %s", err)
	}
	if !strings.HasPrefix(dmm.Name, "FLUKE") {
		return nil, fmt.Errorf("port %s has not a Fluke multimeter connected", port)
	}
	return dmm, nil
}

// Close will set the instrument to local and close connection
func (f *Fluke) Close() {
	time.Sleep(time.Millisecond * 50)
	_ = f.Write("SYST:LOC")
}

// Measure will do a measurement according to Configure(setup)
func (f *Fluke) Measure() (float64, error) {
	if f.setup.Unit == instr.Illegal {
		return 0.0, fmt.Errorf("undefined setup")
	}
	// A wait of 50mS is needed to avoid error on the instrument
	time.Sleep(time.Millisecond * 50)
	// Do actual measurement
	response, err := f.Ask(f.request)
	if err != nil {
		return 0.0, err
	}
	response = strings.TrimRight(response, "AV\n\000 ")
	volt, err := strconv.ParseFloat(response, 64)
	return volt, err
}

// Configure will select unit to measure and range etc.
func (f *Fluke) Configure(s instr.Setup) error {
	if s.Chan == 0 {
		s.Chan = 1
	}
	if s.Chan > 1 || s.Chan < 1 {
		return fmt.Errorf("%d is illegal channel", f.setup.Chan)
	}
	r := s.Range
	if s.Unit == instr.VoltDc {
		if r == "" {
			r = "100.0"
		}
		f.request = fmt.Sprintf("MEAS:VOLT:DC? %s", r)
	} else if s.Unit == instr.VoltAcRms {
		if r == "" {
			r = "10.0"
		}
		f.request = fmt.Sprintf("MEAS:VOLT:AC? %s", r)
	} else if s.Unit == instr.CurrentDc {
		if r == "" {
			r = "10.0"
		}
		f.request = fmt.Sprintf("MEAS:CURR:DC? %s", r)
	} else if s.Unit == instr.CurrentAcRms {
		if r == "" {
			r = "10.0"
		}
		f.request = fmt.Sprintf("MEAS:CURR:AC? %s", r)
	} else if s.Unit == instr.Hz {
		f.request = "MEAS:FREQ?"
	} else if s.Unit == instr.Ohm {
		if r == "" {
			r = "10000000.0"
		}
		f.request = fmt.Sprintf("MEAS:RES? %s", r)
	} else {
		return fmt.Errorf("illegal unit")
	}
	f.setup = s
	return nil
}
