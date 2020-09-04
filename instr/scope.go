package instr

// Package instr defines common interfaces

// SampleMode indicates the decimation mode going from the raw sampling interval
// to the time between stored samples
type SampleMode int

const (
	MinMax SampleMode = iota
	Sample
	Average
)

// Coupling is for oscilloscope prope coupling
type Coupling int

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

const (
	Rising Slope = iota
	Falling
	Either
)

// Scope is an oscilloscope definition
type Scope interface {
	// GetName will return the *IDN? string
	QueryIdn() (string, error)
	// SetupChannel where rng is 10*volt/div
	SetupChannel(ch Chan, rng float64, offset float64, coupling Coupling) error
	// SetupTime will set sampling time and offset
	// sample time is typicaly timePrDiv/samplesPrDiv or f.ex. 10e-6/250 for 10uS/div
	// mode is
	SetupTime(sampleTime float64, offs float64, sampleMode SampleMode) error
	// SetupTrigger will set main trigger parameters
	SetupTrigger(sourceChan Chan, coupling Coupling, slope Slope, trigLevel float64, auto bool, holdoff float64)
	// Measure data on channel. Type may vary, typical FREQUENCY, CRMS etc
	Measure(ch Chan, typ string) (float64, error)
	// Return the data points for a single scan on selected channels
	Curve(channels []Chan, samples int) ([][]float64, error)
	// Close will close communication channel
	Close()
}
