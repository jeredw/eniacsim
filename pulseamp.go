package main

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
	"strings"
)

// PulseAmps models ENIAC's pulse amplifiers, eight, 11-wide buffered line
// drivers.  Pulse amplifiers are useful to establish directional connections
// between trunks.
//
// Typical usage wires both the inputs and outputs of a given pulse amplifier
// to 11 program lines or to 11 digit lines.  Physically, though, pulse
// amplifiers were boxes with two 12-pin terminals either of which could be
// connected to either digit or program lines with suitable cabling.  So this
// implementation supports connecting a digit output to input jack "sa" and
// parting out individual pulses on output "sb", or composing separate pulses
// on input jack "sa" onto one digit output jack "sb".  In principle this might
// be used to replace digit-pulse adapters.
type PulseAmps struct {
	// Port 0 is an 11-wide digit connection with fan-in/out from individual
	// terminals on ports 1-11.
	input  [8][12]*Jack
	output [8][12]*Jack
}

func NewPulseAmps() *PulseAmps {
	pa := &PulseAmps{}
	for unit := 0; unit < 8; unit++ {
		for port := 0; port < 12; port++ {
			var sa, sb string
			if port == 0 {
				sa = fmt.Sprintf("pa.%d.sa", unit+1)
				sb = fmt.Sprintf("pa.%d.sb", unit+1)
			} else {
				sa = fmt.Sprintf("pa.%d.sa.%d", unit+1, port)
				sb = fmt.Sprintf("pa.%d.sb.%d", unit+1, port)
			}
			pa.input[unit][port] = pa.newInput(sa, unit, port)
			pa.output[unit][port] = NewOutput(sb, nil)
		}
	}
	return pa
}

func (pa *PulseAmps) newInput(name string, unit, port int) *Jack {
	return NewInput(name, func(j *Jack, val int) {
		pa.output[unit][port].Transmit(val)
		if port == 0 {
			// Fanout the 11-wide digit input to individual outputs.
			for i := 0; i < 11; i++ {
				if val&(1<<i) != 0 {
					pa.output[unit][i+1].Transmit(1)
				}
			}
		} else {
			// Fanin input to the appropriate pin of the 11-wide digit output.
			pa.output[unit][0].Transmit(1 << (port - 1))
		}
	})
}

func (pa *PulseAmps) FindJack(name string) (*Jack, error) {
	// (pa.)%d.s[ab]{.%d}
	p := strings.Split(name, ".")
	if len(p) != 2 && len(p) != 3 {
		return nil, fmt.Errorf("invalid jack %s", name)
	}
	unit, err := strconv.Atoi(p[0])
	if err != nil || !(unit >= 1 && unit <= 8) {
		return nil, fmt.Errorf("invalid jack %s", name)
	}
	unit--
	port := 0
	if len(p) == 3 {
		port, err = strconv.Atoi(p[2])
		if err != nil || !(port >= 1 && port <= 11) {
			return nil, fmt.Errorf("invalid jack %s", name)
		}
	}
	switch p[1] {
	case "sa":
		return pa.input[unit][port], nil
	case "sb":
		return pa.output[unit][port], nil
	}
	return nil, fmt.Errorf("invalid jack %s", name)
}
