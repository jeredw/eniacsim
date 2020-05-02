package units

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

// Simulates an ENIAC function table unit.
type Ft struct {
	jack [27]chan Pulse

	unit int

	inff1, inff2 [11]bool
	opsw         [11]int
	rptsw        [11]int
	argsw        [11]int

	pm1, pm2         int
	cons             [8]int
	del              [8]int
	sub              [12]int
	tab              [104][14]int
	arg              int
	ring             int
	add              bool
	subtr            bool
	argsetup         bool
	gateh42, gatee42 bool
	resp             chan int
	whichrp          bool
	px4119           bool
	prog             int

	rewiring           chan int
	waitingForRewiring chan int

	mu sync.Mutex
}

func NewFt(unit int) *Ft {
	return &Ft{
		unit:               unit,
		rewiring:           make(chan int),
		waitingForRewiring: make(chan int),
		resp:               make(chan int),
	}
}

func (u *Ft) Stat() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := ""
	for i := range u.inff2 {
		if u.inff2[i] {
			s += "1"
		} else {
			s += "0"
		}
	}
	s += fmt.Sprintf(" %d %d", u.arg, u.ring)
	if u.add {
		s += " 1"
	} else {
		s += " 0"
	}
	if u.subtr {
		s += " 1"
	} else {
		s += " 0"
	}
	if u.argsetup {
		s += " 1"
	} else {
		s += " 0"
	}
	return s
}

type ftJson struct {
	ArgUnits int      `json:"argUnits"`
	ArgTens  int      `json:"argTens"`
	Ring     int      `json:"ring"`
	ArgSetup bool     `json:"argSetup"`
	Add      bool     `json:"add"`
	Subtract bool     `json:"subtract"`
	Inff     [11]bool `json:"inff"`
}

func (u *Ft) State() json.RawMessage {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := ftJson{
		Inff:     u.inff2,
		ArgUnits: u.arg % 10,
		ArgTens:  u.arg / 10,
		Ring:     u.ring,
		Add:      u.add,
		Subtract: u.subtr,
		ArgSetup: u.argsetup,
	}
	result, _ := json.Marshal(s)
	return result
}

func (u *Ft) Reset() {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	for i := range u.jack {
		u.jack[i] = nil
	}
	for i := range u.inff1 {
		u.inff1[i] = false
		u.inff2[i] = false
		u.opsw[i] = 0
		u.rptsw[i] = 0
		u.argsw[i] = 0
	}
	u.pm1 = 0
	u.pm2 = 0
	for i := range u.cons {
		u.cons[i] = 0
		u.del[i] = 0
	}
	for i := range u.sub {
		u.sub[i] = 0
	}
	for i := 0; i < len(u.tab); i++ {
		for j := 0; j < len(u.tab[0]); j++ {
			u.tab[i][j] = 0
		}
	}
	u.arg = 0
	u.ring = 0
	u.add = false
	u.subtr = false
	u.argsetup = false
	u.gateh42 = false
	u.gatee42 = false
	u.whichrp = false
	u.px4119 = false
}

func (u *Ft) Plug(jack string, ch chan Pulse, output bool) error {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()

	name := "ft" + strconv.Itoa(u.unit+1) + "." + jack
	switch jack {
	case "arg", "ARG":
		SafePlug(name, &u.jack[0], ch, output)
	case "A":
		SafePlug(name, &u.jack[1], ch, output)
	case "B":
		SafePlug(name, &u.jack[2], ch, output)
	case "NC":
		SafePlug(name, &u.jack[3], ch, output)
	case "C":
		SafePlug(name, &u.jack[4], ch, output)
	default:
		jacks := [22]string{
			"1i", "1o", "2i", "2o", "3i", "3o", "4i", "4o",
			"5i", "5o", "6i", "6o", "7i", "7o", "8i", "8o", "9i", "9o",
			"10i", "10o", "11i", "11o",
		}
		for i, j := range jacks {
			if j == jack {
				SafePlug(name, &u.jack[i+5], ch, output)
				return nil
			}
		}
		return fmt.Errorf("invalid jack %s", name)
	}
	return nil
}

func (u *Ft) lookupSwitch(name string) (Switch, error) {
	switch {
	case len(name) > 2 && name[:2] == "op":
		sw, _ := strconv.Atoi(name[2:])
		if !(sw >= 1 && sw <= 11) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.opsw[sw-1], ftOpSettings()}, nil
	case len(name) > 2 && name[:2] == "cl":
		sw, _ := strconv.Atoi(name[2:])
		if !(sw >= 1 && sw <= 11) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.argsw[sw-1], ftArgSettings()}, nil
	case len(name) > 2 && name[:2] == "rp":
		sw, _ := strconv.Atoi(name[2:])
		if !(sw >= 1 && sw <= 11) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.rptsw[sw-1], rpSettings()}, nil
	case name == "mpm1":
		return &IntSwitch{name, &u.pm1, signSettings()}, nil
	case name == "mpm2":
		return &IntSwitch{name, &u.pm2, signSettings()}, nil
	case len(name) > 1 && name[0] == 'A', len(name) > 1 && name[0] == 'B':
		var bank, digit, ilk int
		fmt.Sscanf(name, "%c%d%c", &bank, &digit, &ilk)
		var offset int
		switch bank {
		case 'A':
			offset = 0
		case 'B':
			offset = 1
		default:
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		switch ilk {
		case 'd', 'D':
			if !(digit >= 1 && digit <= 4) {
				return nil, fmt.Errorf("invalid switch %s", name)
			}
			return &IntSwitch{name, &u.del[4*offset+digit-1], delSettings()}, nil
		case 'c', 'C':
			if !(digit >= 1 && digit <= 4) {
				return nil, fmt.Errorf("invalid switch %s", name)
			}
			return &IntSwitch{name, &u.cons[4*offset+digit-1], consSettings()}, nil
		case 's', 'S':
			if !(digit >= 4 && digit <= 10) {
				return nil, fmt.Errorf("invalid switch %s", name)
			}
			return &IntSwitch{name, &u.sub[6*offset+digit-5], subSettings()}, nil
		default:
			return nil, fmt.Errorf("invalid switch %s", name)
		}
	case len(name) > 1 && name[0] == 'R':
		var bank, row, digit int
		n, _ := fmt.Sscanf(name, "R%c%dL%d", &bank, &row, &digit)
		if n == 3 {
			if !(row >= -2 && row <= 101) {
				return nil, fmt.Errorf("invalid switch %s", name)
			}
			if !(digit >= 1 && digit <= 6) {
				return nil, fmt.Errorf("invalid switch %s", name)
			}
			switch bank {
			case 'A':
				return &IntSwitch{name, &u.tab[row+2][7-digit], valSettings()}, nil
			case 'B':
				return &IntSwitch{name, &u.tab[row+2][13-digit], valSettings()}, nil
			default:
				return nil, fmt.Errorf("invalid switch %s", name)
			}
		} else {
			fmt.Sscanf(name, "R%c%dS", &bank, &row)
			if !(row >= -2 && row <= 101) {
				return nil, fmt.Errorf("invalid switch %s", name)
			}
			switch bank {
			case 'A':
				return &IntSwitch{name, &u.tab[row+2][0], pmSettings()}, nil
			case 'B':
				return &IntSwitch{name, &u.tab[row+2][13], pmSettings()}, nil
			default:
				return nil, fmt.Errorf("invalid switch %s", name)
			}
		}
	case name == "ninep" || name == "Ninep":
		return &BoolSwitch{name, &u.px4119, ninepSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

func (u *Ft) SetSwitch(name, value string) error {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return err
	}
	return sw.Set(value)
}

func (u *Ft) GetSwitch(name string) (string, error) {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return "", err
	}
	return sw.Get(), nil
}

func (u *Ft) addlookup(c int) {
	a := 0
	b := 0
	arg := u.arg
	if c&Ninep != 0 {
		as := u.pm1 == 1 || u.pm1 == 2 && u.tab[arg][0] == 1
		bs := u.pm2 == 1 || u.pm2 == 2 && u.tab[arg][13] == 1
		if as {
			a |= 1 << 10
		}
		if bs {
			b |= 1 << 10
		}
		for i := 0; i < 4; i++ {
			if u.del[i] == 0 {
				if u.cons[i] == 10 && as {
					a |= 1 << (9 - uint(i))
				} else if u.cons[i] == 11 && bs {
					a |= 1 << (9 - uint(i))
				}
			}
			if u.del[i+4] == 0 {
				if u.cons[i+4] == 10 && as {
					b |= 1 << (9 - uint(i))
				} else if u.cons[i+4] == 11 && bs {
					b |= 1 << (9 - uint(i))
				}
			}
		}
		for i := 0; i < 4; i++ {
			if u.cons[i] == 9 {
				a |= 1 << (9 - uint(i))
			}
			if u.cons[i+4] == 9 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if u.tab[arg][i+1] == 9 {
				a |= 1 << (5 - uint(i))
			}
			if u.tab[arg][i+7] == 9 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Fourp != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x >= 4 && x <= 8 {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x >= 4 && x <= 8 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x >= 4 && x <= 8 {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x >= 4 && x <= 8 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Twopp != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x == 8 || x == 8 {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x == 8 || x == 8 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x == 8 || x == 8 {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x == 8 || x == 8 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Twop != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x == 2 || x == 3 || (x > 5 && x < 9) {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x == 2 || x == 3 || (x > 5 && x < 9) {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x == 2 || x == 3 || (x > 5 && x < 9) {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x == 2 || x == 3 || (x > 5 && x < 9) {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Onep != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x < 9 && x%2 == 1 {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x < 9 && x%2 == 1 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x < 9 && x%2 == 1 {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x < 9 && x%2 == 1 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if a != 0 {
		Handshake(a, u.jack[1], u.resp)
	}
	if b != 0 {
		Handshake(b, u.jack[1], u.resp)
	}
}

func (u *Ft) subtrlookup(c int) {
	a := 0
	b := 0
	arg := u.arg
	if c&Ninep != 0 {
		as := u.pm1 == 0 || u.pm1 == 2 && u.tab[arg][0] == 0
		bs := u.pm2 == 0 || u.pm2 == 2 && u.tab[arg][13] == 0
		if as {
			a |= 1 << 10
		}
		if bs {
			b |= 1 << 10
		}
		for i := 0; i < 4; i++ {
			if u.del[i] == 0 {
				if u.cons[i] == 10 && as {
					a |= 1 << (9 - uint(i))
				} else if u.cons[i] == 11 && bs {
					a |= 1 << (9 - uint(i))
				}
			}
			if u.del[i+4] == 0 {
				if u.cons[i+4] == 10 && as {
					b |= 1 << (9 - uint(i))
				} else if u.cons[i+4] == 11 && bs {
					b |= 1 << (9 - uint(i))
				}
			}
		}
	}
	if c&Fourp != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x < 6 {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x < 6 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x < 6 {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x < 6 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Twopp != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x < 2 {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x < 2 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x < 2 {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x < 2 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Twop != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x == 6 || x == 7 || x < 4 {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x == 6 || x == 7 || x < 4 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x == 6 || x == 7 || x < 4 {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x == 6 || x == 7 || x < 4 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Onep != 0 {
		for i := 0; i < 4; i++ {
			if x := u.cons[i]; x < 10 && x%2 == 0 {
				a |= 1 << (9 - uint(i))
			}
			if x := u.cons[i+4]; x < 10 && x%2 == 0 {
				b |= 1 << (9 - uint(i))
			}
		}
		for i := 0; i < 6; i++ {
			if x := u.tab[arg][i+1]; x < 10 && x%2 == 0 {
				a |= 1 << (5 - uint(i))
			}
			if x := u.tab[arg][i+7]; x < 10 && x%2 == 0 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if c&Onepp != 0 {
		for i := 0; i < 6; i++ {
			if u.sub[i] == 1 {
				a |= 1 << (5 - uint(i))
			}
			if u.sub[i+6] == 1 {
				b |= 1 << (5 - uint(i))
			}
		}
	}
	if a != 0 {
		Handshake(a, u.jack[1], u.resp)
	}
	if b != 0 {
		Handshake(b, u.jack[2], u.resp)
	}
}

func (u *Ft) clock(p Pulse) {
	//  u.mu.Lock()
	//  defer u.mu.Unlock()
	if u.px4119 {
		if p.Val&Cpp != 0 {
			p.Val |= Ninep
		} else {
			p.Val &= ^Ninep
		}
	}
	c := p.Val
	if u.gatee42 {
		sw := u.opsw[u.prog]
		if c&Onep != 0 && (sw == 1 || sw == 3 || sw == 6 || sw == 8) {
			u.arg++
		}
		if c&Twop != 0 && (sw == 2 || sw == 3 || sw == 6 || sw == 7) {
			u.arg++
		}
		if c&Fourp != 0 && (sw == 4 || sw == 5) {
			u.arg++
		}
	}
	if u.add {
		if u.arg >= 0 && u.arg < 104 {
			u.addlookup(c)
		} else {
			fmt.Println("Invalid function table argument", u.arg)
		}
	}
	if u.subtr {
		if u.arg >= 0 && u.arg < 104 {
			u.subtrlookup(c)
		} else {
			fmt.Println("Invalid function table argument", u.arg)
		}
	}
	if c&Cpp != 0 {
		switch u.ring {
		case 0: // Stage -3
			for u.prog = 0; u.prog < 11 && !u.inff2[u.prog]; u.prog++ {
			}
			if u.prog >= 11 {
				break
			}
			switch u.argsw[u.prog] {
			case 1:
				Handshake(1, u.jack[3], u.resp)
			case 2:
				Handshake(1, u.jack[4], u.resp)
			}
			u.ring++ // Stage -2 begins
			u.gateh42 = true
		case 1:
			u.ring++ // Stage -1 begins
			u.gateh42 = false
			u.gatee42 = true
		case 2:
			u.ring++ // Stage 0 begins
			u.gatee42 = false
			/*
				if u.opsw[u.prog] < 5 {
					u.add = true
				} else {
					u.subtr = true
				}
			*/
		case 3: // Stage 0
			u.ring++ // Stage 1 begins
			if u.opsw[u.prog] < 5 {
				u.add = true
			} else {
				u.subtr = true
			}
		default: // Stages 1-9
			if u.rptsw[u.prog] == u.ring-4 {
				Handshake(1, u.jack[u.prog*2+6], u.resp)
				u.arg = 0
				u.add = false
				u.subtr = false
				u.inff2[u.prog] = false
				u.argsetup = false
				u.ring = 0
			} else {
				u.ring++
			}
		}
	}
	if c&Ccg != 0 {
		u.whichrp = false
	}
	if c&Rp != 0 {
		if u.whichrp {
			for i, _ := range u.inff1 {
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
	if u.ring == 2 && c&Onepp != 0 {
		u.argsetup = true
	}
}

func (u *Ft) MakeClockFunc() ClockFunc {
	return func(p Pulse) {
		u.clock(p)
	}
}

func (u *Ft) trigger(input int) {
	u.mu.Lock()
	u.inff1[input] = true
	u.mu.Unlock()
}

func (u *Ft) Run() {
	var p Pulse

	for {
		p.Resp = nil
		select {
		case <-u.rewiring:
			u.waitingForRewiring <- 1
			<-u.rewiring
		case p = <-u.jack[5]:
			u.trigger(0)
		case p = <-u.jack[7]:
			u.trigger(1)
		case p = <-u.jack[9]:
			u.trigger(2)
		case p = <-u.jack[11]:
			u.trigger(3)
		case p = <-u.jack[13]:
			u.trigger(4)
		case p = <-u.jack[15]:
			u.trigger(5)
		case p = <-u.jack[17]:
			u.trigger(6)
		case p = <-u.jack[19]:
			u.trigger(7)
		case p = <-u.jack[21]:
			u.trigger(8)
		case p = <-u.jack[23]:
			u.trigger(9)
		case p = <-u.jack[25]:
			u.trigger(10)
		case p = <-u.jack[0]:
			u.mu.Lock()
			if u.gateh42 {
				if p.Val&0x01 != 0 {
					u.arg++
				}
				if p.Val&0x02 != 0 {
					u.arg += 10
				}
			}
			u.mu.Unlock()
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}
