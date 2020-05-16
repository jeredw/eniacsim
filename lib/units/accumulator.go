package units

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

/*
 * Bit positions for the ST1 and ST2 connectors
 */
const (
	pin1 = 1 << iota
	pin2
	pin3
	pin4
	pin5
	pin6
	pin7
	pin8
	pin9
	pin10
	_
	_
	_
	pin14
	pin15
	pin16
	pin17
)

/*
 * Signal names on ST1 and ST2 connectors
 */
const (
	stα        = pin1
	stβ        = pin2
	stγ        = pin3
	stδ        = pin4
	stε        = pin5
	stA        = pin6
	stAS       = pin7
	stS        = pin8
	stCLR      = pin9
	stCORR     = pin10
	stPMinp    = pin14
	stCORRsrc  = pin14
	stDEC10CAR = pin15
	stDEC1inp  = pin15
	stSF0out   = pin16
	stDEC1sub  = pin16
	stSFSWinp  = pin17
	stSF10out  = pin17
)

// Simulates an ENIAC accumulator unit.
type Accumulator struct {
	Io AccumulatorConn

	unit                int
	α, β, γ, δ, ε, A, S *Jack
	ctlterm             [20]*Jack
	inff1, inff2        [12]bool
	opsw                [12]int
	clrsw               [12]bool
	rptsw               [8]int
	sigfig              int
	sc                  byte
	val                 [10]byte
	decff               [10]bool
	sign                bool
	h50                 bool
	rep                 int
	whichrp             bool
	lbuddy, rbuddy      int
	plbuddy, prbuddy    *Accumulator
	su1Cache            int

	tracePulse TraceFunc

	mu sync.Mutex
}

// Connections to other units.
type AccumulatorConn struct {
	Sv    func() int
	Su2   func() int
	Su3   func() int
	Multl func() bool
	Multr func() bool
}

func NewAccumulator(unit int) *Accumulator {
	unitDot := fmt.Sprintf("a%d.", unit+1)
	u := &Accumulator{
		unit:   unit,
		sigfig: 10,
		lbuddy: unit,
		rbuddy: unit,
	}
	digitInput := func(st1Pin int) JackHandler {
		return func(j *Jack, val int) {
			u.mu.Lock()
			defer u.mu.Unlock()
			if u.st1()&st1Pin != 0 {
				u.receive(val)
				if u.tracePulse != nil {
					u.tracePulse(j.Name, 11, int64(val))
				}
			}
		}
	}
	programInput := func(prog int) JackHandler {
		return func(j *Jack, val int) {
			if val == 1 {
				u.trigger(prog)
				if u.tracePulse != nil {
					u.tracePulse(j.Name, 1, int64(val))
				}
			}
		}
	}
	output := func(width int) JackHandler {
		return func(j *Jack, val int) {
			if u.tracePulse != nil {
				u.tracePulse(j.Name, width, int64(val))
			}
		}
	}
	u.α = NewInput(unitDot+"α", digitInput(stα))
	u.β = NewInput(unitDot+"β", digitInput(stβ))
	u.γ = NewInput(unitDot+"γ", digitInput(stγ))
	u.δ = NewInput(unitDot+"δ", digitInput(stδ))
	u.ε = NewInput(unitDot+"ε", digitInput(stε))
	u.A = NewOutput(unitDot+"A", output(11))
	u.S = NewOutput(unitDot+"S", output(11))
	u.ctlterm[0] = NewInput(unitDot+"1i", programInput(0))
	u.ctlterm[1] = NewInput(unitDot+"2i", programInput(1))
	u.ctlterm[2] = NewInput(unitDot+"3i", programInput(2))
	u.ctlterm[3] = NewInput(unitDot+"4i", programInput(3))
	u.ctlterm[4] = NewInput(unitDot+"5i", programInput(4))
	u.ctlterm[5] = NewOutput(unitDot+"5o", output(1))
	u.ctlterm[6] = NewInput(unitDot+"6i", programInput(5))
	u.ctlterm[7] = NewOutput(unitDot+"6o", output(1))
	u.ctlterm[8] = NewInput(unitDot+"7i", programInput(6))
	u.ctlterm[9] = NewOutput(unitDot+"7o", output(1))
	u.ctlterm[10] = NewInput(unitDot+"8i", programInput(7))
	u.ctlterm[11] = NewOutput(unitDot+"8o", output(1))
	u.ctlterm[12] = NewInput(unitDot+"9i", programInput(8))
	u.ctlterm[13] = NewOutput(unitDot+"9o", output(1))
	u.ctlterm[14] = NewInput(unitDot+"10i", programInput(9))
	u.ctlterm[15] = NewOutput(unitDot+"10o", output(1))
	u.ctlterm[16] = NewInput(unitDot+"11i", programInput(10))
	u.ctlterm[17] = NewOutput(unitDot+"11o", output(1))
	u.ctlterm[18] = NewInput(unitDot+"12i", programInput(11))
	u.ctlterm[19] = NewOutput(unitDot+"12o", output(1))
	return u
}

func (u *Accumulator) Sign() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.sign {
		return "M"
	}
	return "P"
}

func (u *Accumulator) Value() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	var s string
	if u.sign {
		s += "M "
	} else {
		s += "P "
	}
	for i := 9; i >= 0; i-- {
		s += fmt.Sprintf("%d", u.val[i])
	}
	return s
}

func (u *Accumulator) Stat() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	var s string
	if u.sign {
		s += "M "
	} else {
		s += "P "
	}
	for i := 9; i >= 0; i-- {
		s += fmt.Sprintf("%d", u.val[i])
	}
	s += " "
	for i := 9; i >= 0; i-- {
		s += ToBin(u.decff[i])
	}
	s += fmt.Sprintf(" %d ", u.rep)
	for _, f := range u.inff1 {
		s += ToBin(f)
	}
	return s
}

type accJson struct {
	Sign    bool     `json:"sign"`
	Decade  [10]int  `json:"decade"`
	Decff   [10]bool `json:"decff"`
	Repeat  int      `json:"repeat"`
	Program [12]bool `json:"program"`
}

func (u *Accumulator) State() json.RawMessage {
	u.mu.Lock()
	defer u.mu.Unlock()
	digits := [10]int{}
	for i := range u.val {
		digits[i] = int(u.val[i])
	}
	s := accJson{
		Sign:    u.sign,
		Decade:  digits,
		Decff:   u.decff,
		Repeat:  u.rep,
		Program: u.inff1,
	}
	result, _ := json.Marshal(s)
	return result
}

func (u *Accumulator) AttachTrace(tracePulse TraceFunc) []func(TraceFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.tracePulse = tracePulse
	sign := fmt.Sprintf("a%d.sign", u.unit+1)
	value := fmt.Sprintf("a%d.value", u.unit+1)
	return []func(t TraceFunc){
		func(traceReg TraceFunc) {
			s := int64(0)
			if u.sign {
				s = 1
			}
			traceReg(sign, 1, s)
		},
		func(traceReg TraceFunc) {
			var n int64
			for i := 9; i >= 0; i-- {
				n <<= 4
				n += int64(u.val[i])
			}
			traceReg(value, 40, n)
		},
	}
}

func (u *Accumulator) Reset() {
	u.mu.Lock()
	u.α.Disconnect()
	u.β.Disconnect()
	u.γ.Disconnect()
	u.δ.Disconnect()
	u.ε.Disconnect()
	for i := 0; i < 12; i++ {
		u.ctlterm[i].Disconnect()
		u.inff1[i] = false
		u.inff2[i] = false
		u.opsw[i] = stα
		u.clrsw[i] = false
	}
	for i := 0; i < 8; i++ {
		u.rptsw[i] = 0
	}
	u.sigfig = 10
	u.sc = 0
	u.h50 = false
	u.rep = 0
	u.whichrp = false
	u.lbuddy = u.unit
	u.rbuddy = u.unit
	u.mu.Unlock()
	u.Clear()
}

func (u *Accumulator) Clear() {
	u.mu.Lock()
	defer u.mu.Unlock()
	for i := 0; i < 10; i++ {
		u.val[i] = 0
		u.decff[i] = false
	}
	if u.sigfig < 10 {
		u.val[9-u.sigfig] = 5
	}
	u.sign = false
}

func (u *Accumulator) Set(value int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.sign = value < 0
	if value < 0 {
		value = -value
	}
	for i := 0; i < 10; i++ {
		u.val[i] = byte(value % 10)
		u.decff[i] = false
		value /= 10
	}
}

// Interconnect connects accumulators with other units statically.
// FIXME This seems to be totally unused/untested.
func Interconnect(units [20]*Accumulator, p1 []string, p2 []string) error {
	unit1, _ := strconv.Atoi(p1[0][1:])
	unit2 := -1
	if len(p2) > 1 && p2[0][0] == 'a' {
		unit2, _ = strconv.Atoi(p2[0][1:])
	}
	switch {
	case p2[0] == "m" && p2[1] == "l":
		units[unit1-1].lbuddy = -1
	case p2[0] == "m" && p2[1] == "r":
		units[unit1-1].lbuddy = -2
	case p2[0] == "d" && p2[1] == "sv":
		units[unit1-1].lbuddy = -3
	case p2[0] == "d" && p2[1] == "su2q":
		units[unit1-1].lbuddy = -4
	case p2[0] == "d" && p2[1] == "su2s":
		units[unit1-1].lbuddy = -5
	case p2[0] == "d" && p2[1] == "su3":
		units[unit1-1].lbuddy = -6
	case p1[1] == "st1" || p1[1] == "il1":
		if unit2 != -1 && unit1 != unit2 {
			units[unit1-1].lbuddy = unit2 - 1
			units[unit2-1].rbuddy = unit1 - 1
		}
	case p1[1] == "st2" || p1[1] == "ir1":
		if unit2 != -1 && unit1 != unit2 {
			units[unit1-1].rbuddy = unit2 - 1
			units[unit2-1].lbuddy = unit1 - 1
		}
	case p1[1] == "su1" || p1[1] == "il2":
		if unit2 != -1 && unit1 != unit2 {
			units[unit1-1].lbuddy = unit2 - 1
			units[unit2].rbuddy = unit1 - 1
		}
	case p1[1] == "su2" || p1[1] == "ir2":
		if unit2 != -1 && unit1 != unit2 {
			units[unit1-1].rbuddy = unit2 - 1
			units[unit2-1].lbuddy = unit1 - 1
		}
	}
	if unit2 != -1 && unit1 != unit2 {
		//		units[unit1-1].change <- 1
		//		units[unit2-1].change <- 1
	}
	return nil
}

func (u *Accumulator) FindJack(jack string) (*Jack, error) {
	switch {
	case jack == "α", jack == "a", jack == "alpha":
		return u.α, nil
	case jack == "β", jack == "b", jack == "beta":
		return u.β, nil
	case jack == "γ", jack == "g", jack == "gamma":
		return u.γ, nil
	case jack == "δ", jack == "d", jack == "delta":
		return u.δ, nil
	case jack == "ε", jack == "e", jack == "epsilon":
		return u.ε, nil
	case jack == "A":
		return u.A, nil
	case jack == "S":
		return u.S, nil
	}
	jacks := [20]string{
		"1i", "2i", "3i", "4i", "5i", "5o", "6i", "6o", "7i", "7o",
		"8i", "8o", "9i", "9o", "10i", "10o", "11i", "11o", "12i", "12o",
	}
	for i, j := range jacks {
		if j == jack {
			return u.ctlterm[i], nil
		}
	}
	return nil, fmt.Errorf("invalid jack: %s", jack)
}

func accOpSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"α", stα}, {"a", stα}, {"alpha", stα},
		{"β", stβ}, {"b", stβ}, {"beta", stβ},
		{"γ", stγ}, {"g", stγ}, {"gamma", stγ},
		{"δ", stδ}, {"d", stδ}, {"delta", stδ},
		{"ε", stε}, {"e", stε}, {"epsilon", stε},
		{"0", 0}, // Must be 0 for su1().
		{"A", stA},
		{"AS", stAS},
		{"S", stS},
	}
}

func (u *Accumulator) lookupSwitch(name string) (Switch, error) {
	if name == "sf" {
		return &IntSwitch{name, &u.sigfig, sfSettings()}, nil
	}
	if name == "sc" {
		return &ByteSwitch{name, &u.sc, scSettings()}, nil
	}
	if len(name) < 3 {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	prog, _ := strconv.Atoi(name[2:])
	if !(prog >= 1 && prog <= 12) {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	prog--
	switch name[:2] {
	case "op":
		return &IntSwitch{name, &u.opsw[prog], accOpSettings()}, nil
	case "cc":
		return &BoolSwitch{name, &u.clrsw[prog], clearSettings()}, nil
	case "rp":
		if !(prog >= 4 && prog <= 11) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.rptsw[prog-4], rpSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

func (u *Accumulator) SetSwitch(name, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return err
	}
	u.updateSu1Cache()
	return sw.Set(value)
}

func (u *Accumulator) GetSwitch(name string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return "?", err
	}
	return sw.Get(), nil
}

/*
 * Implement the PX-5-109 terminator and PX-5-110
 * and PX-5-121 interconnect cables
 */
func (u *Accumulator) su1() int {
	x := 0
	if u.rbuddy >= 0 && u.rbuddy != u.unit {
		x = u.prbuddy.su1()
	}
	return x | u.su1Cache
}

func (u *Accumulator) updateSu1Cache() {
	x := 0
	for i := range u.inff1 {
		if u.inff1[i] {
			x |= u.opsw[i]
			if u.clrsw[i] {
				if u.opsw[i] == 0 || u.opsw[i] >= stA {
					if i < 4 || u.rep == int(u.rptsw[i-4]) {
						x |= stCLR
					}
				} else {
					x |= stCORR
				}
			}
		}
	}
	u.su1Cache = x
}

func (u *Accumulator) st1() int {
	x := 0
	if u.lbuddy == u.unit {
		x = u.su1()
	} else if u.lbuddy == -1 {
		x = u.su1()
		if u.Io.Multl() {
			x |= stα
		}
	} else if u.lbuddy == -2 {
		x = u.su1()
		if u.Io.Multr() {
			x |= stα
		}
	} else if u.lbuddy == -3 {
		x = u.su1()
		x |= u.Io.Sv()
	} else if u.lbuddy == -4 {
		x = u.su1()
		/* Wiring for PX-5-134 for quotient */
		su2 := u.Io.Su2()
		x |= su2 & stα
		x |= (su2 & su2qA) << 2
		x |= (su2 & su2qS) << 3
		x |= (su2 & su2qCLR) << 3
	} else if u.lbuddy == -5 {
		x = u.su1()
		/* Wiring for PX-5-135 for shift */
		su2 := u.Io.Su2()
		x |= (su2 & su2sα) >> 1
		x |= (su2 & su2sA) << 3
		x |= (su2 & su2sCLR) << 2
	} else if u.lbuddy == -6 {
		x = u.su1()
		/* Wiring for PX-5-136 for denominator */
		su3 := u.Io.Su3()
		x |= su3 & (stα | stβ | stγ)
		x |= (su3 & su3A) << 2
		x |= (su3 & su3S) << 3
		x |= (su3 & su3CLR) << 3
	} else {
		x = u.plbuddy.st2() & 0x1c3ff
	}
	return x
}

func (u *Accumulator) st2() int {
	x := u.st1() & 0x03ff

	return x
}

func (u *Accumulator) docpp(cyc Pulse) {
	su1Dirty := false
	for i := 0; i < 4; i++ {
		if u.inff2[i] {
			u.inff1[i] = false
			u.inff2[i] = false
			su1Dirty = true
		}
	}
	if u.h50 {
		u.rep++
		su1Dirty = true
		rstrep := false
		for i := 4; i < 12; i++ {
			if u.inff2[i] && u.rep == int(u.rptsw[i-4])+1 {
				u.inff1[i] = false
				u.inff2[i] = false
				rstrep = true
				t := (i-4)*2 + 5
				u.ctlterm[t].Transmit(1)
			}
		}
		if rstrep {
			u.rep = 0
			u.h50 = false
		}
	}
	if su1Dirty {
		u.updateSu1Cache()
	}
}

func (u *Accumulator) ripple() {
	for i := 0; i < 9; i++ {
		if u.decff[i] {
			u.val[i+1]++
			if u.val[i+1] == 10 {
				u.val[i+1] = 0
				u.decff[i+1] = true
			}
		}
	}
	if u.lbuddy < 0 || u.lbuddy == u.unit {
		if u.decff[9] {
			/*
			 * Connection PX-5-121 pins 14, 15
			 */
			u.sign = !u.sign
		}
	} else {
		/*
		 * PX-5-110, pin 15 straight through
		 */
		if u.decff[9] {
			u.plbuddy.val[0]++
			if u.plbuddy.val[0] == 10 {
				u.plbuddy.val[0] = 0
				u.plbuddy.decff[0] = true
			}
		}
		u.plbuddy.ripple()
	}
}

func (u *Accumulator) doccg() {
	curprog := u.st1()
	u.whichrp = false
	if curprog&0x1f != 0 {
		if u.rbuddy == u.unit {
			u.ripple()
		}
	} else if (curprog & stCLR) != 0 {
		for i := 0; i < 10; i++ {
			u.val[i] = byte(0)
		}
		u.sign = false
	}
}

func (u *Accumulator) dorp() {
	if !u.whichrp {
		/*
		 * Ugly hack to avoid races.  Effectively this is
		 * a coarse approximation to the "slow buffer
		 * output" described in 1.2.9 of the Technical
		 * Manual Part 2.
		 */
		for i := 0; i < 12; i++ {
			if u.inff1[i] {
				u.inff2[i] = true
				if i >= 4 {
					u.h50 = true
				}
			}
		}
		for i := 0; i < 10; i++ {
			u.decff[i] = false
		}
		u.whichrp = true
	}
}

func (u *Accumulator) dotenp() {
	curprog := u.st1()
	if curprog&(stA|stAS|stS) != 0 {
		for i := 0; i < 10; i++ {
			u.val[i]++
			if u.val[i] == 10 {
				u.val[i] = 0
				u.decff[i] = true
			}
		}
	}
}

func (u *Accumulator) doninep() {
	curprog := u.st1()
	if curprog&(stA|stAS) != 0 {
		if u.A.Connected() {
			n := 0
			for i := 0; i < 10; i++ {
				if u.decff[i] {
					n |= 1 << uint(i)
				}
			}
			if u.sign {
				n |= 1 << 10
			}
			if n != 0 {
				u.A.Transmit(n)
			}
		}
	}
	if curprog&(stAS|stS) != 0 {
		if u.S.Connected() {
			n := 0
			for i := 0; i < 10; i++ {
				if !u.decff[i] {
					n |= 1 << uint(i)
				}
			}
			if !u.sign {
				n |= 1 << 10
			}
			if n != 0 {
				u.S.Transmit(n)
			}
		}
	}
}

func (u *Accumulator) doonepp() {
	curprog := u.st1()
	if curprog&stCORR != 0 {
		/*
		 * Connection of PX-5-109 pins 14, 15
		 */
		if u.rbuddy == u.unit {
			u.val[0]++
			if u.val[0] > 9 {
				u.val[0] = 0
				u.decff[0] = true
			}
		}
	}
	if curprog&(stAS|stS) != 0 && u.S.Connected() {
		if ((u.lbuddy < 0 || u.lbuddy == u.unit) && u.rbuddy == u.unit && u.sigfig > 0) ||
			(u.rbuddy != u.unit && u.sigfig < 10) ||
			(u.lbuddy != u.unit && u.lbuddy >= 0 && u.plbuddy.sigfig == 10 && u.sigfig > 0) ||
			(u.rbuddy != u.unit && u.sigfig == 10 && u.prbuddy.sigfig == 0) {
			u.S.Transmit(1 << uint(10-u.sigfig))
		}
	}
}

func (u *Accumulator) Clock(cyc Pulse) {
	switch {
	case cyc&Cpp != 0:
		u.docpp(cyc)
	case cyc&Ccg != 0:
		u.doccg()
	case cyc&Scg != 0:
		if u.sc == 1 {
			u.Clear()
		}
	case cyc&Rp != 0:
		u.dorp()
	case cyc&Tenp != 0:
		u.dotenp()
	case cyc&Ninep != 0:
		u.doninep()
	case cyc&Onepp != 0:
		u.doonepp()
	}
}

func (u *Accumulator) receive(value int) {
	for i := 0; i < 10; i++ {
		if value&1 == 1 {
			u.val[i]++
			if u.val[i] >= 10 {
				u.decff[i] = true
				u.val[i] -= 10
			}
		}
		value >>= 1
	}
	if value&1 == 1 {
		u.sign = !u.sign
	}
}

func (u *Accumulator) trigger(input int) {
	u.mu.Lock()
	u.inff1[input] = true
	u.updateSu1Cache()
	u.mu.Unlock()
}
