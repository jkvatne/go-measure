package fluke_test

import (
	"fmt"
	"go-measure/dmm/fluke"
	"go-measure/instr"
	"testing"

	"github.com/stretchr/testify/assert"
)

var FlukePort = "192.168.2.110:3490"

func TestFluke(t *testing.T) {
	fmt.Printf("Test Fluke multimeter using port %s\n", FlukePort)
	d, err := fluke.New(FlukePort)
	assert.NoError(t, err, "Failed to connect to %s, %s", FlukePort, err)
	name, err := d.QueryIdn()
	assert.NoError(t, err, "Failed to get IDN")
	assert.NotEqual(t, name, "")
	fmt.Printf("Found %s\n", name)

	err = d.Configure(instr.Setup{Unit: instr.VoltDc})
	assert.NoError(t, err, "Configuration failed")
	volt, err := d.Measure()
	fmt.Printf("Measured voltage is %0.6f\n", volt)
	assert.NoError(t, err, "Measurement failed")
	assert.InDelta(t, 0.0, volt, 0.01, "open input voltage")

	volt, err = d.Measure()
	fmt.Printf("Measured voltage is %0.6f\n", volt)
	assert.NoError(t, err, "Measurement failed")
	assert.InDelta(t, 0.0, volt, 0.01, "open input voltage")

	err = d.Configure(instr.Setup{Unit: instr.Ohm, Range: "10000"})
	assert.NoError(t, err, "Configuration failed")
	ohm, err := d.Measure()
	fmt.Printf("Measured resistance is %0.6f\n", ohm)
	assert.NoError(t, err, "Measurement failed")
	assert.InDelta(t, 10000.0, ohm, 100.0, "10k resistor")

	err = d.Configure(instr.Setup{Unit: instr.CurrentDc, Range: "0.1"})
	assert.NoError(t, err, "Configuration failed")
	cur, err := d.Measure()
	fmt.Printf("Measured current is %0.6f\n", cur)
	assert.NoError(t, err, "Measurement failed")
	assert.InDelta(t, 0.0, cur, 0.01, "open input voltage")

	d.Close()
}
