package manualpsu

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/jkvatne/go-measure/instr"
)

// Check if tti interface satisfies Psu interface
var _ instr.Psu = &ManualPsu{}

// ManualPsu stores setup for a manual PSU
type ManualPsu struct {
	voltage [3]float64
	current [3]float64
	In      *io.Reader
	Out     *io.Writer
}

// NewManualPsu returns a PSU instance for the tti supply
func New(in io.Reader, out io.Writer) (*ManualPsu, error) {
	psu := &ManualPsu{}
	psu.In = &in
	psu.Out = &out
	return psu, nil
}

// GetName will return the instrument name
func (p *ManualPsu) QueryIdn() (string, error) {
	return "Manual power supply control", nil
}

// ChannelCount will return the number of ports.
func (p *ManualPsu) ChannelCount() int {
	return 2
}

// SetOutput will set output voltage and current limit for a given channel
func (p *ManualPsu) SetOutput(ch instr.Chan, voltage float64, current float64) error {
	fmt.Printf("Set PSU output %d to %0.3fV, %0.3fA and press <enter> :", ch, voltage, current)
	r := bufio.NewReader(*p.In)
	_, _ = r.ReadBytes('\n')
	p.voltage[ch] = voltage
	p.current[ch] = current
	return nil
}

// GetSetpoint will return the voltage and current setpoints
func (p *ManualPsu) GetSetpoint(ch instr.Chan) (float64, float64, error) {
	return p.voltage[ch], p.current[ch], nil
}

// Disable will turn off the given output channel
func (p *ManualPsu) Disable(ch instr.Chan) {
	fmt.Printf("Turn off PSU output %d and press <enter> :", ch)
	r := bufio.NewReader(*p.In)
	_, _ = r.ReadBytes('\n')
	return
}

// GetOutput will return the actual output voltage and current from the channel
func (p *ManualPsu) GetOutput(ch instr.Chan) (float64, float64, error) {
	_, err := fmt.Fprintf(*p.Out, "Get PSU current  %d and press <enter> : ", ch)
	if err != nil {
		return 0, 0, err
	}
	r := bufio.NewReader(*p.In)
	b, err := r.ReadBytes('\n')
	if err != nil {
		return 0, 0, err
	}
	s := instr.ToString(b)
	cur, _ := strconv.ParseFloat(s, 64)
	return p.voltage[ch], cur, nil
}

// Close will turn off all outputs and close the communication
func (p *ManualPsu) Close() {
	_, _ = fmt.Fprint(*p.Out, "Turn off power supply\n")
}
