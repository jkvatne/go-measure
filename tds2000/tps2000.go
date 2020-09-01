package tds2000

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jkvatne/go-measure/alog"
	"github.com/jkvatne/go-measure/instr"
)

// Tps2000 defines a Tektronix oscilloscope in the Tps2000 series
type Tps2000 struct {
	instr.Connection
	currentChan     instr.Chan
	measurementType string
}

// Declare conformity with Scope interface
var _ instr.Scope = (*Tps2000)(nil)

// Point is a 2D point, usually containing (Volt, Time) points
type Point struct {
	x, y float64
}

// GetName will read the ID from the instrument.
func (s *Tps2000) GetName() string {
	name, err := s.Ask("*IDN?")
	if name == "" {
		time.Sleep(100 * time.Millisecond)
		name, err = s.Ask("*IDN?")
	}
	if err != nil {
		return fmt.Sprintf("Error, %s", err)
	}
	s.Connection.Name = name
	return name
}

// Close will terminate connection
func (s *Tps2000) Close() {
	s.Connection.Close()
}

// New is a Oscilloscope instance for the tti supply
func New(port string) (instr.Scope, error) {
	if port == "" {
		port = instr.FindSerialPort("TEKTRONIX", 19200)
	}
	conn := instr.Connection{Port: port, Baudrate: 19200, Timeout: 500 * time.Millisecond, Eol: instr.Lf}
	osc := &Tps2000{Connection: conn}
	err := osc.Open(port)
	if err != nil {
		return nil, fmt.Errorf("error opening port, %s", err)
	}
	osc.Connection.Name = osc.GetName()
	alog.Info("Connected to oscilloscope %s", osc.Connection.Name)
	if !strings.HasPrefix(osc.Connection.Name, "TEKTRONIX,TPS 20") {
		return nil, fmt.Errorf("port %s has not a Textronix osciloscope", port)
	}
	return osc, nil
}

// ButtonLights will turn on/off the front panel background lights
func (s *Tps2000) ButtonLights(on bool) {
	if on {
		_ = s.Write("POW:BUTTONLIGHT ON")
	} else {
		_ = s.Write("POW:BUTTONLIGHT ON")
	}
}

// CurveInfo uses WFMP:WFID? command
// Example: "Ch1, DC coupling, 2.0E0 V/div, 5.0E-6 s/div, 2500 points, Pk Detect mode"
func (s *Tps2000) CurveInfo() error {
	resp, err := s.Ask("WFMP:WFID?")
	if err != nil {
		return err
	}
	_ = strings.Split(resp, ",")
	return nil
}

func (s *Tps2000) opc() string {
	s.Write("*opc?")
	return s.Read()
}

// Curve will return a dataset (points) of 2500 points scaled
func (s *Tps2000) Curve(channels []instr.Chan, samples int) (data [][]float64, err error) {
	/*
		s.Write("DESE 1") // enable *opc in device event status enable register
		s.Write("*ESE 1") // enable *opc in event status enable register
		s.Write("*SRE 0")
		s.Write("acquire:stopafter sequence; state on")
		s.Write("trigger force")
		for s.opc() == "0" {
			time.Sleep(time.Millisecond * 200)
		}
	*/

	// Set binary encoding with lsb first
	s.Write("DATA:WIDTH 1;START 1;STOP %d;ENCDG SRI", samples)
	// Read time (horizontal) scaling
	resp, err := s.Ask("WFMPRE:CH%d:XINCR?", chanString[channels[0]])
	if err != nil {
		return nil, err
	}
	xIncr, _ := strconv.ParseFloat(resp, 64)

	// Generate time sequence
	var timeData []float64
	for i := 0; i < samples; i++ {
		timeData = append(timeData, float64(i)*xIncr)
	}
	data = append(data, timeData)

	for _, channel := range channels {
		s.Write("DATA:SOURCE ch" + chanString[channel])
		s.Write("CURVE?")
		s.Timeout = 5 * time.Second
		values := s.Connection.ReadBinary()
		if len(values) < 5 {
			return nil, fmt.Errorf("no data for channel %s", chanString[channel])
		} else if values[0] != 35 || values[1] < 48 || values[1] > 52 {
			return nil, fmt.Errorf("data should start with #1, #2 or #3")
		}
		n := int(values[2] - 48)
		if values[1] >= 50 {
			n = n*10 + int(values[3]) - 48
		}
		if values[1] >= 51 {
			n = n*10 + int(values[4]) - 48
		}
		if values[1] >= 52 {
			n = n*10 + int(values[5]) - 48
		}
		start := int(values[1]) - 48 + 2
		if n != samples {
			return nil, fmt.Errorf("wrong length of data")
		}
		s.Timeout = time.Second
		// Read channel scaling
		resp, _ = s.Ask("WFMPRE:YMULT?")
		yScale, err := strconv.ParseFloat(resp, 64)
		if err != nil {
			return nil, fmt.Errorf("error reading channel scaling")
		}

		resp, err = s.Ask("WFMPRE:YOFF?")
		yOffset, err := strconv.ParseFloat(resp, 64)
		if err != nil {
			return nil, fmt.Errorf("error reading channel scaling")
		}
		var chanData []float64
		for i := start; i < start+samples; i++ {
			chanData = append(chanData, (float64(values[i])-yOffset)*yScale)
		}
		data = append(data, chanData)
	}
	return data, nil
}

// Measure and return value as float64
// typ is one of  FREQuency | MEAN | PERIod |PK2pk | CRMs | MINImum | MAXImum | RISe | FALL |PWIdth | NWIdth
// If Chan=TRIG (0) then the trigger frequency will be returned. This is much more accurate than the
// frequency determined from a channels waveform
func (s *Tps2000) Measure(ch instr.Chan, typ string) (float64, error) {
	if ch < instr.Ch1 || ch > instr.Ch4 {
		return 0.0, fmt.Errorf("%s is illegal channel", chanString[ch])
	}
	var resp string
	if ch == instr.TRIG {
		resp, _ = s.Ask("TRIG:MAI:FREQ?")
	} else {
		s.currentChan = ch
		s.measurementType = typ
		time.Sleep(time.Millisecond * 10)
		_ = s.Write("MEASU:IMM:SOU CH%d", ch)
		time.Sleep(time.Millisecond * 10)
		_ = s.Write("MEASU:IMMED:TYPE " + typ)
		time.Sleep(time.Millisecond * 10)
		resp, _ = s.Ask("MEASU:IMMED:VALUE?")
	}
	f, err := strconv.ParseFloat(resp, 64)
	if err != nil {
		return 0.0, err
	}
	if f == 0.0 {
		f = 1e-38
	}
	return f, nil
}

// SetupChannel where
// rng is the voltage range, or 10xVolt/div. Dvs rng=10V gives +-5V or 1V/div
// offset is the voltage added to the signal before scaling. 0V is center of screen
func (s *Tps2000) SetupChannel(ch instr.Chan, rng float64, offs float64, coupling instr.Coupling) (err error) {
	// Offset is given in divisions
	_ = s.Write("CH%d:POS %0.3g", ch, offs/rng*10.0)
	// scale is the volt pr division setting
	_ = s.Write("CH%d:SCA %0.3g", ch, rng/10.0)
	// Enable channel
	if coupling == instr.OFF {
		err = s.Write("SEL:CH%d OFF", ch)
	} else if coupling == instr.DC {
		_ = s.Write("CH%d:COUP DC", ch)
		err = s.Write("SEL:CH%d ON", ch)
	} else if coupling == instr.AC {
		_ = s.Write("CH%d:COUP AC", ch)
		err = s.Write("SEL:CH%d ON", ch)
	} else if coupling == instr.GND {
		_ = s.Write("CH%d:COUP GND", ch)
		err = s.Write("SEL:CH%d ON", ch)
	}
	return
}

// SetupTime will set samples pr second and trigger offset
func (s *Tps2000) SetupTime(sampleTime float64, offs float64, mode instr.SampleMode) error {
	// The scope allways store 2500 samples for a full 10 divisions or 250 s/div
	// Samples pr sec is then
	nr := fmt.Sprintf("%0.3e", sampleTime*250)
	// Make sure it is in the 1-2.5-5 sequence
	if nr[0:4] != "1.00" && nr[0:4] != "2.50" && nr[0:4] != "5.00" {
		return fmt.Errorf("time pr div must be 1/2.5/5")
	}
	if mode == instr.MinMax {
	} else if mode == instr.Average {
		_ = s.Write("ACQ:MOD AVE")
	} else if mode == instr.Sample {
		_ = s.Write("ACQ:MOD SAM")
	} else {
		_ = s.Write("ACQ:MOD PEAK")
	}
	err := s.Write("HOR:MAI:SCA " + nr)
	if err != nil {
		return err
	}
	_ = s.Write("HOR:POS %0.3g", offs)
	return nil
}

var couplingString = [...]string{"DC", "DC", "AC", "DC", "HFR", "LFR", "NOISE"}
var slopeString = [...]string{"FALL", "RISE"}
var chanString = [...]string{"", "CH1", "CH2", "CH3", "CH4", "EXT", "EXT5", "EXT10", "AC LINE"}

// SetupTrigger will define scope trigger settings
func (s *Tps2000) SetupTrigger(sourceChan instr.Chan, coupling instr.Coupling, slope instr.Slope, trigLevel float64, auto bool, holdoff float64) {
	_ = s.Write("TRIG:MAIN:EDGE:COUP " + couplingString[coupling])
	_ = s.Write("TRIG:MAIN:EDGE:SLOPE " + slopeString[slope])
	_ = s.Write("TRIG:MAIN:EDGE:SOURCE " + slopeString[sourceChan])
	_ = s.Write("TRIG:MAIN:HOLDOFF:VALUE %0.3e", holdoff)
	_ = s.Write("TRIG:MAIN:LEVEL %0.4e", trigLevel)
	if auto {
		_ = s.Write("TRIG:MAIN:MODE AUTO", trigLevel)
	} else {
		_ = s.Write("TRIG:MAIN:MODE NORMAL", trigLevel)
	}
}
