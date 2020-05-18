package units

import (
	"fmt"
	"strconv"
	"sync"
	"unicode"

	. "github.com/jeredw/eniacsim/lib"
)

// Simulates ENIAC constant transmitter unit.
type Constant struct {
	sel          [30]int
	card         [8][10]byte
	signcard     [8][2]byte
	j            [10]byte
	signj        [2]byte
	k            [10]byte
	signk        [2]byte
	out          *Jack
	pin          [30]*Jack
	inff1, inff2 [30]bool
	pout         [30]*Jack

	val     []byte
	sign    byte
	pos1pp  int
	whichrp bool

	tracePulse TraceFunc

	mu sync.Mutex
}

func NewConstant() *Constant {
	u := &Constant{}
	programInput := func(prog int) JackHandler {
		return func(j *Jack, val int) {
			if val == 1 {
				u.trigger(prog)
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
	for i := 0; i < 30; i++ {
		u.pin[i] = NewInput(fmt.Sprintf("c.%di", i+1), programInput(i))
		u.pout[i] = NewOutput(fmt.Sprintf("c.%do", i+1), output(1))
	}
	u.out = NewOutput("c.o", output(11))
	return u
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
		u.pin[i].Disconnect()
		u.inff1[i] = false
		u.inff2[i] = false
		u.pout[i].Disconnect()
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 10; j++ {
			u.card[i][j] = 0
		}
		u.signcard[i][0] = 0
		u.signcard[i][1] = 0
	}
	for i := 0; i < 10; i++ {
		u.j[i] = 0
		u.k[i] = 0
	}
	u.signj[0] = 0
	u.signj[1] = 0
	u.signk[0] = 0
	u.signk[1] = 0
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
		return u.pin[prog-1], nil
	case 'o':
		return u.pout[prog-1], nil
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
			return &ByteSwitch{&u.mu, name, &u.signj[0], constantSignSettings()}, nil
		}
		if name[1] == 'r' {
			return &ByteSwitch{&u.mu, name, &u.signj[1], constantSignSettings()}, nil
		}
		digit, _ := strconv.Atoi(name[1:])
		if !(digit >= 1 && digit <= 10) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &ByteSwitch{&u.mu, name, &u.j[digit-1], constantDigitSettings()}, nil
	case 'k', 'K':
		if name[1] == 'l' {
			return &ByteSwitch{&u.mu, name, &u.signk[0], constantSignSettings()}, nil
		}
		if name[1] == 'r' {
			return &ByteSwitch{&u.mu, name, &u.signk[1], constantSignSettings()}, nil
		}
		digit, _ := strconv.Atoi(name[1:])
		if !(digit >= 1 && digit <= 10) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &ByteSwitch{&u.mu, name, &u.k[digit-1], constantDigitSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

func (u *Constant) ReadCard(c string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	l := len(c)
	if l > 80 {
		l = 80
	}
	for i := 0; i < l/10; i++ {
		u.procfield(i, c[10*i:10*i+10])
	}
}

func (u *Constant) procfield(i int, f string) {
	if len(f) < 10 || f == "          " {
		return
	}
	bank := i / 2
	tendig := true
	for j := bank * 6; j < bank*6+6; j++ {
		switch u.sel[j] {
		case 0, 1:
			if i%2 == 0 && u.pin[j].Connected() {
				tendig = false
			}
		case 3, 4:
			if i%2 == 1 && u.pin[j].Connected() {
				tendig = false
			}
		}
	}
	if tendig {
		u.donum(i, 0, f)
	} else {
		u.donum(i, 0, f[:5])
		u.donum(i, 5, f[5:])
	}
}

func (u *Constant) donum(i, off int, f string) {
	var nz int

	neg := byte(0)
	for _, c := range f {
		if c == '-' || c == ']' || c == '}' || c >= 'J' && c <= 'R' {
			neg = 1
		}
	}
	if off == 0 {
		u.signcard[i][0] = neg
	} else {
		u.signcard[i][1] = neg
	}
	if neg == 0 {
		for j, c := range f {
			if unicode.IsDigit(c) {
				u.card[i][9-(j+off)] = charval(c)
			} else {
				u.card[i][9-(j+off)] = 0
			}
		}
		return
	}
	l := len(f)
	for nz = l - 1; nz >= 0 && f[nz] == '0'; nz-- {
		u.card[i][9-(nz+off)] = '0'
	}
	for ; nz >= 0; nz-- {
		u.card[i][9-(nz+off)] = 9 - charval(rune(f[nz]))
	}
}

func charval(c rune) byte {
	if c == '-' || c == ']' || c == '}' {
		return 0
	}
	if c >= 'J' && c <= 'R' {
		return byte(c - 'J' + 1)
	}
	return byte(c - '0')
}

func (u *Constant) getval(sel int) (sgn byte, val []byte, pos1pp int) {
	if sel >= 24 {
		pos1pp = -1
		switch u.sel[sel] {
		case 0:
			sgn = u.signj[0]
			val = make([]byte, 10)
			copy(val[5:], u.j[5:])
			return
		case 1:
			sgn = u.signj[1]
			val = make([]byte, 10)
			if sgn != 0 {
				for i := 5; i < 10; i++ {
					val[i] = 9
				}
			}
			copy(val[:5], u.j[:5])
			return
		case 2:
			return u.signj[0], u.j[:], -1
		case 3:
			sgn = u.signk[0]
			val = make([]byte, 10)
			copy(val[5:], u.k[5:])
			return
		case 4:
			sgn = u.signk[1]
			val = make([]byte, 10)
			if sgn != 0 {
				for i := 5; i < 10; i++ {
					val[i] = 9
				}
			}
			copy(val[:5], u.k[:5])
			return
		case 5:
			return u.signk[0], u.k[:], -1
		}
	} else {
		bank := sel / 6
		switch u.sel[sel] {
		case 0:
			sgn = u.signcard[2*bank][1]
			val = make([]byte, 10)
			if sgn != 0 {
				for i := 0; i < 5; i++ {
					val[i] = 9
				}
			}
			copy(val[5:], u.card[2*bank][5:])
			pos1pp = 0
			return
		case 1:
			sgn = u.signcard[2*bank][0]
			val = make([]byte, 10)
			copy(val[:5], u.card[2*bank][:5])
			pos1pp = 5
			return
		case 2:
			return u.signcard[2*bank][0], u.card[2*bank][:], 0
		case 3:
			sgn = u.signcard[2*bank+1][1]
			val = make([]byte, 10)
			if sgn != 0 {
				for i := 0; i < 5; i++ {
					val[i] = 9
				}
			}
			copy(val[5:], u.card[2*bank+1][5:])
			pos1pp = 0
			return
		case 4:
			sgn = u.signcard[2*bank+1][0]
			val = make([]byte, 10)
			copy(val[:5], u.card[2*bank+1][:5])
			pos1pp = 5
			return
		case 5:
			return u.signcard[2*bank+1][0], u.card[2*bank+1][:], 0
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
			u.sign, u.val, u.pos1pp = u.getval(i)
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
			u.pout[sending].Transmit(1)
			u.inff2[sending] = false
			sending = -1
		} else if cyc&Ninep != 0 {
			n := 0
			for i := uint(0); i < uint(10); i++ {
				if cyc&digitcons[u.val[i]] != 0 {
					n |= 1 << i
				}
			}
			if u.sign == 1 {
				n |= 1 << 10
			}
			if n != 0 {
				u.out.Transmit(n)
			}
		} else if cyc&Onepp != 0 && u.pos1pp >= 0 && u.sign == 1 {
			u.out.Transmit(1 << uint(u.pos1pp))
		}
	}
}

func (u *Constant) trigger(input int) {
	u.mu.Lock()
	u.inff1[input] = true
	u.mu.Unlock()
}
