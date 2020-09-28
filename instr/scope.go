// Copyright 2020 Jan Kåre Vatne. All rights reserved.

package instr

// SampleMode indicates the decimation mode going from the raw sampling interval
// to the time between stored samples
type SampleMode int

const (
	// MinMax uses two samples for the minimum or  maximum sample values
	MinMax SampleMode = iota
	// Sample uses a single sample for each sample interval
	Sample
	// Average will average all samples in the sample interval
	Average
)

// Coupling is for oscilloscope prope coupling
type Coupling int

// Coupling constants are normal scope types
const (
	OFF = iota
	DC
	AC
	GND
	HfReject
	LfReject
	NoiseReject
)

// Slope is the trigger slope
type Slope int

// Slope types are rising/falling or either edge
const (
	Rising Slope = iota
	Falling
	Either
)

// Scope is an oscilloscope definition
type Scope interface {
	// GetName will return the *IDN? string
	QueryIdn() (string, error)
	// DisableChannel will turn off channel (will no longer be sampled or returned by Curve())
	DisableChannel(ch Chan)
	// SetupChannel where rng is 10*volt/div
	SetupChannel(ch Chan, rng float64, offset float64, coupling Coupling) error
	// GetChanInfo is..
	GetChanInfo() []string
	// SetupTime will set sampling time and offset
	// sample time is typicaly timePrDiv/samplesPrDiv or f.ex. 10e-6/250 for 10uS/div
	// mode is
	SetupTime(sampleTime float64, offs float64, sampleMode SampleMode, sampleCount int) error
	// SetupTrigger will set main trigger parameters
	SetupTrigger(sourceChan Chan, coupling Coupling, slope Slope, trigLevel float64, auto bool, xPos float64) error
	// Measure data on channel. Type may vary, typical FREQUENCY, CRMS etc
	Measure(ch Chan, typ string) (float64, error)
	// Return the data points for a single scan on selected channels
	GetSamples() ([][]float64, error)
	// GetTime will return horizontal settings
	GetTime() (sampleIntervalSec float64, xPosSec float64)
	// Close will close communication channel
	Close()
	// ChannelCount is the maximum number of channels on this instrument
	ChannelCount() int
}
