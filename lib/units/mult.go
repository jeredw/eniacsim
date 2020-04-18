package units

import (
	"fmt"
	"strconv"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

// Simulates ENIAC multiplier unit.
type Multiplier struct {
	Io	MultiplierConn

	multin, multout [24]chan Pulse
	R, D [5]chan Pulse
	A, S, AS, AC, SC, ASC, RS, DS, F chan Pulse
	lhppI, lhppII, rhppI, rhppII chan Pulse
	stage int
	multff [24]bool
	iersw, iercl, icandsw, icandcl, sigsw, placsw, prodsw [24]int
	reset1ff, reset3ff bool
	multl, multr bool
	buffer61, f44 bool
	ier, icand string
	sigfig int

	rewiring chan int
	waitingForRewiring chan int

	mu sync.Mutex
}

// Connections to other units.
type MultiplierConn struct {
	Acc8Clear	func()
	Acc9Clear	func()
	Acc8Value	func() string
	Acc9Value	func() string
}

type pulseset struct {
	one, two, twop, four int
}

var table10 [10][10]pulseset = [10][10]pulseset{{},
	{},
	{{}, {}, {}, {}, {},
		{1, 0, 0, 0}, {1, 0, 0, 0}, {1, 0, 0, 0}, {1, 0, 0, 0}, {1, 0, 0, 0}},
	{{}, {}, {}, {}, {1, 0, 0, 0},
		{1, 0, 0, 0}, {1, 0, 0, 0}, {0, 1, 0, 0}, {0, 1, 0, 0}, {0, 1, 0, 0}},
	{{}, {}, {}, {1, 0, 0, 0}, {1, 0, 0, 0},
		{0, 1, 0, 0}, {0, 1, 0, 0}, {0, 1, 0, 0}, {1, 1, 0, 0}, {1, 1, 0, 0}},
	{{}, {}, {1, 0, 0, 0}, {1, 0, 0, 0}, {0, 1, 0, 0},
		{0, 1, 0, 0}, {1, 1, 0, 0}, {1, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 1, 0}},
	{{}, {}, {1, 0, 0, 0}, {1, 0, 0, 0}, {0, 1, 0, 0},
		{1, 1, 0, 0}, {1, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 1, 0}, {1, 1, 1, 0}},
	{{}, {}, {1, 0, 0, 0}, {0, 1, 0, 0}, {0, 1, 0, 0},
		{1, 1, 0, 0}, {0, 0, 0, 1}, {0, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 0, 1}},
	{{}, {}, {1, 0, 0, 0}, {0, 1, 0, 0}, {1, 1, 0, 0},
		{0, 0, 0, 1}, {0, 0, 0, 1}, {1, 0, 0, 1}, {0, 1, 0, 1}, {1, 1, 0, 1}},
	{{}, {}, {1, 0, 0, 0}, {0, 1, 0, 0}, {1, 0, 1, 0},
		{0, 0, 0, 1}, {1, 1, 1, 0}, {0, 1, 0, 1}, {1, 0, 1, 1}, {0, 1, 1, 1}},
}

var table1 [10][10]pulseset = [10][10]pulseset{{},
	{{}, {1, 0, 0, 0}, {0, 1, 0, 0}, {1, 0, 1, 0}, {0, 1, 1, 0},
		{1, 0, 0, 1}, {0, 1, 0, 1}, {1, 0, 1, 1}, {0, 1, 1, 1}, {1, 1, 1, 1}},
	{{}, {0, 1, 0, 0}, {0, 0, 0, 1}, {0, 0, 1, 1}, {0, 1, 1, 1},
		{}, {0, 1, 0, 0}, {0, 1, 1, 0}, {0, 0, 1, 1}, {0, 1, 1, 1}},
	{{}, {1, 1, 0, 0}, {0, 0, 1, 1}, {1, 1, 1, 1}, {0, 1, 0, 0},
		{1, 1, 1, 0}, {0, 1, 1, 1}, {1, 0, 0, 0}, {0, 0, 0, 1}, {1, 0, 1, 1}},
	{{}, {0, 1, 1, 0}, {0, 1, 1, 1}, {0, 1, 0, 0}, {0, 0, 1, 1},
		{}, {0, 1, 1, 0}, {0, 1, 1, 1}, {0, 1, 0, 0}, {0, 0, 1, 1}},
	{{}, {1, 0, 0, 1}, {}, {1, 0, 0, 1}, {},
		{1, 0, 0, 1}, {}, {1, 0, 0, 1}, {}, {1, 0, 0, 1}},
	{{}, {0, 0, 1, 1}, {0, 1, 0, 0}, {0, 1, 1, 1}, {0, 1, 1, 0},
		{}, {0, 0, 1, 1}, {0, 1, 0, 0}, {0, 1, 1, 1}, {0, 0, 0, 1}},
	{{}, {1, 0, 1, 1}, {0, 1, 1, 0}, {1, 0, 0, 0}, {0, 1, 1, 1},
		{1, 1, 1, 0}, {0, 1, 0, 0}, {1, 1, 1, 1}, {0, 0, 1, 1}, {1, 1, 0, 0}},
	{{}, {0, 1, 1, 1}, {0, 0, 1, 1}, {0, 1, 1, 0}, {0, 1, 0, 0},
		{}, {0, 1, 1, 1}, {0, 0, 1, 1}, {0, 0, 0, 1}, {0, 1, 0, 0}},
	{{}, {1, 1, 1, 1}, {0, 1, 1, 1}, {1, 0, 1, 1}, {0, 1, 0, 1},
		{1, 0, 0, 1}, {0, 1, 1, 0}, {1, 0, 1, 0}, {0, 1, 0, 0}, {1, 0, 0, 0}},
}

func NewMultiplier() *Multiplier {
	return &Multiplier{
		rewiring: make(chan int),
		waitingForRewiring: make(chan int),
	}
}

func (u *Multiplier) Multl() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.multl
}

func (u *Multiplier) Multr() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.multr
}

func (u *Multiplier) Stat() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := fmt.Sprintf("%d ", u.stage)
	for i, _ := range u.multff {
		if u.multff[i] {
			s += "1"
		} else {
			s += "0"
		}
	}
	if u.reset1ff {
		s += " 1"
	} else {
		s += " 0"
	}
	if u.reset3ff {
		s += " 1"
	} else {
		s += " 0"
	}
	return s
}

func (u *Multiplier) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.rewiring <- 1
	<-u.waitingForRewiring
	for i := 0; i < 24; i++ {
		u.multin[i] = nil
		u.multout[i] = nil
		u.multff[i] = false
		u.iersw[i] = 0
		u.iercl[i] = 0
		u.icandsw[i] = 0
		u.icandcl[i] = 0
		u.sigsw[i] = 0
		u.placsw[i] = 0
		u.prodsw[i] = 0
	}
	for i := 0; i < 5; i++ {
		u.R[i] = nil
		u.D[i] = nil
	}
	u.A = nil
	u.S = nil
	u.AS = nil
	u.AC = nil
	u.SC = nil
	u.ASC = nil
	u.RS = nil
	u.DS = nil
	u.F = nil
	u.lhppI = nil
	u.lhppII = nil
	u.rhppI = nil
	u.rhppII = nil
	u.stage = 0
	u.reset1ff = false
	u.reset3ff = false
	u.multl = false
	u.multr = false
	u.buffer61 = false
	u.f44 = false
	u.rewiring <- 1
}

func (u *Multiplier) Plug(jack string, ch chan Pulse) error {
	if len(jack) == 0 {
		return fmt.Errorf("invalid jack")
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	name := "m." + jack
	switch jack {
	case "Rα", "Ra", "rα", "ra":
		SafePlug(name, &u.R[0], ch)
	case "Rβ", "Rb", "rβ", "rb":
		SafePlug(name, &u.R[1], ch)
	case "Rγ", "Rg", "rγ", "rg":
		SafePlug(name, &u.R[2], ch)
	case "Rδ", "Rd", "rδ", "rd":
		SafePlug(name, &u.R[3], ch)
	case "Rε", "Re", "rε", "re":
		SafePlug(name, &u.R[4], ch)
	case "Dα", "Da", "dα", "da":
		SafePlug(name, &u.D[0], ch)
	case "Dβ", "Db", "dβ", "db":
		SafePlug(name, &u.D[1], ch)
	case "Dγ", "Dg", "dγ", "dg":
		SafePlug(name, &u.D[2], ch)
	case "Dδ", "Dd", "dδ", "dd":
		SafePlug(name, &u.D[3], ch)
	case "Dε", "De", "dε", "de":
		SafePlug(name, &u.D[4], ch)
	case "A", "a":
		SafePlug(name, &u.A, ch)
	case "S", "s":
		SafePlug(name, &u.S, ch)
	case "AS", "as":
		SafePlug(name, &u.AS, ch)
	case "AC", "ac":
		SafePlug(name, &u.AC, ch)
	case "SC", "sc":
		SafePlug(name, &u.SC, ch)
	case "ASC", "asc":
		SafePlug(name, &u.ASC, ch)
	case "RS", "rs":
		SafePlug(name, &u.RS, ch)
	case "DS", "ds":
		SafePlug(name, &u.DS, ch)
	case "F", "f":
		SafePlug(name, &u.F, ch)
	case "LHPPI", "lhppi", "lhppI":
		SafePlug(name, &u.lhppI, ch)
	case "LHPPII", "lhppii", "lhppII":
		SafePlug(name, &u.lhppII, ch)
	case "RHPPI", "rhppi", "rhppI":
		SafePlug(name, &u.rhppI, ch)
	case "RHPPII", "rhppii", "rhppII":
		SafePlug(name, &u.rhppII, ch)
	default:
		prog, err := strconv.Atoi(jack[:len(jack)-1])
		if err != nil {
			return fmt.Errorf("invalid jack %s", jack)
		}
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid jack %s", jack)
		}
		switch jack[len(jack)-1] {
		case 'i':
			SafePlug(name, &u.multin[prog-1], ch)
		case 'o':
			SafePlug(name, &u.multout[prog-1], ch)
		default:
			return fmt.Errorf("invalid jack %s", jack)
		}
	}
	return nil
}

func (u *Multiplier) Switch(name, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	switch {
	case len(name) > 6 && name[:6] == "ieracc":
		prog, _ := strconv.Atoi(name[6:])
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid switch %s", name)
		}
		u.iersw[prog-1] = recv2val(value)
	case len(name) > 5 && name[:5] == "iercl":
		prog, _ := strconv.Atoi(name[5:])
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid switch %s", name)
		}
		switch value {
		case "C":
			u.iercl[prog-1] = 1
		case "0":
			u.iercl[prog-1] = 0
		default:
			return fmt.Errorf("invalid switch %s setting %s", name, value)
		}
	case len(name) > 8 && name[:8] == "icandacc":
		prog, _ := strconv.Atoi(name[8:])
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid switch %s", name)
		}
		u.icandsw[prog-1] = recv2val(value)
	case len(name) > 7 && name[:7] == "icandcl":
		prog, _ := strconv.Atoi(name[7:])
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid switch %s", name)
		}
		switch value {
		case "C":
			u.icandcl[prog-1] = 1
		case "0":
			u.icandcl[prog-1] = 0
		default:
			return fmt.Errorf("invalid switch %s setting %s", name, value)
		}
	case len(name) > 2 && name[:2] == "sf":
		prog, _ := strconv.Atoi(name[2:])
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid switch %s", name)
		}
		val, _ := strconv.Atoi(value)
		if !(val == 0 || val >= 2 && val <= 10) {
			return fmt.Errorf("invalid switch %s setting %s", name, value)
		}
		if val == 0 {
			val = 1
		}
		u.sigsw[prog-1] = 10 - val
	case len(name) > 5 && name[:5] == "place":
		prog, _ := strconv.Atoi(name[5:])
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid switch %s", name)
		}
		val, _ := strconv.Atoi(value)
		if !(val >= 2 && val <= 10) {
			return fmt.Errorf("invalid switch %s setting %s", name, value)
		}
		u.placsw[prog-1] = val - 2
	case len(name) > 4 && name[:4] == "prod":
		prog, _ := strconv.Atoi(name[4:])
		if !(prog >= 1 && prog <= 24) {
			return fmt.Errorf("invalid switch %s", name)
		}
		products := [7]string{"A", "S", "AS", "0", "AC", "SC", "ASC"}
		for i, p := range products {
			if p == value {
				u.prodsw[prog-1] = i
				return nil
			}
		}
		return fmt.Errorf("invalid switch %s setting %s", name, value)
	default:
		return fmt.Errorf("invalid switch %s", name)
	}
	return nil
}

func recv2val(recv string) int {
	switch recv {
	case "α", "a", "alpha":
		return 0
	case "β", "b", "beta":
		return 1
	case "γ", "g", "gamma":
		return 2
	case "δ", "d", "delta":
		return 3
	case "ε", "e", "epsilon":
		return 4
	case "0":
		return 5
	}
	return 5
}

func (u *Multiplier) shiftprod(lhpp, rhpp int, resp1, resp2, resp3, resp4 chan int) {
	if u.lhppI != nil && lhpp != 0 {
		u.lhppI <- Pulse{lhpp >> uint(u.stage-2), resp1}
	}
	if u.lhppII != nil && lhpp != 0 {
		u.lhppII <- Pulse{(lhpp << uint(12-u.stage)) & 0x3ff, resp2}
	}
	if u.rhppI != nil && rhpp != 0 {
		u.rhppI <- Pulse{rhpp >> uint(u.stage-1), resp3}
	}
	if u.rhppII != nil && rhpp != 0 {
		u.rhppII <- Pulse{(rhpp << uint(11-u.stage)) & 0x3ff, resp4}
	}
	if u.lhppI != nil && lhpp != 0 {
		<-resp1
	}
	if u.lhppII != nil && lhpp != 0 {
		<-resp2
	}
	if u.rhppI != nil && rhpp != 0 {
		<-resp3
	}
	if u.rhppII != nil && rhpp != 0 {
		<-resp4
	}
}

func (u *Multiplier) clock(c Pulse, resp1, resp2, resp3, resp4 chan int) {
//	u.mu.Lock()
//	defer u.mu.Unlock()
	switch {
	case c.Val&Cpp != 0:
		if u.f44 {
			u.stage = 1
			u.f44 = false
		} else if u.stage == 12 {
			u.reset1ff = true
			u.reset3ff = true
			Handshake(1, u.F, resp1)
			u.stage++
		} else if u.stage == 13 {
			which := -1
			for i, f := range u.multff {
				if f {
					which = i
					break
				}
			}
			if which != -1 {
				Handshake(1, u.multout[which], resp1)
				u.multff[which] = false
				switch u.prodsw[which] {
				case 0:
					Handshake(1, u.A, resp1)
				case 1:
					Handshake(1, u.S, resp1)
				case 2:
					Handshake(1, u.AS, resp1)
				case 4:
					Handshake(1, u.AC, resp1)
				case 5:
					Handshake(1, u.SC, resp1)
				case 6:
					Handshake(1, u.ASC, resp1)
				}
			}
			u.reset1ff = false
			u.reset3ff = false
			u.stage = 0
		} else if u.stage != 0 {
			minplace := 10
			for i := 0; i < 24; i++ {
				if u.multff[i] && u.placsw[i]+2 < minplace {
					minplace = u.placsw[i] + 2
				}
			}
			if u.stage == minplace+1 {
				if u.ier[0] == 'M' {
					Handshake(1, u.DS, resp1)
				}
				if u.icand[0] == 'M' {
					Handshake(1, u.RS, resp1)
				}
				u.multl = false
				u.multr = false
				u.stage = 12
			} else {
				u.stage++
			}
		}
	case c.Val&Ccg != 0 && u.stage == 13:
		which := -1
		for i, f := range u.multff {
			if f {
				which = i
				break
			}
		}
		if u.iercl[which] == 1 {
			u.Io.Acc8Clear()
		}
		if u.icandcl[which] == 1 {
			u.Io.Acc9Clear()
		}
	case c.Val&Onep != 0 && u.stage == 1:
		u.multl = true
		u.multr = true
		u.sigfig = -1
		for i := 0; i < 24; i++ {
			if u.multff[i] {
				u.sigfig = u.sigsw[i]
			}
		}
		if u.sigfig == 0 && u.lhppII != nil {
			Handshake(1<<10, u.lhppII, resp1)
		} else if u.sigfig > 0 && u.sigfig < 9 && u.lhppI != nil {
			Handshake(1<<uint(u.sigfig-1), u.lhppI, resp1)
		}
	case c.Val&Fourp != 0 && u.stage == 1:
		if u.sigfig == 0 && u.lhppII != nil {
			Handshake(1<<10, u.lhppII, resp1)
		} else if u.sigfig > 0 && u.sigfig < 9 && u.lhppI != nil {
			Handshake(1<<uint(u.sigfig-1), u.lhppI, resp1)
		}
	case c.Val&Onep != 0 && u.stage >= 2 && u.stage < 12:
		u.ier = u.Io.Acc8Value()
		u.icand = u.Io.Acc9Value()
		lhpp := 0
		rhpp := 0
		for i := 0; i < 10; i++ {
			ps10 := table10[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			ps1 := table1[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			if ps10.one == 1 {
				lhpp |= 1 << uint(9-i)
			}
			if ps1.one == 1 {
				rhpp |= 1 << uint(9-i)
			}
		}
		u.shiftprod(lhpp, rhpp, resp1, resp2, resp3, resp4)
	case c.Val&Twop != 0 && u.stage >= 2 && u.stage < 12:
		lhpp := 0
		rhpp := 0
		for i := 0; i < 10; i++ {
			ps10 := table10[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			ps1 := table1[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			if ps10.two == 1 {
				lhpp |= 1 << uint(9-i)
			}
			if ps1.two == 1 {
				rhpp |= 1 << uint(9-i)
			}
		}
		u.shiftprod(lhpp, rhpp, resp1, resp2, resp3, resp4)
	case c.Val&Twopp != 0 && u.stage >= 2 && u.stage < 12:
		lhpp := 0
		rhpp := 0
		for i := 0; i < 10; i++ {
			ps10 := table10[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			ps1 := table1[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			if ps10.twop == 1 {
				lhpp |= 1 << uint(9-i)
			}
			if ps1.twop == 1 {
				rhpp |= 1 << uint(9-i)
			}
		}
		u.shiftprod(lhpp, rhpp, resp1, resp2, resp3, resp4)
	case c.Val&Fourp != 0 && u.stage >= 2 && u.stage < 12:
		lhpp := 0
		rhpp := 0
		for i := 0; i < 10; i++ {
			ps10 := table10[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			ps1 := table1[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
			if ps10.four == 1 {
				lhpp |= 1 << uint(9-i)
			}
			if ps1.four == 1 {
				rhpp |= 1 << uint(9-i)
			}
		}
		u.shiftprod(lhpp, rhpp, resp1, resp2, resp3, resp4)
	case c.Val&Onepp != 0 && u.stage >= 2 && u.stage < 12:
		minplace := 10
		for i := 0; i < 24; i++ {
			if u.multff[i] && u.placsw[i]+2 < minplace {
				minplace = u.placsw[i] + 2
			}
		}
		if u.stage == minplace+1 && u.ier[0] == 'M' && u.icand[0] == 'M' {
			Handshake(1<<10, u.rhppI, resp1)
		}
	case c.Val&Rp != 0 && u.buffer61:
		u.buffer61 = false
		u.f44 = true
	}
}

func (u *Multiplier) MakeClockFunc() ClockFunc {
	resp1 := make(chan int)
	resp2 := make(chan int)
	resp3 := make(chan int)
	resp4 := make(chan int)
	return func(p Pulse) {
		u.clock(p, resp1, resp2, resp3, resp4)
	}
}

func (u *Multiplier) multargs(prog int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	resp1 := make(chan int)
	resp2 := make(chan int)
	ier := u.iersw[prog]
	icand := u.icandsw[prog]
	if ier < 5 && u.R[ier] != nil {
		u.R[ier] <- Pulse{1, resp1}
	}
	if icand < 5 && u.D[icand] != nil {
		u.D[icand] <- Pulse{1, resp2}
	}
	if ier < 5 && u.R[ier] != nil {
		<-resp1
	}
	if icand < 5 && u.D[icand] != nil {
		<-resp2
	}
	u.multff[prog] = true
	u.buffer61 = true
}

func (u *Multiplier) Run() {
	var p Pulse

	for {
		p.Resp = nil
		select {
		case <-u.rewiring:
			u.waitingForRewiring <- 1
			<-u.rewiring
		case p = <-u.multin[0]:
			u.multargs(0)
		case p = <-u.multin[1]:
			u.multargs(1)
		case p = <-u.multin[2]:
			u.multargs(2)
		case p = <-u.multin[3]:
			u.multargs(3)
		case p = <-u.multin[4]:
			u.multargs(4)
		case p = <-u.multin[5]:
			u.multargs(5)
		case p = <-u.multin[6]:
			u.multargs(6)
		case p = <-u.multin[7]:
			u.multargs(7)
		case p = <-u.multin[8]:
			u.multargs(8)
		case p = <-u.multin[9]:
			u.multargs(9)
		case p = <-u.multin[10]:
			u.multargs(10)
		case p = <-u.multin[11]:
			u.multargs(11)
		case p = <-u.multin[12]:
			u.multargs(12)
		case p = <-u.multin[13]:
			u.multargs(13)
		case p = <-u.multin[14]:
			u.multargs(14)
		case p = <-u.multin[15]:
			u.multargs(15)
		case p = <-u.multin[16]:
			u.multargs(16)
		case p = <-u.multin[17]:
			u.multargs(17)
		case p = <-u.multin[18]:
			u.multargs(18)
		case p = <-u.multin[19]:
			u.multargs(19)
		case p = <-u.multin[20]:
			u.multargs(20)
		case p = <-u.multin[21]:
			u.multargs(21)
		case p = <-u.multin[22]:
			u.multargs(22)
		case p = <-u.multin[23]:
			u.multargs(23)
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}
