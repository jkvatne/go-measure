package cpx400

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jkvatne/go-measure/instr"
)

// Check if tti interface satisfies Psu interface
var _ instr.Psu = &Cpx400{}

// Cpx400 stores setup for a TTI CPX4000 power supply
type Cpx400 struct {
	instr.Connection
}

// New returns a PSU instance for the tti supply
func New(port string) (*Cpx400, error) {
	conn := instr.Connection{Port: port, Timeout: 200 * time.Millisecond, Eol: instr.Lf}
	psu := &Cpx400{conn}
	err := psu.Open(port)
	if err != nil {
		return nil, fmt.Errorf("error opening port, %s", err)
	}
	psu.Connection.Name, err = psu.QueryIdn()
	if err != nil || !strings.HasPrefix(psu.Connection.Name, "THURLBY THANDAR, CPX400DP") {
		return nil, fmt.Errorf("port %s has not a TTi supply connected", port)
	}
	return psu, nil
}

// ChannelCount returns the number of channels
func (psu *Cpx400) ChannelCount() int {
	return 2
}

// SetOutput will set output voltage and current limit for a given channel
func (psu *Cpx400) SetOutput(ch instr.Chan, voltage float64, current float64) error {
	// Set output voltage
	err := psu.Write("V%d %0.3f", ch, voltage)
	if err != nil {
		return err
	}
	// Set current limit
	err = psu.Write("I%d %0.2f", ch, current)
	if err != nil {
		return err
	}
	return psu.Write("OP%d 1", ch)
}

// Disable will turn off the given output channel
func (psu *Cpx400) Disable(ch instr.Chan) {
	if ch < 1 || ch > 2 {
		return
	}
	psu.Write("OP%d 0", ch)
}

// GetOutput will return the actual output voltage and current from the channel
func (psu *Cpx400) GetOutput(ch instr.Chan) (float64, float64, error) {
	if ch < 1 || ch > 2 {
		return 0.0, 0.0, fmt.Errorf("channel %d illegal", ch)
	}
	// Read back output voltage
	voltageString, err1 := psu.Ask("V%dO?", ch)
	voltageString = strings.TrimRight(voltageString, "V\n")
	// Read back output current
	currentString, err2 := psu.Ask("I%dO?", ch)
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

// GetSetpoint will return the setpoint voltage and current from the channel
func (psu *Cpx400) GetSetpoint(ch instr.Chan) (float64, float64, error) {
	if ch < 1 || ch > 2 {
		return 0.0, 0.0, fmt.Errorf("channel %d illegal", ch)
	}
	// Read back output voltage setpoint
	voltageString, err1 := psu.Ask("V%d?", ch)
	voltageString = strings.TrimPrefix(voltageString, fmt.Sprintf("V%d ", ch))
	voltageString = strings.TrimRight(voltageString, "V\n")
	// Read back output current setpoint
	currentString, err2 := psu.Ask("I%d?", ch)
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
