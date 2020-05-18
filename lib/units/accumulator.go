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
	carry  [10]bool // Carry ff per decade
	carry2 [10]bool // Temp for ripple carries

	inff1, inff2 [12]bool
	repeating    bool
	repeatCount  int
	afterFirstRp bool
	programCache int

	left  *Accumulator
	right *Accumulator

	unit       int // Unit number 0-19
	tracePulse TraceFunc

	mu sync.Mutex
}

// NewAccumulator returns an Accumulator with I/O jacks configured.
func NewAccumulator(unit int) *Accumulator {
	u := &Accumulator{
		unit:    unit,
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

// Clock updates unit state in response to a pulse on the cycling trunk.
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
		if !u.afterFirstRp {
			u.doRp1()
		} else {
			u.doRp2()
		}
		u.afterFirstRp = !u.afterFirstRp
	case cyc&Scg != 0:
		if u.selectiveClear {
			u.clearInternal()
		}
	case cyc&Cpp != 0:
		u.doCpp(cyc)
	}
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
	if program&(opA|opAS) != 0 && u.A.Connected() {
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
	if program&(opAS|opS) != 0 && u.S.Connected() {
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

func (u *Accumulator) doOnepp() {
	program := u.activeProgram()
	if program&opCorrect != 0 {
		// Apply tens' complement correction to the lowest order decade.
		if u.right == nil {
			u.decade[0]++
			if u.decade[0] > 9 {
				u.decade[0] = 0
				u.carry[0] = true
			}
		}
	}
	if program&(opAS|opS) != 0 && u.S.Connected() {
		// Transmit a final +1 in the least significant decade on S.
		//
		// Behavior of figures switches for interconnected accumulators is
		// explained in Technical Manual IV-22: "The significant figures switch of
		// the left hand accumulator should be set to 10 and in the right hand
		// accumulator to s' where 0 <= s' < 10 if 10 + s' significant figures are
		// desired.  If fewer than 10 significant figures are desired, the left
		// hand switch is set to this number and the right hand switch to 10."
		if (u.left == nil && u.right == nil && u.figures > 0) ||
			(u.right != nil && u.figures < 10) ||
			(u.left != nil && u.left.figures == 10 && u.figures > 0) ||
			(u.right != nil && u.figures == 10 && u.right.figures == 0) {
			u.S.Transmit(1 << uint(10-u.figures))
		}
	}
}

func (u *Accumulator) doCcg() {
	program := u.activeProgram()
	if program&opClear != 0 {
		u.clearInternal()
	}
	// (Carry is actually initated on Rp-1.)
}

func (u *Accumulator) clearInternal() {
	for i := 0; i < 10; i++ {
		u.decade[i] = 0
		u.carry[i] = false
		u.carry2[i] = false
	}
	if u.figures < 10 {
		u.decade[9-u.figures] = 5
	}
	u.sign = false
}

func (u *Accumulator) doRp1() {
	// The first Rp initiates carry-over and clears any carries set during
	// digit reception.
	program := u.activeProgram()
	if program&(opα|opβ|opγ|opδ|opε) != 0 {
		// Carry-over starts from the right accumulator of a pair.
		if u.right == nil {
			u.ripple()
		}
	}
	// Carry propagation will set some carry ffs again as ripple occurs.  We
	// track these in carry2.  This would happen over the next 5 pulse times
	// but just do it here.
	for i := 0; i < 10; i++ {
		u.carry[i] = u.carry2[i]
	}
}

func (u *Accumulator) ripple() {
	for i := 0; i < 9; i++ {
		if u.carry[i] || u.carry2[i] {
			u.decade[i+1]++
			if u.decade[i+1] == 10 {
				u.decade[i+1] = 0
				u.carry2[i+1] = true
			}
		}
	}
	if u.left == nil {
		// Carry from final decade into sign.
		if u.carry[9] || u.carry2[9] {
			u.sign = !u.sign
		}
	} else {
		// This is the right half of a pair of interconnected accumulators.
		// Carry from final decade into decade 1 of left half.
		if u.carry[9] || u.carry2[9] {
			u.left.decade[0]++
			if u.left.decade[0] == 10 {
				u.left.decade[0] = 0
				u.left.carry2[0] = true
			}
		}
		u.left.ripple()
	}
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

func (u *Accumulator) doRp2() {
	// The second Rp clears any leftover ripple carries.
	for i := 0; i < 10; i++ {
		u.carry[i] = false
		u.carry2[i] = false
	}
	// Apply programs setup on Cpp; see the note in trigger().
	for i := 0; i < 12; i++ {
		if u.inff1[i] {
			u.inff1[i] = false
			u.inff2[i] = true
			if i >= 4 {
				u.repeating = true
			}
		}
	}
	u.updateActiveProgram()
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
	if u.left != nil {
		return u.programCache | u.left.programCache
	}
	if u.right != nil {
		return u.programCache | u.right.programCache
	}
	return u.programCache
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
	u.left = nil
	u.right = nil
	u.clearInternal()
}

// Static connections to other non-accumulator units.
type StaticWiring interface {
	Sign() string
	Value() string
	Clear()
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
		u.carry2[i] = false
		value /= 10
	}
}

// Interconnect joins two accumulators to form a 20 digit accumulator.
func Interconnect(accumulator [20]*Accumulator, p1 []string, p2 []string) error {
	unit1, err := strconv.Atoi(p1[0][1:])
	if err != nil {
		return err
	}
	unit2, err := strconv.Atoi(p2[0][1:])
	if err != nil {
		return err
	}
	port1 := p1[1]
	port2 := p2[1]
	if unit1 == unit2 {
		if (port1 == "il1" && port2 == "il2") || (port1 == "il2" && port2 == "il1") {
			return nil
		}
		return fmt.Errorf("illegal interconnection")
	}
	a1 := accumulator[unit1-1]
	a2 := accumulator[unit2-1]
	a1.mu.Lock()
	defer a1.mu.Unlock()
	a2.mu.Lock()
	defer a2.mu.Unlock()
	switch {
	case port1 == "il1" && port2 == "ir1", port1 == "il2" && port2 == "ir2":
		a1.left = a2
		a2.right = a1
	case port1 == "ir1" && port2 == "il1", port1 == "ir2" && port2 == "il2":
		a1.right = a2
		a2.left = a1
	default:
		return fmt.Errorf("illegal interconnection")
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

func (u *Accumulator) FindSwitch(name string) (Switch, error) {
	if name == "sf" {
		return &IntSwitch{&u.mu, name, &u.figures, sfSettings()}, nil
	}
	if name == "sc" {
		return &BoolSwitch{&u.mu, name, &u.selectiveClear, scSettings()}, nil
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
		return &IntSwitch{&u.mu, name, &u.operation[prog], accOpSettings()}, nil
	case "cc":
		return &BoolSwitch{&u.mu, name, &u.clear[prog], clearSettings()}, nil
	case "rp":
		if !(prog >= 4 && prog <= 11) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.repeat[prog-4], rpSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}
