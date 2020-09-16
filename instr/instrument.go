// Package instrument contains the basic communication routines
// used by all instruments. USB, go-measure/serialSerial and TCP/IP is supported

// Copyright 2020 Jan Kåre Vatne. All rights reserved.

package instr

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jkvatne/serial"
)

// Chan is used for channel numbers
type Chan int

// Channel constants
const (
	TRIG Chan = iota
	Ch1
	Ch2
	Ch3
	Ch4
	Ch5
	EXT
	EXT5
	EXT10
	MAINS
)

type eol int

// Character constants for line terminators
const (
	Lf eol = iota
	Cr
	CrLf
	LfCr
	None
)

// Connection contains the local data for the connection to an instrument.
type Connection struct {
	Port     string
	Eol      eol                // Command string terminator
	Timeout  time.Duration      // Timeout on read operations
	Baudrate int                // Baudrate for serial ports
	Name     string             // Identifier read from the instrument by *IDN? or similar
	conn     io.ReadWriteCloser // Can be a net connection or a serial port
}

// Open will open a connection defined by portName
// The parameter is a TCP/IP address or a com port name
func (i *Connection) Open(portName string) error {
	var err error
	if i.Timeout == 0 {
		i.Timeout = time.Second
	}
	if strings.HasPrefix(portName, "COM") {
		if i.Baudrate == 0 {
			i.Baudrate = 115200
		}
		// Default to a interval timeout equal to one character length  (11 bits)
		c := &serial.Config{Name: portName, Baud: i.Baudrate, ReadTimeout: i.Timeout, IntervalTimeout: time.Duration(1e12 / i.Baudrate)}
		i.conn, err = serial.OpenPort(c)
	} else {
		i.conn, err = net.DialTimeout("tcp", portName, 1000*time.Millisecond)
	}
	if err != nil {
		return fmt.Errorf("could not connect to %s, error=%s", portName, err)
	}
	return nil
}

// Close will close a connection already opened
func (i *Connection) Close() {
	if i.conn != nil {
		_ = i.conn.Close()
		i.conn = nil
	}
}

// Write will send a commend to the instrument, adding end of line characters
func (i *Connection) Write(s string, args ...interface{}) error {
	if args != nil {
		s = fmt.Sprintf(s, args...)
	}
	if i.conn == nil {
		return fmt.Errorf("writing to invalic port")
	}
	b := []byte(i.addEol(s))
	if conn, ok := i.conn.(net.Conn); ok {
		_ = conn.SetWriteDeadline(time.Now().Add(i.Timeout))
	}
	n, err := i.conn.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("did not send all characters")
	}
	return nil
}

// ReadByte will return an array of bytes
func (i *Connection) ReadByte() byte {
	b := make([]byte, 1)
	_, _ = i.conn.Read(b)
	return b[0]
}

// ReadBinary will return an array of bytes
func (i *Connection) Read(b []byte) int {
	n, _ := i.conn.Read(b)
	return n
}

// ReadString will read any response from the instrument, with given timeout
func (i *Connection) ReadString() string {
	if i.conn == nil {
		return ""
	}
	if conn, ok := i.conn.(net.Conn); ok {
		_ = conn.SetReadDeadline(time.Now().Add(i.Timeout))
	}
	b := make([]byte, 1024)
	n, err := i.conn.Read(b)
	if n == 0 || err != nil {
		return ""
	}
	return ToString(b[0:n])
}

// SetTimeout sets the read timeout
func (i *Connection) SetTimeout(t time.Duration) {
	i.Timeout = t
}

// Flush will empty the read queue
func (i *Connection) Flush() {
	if c, ok := i.conn.(*serial.Port); ok {
		_ = c.Flush()
	}
	if c, ok := i.conn.(net.Conn); ok {
		b := make([]byte, 1024)
		_ = c.SetReadDeadline(time.Now().Add(time.Millisecond))
		_, _ = c.Read(b)
	}
}

// Ask will query the instrument for a string response
func (i *Connection) Ask(query string, args ...interface{}) (string, error) {
	i.Flush()
	if args != nil {
		query = fmt.Sprintf(query, args...)
	}
	query = i.addEol(query)
	err := i.Write(query)
	if err != nil {
		return "", err
	}
	response := i.ReadString()
	return response, nil
}

// PollFloat will read a float64 value
func (i *Connection) PollFloat(query string, args ...interface{}) (float64, error) {
	s, err := i.Ask(query, args...)
	if err != nil {
		return 0.0, err
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return f, nil
}

// CheckSerialPort takes a serial port name
// and test if it exists and are not used already.
// It returns nil if the port is ok, and an error message if not
func CheckSerialPort(port string) error {
	c := &serial.Config{Name: port, Baud: 115200, ReadTimeout: 100}
	p, e1 := serial.OpenPort(c)
	if p != nil {
		_ = p.Close()
	}
	if e1 != nil || p == nil {
		return fmt.Errorf("port \"%s\" fails (might be in use), %s", port, e1)
	}
	return nil
}

// FindSerialPort will return the name of the last (highest numbered)
// serial port that is not in use already
func FindSerialPort(id string, baudrate int, eol eol) string {
	list, desc, _ := EnumerateSerialPorts()
	highest := ""
	for i := len(list) - 1; i >= 0; i-- {
		if !strings.Contains(desc[i], "Bluetooth") && CheckSerialPort(list[i]) == nil {
			c := &Connection{Name: list[i], Baudrate: baudrate, Timeout: time.Second / 5, Eol: eol}
			err := c.Open(list[i])
			if err == nil {
				highest = list[i]
				resp, _ := c.Ask("*IDN?")
				c.Close()
				if strings.Contains(resp, id) {
					return list[i]
				}
			}
		}
	}
	// if no port responded to *IDN?, return highest available port
	return highest
}

// ToString converts a byte array to string and strips invalid characters
func ToString(buf []byte) string {
	s := ""
	for _, ch := range buf {
		if ch != '\000' && ch != '\n' && ch != '\r' {
			s = s + string(ch)
		}
	}
	return s
}

func (i *Connection) addEol(s string) string {
	switch i.Eol {
	case None:
		return s
	case Cr:
		if strings.HasSuffix(s, "\r") {
			return s
		}
		return s + "\r"
	case CrLf:
		if strings.HasSuffix(s, "\r\n") {
			return s
		}
		return s + "\r\n"
	case LfCr:
		if strings.HasSuffix(s, "\n\r") {
			return s
		}
		return s + "\n\r"
	default:
		if strings.HasSuffix(s, "\n") {
			return s
		}
		return s + "\n"
	}
}

// QueryIdn will read the instrument identification string
// It uses *IDN? which most instruments implement
func (i *Connection) QueryIdn() (string, error) {
	name, err := i.Ask("*IDN?")
	if name == "" {
		time.Sleep(100 * time.Millisecond)
		name, err = i.Ask("*IDN?")
	}
	if err != nil {
		return "", err
	}
	i.Name = name
	return name, nil
}
