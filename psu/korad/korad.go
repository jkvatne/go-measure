// korad is an interface to the KD3005 series power supplies.
// An example is the Elfa RND320 supply
// It is special in that it does not use CR/LF as command endings, but depends on timeouts.

package korad

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jkvatne/go-measure/instr"
)

// Check if Psu interface satisfies Psu interface
var _ instr.Psu = &Psu{}

// Psu stores setup for a Korad KD3005 power supply
type Psu struct {
	instr.Connection
	voltage float64
	current float64
}

// New returns a PSU instance for the korad supply
func New(port string) (*Psu, error) {
	if port == "" {
		instr.FindSerialPort("KD3005P", 9600)
	}
	psu := &Psu{}
	psu.Connection.Baudrate = 9600
	err := psu.Open(port)
	if err != nil {
		return nil, fmt.Errorf("error opening port, %s", err)
	}
	name, err := psu.QueryIdn()
	if !strings.Contains(name, "KD3005P") {
		return nil, fmt.Errorf("unknown instrument %s", name)
	}
	return psu, nil
}

// ChannelCount will return the number of channels.
func (psu *Psu) ChannelCount() int {
	return 2
}

// SetOutput will set output voltage and current limit for a given channel
func (psu *Psu) SetOutput(ch instr.Chan, voltage float64, current float64) error {
	// Korad has no enable command, so we save the setpoints and set outputs to zero if not enabled
	psu.voltage = voltage
	psu.current = current
	// The output voltage rate of change is ca 10V/sec
	var wait time.Duration
	if voltage > psu.voltage {
		wait = 100 * time.Millisecond
	} else {
		wait = 50*time.Millisecond + time.Duration(math.Round(math.Abs(voltage-psu.voltage)*30))*time.Millisecond
	}
	// Set output voltage
	err := psu.Connection.Write("VSET%d:%0.2f", ch, voltage)
	if err != nil {
		return err
	}
	// Set current limit
	err = psu.Write("ISET%d:%0.3f", ch, current)
	if err != nil {
		return err
	}
	time.Sleep(wait)
	return nil
}

// Disable will turn off the given output channel
// Korad has no disable option, so we set the output to 0V/0A
func (psu *Psu) Disable(ch instr.Chan) {
	_ = psu.SetOutput(ch, 0, 0)
}

// GetOutput will return the actual output voltage and current from the channel
func (psu *Psu) GetOutput(ch instr.Chan) (float64, float64, error) {
	// Read back output voltage
	voltageString, err1 := psu.Ask("VOUT%d?", ch)
	voltageString = strings.TrimRight(voltageString, "V\n")
	// Read back output current
	currentString, err2 := psu.Ask("IOUT%d?", ch)
	currentString = strings.TrimRight(currentString, "A\n")
	volt, err3 := strconv.ParseFloat(voltageString, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, fmt.Errorf("error reding voltage, %s", err1)
	}
	curr, err := strconv.ParseFloat(currentString, 64)
	if err != nil {
		return volt, 0, fmt.Errorf("error reding current, %s", err)
	}
	return volt, curr, nil
}

// GetSetpoint will return the voltage and current setpoints for the channel
func (psu *Psu) GetSetpoint(ch instr.Chan) (float64, float64, error) {
	// Read back output voltage setpoint
	voltageString, err1 := psu.Ask("VSET%d?", ch)
	voltageString = strings.TrimPrefix(voltageString, fmt.Sprintf("V%d ", ch))
	voltageString = strings.TrimRight(voltageString, "V\n")
	// Read back output current setpoint
	currentString, err2 := psu.Ask("ISET%d?", ch)
	currentString = strings.TrimPrefix(currentString, fmt.Sprintf("I%d ", ch))
	currentString = strings.TrimRight(currentString, "A\n")

	volt, err3 := strconv.ParseFloat(voltageString, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, fmt.Errorf("error reding voltage, %s", err1)
	}

	curr, err := strconv.ParseFloat(currentString, 64)
	if err != nil {
		return volt, 0, fmt.Errorf("error reding current, %s", err)
	}
	return volt, curr, nil
}

// Close will turn off all outputs and close the communication
func (psu *Psu) Close() {
	_ = psu.SetOutput(1, 0, 0)
	psu.Connection.Close()
}
