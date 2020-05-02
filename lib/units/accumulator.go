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
	α, β, γ, δ, ε, A, S chan Pulse
	ctlterm             [20]chan Pulse
	inff1, inff2        [12]bool
	opsw                [12]byte
	clrsw               [12]bool
	rptsw               [8]byte
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

	rewiring           chan int
	waitingForRewiring chan int
	resp               chan int

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
	return &Accumulator{
		unit:               unit,
		rewiring:           make(chan int),
		waitingForRewiring: make(chan int),
		resp:               make(chan int),
		sigfig:             10,
		lbuddy:             unit,
		rbuddy:             unit,
	}
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

func (u *Accumulator) Reset() {
	u.rewiring <- 1
	<-u.waitingForRewiring
	u.mu.Lock()
	u.α = nil
	u.β = nil
	u.γ = nil
	u.δ = nil
	u.ε = nil
	for i := 0; i < 12; i++ {
		u.ctlterm[i] = nil
		u.inff1[i] = false
		u.inff2[i] = false
		u.opsw[i] = 0
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
	u.rewiring <- 1
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

func (u *Accumulator) Plug(jack string, ch chan Pulse, output bool) error {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	name := "a" + strconv.Itoa(u.unit+1) + "." + jack
	switch {
	case jack == "α", jack == "a", jack == "alpha":
		SafePlug(name, &u.α, ch, output)
	case jack == "β", jack == "b", jack == "beta":
		SafePlug(name, &u.β, ch, output)
	case jack == "γ", jack == "g", jack == "gamma":
		SafePlug(name, &u.γ, ch, output)
	case jack == "δ", jack == "d", jack == "delta":
		SafePlug(name, &u.δ, ch, output)
	case jack == "ε", jack == "e", jack == "epsilon":
		SafePlug(name, &u.ε, ch, output)
	case jack == "A":
		SafePlug(name, &u.A, ch, output)
	case jack == "S":
		SafePlug(name, &u.S, ch, output)
	case jack[0] == 'I':
	default:
		jacks := [20]string{
			"1i", "2i", "3i", "4i", "5i", "5o", "6i", "6o", "7i", "7o",
			"8i", "8o", "9i", "9o", "10i", "10o", "11i", "11o", "12i", "12o",
		}
		for i, j := range jacks {
			if j == jack {
				u.ctlterm[i] = Tee(u.ctlterm[i], ch)
				return nil
			}
		}
		return fmt.Errorf("invalid jack: %s", jack)
	}
	return nil
}

type sfSwitch struct {
	name string
	data *int
}

func (s *sfSwitch) Get() string {
	return fmt.Sprintf("%d", *s.data)
}

func (s *sfSwitch) Set(value string) error {
	n, _ := strconv.Atoi(value)
	if !(n >= 0 && n <= 10) {
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	*s.data = n
	return nil
}

type scSwitch struct {
	name string
	data *byte
}

func (s *scSwitch) Get() string {
	switch *s.data {
	case 0:
		return "0"
	case 1:
		return "SC"
	}
	return "?"
}

func (s *scSwitch) Set(value string) error {
	switch value {
	case "0":
		*s.data = 0
	case "SC", "sc":
		*s.data = 1
	default:
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	return nil
}

type opSwitch struct {
	name string
	data *byte
}

func (s *opSwitch) Set(value string) error {
	switch value {
	case "α", "a", "alpha":
		*s.data = 0
	case "β", "b", "beta":
		*s.data = 1
	case "γ", "g", "gamma":
		*s.data = 2
	case "δ", "d", "delta":
		*s.data = 3
	case "ε", "e", "epsilon":
		*s.data = 4
	case "0":
		*s.data = 5
	case "A":
		*s.data = 6
	case "AS":
		*s.data = 7
	case "S":
		*s.data = 8
	default:
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	return nil
}

func (s *opSwitch) Get() string {
	switch *s.data {
	case 0:
		return "α"
	case 1:
		return "β"
	case 2:
		return "γ"
	case 3:
		return "δ"
	case 4:
		return "ε"
	case 5:
		return "0"
	case 6:
		return "A"
	case 7:
		return "AS"
	case 8:
		return "S"
	}
	return "?"
}

type ccSwitch struct {
	name string
	data *bool
}

func (s *ccSwitch) Get() string {
	if *s.data {
		return "C"
	}
	return "0"
}

func (s *ccSwitch) Set(value string) error {
	switch value {
	case "0":
		*s.data = false
	case "C", "c":
		*s.data = true
	default:
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	return nil
}

type rpSwitch struct {
	name string
	data *byte
}

func (s *rpSwitch) Get() string {
	return fmt.Sprintf("%d", int(1+*s.data))
}

func (s *rpSwitch) Set(value string) error {
	repeatCount, _ := strconv.Atoi(value)
	if !(repeatCount >= 1 && repeatCount <= 9) {
		return fmt.Errorf("invalid switch %s setting %s", s.name, value)
	}
	*s.data = byte(repeatCount - 1)
	return nil
}

func (u *Accumulator) lookupSwitch(name string) (Switch, error) {
	if name == "sf" {
		return &sfSwitch{name: name, data: &u.sigfig}, nil
	}
	if name == "sc" {
		return &scSwitch{name: name, data: &u.sc}, nil
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
		return &opSwitch{name: name, data: &u.opsw[prog]}, nil
	case "cc":
		return &ccSwitch{name: name, data: &u.clrsw[prog]}, nil
	case "rp":
		if !(prog >= 4 && prog <= 11) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &rpSwitch{name: name, data: &u.rptsw[prog-4]}, nil
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
	for i := range u.inff1 {
		if u.inff1[i] {
			switch u.opsw[i] {
			case 0:
				x |= stα
			case 1:
				x |= stβ
			case 2:
				x |= stγ
			case 3:
				x |= stδ
			case 4:
				x |= stε
			case 6:
				x |= stA
			case 7:
				x |= stAS
			case 8:
				x |= stS
			}
			if u.clrsw[i] {
				if u.opsw[i] >= 5 {
					if i < 4 || u.rep == int(u.rptsw[i-4]) {
						x |= stCLR
					}
				} else {
					x |= stCORR
				}
			}
		}
	}
	return x
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

func (u *Accumulator) docpp(cyc int) {
	for i := 0; i < 4; i++ {
		if u.inff2[i] {
			u.inff1[i] = false
			u.inff2[i] = false
		}
	}
	if u.h50 {
		u.rep++
		rstrep := false
		for i := 4; i < 12; i++ {
			if u.inff2[i] && u.rep == int(u.rptsw[i-4])+1 {
				u.inff1[i] = false
				u.inff2[i] = false
				rstrep = true
				t := (i-4)*2 + 5
				Handshake(1, u.ctlterm[t], u.resp)
			}
		}
		if rstrep {
			u.rep = 0
			u.h50 = false
		}
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
		if u.A != nil {
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
				Handshake(n, u.A, u.resp)
			}
		}
	}
	if curprog&(stAS|stS) != 0 {
		if u.S != nil {
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
				Handshake(n, u.S, u.resp)
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
	if curprog&(stAS|stS) != 0 && u.S != nil {
		if ((u.lbuddy < 0 || u.lbuddy == u.unit) && u.rbuddy == u.unit && u.sigfig > 0) ||
			(u.rbuddy != u.unit && u.sigfig < 10) ||
			(u.lbuddy != u.unit && u.lbuddy >= 0 && u.plbuddy.sigfig == 10 && u.sigfig > 0) ||
			(u.rbuddy != u.unit && u.sigfig == 10 && u.prbuddy.sigfig == 0) {
			Handshake(1<<uint(10-u.sigfig), u.S, u.resp)
		}
	}
}

func (u *Accumulator) clock(p Pulse) {
	cyc := p.Val
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

func (u *Accumulator) MakeClockFunc() ClockFunc {
	return func(p Pulse) {
		u.clock(p)
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
	u.mu.Unlock()
}

func (u *Accumulator) Run() {
	var p Pulse

	for {
		p.Resp = nil
		select {
		case <-u.rewiring:
			u.waitingForRewiring <- 1
			<-u.rewiring
		case p = <-u.α:
			u.mu.Lock()
			if u.st1()&stα != 0 {
				u.receive(p.Val)
			}
			u.mu.Unlock()
		case p = <-u.β:
			u.mu.Lock()
			if u.st1()&stβ != 0 {
				u.receive(p.Val)
			}
			u.mu.Unlock()
		case p = <-u.γ:
			u.mu.Lock()
			if u.st1()&stγ != 0 {
				u.receive(p.Val)
			}
			u.mu.Unlock()
		case p = <-u.δ:
			u.mu.Lock()
			if u.st1()&stδ != 0 {
				u.receive(p.Val)
			}
			u.mu.Unlock()
		case p = <-u.ε:
			u.mu.Lock()
			if u.st1()&stε != 0 {
				u.receive(p.Val)
			}
			u.mu.Unlock()
		case p = <-u.ctlterm[0]:
			if p.Val == 1 {
				u.trigger(0)
			}
		case p = <-u.ctlterm[1]:
			if p.Val == 1 {
				u.trigger(1)
			}
		case p = <-u.ctlterm[2]:
			if p.Val == 1 {
				u.trigger(2)
			}
		case p = <-u.ctlterm[3]:
			if p.Val == 1 {
				u.trigger(3)
			}
		case p = <-u.ctlterm[4]:
			if p.Val == 1 {
				u.trigger(4)
			}
		case p = <-u.ctlterm[6]:
			if p.Val == 1 {
				u.trigger(5)
			}
		case p = <-u.ctlterm[8]:
			if p.Val == 1 {
				u.trigger(6)
			}
		case p = <-u.ctlterm[10]:
			if p.Val == 1 {
				u.trigger(7)
			}
		case p = <-u.ctlterm[12]:
			if p.Val == 1 {
				u.trigger(8)
			}
		case p = <-u.ctlterm[14]:
			if p.Val == 1 {
				u.trigger(9)
			}
		case p = <-u.ctlterm[16]:
			if p.Val == 1 {
				u.trigger(10)
			}
		case p = <-u.ctlterm[18]:
			if p.Val == 1 {
				u.trigger(11)
			}
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}
