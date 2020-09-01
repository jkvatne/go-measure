// Package alog extends the default log package with levels and other features
// Program command line arguments can be used to override defaults
// Logging level is taken from the flag "-level=2"
// Format is taken from the flag "-format=3"
// File is taken from the flag "-logfile=stdout"
package alog

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
)

const (
	//NoneLevel is used to suppress all messages
	NoneLevel = 0
	//ErrorLevel is used to show only errors
	ErrorLevel = 1
	//WarningLevel is used to show only warnings and errors
	WarningLevel = 2
	//InfoLevel is used to show Info,Warning and errors
	InfoLevel = 3
	//DebugLevel is used to show all messages
	DebugLevel = 4
	//AllLevels is just a spare value in case of more levels.
	AllLevels = 99
)

var (
	// D is the debug logger
	d *log.Logger
	// I is the info logger
	i *log.Logger
	// W is the warning logger
	w *log.Logger
	// E is the error logger
	e *log.Logger
	// Variables for flags
	outFile     string
	level       int
	format      int
	initialized bool
)

func createLoggers(logTo io.Writer, level int, format int) {
	if level > 0 {
		e = log.New(logTo, "E ", format)
	} else {
		e = log.New(ioutil.Discard, "E ", format)
	}
	if level > 1 {
		w = log.New(logTo, "W ", format)
	} else {
		w = log.New(ioutil.Discard, "W ", format)
	}
	if level > 2 {
		i = log.New(logTo, "I ", format)
	} else {
		i = log.New(ioutil.Discard, "I ", format)
	}
	if level > 3 {
		d = log.New(logTo, "D ", format)
	} else {
		d = log.New(ioutil.Discard, "D ", format)
	}
}

// Setup can be called with different default levels
// The levels given will be overridden by command line arguments
func Setup(defaultWriter io.Writer, defaultLevel int, defaultFormat int) {
	var logTo io.Writer
	var err error
	flag.Parse()
	// Set to default values if not given by flags
	if format < 0 {
		format = defaultFormat
	}
	if level < 0 {
		level = defaultLevel
	}
	if outFile == "" {
		logTo = defaultWriter
	} else if outFile == "stdout" {
		logTo = os.Stdout
	} else if outFile == "stderr" {
		logTo = os.Stderr
	} else {
		// Open file
		logTo, err = os.OpenFile(outFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logTo = os.Stdout
		}
	}

	createLoggers(logTo, level, format)

}

// Fatal is used to generate exit program with fatal error
func Fatal(s string, args ...interface{}) {
	e.Printf(fmt.Sprintf(s, args...))
	_, file, line, ok := runtime.Caller(2)
	if ok {
		e.Printf(fmt.Sprintf("Fatal error on line %d, file: %s", line, file))
	}
	os.Exit(1)
}

func checkSetup() {
	if !initialized {
		initialized = true
		Setup(os.Stdout, ErrorLevel, log.Ldate|log.Ltime)
	}
}

// Debug is for debugging messages
func Debug(s string, args ...interface{}) {
	checkSetup()
	d.Printf(fmt.Sprintf(s, args...))
}

// Info is for insignificant info
func Info(s string, args ...interface{}) {
	checkSetup()
	i.Printf(fmt.Sprintf(s, args...))
}

// Warn is for significant warnings
func Warn(s string, args ...interface{}) {
	checkSetup()
	w.Printf(fmt.Sprintf(s, args...))
}

// Error is for serious errors
func Error(s string, args ...interface{}) {
	checkSetup()
	e.Printf(fmt.Sprintf(s, args...))
}

func init() {
	// Setup command line variables. They will override defaults
	// Note that flag.Parse() MUST be called at start of main(), possibly after defining more flags.
	flag.IntVar(&level, "level", 1, "Select logging level, 0 for no logging, 5 for all messages")
	flag.IntVar(&format, "format", 3, "Select logging format, 1=Date, 2=Time, 4=MicroSec, 16=Filename")
	flag.StringVar(&outFile, "logfile", "", "Specify output file")
}

// Level returns the current logging level
func Level() int {
	return level
}
