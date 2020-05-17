package units

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

// An Accumulator is a unit of the ENIAC which is capable of performing the
// following operations:
// 1) Storing a ten digit number along with the proper indication of its sign.
// 2) Receiving numbers (positive or negative) from other units of the ENIAC
//    and adding them to numbers already stored, properly indicating the sign of
//    the sum.
// 3) Round off its contents to a previously determined number of
//    places.
// 4) Transmitting the number held, or its complement with respect to 10^10,
//    without losing its contents (this makes it possible to add and/or subtract
//    from the contents of one accumulator those of another).
// 5) Clear its contents to zero (except for a possible round off five).
// 6) Information stored in certain accumulators may be transmitted statically
//    to certain other units.
// -- ENIAC Technical Manual, Part II (Ch IV)
type Accumulator struct {
	α, β, γ, δ, ε *Jack     // Digit inputs
	A, S          *Jack     // Digit outputs
	program       [20]*Jack // Program terminals

	operation      [12]int  // Operation switches
	clear          [12]bool // Clear switches
	repeat         [8]int   // Repeat count selection
	figures        int      // Significant figures
	selectiveClear bool     //

	sign   bool     // True if negative
	decade [10]int  // Ten digits 0-9
	carry  [10]bool // Carry out of each digit position

	inff1, inff2     [12]bool //
	repeating        bool
	repeatCount      int
	afterFirstRp     bool
	lbuddy, rbuddy   int
	plbuddy, prbuddy *Accumulator
	programCache     int

	unit       int // Unit number 0-19
	Io         AccumulatorConn
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

// Static connections to other non-accumulator units.
type StaticWiring interface {
	Sign() string
	Value() string
	Clear()
}

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

func NewAccumulator(unit int) *Accumulator {
	unitDot := fmt.Sprintf("a%d.", unit+1)
	u := &Accumulator{
		unit:    unit,
		figures: 10,
		lbuddy:  unit,
		rbuddy:  unit,
	}
	digitInput := func(st1Pin int) JackHandler {
		return func(j *Jack, val int) {
			u.mu.Lock()
			defer u.mu.Unlock()
			if u.activeProgram()&st1Pin != 0 {
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
	u.program[0] = NewInput(unitDot+"1i", programInput(0))
	u.program[1] = NewInput(unitDot+"2i", programInput(1))
	u.program[2] = NewInput(unitDot+"3i", programInput(2))
	u.program[3] = NewInput(unitDot+"4i", programInput(3))
	u.program[4] = NewInput(unitDot+"5i", programInput(4))
	u.program[5] = NewOutput(unitDot+"5o", output(1))
	u.program[6] = NewInput(unitDot+"6i", programInput(5))
	u.program[7] = NewOutput(unitDot+"6o", output(1))
	u.program[8] = NewInput(unitDot+"7i", programInput(6))
	u.program[9] = NewOutput(unitDot+"7o", output(1))
	u.program[10] = NewInput(unitDot+"8i", programInput(7))
	u.program[11] = NewOutput(unitDot+"8o", output(1))
	u.program[12] = NewInput(unitDot+"9i", programInput(8))
	u.program[13] = NewOutput(unitDot+"9o", output(1))
	u.program[14] = NewInput(unitDot+"10i", programInput(9))
	u.program[15] = NewOutput(unitDot+"10o", output(1))
	u.program[16] = NewInput(unitDot+"11i", programInput(10))
	u.program[17] = NewOutput(unitDot+"11o", output(1))
	u.program[18] = NewInput(unitDot+"12i", programInput(11))
	u.program[19] = NewOutput(unitDot+"12o", output(1))
	return u
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
		s += fmt.Sprintf("%d", u.decade[i])
	}
	s += " "
	for i := 9; i >= 0; i-- {
		s += ToBin(u.carry[i])
	}
	s += fmt.Sprintf(" %d ", u.repeatCount)
	for _, f := range u.inff2 {
		s += ToBin(f)
	}
	return s
}

type accJson struct {
	Sign    bool     `json:"sign"`
	Decade  [10]int  `json:"decade"`
	Carry   [10]bool `json:"carry"`
	Repeat  int      `json:"repeat"`
	Program [12]bool `json:"program"`
}

func (u *Accumulator) State() json.RawMessage {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := accJson{
		Sign:    u.sign,
		Decade:  u.decade,
		Carry:   u.carry,
		Repeat:  u.repeatCount,
		Program: u.inff2,
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
				n += int64(u.decade[i])
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
		u.program[i].Disconnect()
		u.inff1[i] = false
		u.inff2[i] = false
		u.operation[i] = stα
		u.clear[i] = false
	}
	for i := 0; i < 8; i++ {
		u.repeat[i] = 0
	}
	u.figures = 10
	u.selectiveClear = false
	u.repeating = false
	u.repeatCount = 0
	u.afterFirstRp = false
	u.lbuddy = u.unit
	u.rbuddy = u.unit
	u.mu.Unlock()
	u.Clear()
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
		s += fmt.Sprintf("%d", u.decade[i])
	}
	return s
}

func (u *Accumulator) Clear() {
	u.mu.Lock()
	defer u.mu.Unlock()
	for i := 0; i < 10; i++ {
		u.decade[i] = 0
		u.carry[i] = false
	}
	if u.figures < 10 {
		u.decade[9-u.figures] = 5
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
		u.decade[i] = int(value % 10)
		u.carry[i] = false
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
			return u.program[i], nil
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
		{"0", 0}, // Must be 0 for userProgram().
		{"A", stA},
		{"AS", stAS},
		{"S", stS},
	}
}

func (u *Accumulator) lookupSwitch(name string) (Switch, error) {
	if name == "sf" {
		return &IntSwitch{name, &u.figures, sfSettings()}, nil
	}
	if name == "sc" {
		return &BoolSwitch{name, &u.selectiveClear, scSettings()}, nil
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
		return &IntSwitch{name, &u.operation[prog], accOpSettings()}, nil
	case "cc":
		return &BoolSwitch{name, &u.clear[prog], clearSettings()}, nil
	case "rp":
		if !(prog >= 4 && prog <= 11) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.repeat[prog-4], rpSettings()}, nil
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
	u.updateActiveProgram()
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

func (u *Accumulator) userProgram() int {
	x := 0
	if u.rbuddy >= 0 && u.rbuddy != u.unit {
		x = u.prbuddy.userProgram()
	}
	return x | u.programCache
}

func (u *Accumulator) updateActiveProgram() {
	x := 0
	for i := range u.inff2 {
		if u.inff2[i] {
			x |= u.operation[i]
			if u.clear[i] {
				if u.operation[i] == 0 || u.operation[i] >= stA {
					if i < 4 || u.repeatCount == int(u.repeat[i-4]) {
						x |= stCLR
					}
				} else {
					x |= stCORR
				}
			}
		}
	}
	u.programCache = x
}

func (u *Accumulator) activeProgram() int {
	x := 0
	if u.lbuddy == u.unit {
		x = u.userProgram()
	} else if u.lbuddy == -1 {
		x = u.userProgram()
		if u.Io.Multl() {
			x |= stα
		}
	} else if u.lbuddy == -2 {
		x = u.userProgram()
		if u.Io.Multr() {
			x |= stα
		}
	} else if u.lbuddy == -3 {
		x = u.userProgram()
		x |= u.Io.Sv()
	} else if u.lbuddy == -4 {
		x = u.userProgram()
		/* Wiring for PX-5-134 for quotient */
		su2 := u.Io.Su2()
		x |= su2 & stα
		x |= (su2 & su2qA) << 2
		x |= (su2 & su2qS) << 3
		x |= (su2 & su2qCLR) << 3
	} else if u.lbuddy == -5 {
		x = u.userProgram()
		/* Wiring for PX-5-135 for shift */
		su2 := u.Io.Su2()
		x |= (su2 & su2sα) >> 1
		x |= (su2 & su2sA) << 3
		x |= (su2 & su2sCLR) << 2
	} else if u.lbuddy == -6 {
		x = u.userProgram()
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
	x := u.activeProgram() & 0x03ff

	return x
}

func (u *Accumulator) doCpp(cyc Pulse) {
	for i := 0; i < 4; i++ {
		if u.inff2[i] {
			u.inff2[i] = false
		}
	}
	if u.repeating {
		u.repeatCount++
		done := false
		for i := 4; i < 12; i++ {
			if u.inff2[i] && u.repeatCount == int(u.repeat[i-4])+1 {
				u.inff2[i] = false
				done = true
				t := (i-4)*2 + 5
				u.program[t].Transmit(1)
			}
		}
		if done {
			u.repeatCount = 0
			u.repeating = false
		}
	}
}

func (u *Accumulator) ripple() {
	for i := 0; i < 9; i++ {
		if u.carry[i] {
			u.decade[i+1]++
			if u.decade[i+1] == 10 {
				u.decade[i+1] = 0
				u.carry[i+1] = true
			}
		}
	}
	if u.lbuddy < 0 || u.lbuddy == u.unit {
		if u.carry[9] {
			/*
			 * Connection PX-5-121 pins 14, 15
			 */
			u.sign = !u.sign
		}
	} else {
		/*
		 * PX-5-110, pin 15 straight through
		 */
		if u.carry[9] {
			u.plbuddy.decade[0]++
			if u.plbuddy.decade[0] == 10 {
				u.plbuddy.decade[0] = 0
				u.plbuddy.carry[0] = true
			}
		}
		u.plbuddy.ripple()
	}
}

func (u *Accumulator) doCcg() {
	program := u.activeProgram()
	if program&(stα|stβ|stγ|stδ|stε) != 0 {
		if u.rbuddy == u.unit {
			u.ripple()
		}
	} else if program&stCLR != 0 {
		for i := 0; i < 10; i++ {
			u.decade[i] = 0
		}
		u.sign = false
	}
}

func (u *Accumulator) doRp() {
	if u.afterFirstRp {
		for i := 0; i < 12; i++ {
			if u.inff1[i] {
				u.inff1[i] = false
				u.inff2[i] = true
				if i >= 4 {
					u.repeating = true
				}
			}
		}
		for i := 0; i < 10; i++ {
			u.carry[i] = false
		}
		u.updateActiveProgram()
	}
	u.afterFirstRp = !u.afterFirstRp
}

func (u *Accumulator) doTenp() {
	program := u.activeProgram()
	if program&(stA|stAS|stS) != 0 {
		for i := 0; i < 10; i++ {
			u.decade[i]++
			if u.decade[i] == 10 {
				u.decade[i] = 0
				u.carry[i] = true
			}
		}
	}
}

func (u *Accumulator) doNinep() {
	program := u.activeProgram()
	if program&(stA|stAS) != 0 {
		if u.A.Connected() {
			n := 0
			for i := 0; i < 10; i++ {
				if u.carry[i] {
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
	if program&(stAS|stS) != 0 {
		if u.S.Connected() {
			n := 0
			for i := 0; i < 10; i++ {
				if !u.carry[i] {
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

func (u *Accumulator) doOnepp() {
	program := u.activeProgram()
	if program&stCORR != 0 {
		if u.rbuddy == u.unit {
			u.decade[0]++
			if u.decade[0] > 9 {
				u.decade[0] = 0
				u.carry[0] = true
			}
		}
	}
	if program&(stAS|stS) != 0 && u.S.Connected() {
		if ((u.lbuddy < 0 || u.lbuddy == u.unit) && u.rbuddy == u.unit && u.figures > 0) ||
			(u.rbuddy != u.unit && u.figures < 10) ||
			(u.lbuddy != u.unit && u.lbuddy >= 0 && u.plbuddy.figures == 10 && u.figures > 0) ||
			(u.rbuddy != u.unit && u.figures == 10 && u.prbuddy.figures == 0) {
			u.S.Transmit(1 << uint(10-u.figures))
		}
	}
}

func (u *Accumulator) Clock(cyc Pulse) {
	switch {
	case cyc&Tenp != 0:
		u.doTenp()
	case cyc&Ninep != 0:
		u.doNinep()
	case cyc&Onepp != 0:
		u.doOnepp()
	case cyc&Ccg != 0:
		u.doCcg()
	case cyc&Rp != 0:
		u.doRp()
	case cyc&Scg != 0:
		if u.selectiveClear {
			u.Clear()
		}
	case cyc&Cpp != 0:
		u.doCpp(cyc)
	}
}

func (u *Accumulator) receive(value int) {
	for i := 0; i < 10; i++ {
		if value&1 == 1 {
			u.decade[i]++
			if u.decade[i] >= 10 {
				u.carry[i] = true
				u.decade[i] -= 10
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
	if u.afterFirstRp {
		// This is a normal input during Cpp, either before or after u.Clock(Cpp).
		// Set a flag to update the program on Rp following the Cpp so that
		// clocking order doesn't matter.
		u.inff1[input] = true
	} else {
		// Inputs before the Cpp, e.g. from digit pulse adapters to dummy programs,
		// should take effect on the current Cpp.
		u.inff2[input] = true
		if input >= 4 {
			u.repeating = true
		}
		// fttest2.e wires digit pulse adapters to non-dummy programs.
		u.updateActiveProgram()
	}
	u.mu.Unlock()
}
