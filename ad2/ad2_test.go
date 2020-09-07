package ad2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jkvatne/go-measure/instr"

	"github.com/jkvatne/go-measure/ad2"

	"github.com/stretchr/testify/assert"
)

func TestOpenUnknown(t *testing.T) {
	_, err := ad2.New("illegal")
	assert.Error(t, err, "Invalid sno should fail")
}

func setQuadratureOut(t *testing.T, a *ad2.Ad2, freq float64) {
	err := a.SetOutput(0, 3.0, 0.0)
	assert.NoError(t, err)
	err = a.SetOutput(1, -2.0, 0.0)
	assert.NoError(t, err)
	err = a.SetAnalogOut(instr.Ch1, freq, 0.0, ad2.WfSine, 2.0, 0)
	assert.NoError(t, err)
	err = a.SetAnalogOut(instr.Ch2, freq, 90.0, ad2.WfSine, 2.0, 0)
	assert.NoError(t, err)
	a.StartAnalogOut(instr.TRIG)
}

const periode = 1e-3 // Generate sine with this periode (1khz)
const samples = 48

func TestScope(t *testing.T) {
	a, err := ad2.New("")
	assert.NoError(t, err)
	if err == nil {
		fmt.Printf("Min buffer=%d, max buffer=%d\n", a.MinBuffer, a.MaxBuffer)
		assert.NoError(t, err)
		setQuadratureOut(t, a, 1/periode)
		err = a.SetupChannel(instr.Ch1, 5.0, 0.0, instr.DC)
		assert.NoError(t, err)
		err = a.SetupChannel(instr.Ch2, 5.0, 0.0, instr.DC)
		assert.NoError(t, err)
		err = a.SetupTrigger(instr.Ch1, instr.DC, instr.Rising, 0.0, false, 0.0)
		assert.NoError(t, err)
		sampleInterval := periode / samples
		err = a.SetupTime(sampleInterval, 0.0, instr.Sample)
		assert.NoError(t, err)
		time.Sleep(2 * time.Second)
		values, err := a.Curve([]instr.Chan{instr.Ch1, instr.Ch2}, samples)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(values))
		assert.Equal(t, samples, len(values[0]))
		assert.Equal(t, samples, len(values[1]))
		assert.Equal(t, samples, len(values[2]))
		assert.InDelta(t, 0.0005, values[0][24], 1e-6)
		assert.InDelta(t, -0.23, values[1][24], 0.15)
		assert.InDelta(t, 0.29, values[1][26], 0.15)
		assert.InDelta(t, -2.0, values[1][13], 3e-2)
		assert.InDelta(t, 2.0, values[1][37], 3e-2)
		assert.InDelta(t, -2.0, values[2][1], 3e-2)
		assert.InDelta(t, 2.0, values[2][24], 3e-2)
	}
}
