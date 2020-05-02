package units

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

const (
	svα   = 1 << 0
	svβ   = 1 << 1
	svγ   = 1 << 2
	svA   = 1 << 5
	svCLR = 1 << 8

	su2qα   = 1 << 0
	su2qA   = 1 << 3
	su2qS   = 1 << 4
	su2qCLR = 1 << 5
	su2sα   = 1 << 1
	su2sA   = 1 << 2
	su2sCLR = 1 << 6

	su3α   = 1 << 0
	su3β   = 1 << 1
	su3γ   = 1 << 2
	su3A   = 1 << 3
	su3S   = 1 << 4
	su3CLR = 1 << 5
)

// Divsr simulates the ENIAC divider/square rooter unit.
type Divsr struct {
	Io DivsrConn

	progin, progout, ilock                                                     [8]chan Pulse
	answer                                                                     chan Pulse
	numarg, denarg, roundoff, places, ilocksw, anssw                           [8]int
	numcl, dencl                                                               [8]bool
	preff, progff                                                              [8]bool
	placering, progring                                                        int
	divff, clrff, ilockff, coinff, dpγ, nγ, psrcff, pringff, denomff, numrplus bool
	numrmin, qα, sac, m2, m1, nac, da, nα, dα, dγ, npγ, p2, p1, sα, ds, nβ, dβ bool
	ans1, ans2, ans3, ans4                                                     bool
	curprog, divadap, sradap                                                   int
	sv, su2, su3                                                               int

	rewiring           chan int
	waitingForRewiring chan int

	mu sync.Mutex
}

// Connections to dedicated accumulators.
type DivsrConn struct {
	A2Sign  func() string
	A2Clear func()
	A4Sign  func() string
	A4Clear func()
}

func NewDivsr() *Divsr {
	return &Divsr{
		rewiring:           make(chan int),
		waitingForRewiring: make(chan int),
	}
}

func (u *Divsr) Sv() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.sv
}

func (u *Divsr) Su2() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.su2
}

func (u *Divsr) Su3() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.su3
}

type divsrJson struct {
	PlaceRing int     `json:"progRing"`
	ProgRing  int     `json:"progRing"`
	Program   [8]bool `json:"program"`
	Ffs       string  `json:"ffs"`
}

func (u *Divsr) State() json.RawMessage {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := divsrJson{
		PlaceRing: u.placering,
		ProgRing:  u.progring,
		Program:   u.progff,
		Ffs:       u.ffs(),
	}
	result, _ := json.Marshal(s)
	return result
}

func (u *Divsr) Stat() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := fmt.Sprintf("%d %d ", u.placering, u.progring)
	for i := range u.progff {
		if u.progff[i] {
			s += "1"
		} else {
			s += "0"
		}
	}
	s += " " + u.ffs()
	return s
}

func (u *Divsr) ffs() string {
	return ToBin(u.divff) + ToBin(u.clrff) + ToBin(u.coinff) + ToBin(u.dpγ) +
		ToBin(u.nγ) + ToBin(u.psrcff) + ToBin(u.pringff) + ToBin(u.denomff) +
		ToBin(u.numrplus) + ToBin(u.numrmin) + ToBin(u.qα) + ToBin(u.sac) +
		ToBin(u.m2) + ToBin(u.m1) + ToBin(u.nac) + ToBin(u.da) + ToBin(u.nα) +
		ToBin(u.dα) + ToBin(u.dγ) + ToBin(u.npγ) + ToBin(u.p2) + ToBin(u.p1) +
		ToBin(u.sα) + ToBin(u.ds) + ToBin(u.nβ) + ToBin(u.dβ) + ToBin(u.ans1) +
		ToBin(u.ans2) + ToBin(u.ans3) + ToBin(u.ans4)
}

func (u *Divsr) Stat2() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := fmt.Sprintf("%d %d ", u.placering, u.progring)
	for i := range u.progff {
		if u.progff[i] {
			s += "1"
		} else {
			s += "0"
		}
	}
	if u.divff {
		s += " divff"
	}
	if u.clrff {
		s += " clrff"
	}
	if u.coinff {
		s += " coinff"
	}
	if u.dpγ {
		s += " dpg"
	}
	if u.nγ {
		s += " ng"
	}
	if u.psrcff {
		s += " psrcff"
	}
	if u.denomff {
		s += " denomff"
	}
	if u.numrplus {
		s += " n+"
	}
	if u.numrmin {
		s += " n-"
	}
	if u.qα {
		s += " qa"
	}
	if u.sac {
		s += " SAC"
	}
	if u.m2 {
		s += " -2"
	}
	if u.m1 {
		s += " -1"
	}
	if u.nac {
		s += " NAC"
	}
	if u.da {
		s += " dA"
	}
	if u.nα {
		s += " na"
	}
	if u.dα {
		s += " da"
	}
	if u.dγ {
		s += " dg"
	}
	if u.npγ {
		s += " npg"
	}
	if u.p2 {
		s += " +2"
	}
	if u.p1 {
		s += " +1"
	}
	if u.sα {
		s += " sa"
	}
	if u.ds {
		s += " dS"
	}
	if u.nβ {
		s += " nb"
	}
	if u.dβ {
		s += " db"
	}
	if u.ans1 {
		s += " A1"
	}
	if u.ans2 {
		s += " A2"
	}
	if u.ans3 {
		s += " A3"
	}
	if u.ans4 {
		s += " A4"
	}
	return s
}

func (u *Divsr) Reset() {
	u.mu.Lock()
	u.rewiring <- 1
	<-u.waitingForRewiring
	for i := 0; i < 8; i++ {
		u.progin[i] = nil
		u.progout[i] = nil
		u.ilock[i] = nil
		u.numarg[i] = 0
		u.numcl[i] = false
		u.denarg[i] = 0
		u.dencl[i] = false
		u.roundoff[i] = 0
		u.places[i] = 0
		u.ilocksw[i] = 0
		u.anssw[i] = 0
		u.preff[i] = false
		u.progff[i] = false
	}
	u.answer = nil
	u.divff = false
	u.ilockff = false
	u.ans1 = false
	u.ans2 = false
	u.ans3 = false
	u.ans4 = false
	u.divadap = 0
	u.sradap = 0
	u.mu.Unlock()
	u.Clear()
	u.rewiring <- 1
}

func (u *Divsr) Clear() {
	u.intclear()
	u.mu.Lock()
	defer u.mu.Unlock()
	u.sv = 0
	u.su2 = 0
	u.su3 = 0
}

func (u *Divsr) intclear() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.progring = 0
	u.placering = 0
	u.numrplus = true
	u.numrmin = false
	u.denomff = false
	u.psrcff = false
	u.pringff = false
	u.curprog = -1
	u.coinff = false
	u.clrff = false
	u.dpγ = false
	u.nγ = false
	u.qα = false
	u.sac = false
	u.m2 = false
	u.m1 = false
	u.nac = false
	u.da = false
	u.nα = false
	u.dα = false
	u.dγ = false
	u.npγ = false
	u.p2 = false
	u.p1 = false
	u.sα = false
	u.ds = false
	u.nβ = false
	u.dβ = false
}

func (u *Divsr) Plug(jack string, ch chan Pulse, output bool) error {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()

	name := "d." + jack
	if jack == "ans" || jack == "ANS" {
		SafePlug(name, &u.answer, ch, output)
	} else {
		var prog int
		var ilk rune
		fmt.Sscanf(jack, "%d%c", &prog, &ilk)
		if !(prog >= 1 && prog <= 8) {
			return fmt.Errorf("invalid jack %s", jack)
		}
		switch ilk {
		case 'i':
			SafePlug(name, &u.progin[prog-1], ch, output)
		case 'o':
			SafePlug(name, &u.progout[prog-1], ch, output)
		case 'l':
			SafePlug(name, &u.ilock[prog-1], ch, output)
		default:
			return fmt.Errorf("invalid jack %s", jack)
		}
	}
	return nil
}

func adSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"A", 0},
		{"B", 1},
		{"C", 2},
	}
}

func argSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"α", 0}, {"a", 0}, {"alpha", 0},
		{"β", 1}, {"b", 1}, {"beta", 1},
		{"0", 2},
	}
}

func placeSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"D4", 0}, {"d4", 0},
		{"D7", 1}, {"d7", 1},
		{"D8", 2}, {"d8", 2},
		{"D9", 3}, {"d9", 3},
		{"D10", 4}, {"d10", 4},
		{"S4", 5}, {"s4", 5}, {"R4", 5}, {"r4", 5},
		{"S7", 6}, {"s7", 6}, {"R7", 6}, {"r7", 6},
		{"S8", 7}, {"s8", 7}, {"R8", 7}, {"r8", 7},
		{"S9", 8}, {"s9", 8}, {"R9", 8}, {"r9", 8},
		{"S10", 9}, {"s10", 9}, {"R10", 9}, {"r10", 9},
	}
}

func roSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"RO", 1}, {"ro", 1},
		{"NRO", 0}, {"nro", 0},
	}
}

func ilSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"I", 1}, {"i", 1},
		{"NI", 0}, {"ni", 0},
	}
}

func anSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"1", 0},
		{"2", 1},
		{"3", 2},
		{"4", 3},
		{"OFF", 4}, {"off", 4},
	}
}

func (u *Divsr) lookupSwitch(name string) (Switch, error) {
	if name == "da" {
		return &IntSwitch{name, &u.divadap, adSettings()}, nil
	}
	if name == "ra" {
		return &IntSwitch{name, &u.sradap, adSettings()}, nil
	}
	if len(name) < 3 {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	sw, _ := strconv.Atoi(name[2:])
	if !(sw >= 1 && sw <= 8) {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	switch name[:2] {
	case "nr":
		return &IntSwitch{name, &u.numarg[sw-1], argSettings()}, nil
	case "nc":
		return &ClearSwitch{name, &u.numcl[sw-1]}, nil
	case "dr":
		return &IntSwitch{name, &u.denarg[sw-1], argSettings()}, nil
	case "dc":
		return &ClearSwitch{name, &u.dencl[sw-1]}, nil
	case "pl":
		return &IntSwitch{name, &u.places[sw-1], placeSettings()}, nil
	case "ro":
		return &IntSwitch{name, &u.roundoff[sw-1], roSettings()}, nil
	case "an":
		return &IntSwitch{name, &u.anssw[sw-1], anSettings()}, nil
	case "il":
		return &IntSwitch{name, &u.ilocksw[sw-1], ilSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

func (u *Divsr) SetSwitch(name, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return err
	}
	return sw.Set(value)
}

func (u *Divsr) GetSwitch(name string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return "", err
	}
	return sw.Get(), nil
}

func (u *Divsr) divargs(prog int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.preff[prog] = true
	if u.places[prog] < 5 {
		u.divff = true
	} else {
		u.divff = false
	}
	switch u.numarg[prog] {
	case 0:
		u.nα = true
		u.sv |= svα
	case 1:
		u.nβ = true
		u.sv |= svβ
	}
	switch u.denarg[prog] {
	case 0:
		u.dα = true
		u.su3 |= su3α
	case 1:
		u.dβ = true
		u.su3 |= su3β
	}
}

func (u *Divsr) doP() {
	u.nγ = true
	u.sv |= svγ
	if u.samesign() {
		u.ds = true
		u.su3 |= su3S
	} else {
		u.da = true
		u.su3 |= su3A
	}
}

func (u *Divsr) doS() {
	u.sα = true
	u.su2 |= su2sα
	u.nac = true
	u.sv |= svA | svCLR
	if !u.divff {
		if u.samesign() {
			u.m1 = true
		} else {
			u.p1 = true
		}
		u.dpγ = true
		u.su3 |= su3γ
	}
	p := u.places[u.curprog] % 5
	if p == 0 {
		p = 4
	} else {
		p += 6
	}
	if u.placering == p-2 { // Gate E6
		u.psrcff = true
	}
}

func (u *Divsr) samesign() bool {
	return u.denomff && u.numrmin || !u.denomff && u.numrplus
}

func (u *Divsr) overflow() bool {
	s := u.Io.A2Sign()
	return s[0] == 'P' && u.numrmin || s[0] == 'M' && u.numrplus
}

func (u *Divsr) interlock() {
	u.mu.Lock()
	u.ilockff = true
	u.mu.Unlock()
}

func (u *Divsr) doGP(resp chan int) {
	if u.coinff { // Gate E50
		if u.ilocksw[u.curprog] == 0 || u.ilockff {
			u.coinff = false
			u.clrff = true
			return
		}
	} else if u.clrff {
		u.progff[u.curprog] = false
		Handshake(1, u.progout[u.curprog], resp)
		if u.ilocksw[u.curprog] == 1 {
			u.ilockff = false
		}
		/*
		 * Implement the PX-4-114 adapters
		 */
		switch u.anssw[u.curprog] {
		case 0:
			u.ans1 = true
			u.su2 |= su2qA
			if u.divadap == 2 {
				u.su2 |= su2qCLR
			}
		case 1:
			u.ans2 = true
			switch u.divadap {
			case 0:
				u.su2 |= su2qA | su2qCLR
			case 1:
				u.su2 |= su2qS
			case 2:
				u.su2 |= su2qS | su2qCLR
			}
		case 2:
			u.ans3 = true
			u.su3 |= su3A
			if u.sradap == 2 {
				u.su3 |= su3CLR
			}
		case 3:
			u.ans4 = true
			switch u.sradap {
			case 0:
				u.su3 |= su3A | su3CLR
			case 1:
				u.su3 |= su3S
			case 2:
				u.su3 |= su3S | su3CLR
			}
		}
		if u.numcl[u.curprog] {
			u.Io.A2Clear()
		}
		if u.dencl[u.curprog] {
			u.Io.A4Clear()
		}
		u.intclear()
		return
	}
	if u.qα {
		u.p1 = false
		u.m1 = false
		if u.overflow() { // Gates D9, D11, D12
			u.doS()
		} else {
			u.doP()
		}
		u.qα = false
		u.su2 &^= su2qα
	} else if u.nγ { //  Gates L10, G11, H11
		u.nγ = false
		u.sv &^= svγ
		if u.divff {
			u.qα = true
			u.su2 |= su2qα
			if u.ds {
				u.ds = false
				u.su3 &^= su3S
				u.p1 = true
			} else if u.da {
				u.da = false
				u.su3 &^= su3A
				u.m1 = true
			}
		} else {
			u.dγ = true
			u.su3 |= su3γ
			if u.ds {
				u.ds = false
				u.su3 &^= su3S
				u.p2 = true
			} else if u.da {
				u.da = false
				u.su3 &^= su3A
				u.m2 = true
			}
		}
	} else if u.npγ { // Gate C9
		u.npγ = false
		u.sv &^= svγ
		u.sac = false
		u.su2 &^= su2sA | su2sCLR
		u.m1 = false
		u.p1 = false
		u.dpγ = false
		u.su3 &^= su3γ
		u.doP()
	} else if u.sα { // Gates K7, L1
		u.sα = false
		u.su2 &^= su2sα
		u.nac = false
		u.sv &^= svA | svCLR
		u.sac = true
		u.su2 |= su2sA | su2sCLR
		u.npγ = true
		u.sv |= svγ
		u.numrplus, u.numrmin = u.numrmin, u.numrplus
	} else if u.dγ {
		u.dγ = false
		u.su3 &^= su3γ
		u.p2 = false
		u.m2 = false
		if u.overflow() {
			u.doS()
		} else {
			u.doP()
		}
	}
	switch u.progring {
	case 0:
		u.nα = false
		u.nβ = false
		u.sv &^= svα | svβ
		u.dα = false
		u.dβ = false
		u.su3 &^= su3α | su3β
		if !u.pringff {
			u.progring++
		}
	case 1: // Gate D6
		s := u.Io.A2Sign()
		if s[0] == 'M' {
			u.numrplus, u.numrmin = u.numrmin, u.numrplus
		}
		s = u.Io.A4Sign()
		if s[0] == 'M' {
			u.denomff = true
		}
		if !u.pringff {
			u.progring++
		}
	case 2: // Gate A7, B7, B8
		if u.divff {
			u.doP()
			u.pringff = true
			u.progring = 0
		} else {
			u.p1 = true
			u.dγ = true
			u.su3 |= su3γ
			u.progring++
		}
	case 3:
		u.p1 = false
		u.dγ = false
		u.su3 &^= su3γ
		u.doP()
		u.pringff = true
		u.progring = 0
	}
}

func (u *Divsr) doIIIP() {
	if u.npγ { // Gate C9
		u.npγ = false
		u.sv &^= svγ
		u.sac = false
		u.su2 &^= su2sA | su2sCLR
		u.m1 = false
		u.p1 = false
		u.dpγ = false
		u.su3 &^= su3γ
	} else if u.sα {
		u.sα = false
		u.su2 &^= su2sα
		u.nac = false
		u.sv &^= svA | svCLR
		u.sac = true
		u.su2 |= su2sA | su2sCLR
		u.npγ = true
		u.sv |= svγ
		if u.psrcff {
			u.dpγ = false
			u.su3 &^= su3γ
			u.m1 = false
			u.p1 = false
		}
		u.numrplus, u.numrmin = u.numrmin, u.numrplus
	} else if u.qα {
		u.qα = false
		u.su2 &^= su2qα
		u.m1 = false
		u.p1 = false
	} else if u.dγ {
		u.dγ = false
		u.su3 &^= su3γ
		u.m2 = false
		u.p2 = false
	}
	switch u.progring {
	case 1:
		u.doP()
	case 6: // Gate D4
		u.nγ = false
		u.sv &^= svγ
		u.da = false
		u.ds = false
		u.su3 &^= su3A | su3S
	case 7: // Gate J13
		if !u.overflow() && u.roundoff[u.curprog] == 1 { // Gate K12
			if u.divff {
				u.qα = true
				u.su2 |= su2qα
				if u.samesign() {
					u.p1 = true
				} else {
					u.m1 = true
				}
			} else {
				u.dγ = true
				u.su3 |= su3γ
				if u.samesign() {
					u.p2 = true
				} else {
					u.m2 = true
				}
			}
		}
	case 8: // Gate E3. L50
		u.psrcff = false
		u.coinff = true
	}
	u.progring++
}

func (u *Divsr) clock(p Pulse, resp chan int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	switch {
	case p.Val&Cpp != 0:
		if u.progring == 0 {
			u.ans1 = false
			u.ans2 = false
			u.ans3 = false
			u.ans4 = false
			u.su2 &^= su2qA | su2qS | su2qCLR
			u.su3 &^= su3A | su3S | su3CLR
		}
		if u.curprog >= 0 {
			if u.psrcff == false { // Gate F4
				u.doGP(resp)
			} else { // Gate F5
				u.doIIIP()
			}
		}
	case p.Val&Rp != 0:
		/*
		 * Ugly hack to avoid races
		 */
		for i := 0; i < 8; i++ {
			if u.preff[i] {
				u.preff[i] = false
				u.progff[i] = true
				u.curprog = i
			}
		}
	case p.Val&Onep != 0 && u.p1 || p.Val&Twop != 0 && u.p2:
		if u.placering < 9 {
			Handshake(1<<uint(8-u.placering), u.answer, resp)
		}
	case p.Val&Onep != 0 && u.m2 || p.Val&Twopp != 0 && u.m1:
		Handshake(0x7ff, u.answer, resp)
	case p.Val&Onep != 0 && u.m1 || p.Val&Twopp != 0 && u.m2:
		if u.placering < 9 {
			Handshake(0x7ff^(1<<uint(8-u.placering)), u.answer, resp)
		} else {
			Handshake(0x7ff, u.answer, resp)
		}
	case (p.Val&Fourp != 0 || p.Val&Twop != 0) && (u.m1 || u.m2):
		Handshake(0x7ff, u.answer, resp)
	case p.Val&Onepp != 0:
		if u.m1 || u.m2 {
			Handshake(1, u.answer, resp)
		}
		if u.psrcff == false && u.sα { // Gate L45
			u.placering++
		}
	}
}

func (u *Divsr) MakeClockFunc() ClockFunc {
	resp := make(chan int)
	return func(p Pulse) {
		u.clock(p, resp)
	}
}

func (u *Divsr) Run() {
	u.intclear()
	go u.readInputs()
}

func (u *Divsr) readInputs() {
	var p Pulse

	for {
		p.Resp = nil
		select {
		case <-u.rewiring:
			u.waitingForRewiring <- 1
			<-u.rewiring
		case p = <-u.progin[0]:
			u.divargs(0)
		case p = <-u.progin[1]:
			u.divargs(1)
		case p = <-u.progin[2]:
			u.divargs(2)
		case p = <-u.progin[3]:
			u.divargs(3)
		case p = <-u.progin[4]:
			u.divargs(4)
		case p = <-u.progin[5]:
			u.divargs(5)
		case p = <-u.progin[6]:
			u.divargs(6)
		case p = <-u.progin[7]:
			u.divargs(7)
		case p = <-u.ilock[0]:
			u.interlock()
		case p = <-u.ilock[1]:
			u.interlock()
		case p = <-u.ilock[2]:
			u.interlock()
		case p = <-u.ilock[3]:
			u.interlock()
		case p = <-u.ilock[4]:
			u.interlock()
		case p = <-u.ilock[5]:
			u.interlock()
		case p = <-u.ilock[6]:
			u.interlock()
		case p = <-u.ilock[7]:
			u.interlock()
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}
