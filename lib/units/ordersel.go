package units

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
)

// OrderSelector implements a 2 of 12 digit multiplexer/counter used for
// selecting instructions from function table rows, described in Clippinger '48
// "A Logical Coding System Applied to the ENIAC".
//
// Clippinger does not describe the order selector in detail, but we can infer
// from the timing chart (Figure 2.2) that there is an "i" control input to
// gate transmitting output and a separate input/output for advancing the ring
// counter.  Also we can infer there must be a way to reset it, otherwise it
// would be impossible to ensure a jump to the start of an FT line.  This
// suggests the following theory of operation.
//
// When enabled by os.i, 2 of 12 input digits from os.A and os.B are routed to
// the two lowest order digits of os.o - which two is controlled by a
// 6-position ring counter.  The counter is clearable by os.Ci, steppable by
// os.Ri, emitting os.Ro on overflow.  Since the counter is incremented before
// data returns in the fetch cycle, assume selection is wired to pass the
// "previous" two digit positions.
//
// Even money says like the FT selector, this was built on another spare
// 6-stage MP stepper with some gate tubes attached to the stage outputs.
type OrderSelector struct {
	a, b, out       *Jack
	en              *Jack
	ringIn, ringOut *Jack
	ringClear       *Jack
	ring            int

	enff1, enff2 bool
	rff1, rff2   bool
	afterFirstRp bool
}

func NewOrderSelector() *OrderSelector {
	u := &OrderSelector{}
	// Actually cross-wire digits, instead of trying to reset ring to -1 or 5 or
	// somesuch, to model what would have happened had Ri not been stepped.
	u.a = NewInput("os.A", func(j *Jack, val int) {
		if u.enff2 {
			switch u.ring {
			case 1:
				u.out.Transmit((val >> 4) & 3)
			case 2:
				u.out.Transmit((val >> 2) & 3)
			case 3:
				u.out.Transmit(val & 3)
			}
		}
	})
	u.b = NewInput("os.B", func(j *Jack, val int) {
		if u.enff2 {
			switch u.ring {
			case 4:
				u.out.Transmit((val >> 4) & 3)
			case 5:
				u.out.Transmit((val >> 2) & 3)
			case 0:
				u.out.Transmit(val & 3)
			}
		}
	})
	u.out = NewOutput("os.o", nil)
	u.en = NewInput("os.i", func(j *Jack, val int) {
		u.enff1 = true
	})
	u.ringClear = NewInput("os.Ci", func(*Jack, int) {
		u.ring = 0
	})
	u.ringIn = NewInput("os.Ri", func(*Jack, int) {
		u.rff1 = true
	})
	u.ringOut = NewOutput("os.Ro", nil)
	return u
}

func (u *OrderSelector) Reset() {
	u.ring = 0
	u.enff1 = false
	u.enff2 = false
	u.rff1 = false
	u.rff2 = false
	u.afterFirstRp = false
}

func (u *OrderSelector) Clock(p Pulse) {
	if p&Cpp != 0 {
		if u.rff2 {
			u.ring++
			if u.ring == 6 {
				// Pulse when the ring overflows (to increment PC).
				u.ringOut.Transmit(1)
				u.ring = 0
			}
		}
	} else if p&Rp != 0 {
		if u.afterFirstRp {
			// NB this either sets or clears *ff2
			u.rff2 = u.rff1
			u.rff1 = false
			u.enff2 = u.enff1
			u.enff1 = false
		}
		u.afterFirstRp = !u.afterFirstRp
	}
}

func (u *OrderSelector) FindJack(name string) (*Jack, error) {
	switch name {
	case "A", "a":
		return u.a, nil
	case "B", "b":
		return u.b, nil
	case "o":
		return u.out, nil
	case "i":
		return u.en, nil
	case "Ci", "ci":
		return u.ringClear, nil
	case "Ri", "ri":
		return u.ringIn, nil
	case "Ro", "ro":
		return u.ringOut, nil
	}
	return nil, fmt.Errorf("invalid jack %s", name)
}
