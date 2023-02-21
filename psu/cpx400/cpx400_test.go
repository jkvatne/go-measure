// Assumes that a TTi CPX4000 power supply is connected to the port given in "port"

package cpx400_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jkvatne/go-measure/instr"
	"github.com/jkvatne/go-measure/psu/cpx400"

	"github.com/stretchr/testify/assert"
)

var TtiPort = "192.168.2.33:9221"

func TestTtiPsu(t *testing.T) {
	fmt.Printf("Test TTI supply using port %s\n", TtiPort)
	p, err := cpx400.New(TtiPort)
	assert.NoError(t, err, "Failed to open %s", TtiPort)
	commonTest(t, p)
	p.Close()
}

func commonTest(t *testing.T, psu instr.Psu) {
	id, err := psu.QueryIdn()
	assert.NoError(t, err, "get name")
	assert.NotEqual(t, "", id)
	fmt.Printf("Found power supply \"%s\"\n", id)
	// Start by disabling both outputs
	fmt.Printf("Turn off both outputs\n")
	psu.Disable(1)
	psu.Disable(2)
	time.Sleep(500 * time.Millisecond)
	// Verify zero
	volt, current, err := psu.GetOutput(1)
	assert.NoError(t, err, "get output 1")
	assert.InDelta(t, 0.0, volt, 0.3, "voltage 1 setpoint")
	assert.InDelta(t, 0.0, current, 0.1, "current 1 setpoint")
	volt, current, err = psu.GetOutput(2)
	assert.NoError(t, err, "get output 1")
	assert.InDelta(t, 0.0, volt, 0.1, "voltage 2 output")
	assert.InDelta(t, 0.0, current, 0.1, "current 2 output")

	// Set both to 20.0V
	fmt.Printf("Set both outputs to 20.0V\n")
	err = psu.SetOutput(instr.Ch1, 20.0, 0.2)
	assert.NoError(t, err, "set output 1")
	err = psu.SetOutput(instr.Ch2, 20.0, 0.15)
	assert.NoError(t, err, "set output 2")
	time.Sleep(500 * time.Millisecond)

	// Read back and verify output
	volt, current, err = psu.GetOutput(instr.Ch1)
	assert.NoError(t, err, "get output 1")
	assert.InDelta(t, 20.0, volt, 0.1, "voltage 1 output")
	assert.InDelta(t, 0.0, current, 0.1, "current 1 output")

	volt, current, err = psu.GetOutput(instr.Ch2)
	assert.NoError(t, err, "get output 2")
	assert.InDelta(t, 20.0, volt, 0.1, "voltage 2 output")
	assert.InDelta(t, 0.0, current, 0.1, "current 2 output")

	// Read back and verify setpoints
	volt, current, err = psu.GetSetpoint(instr.Ch1)
	assert.NoError(t, err, "get output 1")
	assert.InDelta(t, 20.0, volt, 0.01, "voltage 1 setpoint")
	assert.InDelta(t, 0.2, current, 0.01, "current 1 setpoint")

	volt, current, err = psu.GetSetpoint(instr.Ch2)
	assert.NoError(t, err, "get output 1")
	assert.InDelta(t, 20.0, volt, 0.01, "voltage 2 setpoint")
	assert.InDelta(t, 0.15, current, 0.01, "current 2 setpoint")

	fmt.Printf("Shutdown\n")
	psu.Disable(1)
	psu.Disable(2)
	psu.Close()
}
