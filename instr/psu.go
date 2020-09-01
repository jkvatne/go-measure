package instr

// Psu is a generic power supply interface
type Psu interface {
	SetOutput(c Chan, voltage float64, current float64) error
	GetOutput(c Chan) (float64, float64, error)
	GetSetpoint(c Chan) (float64, float64, error)
	Disable(c Chan)
	QueryIdn() (string, error)
	Close()
	ChannelCount() int
}
