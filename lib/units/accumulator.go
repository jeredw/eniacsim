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
	program       [20]*Jack // Program inputs and outputs

	operation      [12]int  // Operation switches
	clear          [12]bool // Clear switches
	repeat         [8]int   // Repeat count selection
	figures        int      // Significant figures
	selectiveClear bool     // If true, initiate unit may trigger clear

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

// Operations supported by common programming circuits.
const (
	opα = 1 << iota
	opβ
	opγ
	opδ
	opε
	opA
	opAS
	opS
	opClear
	opCorrect
)

func NewAccumulator(unit int) *Accumulator {
	u := &Accumulator{
		unit:    unit,
		lbuddy:  unit,
		rbuddy:  unit,
		figures: 10,
	}
	u.α = u.newDigitInput("α", opα)
	u.β = u.newDigitInput("β", opβ)
	u.γ = u.newDigitInput("γ", opγ)
	u.δ = u.newDigitInput("δ", opδ)
	u.ε = u.newDigitInput("ε", opε)
	u.A = u.newOutput("A", 11)
	u.S = u.newOutput("S", 11)
	u.program[0] = u.newProgramInput("1i", 0)
	u.program[1] = u.newProgramInput("2i", 1)
	u.program[2] = u.newProgramInput("3i", 2)
	u.program[3] = u.newProgramInput("4i", 3)
	u.program[4] = u.newProgramInput("5i", 4)
	u.program[5] = u.newOutput("5o", 1)
	u.program[6] = u.newProgramInput("6i", 5)
	u.program[7] = u.newOutput("6o", 1)
	u.program[8] = u.newProgramInput("7i", 6)
	u.program[9] = u.newOutput("7o", 1)
	u.program[10] = u.newProgramInput("8i", 7)
	u.program[11] = u.newOutput("8o", 1)
	u.program[12] = u.newProgramInput("9i", 8)
	u.program[13] = u.newOutput("9o", 1)
	u.program[14] = u.newProgramInput("10i", 9)
	u.program[15] = u.newOutput("10o", 1)
	u.program[16] = u.newProgramInput("11i", 10)
	u.program[17] = u.newOutput("11o", 1)
	u.program[18] = u.newProgramInput("12i", 11)
	u.program[19] = u.newOutput("12o", 1)
	return u
}

func (u *Accumulator) terminal(name string) string {
	return fmt.Sprintf("a%d.%s", u.unit+1, name)
}

func (u *Accumulator) newDigitInput(name string, programMask int) *Jack {
	return NewInput(u.terminal(name), func(j *Jack, val int) {
		u.mu.Lock()
		defer u.mu.Unlock()
		if u.activeProgram()&programMask != 0 {
			u.receive(val)
			if u.tracePulse != nil {
				u.tracePulse(j.Name, 11, int64(val))
			}
		}
	})
}

func (u *Accumulator) newProgramInput(name string, which int) *Jack {
	return NewInput(u.terminal(name), func(j *Jack, val int) {
		if val == 1 {
			u.trigger(which)
			if u.tracePulse != nil {
				u.tracePulse(j.Name, 1, int64(val))
			}
		}
	})
}

func (u *Accumulator) newOutput(name string, width int) *Jack {
	return NewOutput(u.terminal(name), func(j *Jack, val int) {
		if u.tracePulse != nil {
			u.tracePulse(j.Name, width, int64(val))
		}
	})
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
	sign := u.terminal("sign")
	decade := u.terminal("decade")
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
			traceReg(decade, 40, n)
		},
	}
}

func (u *Accumulator) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.α.Disconnect()
	u.β.Disconnect()
	u.γ.Disconnect()
	u.δ.Disconnect()
	u.ε.Disconnect()
	for i := 0; i < 12; i++ {
		u.program[i].Disconnect()
		u.inff1[i] = false
		u.inff2[i] = false
		u.operation[i] = opα
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
	u.clearInternal()
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
	u.clearInternal()
}

func (u *Accumulator) clearInternal() {
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
		{"α", opα}, {"a", opα}, {"alpha", opα},
		{"β", opβ}, {"b", opβ}, {"beta", opβ},
		{"γ", opγ}, {"g", opγ}, {"gamma", opγ},
		{"δ", opδ}, {"d", opδ}, {"delta", opδ},
		{"ε", opε}, {"e", opε}, {"epsilon", opε},
		{"0", 0}, // Must be 0 for userProgram().
		{"A", opA},
		{"AS", opAS},
		{"S", opS},
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
				if u.operation[i] == 0 || u.operation[i] >= opA {
					if i < 4 || u.repeatCount == u.repeat[i-4] {
						x |= opClear
					}
				} else {
					x |= opCorrect
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
			x |= opα
		}
	} else if u.lbuddy == -2 {
		x = u.userProgram()
		if u.Io.Multr() {
			x |= opα
		}
	} else if u.lbuddy == -3 {
		x = u.userProgram()
		x |= u.Io.Sv()
	} else if u.lbuddy == -4 {
		x = u.userProgram()
		/* Wiring for PX-5-134 for quotient */
		su2 := u.Io.Su2()
		x |= su2 & opα
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
		x |= su3 & (opα | opβ | opγ)
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
			if u.inff2[i] && u.repeatCount == u.repeat[i-4]+1 {
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
	if program&(opα|opβ|opγ|opδ|opε) != 0 {
		if u.rbuddy == u.unit {
			u.ripple()
		}
	} else if program&opClear != 0 {
		u.clearInternal()
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
	if program&(opA|opAS|opS) != 0 {
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
	if program&(opA|opAS) != 0 {
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
	if program&(opAS|opS) != 0 {
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
	if program&opCorrect != 0 {
		if u.rbuddy == u.unit {
			u.decade[0]++
			if u.decade[0] > 9 {
				u.decade[0] = 0
				u.carry[0] = true
			}
		}
	}
	if program&(opAS|opS) != 0 && u.S.Connected() {
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
		// So that simulator clocking order doesn't matter, programs are registered
		// in inff1 and applied on the Rp pulse following Cpp.  (Pulse order is Rp,
		// ..., Cpp, ..., Rp, so this input is during Cpp.)
		u.inff1[input] = true
	} else {
		// Inputs from digit pulse adapters arrive before Cpp, and take effect for
		// the current add cycle - in particular, a dummy program driven by digit
		// pulses triggers its output on Cpp of the same add cycle as its input.
		// (See Technical Manual IV-26, Table 4-3.)
		u.inff2[input] = true
		if input >= 4 {
			u.repeating = true
		}
		// fttest2.e wires digit pulse adapters to non-dummy programs.
		u.updateActiveProgram()
	}
	u.mu.Unlock()
}
