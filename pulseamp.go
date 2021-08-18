package main

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/jeredw/eniacsim/lib"
)

// PulseAmps models ENIAC's pulse amplifiers, eight, 11-wide buffered line
// drivers.  Pulse amplifiers are useful to establish directional connections
// between trunks.
//
// Typical usage wires both the inputs and outputs of a given pulse amplifier
// to 11 program lines or to 11 digit lines.  Physically, though, pulse
// amplifiers were boxes with two 12-pin terminals either of which could be
// connected to either digit or program lines with suitable cabling.
//
// For efficiency, this implementation does not support mixing digit and program
// connections to pulse amplifiers - explicit adapters must be used.
type PulseAmps struct {
	digitInput  [8]*Jack
	digitOutput [8]*Jack
	progInput   [8][11]*Jack
	progOutput  [8][11]*Jack
}

func NewPulseAmps() *PulseAmps {
	pa := &PulseAmps{}
	for unit := 0; unit < 8; unit++ {
		sa := fmt.Sprintf("pa.%d.sa", unit+1)
		sb := fmt.Sprintf("pa.%d.sb", unit+1)
		pa.digitInput[unit] = NewRoutingJack(sa, 1)
		pa.digitOutput[unit] = NewRoutingJack(sb, 2)
		pa.digitInput[unit].Receivers = append(pa.digitInput[unit].Receivers, pa.digitOutput[unit])
		for i := 0; i < 11; i++ {
			sa := fmt.Sprintf("pa.%d.sa.%d", unit+1, i+1)
			sb := fmt.Sprintf("pa.%d.sb.%d", unit+1, i+1)
			pa.progInput[unit][i] = NewRoutingJack(sa, 1)
			pa.progOutput[unit][i] = NewRoutingJack(sb, 2)
			pa.progInput[unit][i].Receivers = append(pa.progInput[unit][i].Receivers, pa.progOutput[unit][i])
		}
	}
	return pa
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
	if len(p) != 3 {
		switch p[1] {
		case "sa":
			return pa.digitInput[unit], nil
		case "sb":
			return pa.digitOutput[unit], nil
		}
		return nil, fmt.Errorf("invalid jack %s", name)
	}
	port, err := strconv.Atoi(p[2])
	if err != nil || !(port >= 1 && port <= 11) {
		return nil, fmt.Errorf("invalid jack %s", name)
	}
	port -= 1
	switch p[1] {
	case "sa":
		return pa.progInput[unit][port], nil
	case "sb":
		return pa.progOutput[unit][port], nil
	}
	return nil, fmt.Errorf("invalid jack %s", name)
}
