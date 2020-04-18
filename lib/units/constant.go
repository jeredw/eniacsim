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
	out          chan Pulse
	pin          [30]chan Pulse
	inff1, inff2 [30]bool
	pout         [30]chan Pulse

	val     []byte
	sign    byte
	pos1pp  int
	whichrp bool

	rewiring           chan int
	waitingForRewiring chan int

	mu sync.Mutex
}

func NewConstant() *Constant {
	return &Constant{
		rewiring:           make(chan int),
		waitingForRewiring: make(chan int),
	}
}

func (u *Constant) Stat() string {
	s := ""
	for _, f := range u.inff2 {
		s += ToBin(f)
	}
	return s
}

func (u *Constant) Reset() {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	for i := 0; i < 30; i++ {
		u.sel[i] = 0
		u.pin[i] = nil
		u.inff1[i] = false
		u.inff2[i] = false
		u.pout[i] = nil
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
	u.out = nil
}

func (u *Constant) Plug(jack string, ch chan Pulse) error {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	name := "c." + jack
	if jack == "o" {
		SafePlug(name, &u.out, ch)
	} else {
		var prog int
		var ilk rune
		fmt.Sscanf(jack, "%d%c", &prog, &ilk)
		if !(prog >= 1 && prog <= 30) {
			return fmt.Errorf("invalid jack %s", name)
		}
		switch ilk {
		case 'i':
			SafePlug(name, &u.pin[prog-1], ch)
		case 'o':
			SafePlug(name, &u.pout[prog-1], ch)
		default:
			return fmt.Errorf("invalid jack %s", name)
		}
	}
	return nil
}

func (u *Constant) Switch(name, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if len(name) < 2 {
		return fmt.Errorf("invalid switch")
	}
	var n int
	switch name[0] {
	case 's':
		prog, _ := strconv.Atoi(name[1:])
		if !(prog >= 1 && prog <= 30) {
			return fmt.Errorf("invalid switch %s", name)
		}
		if len(value) < 2 {
			return fmt.Errorf("invalid switch %s setting %s", name, value)
		}
		switch value[0] {
		case 'a', 'A', 'c', 'C', 'e', 'E', 'g', 'G', 'j', 'J':
			n = 0
		case 'b', 'B', 'd', 'D', 'f', 'F', 'h', 'H', 'k', 'K':
			n = 3
		default:
			return fmt.Errorf("invalid switch %s setting %s", name, value)
		}
		switch value[1:] {
		case "l":
			u.sel[prog-1] = n
		case "r":
			u.sel[prog-1] = n + 1
		case "lr":
			u.sel[prog-1] = n + 2
		default:
			return fmt.Errorf("invalid switch %s setting %s", name, value)
		}
	case 'j', 'J':
		if name[1] == 'l' {
			switch value {
			case "P", "p":
				u.signj[0] = 0
			case "M", "m":
				u.signj[0] = 1
			default:
				return fmt.Errorf("invalid switch %s setting %s", name, value)
			}
		} else if name[1] == 'r' {
			switch value {
			case "P", "p":
				u.signj[1] = 0
			case "M", "m":
				u.signj[1] = 1
			default:
				return fmt.Errorf("invalid switch %s setting %s", name, value)
			}
		} else {
			digit, _ := strconv.Atoi(name[1:])
			if !(digit >= 1 && digit <= 10) {
				return fmt.Errorf("invalid switch %s", name)
			}
			n, _ := strconv.Atoi(value)
			if !(n >= 0 && n <= 9) {
				return fmt.Errorf("invalid switch %s setting %s", name, digit)
			}
			u.j[digit-1] = byte(n)
		}
	case 'k', 'K':
		if name[1] == 'l' {
			switch value {
			case "P", "p":
				u.signk[0] = 0
			case "M", "m":
				u.signk[0] = 1
			default:
				return fmt.Errorf("invalid switch %s setting %s", name, value)
			}
		} else if name[1] == 'r' {
			switch value {
			case "P", "p":
				u.signk[1] = 0
			case "M", "m":
				u.signk[1] = 1
			default:
				return fmt.Errorf("invalid switch %s setting %s", name, value)
			}
		} else {
			digit, _ := strconv.Atoi(name[1:])
			if !(digit >= 1 && digit <= 10) {
				return fmt.Errorf("invalid switch %s", name)
			}
			n, _ := strconv.Atoi(value)
			if !(n >= 0 && n <= 9) {
				return fmt.Errorf("invalid switch %s setting %s", name, digit)
			}
			u.k[digit-1] = byte(n)
		}
	default:
		return fmt.Errorf("invalid switch %s", name)
	}
	return nil
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
			if i%2 == 0 && u.pin[j] != nil {
				tendig = false
			}
		case 3, 4:
			if i%2 == 1 && u.pin[j] != nil {
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

var digitcons = []int{0, Onep, Twop, (Onep | Twop), Fourp, (Onep | Fourp),
	(Twop | Fourp), (Onep | Twop | Fourp), (Twop | Twopp | Fourp),
	(Onep | Twop | Twopp | Fourp)}

func (u *Constant) clock(p Pulse, resp chan int) {
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
	cyc := p.Val
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
			Handshake(1, u.pout[sending], resp)
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
				Handshake(n, u.out, resp)
			}
		} else if cyc&Onepp != 0 && u.pos1pp >= 0 && u.sign == 1 {
			Handshake(1<<uint(u.pos1pp), u.out, resp)
		}
	}
}

func (u *Constant) MakeClockFunc() ClockFunc {
	resp := make(chan int)
	return func(p Pulse) {
		u.clock(p, resp)
	}
}

func (u *Constant) Run() {
	var p Pulse

	for {
		p.Resp = nil
		select {
		case <-u.rewiring:
			u.waitingForRewiring <- 1
			<-u.rewiring
		case p = <-u.pin[0]:
			if p.Val == 1 {
				u.trigger(0)
			}
		case p = <-u.pin[1]:
			if p.Val == 1 {
				u.trigger(1)
			}
		case p = <-u.pin[2]:
			if p.Val == 1 {
				u.trigger(2)
			}
		case p = <-u.pin[3]:
			if p.Val == 1 {
				u.trigger(3)
			}
		case p = <-u.pin[4]:
			if p.Val == 1 {
				u.trigger(4)
			}
		case p = <-u.pin[5]:
			if p.Val == 1 {
				u.trigger(5)
			}
		case p = <-u.pin[6]:
			if p.Val == 1 {
				u.trigger(6)
			}
		case p = <-u.pin[7]:
			if p.Val == 1 {
				u.trigger(7)
			}
		case p = <-u.pin[8]:
			if p.Val == 1 {
				u.trigger(8)
			}
		case p = <-u.pin[9]:
			if p.Val == 1 {
				u.trigger(9)
			}
		case p = <-u.pin[10]:
			if p.Val == 1 {
				u.trigger(10)
			}
		case p = <-u.pin[11]:
			if p.Val == 1 {
				u.trigger(11)
			}
		case p = <-u.pin[12]:
			if p.Val == 1 {
				u.trigger(12)
			}
		case p = <-u.pin[13]:
			if p.Val == 1 {
				u.trigger(13)
			}
		case p = <-u.pin[14]:
			if p.Val == 1 {
				u.trigger(14)
			}
		case p = <-u.pin[15]:
			if p.Val == 1 {
				u.trigger(15)
			}
		case p = <-u.pin[16]:
			if p.Val == 1 {
				u.trigger(16)
			}
		case p = <-u.pin[17]:
			if p.Val == 1 {
				u.trigger(17)
			}
		case p = <-u.pin[18]:
			if p.Val == 1 {
				u.trigger(18)
			}
		case p = <-u.pin[19]:
			if p.Val == 1 {
				u.trigger(19)
			}
		case p = <-u.pin[20]:
			if p.Val == 1 {
				u.trigger(20)
			}
		case p = <-u.pin[21]:
			if p.Val == 1 {
				u.trigger(21)
			}
		case p = <-u.pin[22]:
			if p.Val == 1 {
				u.trigger(22)
			}
		case p = <-u.pin[23]:
			if p.Val == 1 {
				u.trigger(23)
			}
		case p = <-u.pin[24]:
			if p.Val == 1 {
				u.trigger(24)
			}
		case p = <-u.pin[25]:
			if p.Val == 1 {
				u.trigger(25)
			}
		case p = <-u.pin[26]:
			if p.Val == 1 {
				u.trigger(26)
			}
		case p = <-u.pin[27]:
			if p.Val == 1 {
				u.trigger(27)
			}
		case p = <-u.pin[28]:
			if p.Val == 1 {
				u.trigger(28)
			}
		case p = <-u.pin[29]:
			if p.Val == 1 {
				u.trigger(29)
			}
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}

func (u *Constant) trigger(input int) {
	u.mu.Lock()
	u.inff1[input] = true
	u.mu.Unlock()
}
