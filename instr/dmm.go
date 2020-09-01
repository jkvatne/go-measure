package instr

// EngUnit defines the engineering unit of a measurement
type EngUnit int

const (
	Illegal EngUnit = iota
	VoltDc
	VoltAcRms
	VoltAcAvg
	CurrentDc
	CurrentAcRms
	CurrentAcAvg
	Hz
	Ohm
	Celcius
)

type Setup struct {
	Chan       Chan
	Unit       EngUnit
	Range      string
	Resolution float64
	Rate       float64
}

// Dmm is the interface for a digital multilmeter
type Dmm interface {
	Close()
	Configure(setup Setup) error
	Measure() (float64, error)
	QueryIdn() (string, error)
}
