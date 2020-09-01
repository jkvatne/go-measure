package ad2

// #cgo LDFLAGS: -L${SRCDIR} -ldwf
// #include "dwf.h"
// #include "stdlib.h"
import "C"
import (
	"fmt"
	"go-measure/instr"
	"math"
	"os"
	"time"
)

// Ad2 is the state data for the unit
type Ad2 struct {
	hdwf    C.HDWF
	voltage [4]float64
}

// New will create an instance
func New() (*Ad2, error) {
	a := &Ad2{}
	return a, nil

}

// Open will initialize the Digilent Analog Discovery 2 unit
// This function failed for unknown reasons. When moved directly to main it works ok.
func (a *Ad2) Open() {
	ver := C.CString("")
	//defer C.free(unsafe.Pointer(ver))
	C.FDwfGetVersion(ver)
	if C.FDwfDeviceOpen(-1, &a.hdwf) == 0 {
		fmt.Printf("No Digilent device found. Connect via USB.")
		os.Exit(1)
	}
	fmt.Printf("Found Digilent SDK version " + C.GoString(ver) + "\n")
	C.FDwfDeviceReset(a.hdwf)
	//a.SetOutput(0, 0.0, 0.0)
	//a.SetOutput(1, 0.0, 0.0)
	// Master enable analog output
	C.FDwfAnalogIOEnableSet(a.hdwf, 1)
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

// QueryIdn return name of instrument
func (a *Ad2) QueryIdn() (string, error) {
	return "Digilent Analog Discovery 2", nil
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
	if e != 0 {
		return fmt.Errorf("error setting supply output")
	}
	return nil
}

// SetupChannel will configure one channels range and offset and coupling
func (a *Ad2) SetupChannel(ch instr.Chan, rng float64, offs float64, coupling string) error {
	C.FDwfAnalogInChannelRangeSet(a.hdwf, ch, C.double(rng))
	C.FDwfAnalogInChannelOffsetSet(a.hdwf, ch, C.double(offs)())
	C.FDwfAnalogInChannelEnableSet(a.hdwf, ch, C.int(1))
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
	C.FDwfAnalogInFrequencySet(a.hdwf, C.double(20000000.0))
	C.FDwfAnalogInBufferSizeSet(a.hdwf, C.int(avgCnt))
	C.FDwfAnalogInChannelEnableSet(a.hdwf, C.int(ch), C.int(1))
	C.FDwfAnalogInChannelRangeSet(a.hdwf, C.int(ch), C.double(5.0))
	C.FDwfAnalogInConfigure(a.hdwf, C.int(0), C.int(1))
	for {
		var sts C.uchar
		C.FDwfAnalogInStatus(a.hdwf, C.int(1), &sts)
		if sts == C.DwfStateDone {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	var rgdAnalog [avgCnt]C.double
	C.FDwfAnalogInStatusData(a.hdwf, C.int(0), &rgdAnalog[0], C.int(avgCnt)) // get channel 1 data
	result = (rgdAnalog[0] + rgdAnalog[1] + rgdAnalog[2] + rgdAnalog[3] + rgdAnalog[4]) / float64(avgCnt)
	return result, nil
}

/*
// SetQuadrature will set analog output to simulate quadrature encoder
func (a *Ad2) SetQuadrature(freq float64, volt float64, offset float64) {
	// enable both channels
	C.FDwfAnalogOutNodeEnableSet(a.hdwf, 0, C.AnalogOutNodeCarrier, 1)
	C.FDwfAnalogOutNodeEnableSet(a.hdwf, 1, C.AnalogOutNodeCarrier, 1)
	// set sine function
	C.FDwfAnalogOutNodeFunctionSet(a.hdwf, -1, C.AnalogOutNodeCarrier, C.funcSine)
	// set frequency
	if freq >= 0 {
		C.FDwfAnalogOutNodeFrequencySet(a.hdwf, -1, C.AnalogOutNodeCarrier, C.double(freq))
	} else {
		C.FDwfAnalogOutNodeFrequencySet(a.hdwf, -1, C.AnalogOutNodeCarrier, C.double(-freq))
	}
	// 0.5V amplitude (1Vpp)
	C.FDwfAnalogOutNodeAmplitudeSet(a.hdwf, 0, C.AnalogOutNodeCarrier, C.double(volt))
	C.FDwfAnalogOutNodeAmplitudeSet(a.hdwf, 1, C.AnalogOutNodeCarrier, C.double(volt))
	// 2.5V offset
	C.FDwfAnalogOutNodeOffsetSet(a.hdwf, 0, C.AnalogOutNodeCarrier, C.double(offset))
	C.FDwfAnalogOutNodeOffsetSet(a.hdwf, 1, C.AnalogOutNodeCarrier, C.double(offset))
	// Set quadrature phase
	if freq > 0 {
		C.FDwfAnalogOutNodePhaseSet(a.hdwf, 0, C.AnalogOutNodeCarrier, 0)
		C.FDwfAnalogOutNodePhaseSet(a.hdwf, 1, C.AnalogOutNodeCarrier, 90)
	} else {
		C.FDwfAnalogOutNodePhaseSet(a.hdwf, 0, C.AnalogOutNodeCarrier, 90)
		C.FDwfAnalogOutNodePhaseSet(a.hdwf, 1, C.AnalogOutNodeCarrier, 0)
	}
	// start signal generation
	C.FDwfAnalogOutConfigure(a.hdwf, -1, 1)
}

// SetupScope is temporary
func (a *Ad2) SetupScope() {
	C.FDwfAnalogInFrequencySet(a.hdwf, C.double(20000000.0))
	C.FDwfAnalogInBufferSizeSet(a.hdwf, C.int(100))
	C.FDwfAnalogInChannelEnableSet(a.hdwf, C.int(0), C.int(1))
	C.FDwfAnalogInChannelRangeSet(a.hdwf, C.int(0), C.double(5.0))
}

*/

// SetupTime will configure the sample interval and mode
func (a *Ad2) SetupTime(sampleInterval float64, offs float64, mode instr.SampleMode) error {
	// The Ad2 samples at 100Msps. The filter constant determine how to go
	//from n input sample to 1 stored sample.
	n := math.Round(sampleInterval * 100e6)
	sampleInterval = sampleInterval / n
	C.FDwfAnalogInFrequencySet(a.hdwf, C.double(1/sampleInterval))
	if mode == instr.MinMax {
		C.FDwfAnalogInChannelFilterSet(a.hdwf, C.filterMinMax)
	} else if mode == instr.Average {
		C.FDwfAnalogInChannelFilterSet(a.hdwf, C.filterAverage)
	} else if mode == instr.Sample {
		C.FDwfAnalogInChannelFilterSet(a.hdwf, C.filterDecimate)
	}
	return nil
}
