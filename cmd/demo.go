package main

import (
	"fmt"
	"time"

	"github.com/jkvatne/go-measure/dmm/fluke"

	"github.com/jkvatne/go-measure/ad2"
	"github.com/jkvatne/go-measure/psu/cpx400"
	"github.com/jkvatne/go-measure/psu/korad"
	"github.com/jkvatne/go-measure/tds2000"
)

func main() {

	dmm, err := fluke.New("192.168.2.110:3490")
	if err == nil && dmm != nil {
		name, _ := dmm.QueryIdn()
		fmt.Printf("Found " + name + "\n")
		dmm.Close()
	} else {
		fmt.Printf("No Fluke multimeter found\n")
	}

	a, err := ad2.New()
	if err == nil {
		name, _ := a.QueryIdn()
		fmt.Printf("Found %s\n", name)
	}
	fmt.Printf("Testing connection to TTi CPX400DP supply\n")
	p, err := cpx400.New("192.168.2.18:9221")
	if err == nil && p != nil {
		name, _ := p.QueryIdn()
		fmt.Printf("Found %s\n", name)
		_ = p.SetOutput(1, 1.0, 0.1)
		time.Sleep(500 * time.Millisecond)
		u, i, _ := p.GetOutput(1)
		fmt.Printf("Readback %0.3fV, %0.3fA\n", u, i)
	} else {
		fmt.Printf("Error: %s\n", err.Error())
	}

	o, err := tds2000.New("COM11")
	if err == nil && o != nil {
		fmt.Printf("Found " + o.GetName() + "\n")
	} else {
		fmt.Printf("No Osciloscope found\n")
	}

	q, err := korad.New("COM12")
	if err == nil && q != nil {
		name, _ := q.QueryIdn()
		fmt.Printf("Found " + name + "\n")
	} else {
		fmt.Printf("No Korad power supply found\n")
	}
	/*
		m, err := manualpsu.NewManualPsu(os.Stdin, os.Stdout)
		if err == nil && m != nil {
			name, _ := m.QueryIdn()
			fmt.Printf("Connection name: %s\n", name)
			_ = m.SetOutput(1, 1.0, 0.1)
			time.Sleep(500 * time.Millisecond)
			u, i, _ := m.GetOutput(1)
			fmt.Printf("Readback %0.3fV, %0.3fA\n", u, i)
		} else {
			fmt.Printf("Error: %s\n", err.Error())
		}
	*/
	fmt.Printf("Done\n")
}
