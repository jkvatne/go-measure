package bm25x_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jkvatne/go-measure/dmm/bm25x"
	"github.com/stretchr/testify/assert"
)

var Port = "COM15"

func TestDmm(t *testing.T) {
	d, err := bm25x.New(Port)
	assert.NoError(t, err, "Failed to connect to %s, %s", Port, err)
	name, err := d.QueryIdn()
	assert.NoError(t, err, "Failed to get IDN")
	assert.NotEqual(t, name, "")
	fmt.Printf("Found %s\n", name)
	time.Sleep(time.Second * 1)
	volt, err := d.Measure()
	assert.NoError(t, err, "Failed measure()")
	fmt.Printf("Measured voltage is %0.6f\n", volt)
	time.Sleep(time.Second * 1)
	volt, err = d.Measure()
	assert.NoError(t, err, "Failed measure()")
	fmt.Printf("Measured voltage is %0.6f\n", volt)
}
