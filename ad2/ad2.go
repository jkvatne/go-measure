package ad2

// #cgo LDFLAGS: -L${SRCDIR} -ldwf
// #include "dwf.h"
// #include "stdlib.h"
import "C"
import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/jkvatne/go-measure/instr"
)

// WaveForm is the analog output function generator waveforms
type WaveForm int

// WaveForm constants
const (
	WfDc WaveForm = iota
	WfSine
	WfSquare
	WfTriangle
	WfRampUp
	WfRampDown
	WfNoise
	WfPulse
	WfTapeze
	WfSinPower
	WfCustom
	WfPlay
)

// Ad2 is the state data for the unit
type Ad2 struct {
	hdwf    C.HDWF
	voltage [4]float64
}

// New will create an instance
func New() (*Ad2, error) {
	a := &Ad2{}
	a.Open()
	return a, nil

}

// Open will initialize the Digilent Analog Discovery 2 unit
// This function failed for unknown reasons. When moved directly to main it works ok.
func (a *Ad2) Open() {
	//ver := C.CString("")
	//defer C.free(unsafe.Pointer(ver))
	//C.FDwfGetVersion(ver)
	if C.FDwfDeviceOpen(-1, &a.hdwf) == 0 {
		fmt.Printf("No Digilent device found. Connect via USB.")
		os.Exit(1)
	}
	C.FDwfDeviceReset(a.hdwf)
	C.FDwfAnalogIOEnableSet(a.hdwf, 1)
}

// QueryIdn return name of instrument
func (a *Ad2) QueryIdn() (string, error) {
	ver := C.CString("")
	//defer C.free(unsafe.Pointer(ver))
	err := C.FDwfGetVersion(ver)
	if err == 0 {
		return "", fmt.Errorf("AD2 not found")
	}
	return "Digilent Analog Discovery 2, sdk v" + C.GoString(ver), nil
}

// Close will close the device
func (a *Ad2) Close() {
	C.FDwfDeviceClose(a.hdwf)
}

// Disable will turn off channel
func (a *Ad2) Disable(ch instr.Chan) {
	if ch < 0 || ch > 1 {
		return
	}
	C.FDwfAnalogIOChannelNodeSet(a.hdwf, C.int(ch), 0, 0)
}

// ChannelCount returns number of implemented channels
func (a *Ad2) ChannelCount() int {
	return 2
}

// GetSetpoint returns last settpoint
func (a *Ad2) GetSetpoint(ch instr.Chan) (float64, float64, error) {
	return a.voltage[ch], 0.0, nil
}

// GetOutput will return voltage and current for given channel
// Channel 2 = USB power supply (Volt, Current, Temp) read only
// Channel 3 = Aux power supply (Volt, Current, Temp) read only
func (a *Ad2) GetOutput(ch instr.Chan) (float64, float64, error) {
	var voltage C.double
	C.FDwfAnalogIOChannelNodeGet(a.hdwf, C.int(ch), 0, &voltage)
	return float64(voltage), 0.0, nil
}

// SetOutput will set voltages on V+ and V-
// Channel 0 = V+ (positive power slupply) (Node 0 = Enable, Node 1 = Voltage)
// Channel 1 = V- (negative power slupply) (Node 0 = Enable, Node 1 = Voltage)
func (a *Ad2) SetOutput(ch instr.Chan, voltage float64, current float64) error {
	if ch == 0 && (voltage < 0.5 || voltage > 5.0) {
		return fmt.Errorf("voltage setpoint out of range")
	} else if ch == 1 && (voltage < -5.0 || voltage > -0.5) {
		return fmt.Errorf("voltage setpoint out of range")
	} else if ch < 0 || ch > 1 {
		return fmt.Errorf("channel %d invalid", ch)
	}
	// Enable output
	e := C.FDwfAnalogIOChannelNodeSet(a.hdwf, C.int(ch), 0, 1)
	// Set output voltage
	e &= C.FDwfAnalogIOChannelNodeSet(a.hdwf, C.int(ch), 1, C.double(voltage))
	// Set current limit
	e &= C.FDwfAnalogIOChannelNodeSet(a.hdwf, C.int(ch), 2, C.double(current))
	if e == 0 {
		return fmt.Errorf("error setting supply output")
	}
	return nil
}

// SetupChannel will configure one channels range and offset and coupling
func (a *Ad2) SetupChannel(ch instr.Chan, rng float64, offs float64, coupling string) error {
	e := C.FDwfAnalogInChannelRangeSet(a.hdwf, C.int(ch), C.double(rng))
	e &= C.FDwfAnalogInChannelOffsetSet(a.hdwf, C.int(ch), C.double(offs))
	e &= C.FDwfAnalogInChannelEnableSet(a.hdwf, C.int(ch), C.int(1))
	if coupling != "DC" {
		return fmt.Errorf("only DC coupling allowed")
	}
	return nil
}

// Measure (ch int, typ string) (float64, error)
func (a *Ad2) Measure(ch instr.Chan, typ string) (result float64, err error) {
	const avgCnt = 10
	if typ != "VOLT" {
		return 0.0, fmt.Errorf("only voltage measurements possible")
	}
	e := C.FDwfAnalogInFrequencySet(a.hdwf, C.double(20000000.0))
	e &= C.FDwfAnalogInBufferSizeSet(a.hdwf, C.int(avgCnt))
	e &= C.FDwfAnalogInChannelEnableSet(a.hdwf, C.int(ch), C.int(1))
	e &= C.FDwfAnalogInChannelRangeSet(a.hdwf, C.int(ch), C.double(5.0))
	e &= C.FDwfAnalogInConfigure(a.hdwf, C.int(0), C.int(1))
	for {
		var sts C.uchar
		C.FDwfAnalogInStatus(a.hdwf, C.int(1), &sts)
		if sts == C.DwfStateDone {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	var rgdAnalog [avgCnt]C.double
	e &= C.FDwfAnalogInStatusData(a.hdwf, C.int(0), &rgdAnalog[0], C.int(avgCnt)) // get channel 1 data
	if e == 0 {
		return 0.0, fmt.Errorf("error in Measure()")
	}
	result = float64(rgdAnalog[0]+rgdAnalog[1]+rgdAnalog[2]+rgdAnalog[3]+rgdAnalog[4]) / float64(avgCnt)
	return result, nil
}

// SetAnalogOut will set analog output. PhaseDelay for channel is 0-360.0 degrees.
func (a *Ad2) SetAnalogOut(ch instr.Chan, freq float64, phaseDelay float64, waveform WaveForm, volt float64, offset float64) error {
	channel := C.int(ch - instr.Ch1)
	if ch < instr.Ch1 || ch > instr.Ch2 {
		return fmt.Errorf("only channels 1 and 2 allowed")
	}
	if phaseDelay < 0.0 || phaseDelay > 360.0 {
		return fmt.Errorf("phase must be 0-360.0 degrees")
	}
	// enable channel
	e := C.FDwfAnalogOutNodeEnableSet(a.hdwf, channel, C.AnalogOutNodeCarrier, 1)
	// set sine function
	e &= C.FDwfAnalogOutNodeFunctionSet(a.hdwf, channel, C.AnalogOutNodeCarrier, C.uchar(waveform))
	// set frequency
	if freq >= 0 {
		e &= C.FDwfAnalogOutNodeFrequencySet(a.hdwf, channel, C.AnalogOutNodeCarrier, C.double(freq))
	} else {
		e &= C.FDwfAnalogOutNodeFrequencySet(a.hdwf, channel, C.AnalogOutNodeCarrier, C.double(-freq))
	}
	// set amplitude (Vpp)
	e &= C.FDwfAnalogOutNodeAmplitudeSet(a.hdwf, channel, C.AnalogOutNodeCarrier, C.double(volt))
	// set offset in volts
	e &= C.FDwfAnalogOutNodeOffsetSet(a.hdwf, channel, C.AnalogOutNodeCarrier, C.double(offset))
	// Set phase in degrees
	e &= C.FDwfAnalogOutNodePhaseSet(a.hdwf, channel, C.AnalogOutNodeCarrier, C.double(phaseDelay))
	if e == 0 {
		return fmt.Errorf("error in Measure()")
	}
	return nil
}

// StartAnalogOut will start the function generator. Must be preceeded by SetAnalogOut.
// Use ch=-1 to start both channels
func (a *Ad2) StartAnalogOut(ch instr.Chan) {
	// start signal generation. Use ch=-1 (TRIG) to start both at the same time
	channel := C.int(ch - instr.Ch1)
	C.FDwfAnalogOutConfigure(a.hdwf, channel, 1)
}

// SetupTime will configure the sample interval and mode
func (a *Ad2) SetupTime(sampleInterval float64, offs float64, mode instr.SampleMode) error {
	// The Ad2 samples at 100Msps. The filter constant determine how to go
	//from n input sample to 1 stored sample.
	n := math.Round(sampleInterval * 100e6)
	sampleInterval = sampleInterval / n
	C.FDwfAnalogInFrequencySet(a.hdwf, C.double(1/sampleInterval))
	if mode == instr.MinMax {
		C.FDwfAnalogInChannelFilterSet(a.hdwf, -1, C.filterMinMax)
	} else if mode == instr.Average {
		C.FDwfAnalogInChannelFilterSet(a.hdwf, -1, C.filterAverage)
	} else if mode == instr.Sample {
		C.FDwfAnalogInChannelFilterSet(a.hdwf, -1, C.filterDecimate)
	}
	return nil
}
