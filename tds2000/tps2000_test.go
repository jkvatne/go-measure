package tds2000_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jkvatne/go-measure/instr"
	"github.com/jkvatne/go-measure/tds2000"

	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) instr.Scope {
	list, _, err := instr.EnumerateSerialPorts()
	assert.NoError(t, err, "Error fetching com port list")
	assert.True(t, len(list) > 0, "No ports found")
	o, err := tds2000.New("")
	assert.NoError(t, err, "Failed to open COM port")
	assert.NotNil(t, o, "tds2000.New() returned nil")
	return o
}

func TestBasic(t *testing.T) {
	if o := setupTest(t); o != nil {
		name, err := o.QueryIdn()
		assert.NoError(t, err)
		assert.True(t, len(name) > 2, "Name does not exist")
		assert.True(t, strings.HasPrefix(name, "TEKTRONIX,TPS 2"), "Wrong name")
		o.Close()
	}
}

// TestMeasurements assumes that probe for channel 1 is connected to the
// probe comp test output at 1khz 0-5V square wave
func TestMeasurements(t *testing.T) {
	if o := setupTest(t); o != nil {

		err := o.SetupTime(1e-3/250, 20e-6, instr.MinMax)
		assert.NoError(t, err)

		f, err := o.Measure(instr.TRIG, "FREQ")
		assert.NoError(t, err)
		fmt.Printf("Trigger frequency=%0.6f\n", f)
		assert.InDelta(t, 1000, f, 1.0)

		f, err = o.Measure(instr.Ch1, "FREQ")
		assert.NoError(t, err)
		fmt.Printf("Frequency=%0.6f\n", f)
		assert.InDelta(t, 1000, f, 1.0)

		f, err = o.Measure(instr.Ch1, "MEAN")
		assert.NoError(t, err)
		fmt.Printf("MEAN=%0.3f\n", f)
		assert.InDelta(t, 2.5, f, 0.1)

		f, err = o.Measure(instr.Ch1, "CRMS")
		assert.NoError(t, err)
		fmt.Printf("CRMS=%0.3f\n", f)
		assert.InDelta(t, 3.5, f, 0.1)

		f, err = o.Measure(instr.Ch1, "PK2PK")
		assert.NoError(t, err)
		fmt.Printf("PK2PK=%0.3f\n", f)
		assert.InDelta(t, 5.0, f, 0.1)

		f, err = o.Measure(instr.Ch1, "PERIOD")
		assert.NoError(t, err)
		fmt.Printf("PERIOD=%0.3fuS\n", f*1e6)
		assert.InDelta(t, 1e-3, f, 1e-5)

		f, err = o.Measure(1, "MINIMUM")
		assert.NoError(t, err)
		fmt.Printf("MINIMUM=%0.3f\n", f)
		assert.InDelta(t, 0.0, f, 0.1)

		f, err = o.Measure(1, "MAXIMUM")
		assert.NoError(t, err)
		fmt.Printf("MAXIMUM=%0.3f\n", f)
		assert.InDelta(t, 5.0, f, 0.1)
		o.Close()
	}
}

func TestCurve(t *testing.T) {
	if o := setupTest(t); o != nil {
		sampleCount := 50
		data, err := o.Curve([]instr.Chan{instr.Ch1}, sampleCount)
		assert.NoError(t, err, "Must have channel 1 enabled on the scope")
		assert.Equal(t, 2, len(data), "expected one datasets")
		if len(data) > 0 {
			assert.Equal(t, sampleCount, len(data[0]))
			fmt.Println(data[0])
		}
		o.Close()
	}
}

// To show 0-8V signal, use rng=10 and offset=-4
func TestChannelSetup(t *testing.T) {
	if o := setupTest(t); o != nil {
		err := o.SetupChannel(instr.Ch1, 10, -4.0, instr.DC)
		assert.NoError(t, err, "Failed setup channel 1")
		err = o.SetupChannel(instr.Ch2, 0.0, 0.0, instr.OFF)
		assert.NoError(t, err, "Failed setup channel 2")
		err = o.SetupChannel(instr.Ch2, 0.0, 0.0, instr.OFF)
		assert.NoError(t, err, "Failed setup channel 3")
		err = o.SetupChannel(instr.Ch2, 0.0, 0.0, instr.OFF)
		assert.NoError(t, err, "Failed setup channel 4")
		err = o.SetupTime(100e-6/250, 25e-6, instr.Average)
		o.Close()
	}
}

func TestSetupTime(t *testing.T) {
	time.Sleep(time.Second)
	if o := setupTest(t); o != nil {
		err := o.SetupTime(1e-3/250, 2e-3, instr.MinMax)
		assert.NoError(t, err, "Failed SetupTime")
		o.Close()
	}
}
