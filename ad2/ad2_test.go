package ad2_test

import (
	"fmt"
	"testing"

	"github.com/jkvatne/go-measure/instr"

	"github.com/jkvatne/go-measure/ad2"

	"github.com/stretchr/testify/assert"
)

func TestAd2(t *testing.T) {
	a, err := ad2.New()
	assert.NoError(t, err)
	err = a.SetOutput(0, 3.0, 0.0)
	assert.NoError(t, err)
	err = a.SetOutput(1, -2.0, 0.0)
	assert.NoError(t, err)
	err = a.SetAnalogOut(instr.Ch1, 20000000, 0.0, ad2.WfSine, 1.0, 2.5)
	assert.NoError(t, err)
	err = a.SetAnalogOut(instr.Ch2, 20000000, 90.0, ad2.WfNoise, 1.0, 2.5)
	assert.NoError(t, err)
	a.StartAnalogOut(instr.TRIG)
	fmt.Printf("Done\n")
	a.Close()
}
