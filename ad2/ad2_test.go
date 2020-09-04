package ad2_test

import (
	"fmt"
	"testing"

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
	err = a.SetAnalogOut(instr.Ch1, freq, 0.0, ad2.WfSine, 1.0, 2.5)
	assert.NoError(t, err)
	err = a.SetAnalogOut(instr.Ch2, freq, 90.0, ad2.WfNoise, 1.0, 2.5)
	assert.NoError(t, err)
	a.StartAnalogOut(instr.TRIG)
}

func TestAd2(t *testing.T) {
	a, err := ad2.New("")
	assert.NoError(t, err)
	setQuadratureOut(t, a, 20000)
	fmt.Printf("Done\n")
	a.Close()
}

func TestScope(t *testing.T) {
	a, err := ad2.New("")
	assert.NoError(t, err)
	if err == nil {
		fmt.Printf("Min buffer=%d, max buffer=%d\n", a.MinBuffer, a.MaxBuffer)
		assert.NoError(t, err)
		setQuadratureOut(t, a, 1e4)
		err = a.SetupTime(1e-4/20, 0.0, instr.Average)
		assert.NoError(t, err)
	}
}
