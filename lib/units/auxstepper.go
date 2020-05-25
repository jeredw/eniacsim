package units

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
)

// AuxStepper simulates the additional six and ten stage steppers described in
// Clippinger '48 "A Logical Coding System Applied to the ENIAC".
//
// These are essentially two, bolted on versions of a 6-stage MP stepper:  one
// with 6 stages for FT selection, and and one with ten stages for decode.
// This model assumes it has the same I/O jacks but no "clear" set switch.
type AuxStepper struct {
	i, di, cdi      *Jack
	o               []*Jack
	inff1, inff2    bool
	waitForNextTenp bool
	afterFirstRp    bool
	stage           int
	steps           int
}

func NewAuxStepper(name string, steps int) *AuxStepper {
	u := &AuxStepper{steps: steps}
	u.i = NewInput(fmt.Sprintf("%s.i", name), func(*Jack, int) {
		u.inff1 = true
	})
	u.di = NewInput(fmt.Sprintf("%s.di", name), func(*Jack, int) {
		// Only count once per 10P (ignore 9Ps and non-digit pulses).
		if u.waitForNextTenp {
			return
		}
		u.stage++
		if u.stage == 10 {
			u.stage = 0
		}
		u.waitForNextTenp = true
	})
	u.cdi = NewInput(fmt.Sprintf("%s.cdi", name), func(*Jack, int) {
		u.stage = 0
	})
	u.o = make([]*Jack, steps)
	for i := 0; i < steps; i++ {
		u.o[i] = NewOutput(fmt.Sprintf("%s.o%d", name, i+1), nil)
	}
	return u
}

func (u *AuxStepper) Reset() {
	u.i.Disconnect()
	u.di.Disconnect()
	u.cdi.Disconnect()
	for i := range u.o {
		u.o[i].Disconnect()
	}
	u.inff1 = false
	u.inff2 = false
	u.waitForNextTenp = false
	u.afterFirstRp = false
	u.stage = 0
}

func (u *AuxStepper) Clock(p Pulse) {
	if p&Tenp != 0 {
		u.waitForNextTenp = false
	} else if p&Cpp != 0 {
		if u.inff2 {
			u.o[u.stage].Transmit(1)
			u.inff2 = false
		}
	} else if p&Rp != 0 {
		if u.afterFirstRp {
			u.inff2 = u.inff1
			u.inff1 = false
		}
		u.afterFirstRp = !u.afterFirstRp
	}
}

func (u *AuxStepper) FindJack(name string) (*Jack, error) {
	switch {
	case name == "cdi":
		return u.cdi, nil
	case name == "di":
		return u.di, nil
	case name == "i":
		return u.i, nil
	case len(name) >= 2 && name[len(name)-1] == 'o':
		output, err := strconv.Atoi(name[:len(name)-1])
		if err == nil && (output >= 1 && output <= u.steps) {
			return u.o[output-1], nil
		}
	}
	return nil, fmt.Errorf("invalid jack %s", name)
}
