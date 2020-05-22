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
// of five or certain groups ot ten digits and the associated sign.
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

	sel          [30]int
	cardSign     [8][2]bool
	cardDigit    [8][10]int

	jSign        [2]bool
	j            [10]int
	kSign        [2]bool
	k            [10]int

	value   []int
	sign    bool
	pos1pp  int

	inff1, inff2 [30]bool
	whichrp bool

	tracePulse TraceFunc

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
		}
	})
}

func (u *Constant) newOutput(name string, width int) *Jack {
	return NewOutput(name, func(j *Jack, val int) {
		if u.tracePulse != nil {
			u.tracePulse(j.Name, width, int64(val))
		}
	})
}

func (u *Constant) AttachTrace(tracePulse TraceFunc) []func(TraceFunc) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.tracePulse = tracePulse
	return []func(TraceFunc){}
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
			u.cardDigit[i][j] = 0
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
	i := 2 * (s.prog - 1) / 6
	switch *s.data {
	case 0:
		return fmt.Sprintf("%cl", constants[i])
	case 1:
		return fmt.Sprintf("%cr", constants[i])
	case 2:
		return fmt.Sprintf("%clr", constants[i])
	case 3:
		return fmt.Sprintf("%cl", constants[i+1])
	case 4:
		return fmt.Sprintf("%cr", constants[i+1])
	case 5:
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
		n = 0
	case 'b', 'B', 'd', 'D', 'f', 'F', 'h', 'H', 'k', 'K':
		n = 3
	default:
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	switch value[1:] {
	case "l":
		*s.data = n
	case "r":
		*s.data = n + 1
	case "lr":
		*s.data = n + 2
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
		case 0, 1:
			if i%2 == 0 && u.programInput[j].Connected() {
				tenDigitNumber = false
			}
		case 3, 4:
			if i%2 == 1 && u.programInput[j].Connected() {
				tenDigitNumber = false
			}
		}
	}
	if tenDigitNumber {
		u.readConstant(i, 0, field)
	} else {
		u.readConstant(i, 0, field[:5])
		u.readConstant(i, 5, field[5:])
	}
}

func (u *Constant) readConstant(i, offset int, field string) {
	sign, digits := FromIBMCard(field)
	if offset == 0 {
		u.cardSign[i][0] = sign
	} else {
		u.cardSign[i][1] = sign
	}
	for j := range digits {
		u.cardDigit[i][offset+j] = digits[j]
	}
}

func (u *Constant) getval(sel int) (sgn bool, val []int, pos1pp int) {
	if sel >= 24 {
		pos1pp = -1
		switch u.sel[sel] {
		case 0:
			sgn = u.jSign[0]
			val = make([]int, 10)
			copy(val[5:], u.j[5:])
			return
		case 1:
			sgn = u.jSign[1]
			val = make([]int, 10)
			if sgn {
				for i := 5; i < 10; i++ {
					val[i] = 9
				}
			}
			copy(val[:5], u.j[:5])
			return
		case 2:
			return u.jSign[0], u.j[:], -1
		case 3:
			sgn = u.kSign[0]
			val = make([]int, 10)
			copy(val[5:], u.k[5:])
			return
		case 4:
			sgn = u.kSign[1]
			val = make([]int, 10)
			if sgn {
				for i := 5; i < 10; i++ {
					val[i] = 9
				}
			}
			copy(val[:5], u.k[:5])
			return
		case 5:
			return u.kSign[0], u.k[:], -1
		}
	} else {
		bank := sel / 6
		switch u.sel[sel] {
		case 0:
			sgn = u.cardSign[2*bank][1]
			val = make([]int, 10)
			if sgn {
				for i := 0; i < 5; i++ {
					val[i] = 9
				}
			}
			copy(val[5:], u.cardDigit[2*bank][5:])
			pos1pp = 0
			return
		case 1:
			sgn = u.cardSign[2*bank][0]
			val = make([]int, 10)
			copy(val[:5], u.cardDigit[2*bank][:5])
			pos1pp = 5
			return
		case 2:
			return u.cardSign[2*bank][0], u.cardDigit[2*bank][:], 0
		case 3:
			sgn = u.cardSign[2*bank+1][1]
			val = make([]int, 10)
			if sgn {
				for i := 0; i < 5; i++ {
					val[i] = 9
				}
			}
			copy(val[5:], u.cardDigit[2*bank+1][5:])
			pos1pp = 0
			return
		case 4:
			sgn = u.cardSign[2*bank+1][0]
			val = make([]int, 10)
			copy(val[:5], u.cardDigit[2*bank+1][:5])
			pos1pp = 5
			return
		case 5:
			return u.cardSign[2*bank+1][0], u.cardDigit[2*bank+1][:], 0
		}
	}
	return
}

var digitcons = []Pulse{0, Onep, Twop, (Onep | Twop), Fourp, (Onep | Fourp),
	(Twop | Fourp), (Onep | Twop | Fourp), (Twop | Twopp | Fourp),
	(Onep | Twop | Twopp | Fourp)}

func (u *Constant) Clock(cyc Pulse) {
	//	u.mu.Lock()
	//	defer u.mu.Unlock()
	sending := -1
	for i := 0; i < 30; i++ {
		if u.inff2[i] {
			sending = i
			u.sign, u.value, u.pos1pp = u.getval(i)
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
			for i := uint(0); i < uint(10); i++ {
				if cyc&digitcons[u.value[i]] != 0 {
					n |= 1 << i
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
