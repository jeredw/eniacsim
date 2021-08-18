package units

import (
	"encoding/json"
	"fmt"
	"strconv"

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

	sign   bool   // True if negative
	decade uint64 // Ten digits 0-9 in BCD
	carry  uint64 // Carry ffs per decade in BCD
	carry2 uint64 // Temp for ripple carries (also BCD)

	inff1, inff2    int // 12-bit bitmasks
	repeating       bool
	repeatCount     int
	afterFirstRp    bool
	activeProgram   int
	externalProgram int

	left  *Accumulator
	right *Accumulator

	unit        int // Unit number 0-19
	tracer      Tracer
	valueString []byte
}

// NewAccumulator returns an Accumulator with I/O jacks configured.
func NewAccumulator(unit int) *Accumulator {
	u := &Accumulator{
		unit:    unit,
		figures: 10,
	}
	u.valueString = make([]byte, 12)
	u.updateValueString()
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
	if u.activeProgram&(opA|opAS|opS) != 0 {
		increment := u.decade + 0x1111111111
		// Only 9+3=12 will have both bits 2+3 set
		nines := (u.decade + 0x3333333333) & 0xcccccccccc
		nines = (nines & (nines >> 1)) >> 2
		// Subtract 10 from positions that had 9s
		u.decade = increment - ((nines << 1) | (nines << 3))
		u.carry |= nines
	}
}

func (u *Accumulator) doNinep() {
	if u.activeProgram&(opA|opAS) != 0 {
		n := int((u.carry & 0x1) |
			((u.carry >> 3) & 0x2) |
			((u.carry >> 6) & 0x4) |
			((u.carry >> 9) & 0x8) |
			((u.carry >> 12) & 0x10) |
			((u.carry >> 15) & 0x20) |
			((u.carry >> 18) & 0x40) |
			((u.carry >> 21) & 0x80) |
			((u.carry >> 24) & 0x100) |
			((u.carry >> 27) & 0x200))
		if u.sign {
			n |= 1 << 10
		}
		if n != 0 {
			u.A.Transmit(n)
		}
	}
	if u.activeProgram&(opAS|opS) != 0 {
		n := int((u.carry & 0x1) |
			((u.carry >> 3) & 0x2) |
			((u.carry >> 6) & 0x4) |
			((u.carry >> 9) & 0x8) |
			((u.carry >> 12) & 0x10) |
			((u.carry >> 15) & 0x20) |
			((u.carry >> 18) & 0x40) |
			((u.carry >> 21) & 0x80) |
			((u.carry >> 24) & 0x100) |
			((u.carry >> 27) & 0x200))
		n = (^n) & 0x3ff
		if !u.sign {
			n |= 1 << 10
		}
		if n != 0 {
			u.S.Transmit(n)
		}
	}
}

func (u *Accumulator) doOnepp() {
	if u.activeProgram&opCorrect != 0 {
		// Apply tens' complement correction to the lowest order decade.
		if u.right == nil {
			u.decade += 1
			if u.decade&0xf > 9 {
				u.decade &= 0xfffffffff0
				u.carry |= 1
			}
		}
	}
	if u.activeProgram&(opAS|opS) != 0 && u.S.OutputConnected {
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
	if u.activeProgram&opClear != 0 {
		u.clearInternal()
	}
	// (Carry is actually initated on Rp-1.)
}

func (u *Accumulator) clearInternal() {
	u.carry = 0
	u.carry2 = 0
	u.decade = 0
	if u.figures < 10 {
		u.decade |= 5 << (4 * (9 - u.figures))
	}
	u.sign = false
}

func (u *Accumulator) doRp1() {
	// The first Rp initiates carry-over and clears any carries set during
	// digit reception.
	if u.activeProgram&(opα|opβ|opγ|opδ|opε) != 0 {
		// Carry-over starts from the right accumulator of a pair.
		if u.right == nil {
			u.ripple()
		}
	}
	// Carry propagation will set some carry ffs again as ripple occurs.  We
	// track these in carry2.  This would happen over the next 5 pulse times
	// but just do it here.
	u.carry = u.carry2
}

func (u *Accumulator) ripple() {
	digits := ((u.carry | u.carry2) << 4) & 0xfffffffff0
	sum := u.decade + digits + 0x6666666666
	noncarry := ^(sum ^ u.decade ^ digits) & 0x11111111110
	carry := (noncarry ^ 0x11111111110) >> 4
	u.decade = (sum - ((noncarry >> 2) | (noncarry >> 3))) & 0xffffffffff
	u.carry2 |= carry
	if u.left == nil {
		// Carry from final decade into sign.
		if (u.carry|u.carry2)&(1<<(4*9)) != 0 {
			u.sign = !u.sign
		}
	} else {
		// This is the right half of a pair of interconnected accumulators.
		// Carry from final decade into decade 1 of left half.
		if (u.carry|u.carry2)&(1<<(4*9)) != 0 {
			u.left.decade += 1
			if u.left.decade&0xf == 10 {
				u.left.decade &= 0xfffffffff0
				u.left.carry2 |= 1
			}
		}
		u.left.ripple()
	}
}

func (u *Accumulator) doCpp(cyc Pulse) {
	u.inff2 &= 0xff0
	if u.repeating {
		u.repeatCount++
		done := false
		for i := 4; i < 12; i++ {
			if u.inff2&(1<<uint(i)) != 0 && u.repeatCount == u.repeat[i-4]+1 {
				u.inff2 &= ^(1 << uint(i))
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
	u.carry = 0
	u.carry2 = 0
	// Apply programs setup on Cpp; see the note in trigger().
	u.inff2 |= u.inff1
	u.inff1 = 0
	u.repeating = u.inff2&0xff0 != 0
	u.updateActiveProgram()
}

func (u *Accumulator) updateActiveProgram() {
	if u.externalProgram != 0 {
		u.activeProgram = u.externalProgram
	} else {
		u.updateOwnActiveProgram()
		if u.left != nil {
			u.left.updateOwnActiveProgram()
			u.activeProgram |= u.left.activeProgram
		}
		if u.right != nil {
			u.right.updateOwnActiveProgram()
			u.activeProgram |= u.right.activeProgram
		}
	}
	u.enableInputs()
}

func (u *Accumulator) updateOwnActiveProgram() {
	x := 0
	numActivePrograms := 0
	for i := 0; i < 12; i++ {
		if u.inff2&(1<<uint(i)) != 0 {
			numActivePrograms++
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
	if numActivePrograms > 1 {
		repeatCount := 0
		operation := 0
		haveClear := false
		clear := false
		activePrograms := make([]int, 0, 12)
		for i := 0; i < 12; i++ {
			if u.inff2&(1<<uint(i)) != 0 {
				activePrograms = append(activePrograms, i+1)
			}
		}
		for i := 0; i < 12; i++ {
			if u.inff2&(1<<uint(i)) != 0 {
				if operation != 0 && u.operation[i] != 0 {
					panic(fmt.Sprintf("multiple active programs with conflicting op on a%d: %v\n", u.unit+1, activePrograms))
				}
				if !haveClear {
					clear = u.clear[i]
					haveClear = true
				} else if clear != u.clear[i] {
					panic(fmt.Sprintf("multiple active programs with conflicting clear on a%d: %v\n", u.unit+1, activePrograms))
				}
				if i < 4 {
					repeatCount = 1
				} else {
					if repeatCount == 0 {
						repeatCount = u.repeat[i-4]
					} else if repeatCount != u.repeat[i-4] {
						panic(fmt.Sprintf("multiple active programs with conflicting rp on a%d: %v\n", u.unit+1, activePrograms))
					}
				}
			}
		}
	}
	u.activeProgram = x
}

func (u *Accumulator) enableInputs() {
	u.α.Disabled = (u.activeProgram & opα) == 0
	u.β.Disabled = (u.activeProgram & opβ) == 0
	u.γ.Disabled = (u.activeProgram & opγ) == 0
	u.δ.Disabled = (u.activeProgram & opδ) == 0
	u.ε.Disabled = (u.activeProgram & opε) == 0
}

func (u *Accumulator) terminal(name string) string {
	return fmt.Sprintf("a%d.%s", u.unit+1, name)
}

func (u *Accumulator) newDigitInput(name string, programMask int) *Jack {
	return NewInput(u.terminal(name), func(j *Jack, val int) {
		if u.activeProgram&programMask == 0 || j.Disabled {
			panic("inactive acc input should be skipped")
		}
		u.receive(val)
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, 11, int64(val))
		}
	})
}

func (u *Accumulator) receive(value int) {
	x := uint64(value)
	digits := (x & 1) |
		((x << 3) & 0x10) |
		((x << 6) & 0x100) |
		((x << 9) & 0x1000) |
		((x << 12) & 0x10000) |
		((x << 15) & 0x100000) |
		((x << 18) & 0x1000000) |
		((x << 21) & 0x10000000) |
		((x << 24) & 0x100000000) |
		((x << 27) & 0x1000000000)
	sum := u.decade + digits
	// Only digits which overflowed to 9+1=10 will have both bits 2+3 set
	carry := (sum + 0x2222222222) & 0xcccccccccc
	carry = (carry & (carry >> 1)) >> 2
	// Subtract 10 from positions that overflowed
	u.decade = sum - ((carry << 1) | (carry << 3))
	u.carry |= carry
	if value&(1<<10) != 0 {
		u.sign = !u.sign
	}
}

func (u *Accumulator) newProgramInput(name string, which int) *Jack {
	return NewInput(u.terminal(name), func(j *Jack, val int) {
		if val == 1 {
			u.trigger(which)
			if u.tracer != nil {
				u.tracer.LogPulse(j.Name, 1, int64(val))
			}
		}
	})
}

func (u *Accumulator) trigger(input int) {
	if u.afterFirstRp {
		// So that simulator clocking order doesn't matter, programs are registered
		// in inff1 and applied on the Rp pulse following Cpp.  (Pulse order is Rp,
		// ..., Cpp, ..., Rp, so this input is during Cpp.)
		u.inff1 |= 1 << uint(input)
	} else {
		// Inputs from digit pulse adapters arrive before Cpp, and take effect for
		// the current add cycle - in particular, a dummy program driven by digit
		// pulses triggers its output on Cpp of the same add cycle as its input.
		// (See Technical Manual IV-26, Table 4-3.)
		u.inff2 |= 1 << uint(input)
		u.repeating = input >= 4
		if u.operation[input] != 0 {
			// This is probably not intended and won't be simulated correctly.
			panic("attempt to trigger non-dummy acc program from non-cpp")
		}
	}
}

func (u *Accumulator) newOutput(name string, width int) *Jack {
	return NewOutput(u.terminal(name), func(j *Jack, val int) {
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, width, int64(val))
		}
	})
}

func (u *Accumulator) Stat() string {
	var s string
	if u.sign {
		s += "M "
	} else {
		s += "P "
	}
	for i := 9; i >= 0; i-- {
		s += fmt.Sprintf("%d", (u.decade>>(4*i))&0xf)
	}
	s += " "
	for i := 9; i >= 0; i-- {
		s += ToBin(u.carry&(1<<uint(4*i)) != 0)
	}
	s += fmt.Sprintf(" %d ", u.repeatCount)
	for i := 0; i < 12; i++ {
		f := u.inff2&(1<<uint(i)) != 0
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
	carry := [10]bool{}
	decade := [10]int{}
	for i := 0; i < 10; i++ {
		carry[i] = u.carry&(1<<uint(4*i)) != 0
		decade[i] = int((u.decade >> (4 * i)) & 0xf)
	}
	program := [12]bool{}
	for i := 0; i < 12; i++ {
		program[i] = u.inff2&(1<<uint(i)) != 0
	}
	s := accJson{
		Sign:    u.sign,
		Decade:  decade,
		Carry:   carry,
		Repeat:  u.repeatCount,
		Program: program,
	}
	result, _ := json.Marshal(s)
	return result
}

func (u *Accumulator) AttachTracer(tracer Tracer) {
	u.tracer = tracer
	sign := u.terminal("sign")
	decade := u.terminal("decade")
	tracer.RegisterValueCallback(func() {
		tracer.LogValue(sign, 1, BoolToInt64(u.sign))
		tracer.LogValue(decade, 40, int64(u.decade))
	})
}

func (u *Accumulator) Reset() {
	u.inff1 = 0
	u.inff2 = 0
	for i := 0; i < 12; i++ {
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
	Value() []byte
	Clear()
	SetExternalProgram(int)
}

func (u *Accumulator) Sign() string {
	if u.sign {
		return "M"
	}
	return "P"
}

func (u *Accumulator) Value() []byte {
	u.updateValueString()
	return u.valueString
}

func (u *Accumulator) updateValueString() {
	if u.sign {
		u.valueString[0] = 'M'
	} else {
		u.valueString[0] = 'P'
	}
	u.valueString[1] = ' '
	for i := 0; i <= 9; i++ {
		digit := (u.decade >> uint(4*(9-i))) & 0xf
		u.valueString[2+i] = '0' + byte(digit)
	}
}

func (u *Accumulator) Clear() {
	u.clearInternal()
}

func (u *Accumulator) SetExternalProgram(program int) {
	u.externalProgram = program
	u.updateActiveProgram()
}

func (u *Accumulator) Set(value int64) {
	u.sign = value < 0
	if value < 0 {
		value = -value
	}
	u.carry = 0
	u.carry2 = 0
	u.decade = 0
	for i := 0; i < 10; i++ {
		u.decade |= uint64(value%10) << (4 * i)
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
		{"0", 0}, // Must be 0 for updateActiveProgram().
		{"A", opA},
		{"AS", opAS},
		{"S", opS},
	}
}

func (u *Accumulator) FindSwitch(name string) (Switch, error) {
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
