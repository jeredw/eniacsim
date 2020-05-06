package units

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

// Simulates ENIAC multiplier unit.
type Multiplier struct {
	Io MultiplierConn

	multin, multout                                       [24]Wire
	R, D                                                  [5]Wire
	A, S, AS, AC, SC, ASC, RS, DS, F                      Wire
	lhppI, lhppII, rhppI, rhppII                          Wire
	stage                                                 int
	multff                                                [24]bool
	iersw, iercl, icandsw, icandcl, sigsw, placsw, prodsw [24]int
	reset1ff, reset3ff                                    bool
	multl, multr                                          bool
	buffer61, f44                                         bool
	ier, icand                                            string
	sigfig                                                int

	rewiring           chan int
	waitingForRewiring chan int

	mu sync.Mutex
}

// Connections to other units.
type MultiplierConn struct {
	A8Clear func()
	A9Clear func()
	A8Value func() string
	A9Value func() string
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
		rewiring:           make(chan int),
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

type multJson struct {
	Reset1  bool     `json:"reset1"`
	Reset3  bool     `json:"reset3"`
	Stage   int      `json:"stage"`
	Program [24]bool `json:"program"`
}

func (u *Multiplier) State() json.RawMessage {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := multJson{
		Reset1:  u.reset1ff,
		Reset3:  u.reset3ff,
		Stage:   u.stage,
		Program: u.multff,
	}
	result, _ := json.Marshal(s)
	return result
}

func (u *Multiplier) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.rewiring <- 1
	<-u.waitingForRewiring
	for i := 0; i < 24; i++ {
		u.multin[i] = Wire{}
		u.multout[i] = Wire{}
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
		u.R[i] = Wire{}
		u.D[i] = Wire{}
	}
	u.A = Wire{}
	u.S = Wire{}
	u.AS = Wire{}
	u.AC = Wire{}
	u.SC = Wire{}
	u.ASC = Wire{}
	u.RS = Wire{}
	u.DS = Wire{}
	u.F = Wire{}
	u.lhppI = Wire{}
	u.lhppII = Wire{}
	u.rhppI = Wire{}
	u.rhppII = Wire{}
	u.stage = 0
	u.reset1ff = false
	u.reset3ff = false
	u.multl = false
	u.multr = false
	u.buffer61 = false
	u.f44 = false
	u.rewiring <- 1
}

func (u *Multiplier) lookupJack(jack string) (*Wire, error) {
	if len(jack) == 0 {
		return nil, fmt.Errorf("invalid jack")
	}
	switch jack {
	case "Rα", "Ra", "rα", "ra":
		return &u.R[0], nil
	case "Rβ", "Rb", "rβ", "rb":
		return &u.R[1], nil
	case "Rγ", "Rg", "rγ", "rg":
		return &u.R[2], nil
	case "Rδ", "Rd", "rδ", "rd":
		return &u.R[3], nil
	case "Rε", "Re", "rε", "re":
		return &u.R[4], nil
	case "Dα", "Da", "dα", "da":
		return &u.D[0], nil
	case "Dβ", "Db", "dβ", "db":
		return &u.D[1], nil
	case "Dγ", "Dg", "dγ", "dg":
		return &u.D[2], nil
	case "Dδ", "Dd", "dδ", "dd":
		return &u.D[3], nil
	case "Dε", "De", "dε", "de":
		return &u.D[4], nil
	case "A", "a":
		return &u.A, nil
	case "S", "s":
		return &u.S, nil
	case "AS", "as":
		return &u.AS, nil
	case "AC", "ac":
		return &u.AC, nil
	case "SC", "sc":
		return &u.SC, nil
	case "ASC", "asc":
		return &u.ASC, nil
	case "RS", "rs":
		return &u.RS, nil
	case "DS", "ds":
		return &u.DS, nil
	case "F", "f":
		return &u.F, nil
	case "LHPPI", "lhppi", "lhppI":
		return &u.lhppI, nil
	case "LHPPII", "lhppii", "lhppII":
		return &u.lhppII, nil
	case "RHPPI", "rhppi", "rhppI":
		return &u.rhppI, nil
	case "RHPPII", "rhppii", "rhppII":
		return &u.rhppII, nil
	}
	prog, err := strconv.Atoi(jack[:len(jack)-1])
	if err != nil {
		return nil, fmt.Errorf("invalid jack %s", jack)
	}
	if !(prog >= 1 && prog <= 24) {
		return nil, fmt.Errorf("invalid jack %s", jack)
	}
	switch jack[len(jack)-1] {
	case 'i':
		return &u.multin[prog-1], nil
	case 'o':
		return &u.multout[prog-1], nil
	}
	return nil, fmt.Errorf("invalid jack %s", jack)
}

func (u *Multiplier) Plug(jack string, wire Wire) error {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	p, err := u.lookupJack(jack)
	if err != nil {
		return err
	}
	Plug(p, wire)
	return nil
}

func (u *Multiplier) GetPlug(jack string) (Wire, error) {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	p, err := u.lookupJack(jack)
	if err != nil {
		return Wire{}, err
	}
	return *p, nil
}

func (u *Multiplier) lookupSwitch(name string) (Switch, error) {
	switch {
	case len(name) > 6 && name[:6] == "ieracc":
		prog, _ := strconv.Atoi(name[6:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.iersw[prog-1], recvSettings()}, nil
	case len(name) > 5 && name[:5] == "iercl":
		prog, _ := strconv.Atoi(name[5:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.iercl[prog-1], mclSettings()}, nil
	case len(name) > 8 && name[:8] == "icandacc":
		prog, _ := strconv.Atoi(name[8:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.icandsw[prog-1], recvSettings()}, nil
	case len(name) > 7 && name[:7] == "icandcl":
		prog, _ := strconv.Atoi(name[7:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.icandcl[prog-1], mclSettings()}, nil
	case len(name) > 2 && name[:2] == "sf":
		prog, _ := strconv.Atoi(name[2:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.sigsw[prog-1], msfSettings()}, nil
	case len(name) > 5 && name[:5] == "place":
		prog, _ := strconv.Atoi(name[5:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.placsw[prog-1], mplSettings()}, nil
	case len(name) > 4 && name[:4] == "prod":
		prog, _ := strconv.Atoi(name[4:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{name, &u.prodsw[prog-1], prodSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

func (u *Multiplier) SetSwitch(name, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return err
	}
	return sw.Set(value)
}

func (u *Multiplier) GetSwitch(name string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return "", err
	}
	return sw.Get(), nil
}

func (u *Multiplier) shiftprod(lhpp, rhpp int, resp1, resp2, resp3, resp4 chan int) {
	if u.lhppI.Ch != nil && lhpp != 0 {
		u.lhppI.Ch <- Pulse{lhpp >> uint(u.stage-2), resp1}
	}
	if u.lhppII.Ch != nil && lhpp != 0 {
		u.lhppII.Ch <- Pulse{(lhpp << uint(12-u.stage)) & 0x3ff, resp2}
	}
	if u.rhppI.Ch != nil && rhpp != 0 {
		u.rhppI.Ch <- Pulse{rhpp >> uint(u.stage-1), resp3}
	}
	if u.rhppII.Ch != nil && rhpp != 0 {
		u.rhppII.Ch <- Pulse{(rhpp << uint(11-u.stage)) & 0x3ff, resp4}
	}
	if u.lhppI.Ch != nil && lhpp != 0 {
		<-resp1
	}
	if u.lhppII.Ch != nil && lhpp != 0 {
		<-resp2
	}
	if u.rhppI.Ch != nil && rhpp != 0 {
		<-resp3
	}
	if u.rhppII.Ch != nil && rhpp != 0 {
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
			u.Io.A8Clear()
		}
		if u.icandcl[which] == 1 {
			u.Io.A9Clear()
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
		if u.sigfig == 0 && u.lhppII.Ch != nil {
			Handshake(1<<10, u.lhppII, resp1)
		} else if u.sigfig > 0 && u.sigfig < 9 && u.lhppI.Ch != nil {
			Handshake(1<<uint(u.sigfig-1), u.lhppI, resp1)
		}
	case c.Val&Fourp != 0 && u.stage == 1:
		if u.sigfig == 0 && u.lhppII.Ch != nil {
			Handshake(1<<10, u.lhppII, resp1)
		} else if u.sigfig > 0 && u.sigfig < 9 && u.lhppI.Ch != nil {
			Handshake(1<<uint(u.sigfig-1), u.lhppI, resp1)
		}
	case c.Val&Onep != 0 && u.stage >= 2 && u.stage < 12:
		u.ier = u.Io.A8Value()
		u.icand = u.Io.A9Value()
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
	if ier < 5 && u.R[ier].Ch != nil {
		u.R[ier].Ch <- Pulse{1, resp1}
	}
	if icand < 5 && u.D[icand].Ch != nil {
		u.D[icand].Ch <- Pulse{1, resp2}
	}
	if ier < 5 && u.R[ier].Ch != nil {
		<-resp1
	}
	if icand < 5 && u.D[icand].Ch != nil {
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
		case p = <-u.multin[0].Ch:
			u.multargs(0)
		case p = <-u.multin[1].Ch:
			u.multargs(1)
		case p = <-u.multin[2].Ch:
			u.multargs(2)
		case p = <-u.multin[3].Ch:
			u.multargs(3)
		case p = <-u.multin[4].Ch:
			u.multargs(4)
		case p = <-u.multin[5].Ch:
			u.multargs(5)
		case p = <-u.multin[6].Ch:
			u.multargs(6)
		case p = <-u.multin[7].Ch:
			u.multargs(7)
		case p = <-u.multin[8].Ch:
			u.multargs(8)
		case p = <-u.multin[9].Ch:
			u.multargs(9)
		case p = <-u.multin[10].Ch:
			u.multargs(10)
		case p = <-u.multin[11].Ch:
			u.multargs(11)
		case p = <-u.multin[12].Ch:
			u.multargs(12)
		case p = <-u.multin[13].Ch:
			u.multargs(13)
		case p = <-u.multin[14].Ch:
			u.multargs(14)
		case p = <-u.multin[15].Ch:
			u.multargs(15)
		case p = <-u.multin[16].Ch:
			u.multargs(16)
		case p = <-u.multin[17].Ch:
			u.multargs(17)
		case p = <-u.multin[18].Ch:
			u.multargs(18)
		case p = <-u.multin[19].Ch:
			u.multargs(19)
		case p = <-u.multin[20].Ch:
			u.multargs(20)
		case p = <-u.multin[21].Ch:
			u.multargs(21)
		case p = <-u.multin[22].Ch:
			u.multargs(22)
		case p = <-u.multin[23].Ch:
			u.multargs(23)
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}
