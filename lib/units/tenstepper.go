package units

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
)

// TenStepper simulates the ten-stage stepper described in Clippinger '48 "A
// Logical Coding System Applied to the ENIAC".
//
// It's essentially a separate bolted on 10-stage version of a 6-stage MP
// stepper.  This model assumes it has the same I/O jacks but no "clear" set
// switch, since all ten positions are probably going to be used.
type TenStepper struct {
	i, di, cdi      *Jack
	o               [10]*Jack
	inff1, inff2    bool
	waitForNextTenp bool
	afterFirstRp    bool
	stage           int
}

func NewTenStepper() *TenStepper {
	u := &TenStepper{}
	u.i = NewInput("ts.i", func(*Jack, int) {
		u.inff1 = true
	})
	u.di = NewInput("ts.di", func(*Jack, int) {
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
	u.cdi = NewInput("ts.cdi", func(*Jack, int) {
		u.stage = 0
	})
	for i := 0; i < 10; i++ {
		u.o[i] = NewOutput(fmt.Sprintf("ts.o%d", i+1), nil)
	}
	return u
}

func (u *TenStepper) Reset() {
	u.i.Disconnect()
	u.di.Disconnect()
	u.cdi.Disconnect()
	for i := 0; i < 10; i++ {
		u.o[i].Disconnect()
	}
	u.inff1 = false
	u.inff2 = false
	u.waitForNextTenp = false
	u.afterFirstRp = false
	u.stage = 0
}

func (u *TenStepper) Clock(p Pulse) {
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

func (u *TenStepper) FindJack(name string) (*Jack, error) {
	switch {
	case name == "cdi":
		return u.cdi, nil
	case name == "di":
		return u.di, nil
	case name == "i":
		return u.i, nil
	case len(name) >= 2 && name[len(name)-1] == 'o':
		output, err := strconv.Atoi(name[:len(name)-1])
		if err == nil && (output >= 1 && output <= 10) {
			return u.o[output-1], nil
		}
	}
	return nil, fmt.Errorf("invalid jack %s", name)
}
