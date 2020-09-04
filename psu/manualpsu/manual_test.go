package manualpsu_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/jkvatne/go-measure/psu/manualpsu"
	"github.com/stretchr/testify/assert"
)

// Mock stdin and stdout
var InBuffer = bytes.Buffer{}
var OutBuffer = bytes.Buffer{}

func TestManualPsu(t *testing.T) {
	OutBuffer.Grow(4096)
	InBuffer.Grow(4096)
	fmt.Printf("Test manual supply using stdin/stdout\n")
	p, err := manualpsu.New(os.Stdin, os.Stdout)
	assert.NoError(t, err, "Failed to open manual supply")

	p, err = manualpsu.New(&InBuffer, &OutBuffer)
	assert.NoError(t, err, "Failed to open manual supply")
	InBuffer.Write([]byte{13})
	err = p.SetOutput(1, 5.0, 2.0)
	os.Stdout.Write([]byte{10})
	v, c, err := p.GetSetpoint(1)
	assert.Equal(t, 5.0, v, "voltage")
	assert.Equal(t, 2.0, c, "current")
	assert.NoError(t, err, "Failed to set voltage/current")
	InBuffer.Write([]byte{'1', '.', '5', 10})
	v, c, err = p.GetOutput(1)
	assert.Equal(t, 5.0, v, "voltage")
	assert.Equal(t, 1.5, c, "current")
	assert.NoError(t, err, "Failed to set voltage/current")
	p.Close()
}
