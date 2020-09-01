package tds2000_test

import (
	"fmt"
	"go-measure/instr"
	"go-measure/tds2000"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) instr.Scope {
	list, _, err := instr.EnumerateSerialPorts()
	assert.NoError(t, err, "error fetching com port list")
	assert.True(t, len(list) > 0, "no ports found")
	o, err := tds2000.New("")
	assert.NoError(t, err, "Failed to open COM port")
	return o
}

func TestBasic(t *testing.T) {
	o := setupTest(t)
	name := o.GetName()
	assert.True(t, len(name) > 2, "Name does not exist")
	assert.True(t, strings.HasPrefix(name, "TEKTRONIX,TPS 2"), "Wrong name")
	o.Close()
}

func TestMeasurements(t *testing.T) {
	o := setupTest(t)

	f, err := o.Measure(1, "FREQ")
	assert.NoError(t, err)
	fmt.Printf("Frequency=%0.3f\n", f)
	assert.InDelta(t, f, f, 100000.0, 1.0)

	f, err = o.Measure(1, "FREQ")
	assert.NoError(t, err)
	fmt.Printf("Frequency=%0.3f\n", f)
	assert.InDelta(t, f, f, 100000.0, 1.0)

	f, err = o.Measure(1, "MEAN")
	assert.NoError(t, err)
	fmt.Printf("MEAN=%0.3f\n", f)

	f, err = o.Measure(1, "CRMS")
	assert.NoError(t, err)
	fmt.Printf("CRMS=%0.3f\n", f)

	f, err = o.Measure(1, "PK2PK")
	assert.NoError(t, err)
	fmt.Printf("PK2PK=%0.3f\n", f)

	f, err = o.Measure(1, "PERIOD")
	assert.NoError(t, err)
	fmt.Printf("PERIOD=%0.3fuS\n", f*1e6)

	f, err = o.Measure(1, "MINIMUM")
	assert.NoError(t, err)
	fmt.Printf("MINIMUM=%0.3f\n", f)

	f, err = o.Measure(1, "MAXIMUM")
	assert.NoError(t, err)
	fmt.Printf("MAXIMUM=%0.3f\n", f)
	o.Close()
}

func TestCurve(t *testing.T) {
	o := setupTest(t)
	sampleCount := 50
	data, err := o.Curve([]instr.Chan{instr.Ch1, instr.Ch2, instr.Ch3}, sampleCount)
	assert.NoError(t, err, "Must have channels 1,2,3 enabled on the scope")
	assert.Equal(t, 4, len(data), "expected four datasets")
	assert.Equal(t, sampleCount, len(data[0]))
	assert.Equal(t, sampleCount, len(data[1]))
	assert.Equal(t, sampleCount, len(data[2]))
	fmt.Println(data[0])
	fmt.Println(data[1])
	fmt.Println(data[2])
	fmt.Println(data[3])
	o.Close()
}

// To show 0-8V signal, use rng=10 and offset=-4
func TestChannelSetup(t *testing.T) {
	o := setupTest(t)
	err := o.SetupChannel(instr.Ch1, 10, -4.0, instr.DC)
	assert.NoError(t, err, "Failed setup channel 1")
	err = o.SetupChannel(instr.Ch2, 0.0, 0.0, instr.OFF)
	assert.NoError(t, err, "Failed setup channel 2")
	err = o.SetupChannel(instr.Ch2, 0.0, 0.0, instr.OFF)
	assert.NoError(t, err, "Failed setup channel 3")
	err = o.SetupChannel(instr.Ch2, 0.0, 0.0, instr.OFF)
	assert.NoError(t, err, "Failed setup channel 4")
	err = o.SetupTime(100e-6/250, 25e-6, instr.Average)
}

func TestSetupTime(t *testing.T) {
	o := setupTest(t)
	_ = o.SetupTime(10e-6/250, 20e-6, instr.MinMax)
}
