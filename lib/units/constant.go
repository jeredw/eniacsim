package units

import (
	"fmt"
	"strconv"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

// The Constant transmitter stores eighty digits on relays and twenty digits on
// constant set switches.  Provision is made for twenty PM digits, sixteen
// associated with the numbers stored on relays and four associated with the
// numbers set on the switches.  One of these PM signs is associated with each
// five digits.  Thus, the constant transmitter can store twenty five-digit
// numbers and their proper signs.  Provision is made to associate certain
// groups in pairs to form ten digit numbers.  In one addition time the
// constant transmitter can transmit to some other unit of the ENIAC any group
// of five or certain groups of ten digits and the associated sign.
//
// Whenever it is desired to set up new constants on the relays in the constant
// transmitter, the IBM reader must be programmed to read a new card.  This
// does not change the digits set on the constants switches; these have to be
// changed manually.
// -- ENIAC Technical Manual, Part II (Ch VIII)
type Constant struct {
	out           *Jack
	programInput  [30]*Jack
	programOutput [30]*Jack

	sel        [30]int
	cardSign   [8][2]bool
	cardDigits [8][10]int

	jSign [2]bool
	j     [10]int
	kSign [2]bool
	k     [10]int

	sign   bool
	digits []int // Digits for selected constant.  Index 0 is least significant.
	pos1pp int

	inff1, inff2 [30]bool
	whichrp      bool

	tracer Tracer

	mu sync.Mutex
}

func NewConstant() *Constant {
	u := &Constant{}
	for i := 0; i < 30; i++ {
		u.programInput[i] = u.newProgramInput(fmt.Sprintf("c.%di", i+1), i)
		u.programOutput[i] = u.newOutput(fmt.Sprintf("c.%do", i+1), 1)
	}
	u.out = u.newOutput("c.o", 11)
	return u
}

func (u *Constant) newProgramInput(name string, program int) *Jack {
	return NewInput(name, func(j *Jack, val int) {
		if val == 1 {
			u.trigger(program)
			if u.tracer != nil {
				u.tracer.LogPulse(j.Name, 1, int64(val))
			}
		}
	})
}

func (u *Constant) newOutput(name string, width int) *Jack {
	return NewOutput(name, func(j *Jack, val int) {
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, width, int64(val))
		}
	})
}

func (u *Constant) AttachTracer(tracer Tracer) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.tracer = tracer
	u.tracer.RegisterValueCallback(func() {
		u.tracer.LogValue("c.sign", 1, BoolToInt64(u.sign))
		u.tracer.LogValue("c.constant", 40, DigitsToInt64BCD(u.digits))
	})
}

func (u *Constant) Stat() string {
	s := ""
	for _, f := range u.inff2 {
		s += ToBin(f)
	}
	return s
}

func (u *Constant) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	for i := 0; i < 30; i++ {
		u.sel[i] = 0
		u.programInput[i].Disconnect()
		u.inff1[i] = false
		u.inff2[i] = false
		u.programOutput[i].Disconnect()
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 10; j++ {
			u.cardDigits[i][j] = 0
		}
		u.cardSign[i][0] = false
		u.cardSign[i][1] = false
	}
	for i := 0; i < 10; i++ {
		u.j[i] = 0
		u.k[i] = 0
	}
	u.jSign[0] = false
	u.jSign[1] = false
	u.kSign[0] = false
	u.kSign[1] = false
	u.out.Disconnect()
}

func (u *Constant) FindJack(jack string) (*Jack, error) {
	if jack == "o" {
		return u.out, nil
	}
	var prog int
	var ilk rune
	fmt.Sscanf(jack, "%d%c", &prog, &ilk)
	if !(prog >= 1 && prog <= 30) {
		return nil, fmt.Errorf("invalid jack %s", jack)
	}
	switch ilk {
	case 'i':
		return u.programInput[prog-1], nil
	case 'o':
		return u.programOutput[prog-1], nil
	}
	return nil, fmt.Errorf("invalid jack %s", jack)
}

const (
	selLeft  = 0
	selRight = 1
	selBoth  = 2
	selC1    = 0
	selC2    = 3
)

type selSwitch struct {
	owner sync.Locker
	name  string
	prog  int
	data  *int
}

func (s *selSwitch) Get() string {
	s.owner.Lock()
	defer s.owner.Unlock()
	constants := "abcdefghjk"
	i := 2 * ((s.prog - 1) / 6)
	switch *s.data {
	case selC1 + selLeft:
		return fmt.Sprintf("%cl", constants[i])
	case selC1 + selRight:
		return fmt.Sprintf("%cr", constants[i])
	case selC1 + selBoth:
		return fmt.Sprintf("%clr", constants[i])
	case selC2 + selLeft:
		return fmt.Sprintf("%cl", constants[i+1])
	case selC2 + selRight:
		return fmt.Sprintf("%cr", constants[i+1])
	case selC2 + selBoth:
		return fmt.Sprintf("%clr", constants[i+1])
	}
	return "?"
}

func (s *selSwitch) Set(value string) error {
	s.owner.Lock()
	defer s.owner.Unlock()
	if len(value) < 2 {
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	var n int
	switch value[0] {
	case 'a', 'A', 'c', 'C', 'e', 'E', 'g', 'G', 'j', 'J':
		n = selC1
	case 'b', 'B', 'd', 'D', 'f', 'F', 'h', 'H', 'k', 'K':
		n = selC2
	default:
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	switch value[1:] {
	case "l":
		*s.data = n + selLeft
	case "r":
		*s.data = n + selRight
	case "lr":
		*s.data = n + selBoth
	default:
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	return nil
}

func (u *Constant) FindSwitch(name string) (Switch, error) {
	if len(name) < 2 {
		return nil, fmt.Errorf("invalid switch")
	}
	switch name[0] {
	case 's':
		prog, _ := strconv.Atoi(name[1:])
		if !(prog >= 1 && prog <= 30) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &selSwitch{&u.mu, name, prog, &u.sel[prog-1]}, nil
	case 'j', 'J':
		if name[1] == 'l' {
			return &BoolSwitch{&u.mu, name, &u.jSign[0], constantSignSettings()}, nil
		}
		if name[1] == 'r' {
			return &BoolSwitch{&u.mu, name, &u.jSign[1], constantSignSettings()}, nil
		}
		digit, _ := strconv.Atoi(name[1:])
		if !(digit >= 1 && digit <= 10) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.j[digit-1], constantDigitSettings()}, nil
	case 'k', 'K':
		if name[1] == 'l' {
			return &BoolSwitch{&u.mu, name, &u.kSign[0], constantSignSettings()}, nil
		}
		if name[1] == 'r' {
			return &BoolSwitch{&u.mu, name, &u.kSign[1], constantSignSettings()}, nil
		}
		digit, _ := strconv.Atoi(name[1:])
		if !(digit >= 1 && digit <= 10) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.k[digit-1], constantDigitSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

func (u *Constant) ReadCard(card string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	n := len(card)
	if n > 80 {
		n = 80
	}
	for i := 0; i < n/10; i++ {
		u.readCardField(i, card[10*i:10*i+10])
	}
}

func (u *Constant) readCardField(i int, field string) {
	if len(field) < 10 || field == "          " {
		return
	}
	bank := i / 2
	tenDigitNumber := true
	for j := bank * 6; j < bank*6+6; j++ {
		switch u.sel[j] {
		case selC1 + selLeft, selC1 + selRight:
			if i%2 == 0 && u.programInput[j].Connected() {
				tenDigitNumber = false
			}
		case selC2 + selLeft, selC2 + selRight:
			if i%2 == 1 && u.programInput[j].Connected() {
				tenDigitNumber = false
			}
		}
	}
	if tenDigitNumber {
		u.readConstant(i, 0, field)
	} else {
		left5, right5 := field[:5], field[5:]
		u.readConstant(i, 0, right5)
		u.readConstant(i, 5, left5)
	}
}

func (u *Constant) readConstant(i, offset int, field string) {
	sign, digits := IBMCardToNinesComplement(field)
	if offset == 0 {
		u.cardSign[i][0] = sign
	} else {
		u.cardSign[i][1] = sign
	}
	for j := range digits {
		u.cardDigits[i][offset+j] = digits[j]
	}
}

func (u *Constant) selectConstant(program int) {
	if program >= 24 {
		u.pos1pp = -1
		switch u.sel[program] {
		case selC1 + selLeft:
			u.sign = u.jSign[0]
			u.digits = make([]int, 10)
			copy(u.digits[5:], u.j[5:])
		case selC1 + selRight:
			u.sign = u.jSign[1]
			u.digits = make([]int, 10)
			if u.sign {
				for i := 5; i < 10; i++ {
					u.digits[i] = 9
				}
			}
			copy(u.digits[:5], u.j[:5])
		case selC1 + selBoth:
			u.sign, u.digits = u.jSign[0], u.j[:]
		case selC2 + selLeft:
			u.sign = u.kSign[0]
			u.digits = make([]int, 10)
			copy(u.digits[5:], u.k[5:])
		case selC2 + selRight:
			u.sign = u.kSign[1]
			u.digits = make([]int, 10)
			if u.sign {
				for i := 5; i < 10; i++ {
					u.digits[i] = 9
				}
			}
			copy(u.digits[:5], u.k[:5])
		case selC2 + selBoth:
			u.sign, u.digits = u.kSign[0], u.k[:]
		}
	} else {
		bank := program / 6
		switch u.sel[program] {
		case selC1 + selLeft:
			u.sign = u.cardSign[2*bank][1]
			u.digits = make([]int, 10)
			copy(u.digits[5:], u.cardDigits[2*bank][5:])
			u.pos1pp = 5
		case selC1 + selRight:
			u.sign = u.cardSign[2*bank][0]
			u.digits = make([]int, 10)
			if u.sign {
				for i := 5; i < 10; i++ {
					u.digits[i] = 9
				}
			}
			copy(u.digits[:5], u.cardDigits[2*bank][:5])
			u.pos1pp = 0
		case selC1 + selBoth:
			u.sign, u.digits, u.pos1pp = u.cardSign[2*bank][0], u.cardDigits[2*bank][:], 0
		case selC2 + selLeft:
			u.sign = u.cardSign[2*bank+1][1]
			u.digits = make([]int, 10)
			copy(u.digits[5:], u.cardDigits[2*bank+1][5:])
			u.pos1pp = 5
		case selC2 + selRight:
			u.sign = u.cardSign[2*bank+1][0]
			u.digits = make([]int, 10)
			if u.sign {
				for i := 5; i < 10; i++ {
					u.digits[i] = 9
				}
			}
			copy(u.digits[:5], u.cardDigits[2*bank+1][:5])
			u.pos1pp = 0
		case selC2 + selBoth:
			u.sign, u.digits, u.pos1pp = u.cardSign[2*bank+1][0], u.cardDigits[2*bank+1][:], 0
		}
	}
}

func (u *Constant) Clock(cyc Pulse) {
	sending := -1
	for i := 0; i < 30; i++ {
		if u.inff2[i] {
			// FIXME "Simultaneous stimulation of two program controls" is actually
			// allowed in some restricted cases described in ETM, Table 8-6.  This
			// implementation assumes only one active program.
			sending = i
			u.selectConstant(i)
			break
		}
	}
	if cyc&Ccg != 0 {
		u.whichrp = false
	} else if cyc&Rp != 0 {
		if u.whichrp {
			for i := 0; i < 30; i++ {
				if u.inff1[i] {
					u.inff1[i] = false
					u.inff2[i] = true
				}
			}
			u.whichrp = false
		} else {
			u.whichrp = true
		}
	}
	if sending > -1 {
		if cyc&Cpp != 0 {
			u.programOutput[sending].Transmit(1)
			u.inff2[sending] = false
			sending = -1
		} else if cyc&Ninep != 0 {
			n := 0
			for i := 0; i < 10; i++ {
				if cyc&BCD[u.digits[i]] != 0 {
					n |= 1 << uint(i)
				}
			}
			if u.sign {
				n |= 1 << 10
			}
			if n != 0 {
				u.out.Transmit(n)
			}
		} else if cyc&Onepp != 0 && u.pos1pp >= 0 && u.sign {
			u.out.Transmit(1 << uint(u.pos1pp))
		}
	}
}

func (u *Constant) trigger(input int) {
	u.mu.Lock()
	u.inff1[input] = true
	u.mu.Unlock()
}
