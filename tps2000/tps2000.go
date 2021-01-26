// Package tps2000 defines a Tektronix oscilloscope in the Tps2000 series
package tps2000

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jkvatne/go-measure/alog"
	"github.com/jkvatne/go-measure/instr"
)

// Tps2000 contains local state data for the scope
type Tps2000 struct {
	instr.Connection
	currentChan     instr.Chan
	measurementType string
	sampleCount     int
	maxSampleCount  int
	enabled         [4]bool
	offsets         [4]float64
	ranges          [4]float64
	channelCount    int
}

// Declare conformity with Scope interface
var _ instr.Scope = (*Tps2000)(nil)

var callNo int

// Point is a 2D point, usually containing (Volt, Time) points
type Point struct {
	x, y float64
}

// New is a Oscilloscope instance for the tti supply
func New(port string) (instr.Scope, error) {
	if port == "" {
		port = instr.FindSerialPort("TEKTRONIX", 19200, instr.Lf)
	}
	conn := instr.Connection{Port: port, Baudrate: 19200, Timeout: 750 * time.Millisecond, Eol: instr.Lf}
	osc := &Tps2000{Connection: conn}
	err := osc.Open(port)
	if err != nil {
		return nil, fmt.Errorf("error opening port, %s", err)
	}
	osc.Connection.Name, err = osc.QueryIdn()
	alog.Info("Connected to oscilloscope %s", osc.Connection.Name)
	if !strings.HasPrefix(osc.Connection.Name, "TEKTRONIX,TPS 20") {
		return nil, fmt.Errorf("port %s has not a Textronix osciloscope", port)
	}
	osc.sampleCount = 2500
	osc.maxSampleCount = 2500
	osc.channelCount = 4

	return osc, nil
}

// QueryIdn will read the ID from the instrument.
func (s *Tps2000) QueryIdn() (string, error) {
	name, err := s.Ask("*IDN?")
	if err != nil {
		return "", fmt.Errorf("Error, %s", err)
	}
	s.Connection.Name = name
	return name, nil
}

// Close will terminate connection
func (s *Tps2000) Close() {
	s.Connection.Close()
	time.Sleep(200 * time.Millisecond)
}

// ButtonLights will turn on/off the front panel background lights
func (s *Tps2000) ButtonLights(on bool) {
	if on {
		_ = s.Write("POW:BUTTONLIGHT ON")
	} else {
		_ = s.Write("POW:BUTTONLIGHT OFF")
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
	_ = s.Write("*opc?")
	return s.ReadString()
}

// DisableChannel turns the channel off (no longer visible)
func (s *Tps2000) DisableChannel(ch instr.Chan) {
	s.enabled[ch] = false
}

var running int32

func run(start bool) {
	if start {
		if atomic.SwapInt32(&running, 1) == 1 {
			alog.Fatal("GetSamples() was running")
		}
	} else {
		atomic.StoreInt32(&running, 0)
	}
}

// GetSamples will return a dataset (points) of 2500 points scaled
func (s *Tps2000) GetSamples() (data [][]float64, err error) {
	run(true)
	defer run(false)

	// Set binary encoding with lsb first
	err = s.Write("DATA:WIDTH 1;START 1;STOP %d;ENCDG SRI", s.sampleCount)
	if err != nil {
		return nil, err
	}
	// Read time (horizontal) scaling
	resp, err := s.Ask("WFMPRE:CH1:XINCR?")
	if err != nil {
		return nil, err
	}
	xIncr, _ := strconv.ParseFloat(resp, 64)
	if xIncr == 0.0 {
		return nil, fmt.Errorf("time/div is missing")
	}
	// Generate time sequence
	var timeData []float64
	for i := 0; i < s.sampleCount; i++ {
		timeData = append(timeData, float64(i)*xIncr)
	}
	data = append(data, timeData)
	var yMax []float64
	var yMin []float64
	for channel := 0; channel < 4; channel++ {
		if s.enabled[channel] {
			_ = s.Write("DATA:SOURCE " + chanString[channel])
			_ = s.Write("CURVE?")
			s.Timeout = 5 * time.Second
			b := s.Connection.ReadByte()
			if b != 35 {
				return nil, fmt.Errorf("data should start with #1, #2 or #3")
			}
			// Get number of digits in length field
			nd := int(s.Connection.ReadByte())
			if nd < 48 || nd > 52 {
				return nil, fmt.Errorf("data should start with #1, #2 or #3")
			}
			n := 0
			for i := 0; i < nd-48; i++ {
				n = n*10 + int(s.Connection.ReadByte()) - 48
			}
			if n != s.sampleCount {
				return nil, fmt.Errorf("wrong length of data")
			}
			s.Timeout = time.Second
			values := make([]byte, n)
			time.Sleep(time.Second * time.Duration(s.sampleCount*10) / time.Duration(s.Baudrate))
			actual := s.Connection.Read(values)
			if actual != s.sampleCount {
				return nil, fmt.Errorf("wrong length of data")
			}
			// Read channel scaling
			yScale, err := s.PollFloat("WFMPRE:YMULT?")
			if err != nil {
				return nil, fmt.Errorf("error reading channel scaling YMULT")
			}
			yOffset, err := s.PollFloat("WFMPRE:YOFF?")
			if err != nil {
				return nil, fmt.Errorf("error reading channel scaling YOFF")
			}
			yMax = append(yMax, (125-yOffset)*yScale)
			yMin = append(yMin, (-125-yOffset)*yScale)
			var chanData []float64
			for i := 0; i < s.sampleCount; i++ {
				v := (float64(int8(values[i])) - yOffset) * yScale
				chanData = append(chanData, v)
			}
			data = append(data, chanData)
		}
	}
	callNo++
	dataPoints := 0
	if len(data) > 1 {
		dataPoints = len(data[1])
	}
	alog.Info("GetSamplies() n=%d, %d %d", callNo, len(data), dataPoints)
	data = append(data, yMax, yMin)
	return data, nil
}

// Measure and return value as float64
// typ is one of  FREQuency | MEAN | PERIod |PK2pk | CRMs | MINImum | MAXImum | RISe | FALL |PWIdth | NWIdth
// If Chan=TRIG (0) then the trigger frequency will be returned. This is much more accurate than the
// frequency determined from a channels waveform
func (s *Tps2000) Measure(ch instr.Chan, typ string) (float64, error) {
	if (ch < instr.Ch1 || ch > instr.Ch4) && ch != instr.TRIG {
		return 0.0, fmt.Errorf("%d is illegal channel", ch)
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
	c := int(ch - instr.Ch1)
	s.offsets[c] = offs
	s.ranges[c] = rng
	s.enabled[c] = true
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

// GetChanInfo ...
func (s *Tps2000) GetChanInfo() (info []string) {
	for ch := 0; ch < s.channelCount; ch++ {
		if s.enabled[ch] {
			s := fmt.Sprintf("Ch%d %s/div ", int(ch), instr.VoltToStr(s.ranges[ch]/10))
			info = append(info, s)
		}
	}
	return info
}

// SetupTime will set samples pr second and trigger offset
func (s *Tps2000) SetupTime(sampleIntervalSec float64, xPosSec float64, mode instr.SampleMode, sampleCount int) error {
	if sampleCount > s.maxSampleCount {
		return fmt.Errorf("samplecount of %d is larger than max %d", sampleCount, s.maxSampleCount)
	}
	s.sampleCount = sampleCount
	// The scope allways store 2500 samples for a full 10 divisions or 250 s/div
	// Samples pr sec is then
	nr := fmt.Sprintf("%0.3e", sampleIntervalSec*250)
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
	_ = s.Write("HOR:MAI:POS %0.3g", xPosSec)
	return nil
}

// GetTime will return horizontal settings
func (s *Tps2000) GetTime() (sampleIntervalSec float64, xPosSec float64) {
	sampleIntervalSec, _ = s.PollFloat("HOR:MAI:SCA?")
	xPosSec, _ = s.PollFloat("HOR:MAI:POS?")
	return sampleIntervalSec, xPosSec
}

var couplingString = [...]string{"DC", "DC", "AC", "DC", "HFR", "LFR", "NOISE"}
var slopeString = [...]string{"FALL", "RISE"}
var chanString = [...]string{"CH1", "CH2", "CH3", "CH4", "EXT", "EXT5", "EXT10", "AC LINE"}

// SetupTrigger will define scope trigger settings
func (s *Tps2000) SetupTrigger(sourceChan instr.Chan, coupling instr.Coupling, slope instr.Slope, trigLevel float64, auto bool, xPos float64) error {
	_ = s.Write("TRIG:MAIN:EDGE:COUP " + couplingString[coupling])
	_ = s.Write("TRIG:MAIN:EDGE:SLOPE " + slopeString[slope])
	_ = s.Write("TRIG:MAIN:EDGE:SOURCE " + slopeString[sourceChan])
	_ = s.Write("TRIG:MAIN:HOLDOFF:VALUE %0.3e", 0.02)
	_ = s.Write("TRIG:MAIN:LEVEL %0.4e", trigLevel)
	if auto {
		_ = s.Write("TRIG:MAIN:MODE AUTO", trigLevel)
	} else {
		_ = s.Write("TRIG:MAIN:MODE NORMAL", trigLevel)
	}
	err := s.Write("HOR:DELAY:POS %0.4e", xPos)
	return err
}

// ChannelCount is the maximum number of channels for this instrument
func (s *Tps2000) ChannelCount() int {
	return s.channelCount
}
