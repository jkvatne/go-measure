package bm25x

import (
	"fmt"
	"sync"
	"time"

	"github.com/jkvatne/go-measure/instr"
)

// Check if tti interface satisfies Psu interface
var _ instr.Dmm = &Lcd{}

// Lcd stores setup for a BM... multimeter
type Lcd struct {
	instr.Connection
	terminated   bool
	Ok           bool
	CurrentValue float64
	CurrentUnit  string
	CurrentError string
	buf          [16]byte
	mutex        sync.Mutex
}

// QueryIdn returns name
func (dmm *Lcd) QueryIdn() (string, error) {
	dmm.mutex.Lock()
	defer dmm.mutex.Unlock()
	if dmm.Ok {
		return "Lcd multimeter BM25x is online and ok", nil
	}
	return "No data recieved", fmt.Errorf("No data recieved")
}

// Configure implements interface
func (dmm *Lcd) Configure(s instr.Setup) error {
	return nil
}

// Measure returns measured value
func (dmm *Lcd) Measure() (float64, error) {
	dmm.mutex.Lock()
	defer dmm.mutex.Unlock()
	if !dmm.Ok {
		return 0.0, fmt.Errorf(dmm.CurrentError)
	}
	return dmm.CurrentValue, nil
}

func (dmm *Lcd) update(buf []byte, n int) {
	dmm.mutex.Lock()
	defer dmm.mutex.Unlock()
	if n == 15 {
		dmm.Ok = true
		dmm.CurrentValue = dmm.decode(buf)
		dmm.CurrentUnit = unit(buf)
		//fmt.Printf("%0.3f%s\n", dmm.CurrentValue, dmm.CurrentUnit)
	} else {
		dmm.Ok = false
		dmm.CurrentError = "no data received"
	}
}

func (dmm *Lcd) background() {
	for dmm.terminated == false {
		buf := make([]byte, 16)
		n := dmm.Read(buf)
		dmm.update(buf, n)
	}
}

func (dmm *Lcd) isOk() bool {
	dmm.mutex.Lock()
	defer dmm.mutex.Unlock()
	return dmm.Ok
}

// New will return an instrument instance
func New(port string) (*Lcd, error) {
	dmm := &Lcd{}
	dmm.mutex.Lock()
	dmm.Port = port
	dmm.Timeout = 1000 * time.Millisecond
	dmm.Baudrate = 9600
	dmm.Eol = instr.None
	err := dmm.Open(port)
	if err != nil {
		return nil, fmt.Errorf("error opening port, %s", err)
	}
	dmm.mutex.Unlock()
	go dmm.background()
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 10)
		if dmm.isOk() {
			break
		}
	}
	return dmm, nil
}

// Close will terminate go routine
func (dmm *Lcd) Close() {
	dmm.mutex.Lock()
	defer dmm.mutex.Unlock()
	dmm.terminated = true
	dmm.Connection.Close()
}

func toDigt(b1, b2 byte) int {
	b1 = b1 & 0x0E
	b2 = b2 & 0x0F
	switch {
	case b1 == 0x00 && b2 == 0x0A:
		return 1
	case b1 == 0x0A && b2 == 0x0D:
		return 2
	case b1 == 0x08 && b2 == 0x0F:
		return 3
	case b1 == 0x04 && b2 == 0x0e:
		return 4
	case b1 == 0x0c && b2 == 0x07:
		return 5
	case b1 == 0x0e && b2 == 0x07:
		return 6
	case b1 == 0x08 && b2 == 0x0a:
		return 7
	case b1 == 0x0e && b2 == 0x0F:
		return 8
	case b1 == 0x0c && b2 == 0x0F:
		return 9
	case b1 == 0x0e && b2 == 0x0b:
		return 0
	}
	return -1
}

func exp(buf []byte) float64 {
	e := 1.0
	if buf[9]&1 != 0 {
		e = 10.0
	}
	if buf[7]&1 != 0 {
		e = 100.0
	}
	if buf[5]&1 != 0 {
		e = 1000.0
	}
	if buf[3]&1 != 0 {
		e = -e
	}
	return e
}

func unit(buf []byte) string {
	u := ""
	if buf[14]&0x04 != 0 {
		u = "V"
	}
	if buf[12]&0x02 != 0 && buf[11]&0x01 != 0 {
		u = "kHz"
	}
	if buf[12]&0x02 != 0 {
		u = "Hz"
	}
	if buf[12]&0x04 != 0 {
		u = "ohm"
	}
	if buf[13]&0x04 != 0 {
		u = "F"
	}
	if buf[14]&0x02 != 0 {
		u = "A"
	}
	if buf[11]&0x04 != 0 {
		u = "dBm"
	}
	if buf[1]&0x02 != 0 {
		u = u + "ac"
	}
	if buf[1]&0x04 != 0 {
		u = u + "dc"
	}
	// Prefix n/u/m/k/M
	if buf[12]&0x01 != 0 {
		u = "n" + u
	}
	if buf[13]&0x02 != 0 {
		u = "u" + u
	}
	if buf[13]&0x01 != 0 {
		u = "m" + u
	}
	if buf[11]&0x01 != 0 {
		u = "k" + u
	}
	if buf[11]&0x02 != 0 {
		u = "M" + u
	}
	if buf[9] == 0x0E && buf[10] == 0x01 {
		u = "C"
	}
	if buf[9] == 0x0E && buf[10] == 0x04 {
		u = "F"
	}
	return u
}

func (dmm *Lcd) decode(buf []byte) float64 {
	if buf[0] != 0x02 {
		dmm.CurrentError = "message format error"
		return 99999.9
	}
	if buf[11]&0x08 != 0 || buf[12]&0x08 != 0 || buf[13]&0x08 != 0 || buf[14]&0x08 != 0 || buf[1]&0x01 != 0 {
		// Hold, Crest, Min or Max - invalid
		dmm.CurrentError = "hold/crest/min/max not allowed"
		return 99999.9
	}
	d1 := toDigt(buf[3], buf[4])
	d2 := toDigt(buf[5], buf[6])
	d3 := toDigt(buf[7], buf[8])
	d4 := toDigt(buf[9], buf[10])
	if d4 < 0 {
		if d1 < 0 {
			dmm.CurrentError = "could not decode display"
			return 99999.9
		}
		// Special case for temperature where d4 is C or F
		return float64(d1*100 + d2*10 + d3)
	}
	v := float64(d1*1000+d2*100+d3*10+d4) / exp(buf)
	return v
}
