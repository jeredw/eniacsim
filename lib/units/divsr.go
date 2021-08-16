package units

import (
	"encoding/json"
	"fmt"
	"strconv"

	. "github.com/jeredw/eniacsim/lib"
)

// Divsr simulates the ENIAC divider/square rooter unit.
type Divsr struct {
	Io DivsrConn

	progin, progout, ilock                                                     [8]*Jack
	answer                                                                     *Jack
	numarg, denarg, roundoff, places, ilocksw, anssw                           [8]int
	numcl, dencl                                                               [8]bool
	preff, progff                                                              [8]bool
	placering, progring                                                        int
	divff, clrff, ilockff, coinff, dpγ, nγ, psrcff, pringff, denomff, numrplus bool
	numrmin, qα, sac, m2, m1, nac, da, nα, dα, dγ, npγ, p2, p1, sα, ds, nβ, dβ bool
	ans1, ans2, ans3, ans4                                                     bool
	curprog, divadap, sradap                                                   int
	sv, su2q, su2s, su3                                                        int

	tracer Tracer
}

// Connections to dedicated accumulators.
type DivsrConn struct {
	Numerator   StaticWiring
	Denominator StaticWiring
	Quotient    StaticWiring
	Shift       StaticWiring
}

func NewDivsr() *Divsr {
	u := &Divsr{}
	u.intclear()
	for i := 0; i < 8; i++ {
		u.progin[i] = u.newProgramInput(i)
		u.ilock[i] = u.newInterlockInput(i)
		u.progout[i] = u.newOutput(fmt.Sprintf("d.%do", i+1), 1)
	}
	u.answer = u.newOutput("d.ans", 11)
	return u
}

func (u *Divsr) newProgramInput(program int) *Jack {
	return NewInput(fmt.Sprintf("d.%di", program+1), func(j *Jack, val int) {
		u.divargs(program)
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, 1, int64(val))
		}
	})
}

func (u *Divsr) newInterlockInput(program int) *Jack {
	return NewInput(fmt.Sprintf("d.%dl", program+1), func(j *Jack, val int) {
		u.interlock()
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, 1, int64(val))
		}
	})
}

func (u *Divsr) newOutput(name string, width int) *Jack {
	return NewOutput(name, func(j *Jack, val int) {
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, width, int64(val))
		}
	})
}

type divsrJson struct {
	PlaceRing int     `json:"placeRing"`
	ProgRing  int     `json:"progRing"`
	Program   [8]bool `json:"program"`
	Ffs       string  `json:"ffs"`
}

func (u *Divsr) State() json.RawMessage {
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

func (u *Divsr) AttachTracer(tracer Tracer) {
	u.tracer = tracer
	u.tracer.RegisterValueCallback(func() {
		tracer.LogValue("d.place", 5, int64(u.placering))
		tracer.LogValue("d.prog", 5, int64(u.progring))
	})
}

func (u *Divsr) Reset() {
	for i := 0; i < 8; i++ {
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
	u.divff = false
	u.ilockff = false
	u.ans1 = false
	u.ans2 = false
	u.ans3 = false
	u.ans4 = false
	u.divadap = 0
	u.sradap = 0
	u.Clear()
}

func (u *Divsr) Clear() {
	u.intclear()
	u.sv = 0
	u.su2s = 0
	u.su2q = 0
	u.su3 = 0
}

func (u *Divsr) intclear() {
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

func (u *Divsr) FindJack(jack string) (*Jack, error) {
	if jack == "ans" || jack == "ANS" {
		return u.answer, nil
	}
	var prog int
	var ilk rune
	fmt.Sscanf(jack, "%d%c", &prog, &ilk)
	if !(prog >= 1 && prog <= 8) {
		return nil, fmt.Errorf("invalid jack %s", jack)
	}
	switch ilk {
	case 'i':
		return u.progin[prog-1], nil
	case 'o':
		return u.progout[prog-1], nil
	case 'l':
		return u.ilock[prog-1], nil
	}
	return nil, fmt.Errorf("invalid jack %s", jack)
}

func (u *Divsr) FindSwitch(name string) (Switch, error) {
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
		return &BoolSwitch{name, &u.numcl[sw-1], clearSettings()}, nil
	case "dr":
		return &IntSwitch{name, &u.denarg[sw-1], argSettings()}, nil
	case "dc":
		return &BoolSwitch{name, &u.dencl[sw-1], clearSettings()}, nil
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

func (u *Divsr) divargs(prog int) {
	u.preff[prog] = true
	if u.places[prog] < 5 {
		u.divff = true
	} else {
		u.divff = false
	}
	switch u.numarg[prog] {
	case 0:
		u.nα = true
		u.sv |= opα
	case 1:
		u.nβ = true
		u.sv |= opβ
	}
	switch u.denarg[prog] {
	case 0:
		u.dα = true
		u.su3 |= opα
	case 1:
		u.dβ = true
		u.su3 |= opβ
	}
	u.setExternalPrograms()
}

func (u *Divsr) doP() {
	u.nγ = true
	u.sv |= opγ
	if u.samesign() {
		u.ds = true
		u.su3 |= opS
	} else {
		u.da = true
		u.su3 |= opA
	}
}

func (u *Divsr) doS() {
	u.sα = true
	u.su2s |= opα
	u.nac = true
	u.sv |= opA | opClear
	if !u.divff {
		if u.samesign() {
			u.m1 = true
		} else {
			u.p1 = true
		}
		u.dpγ = true
		u.su3 |= opγ
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
	s := u.Io.Numerator.Sign()
	return s[0] == 'P' && u.numrmin || s[0] == 'M' && u.numrplus
}

func (u *Divsr) interlock() {
	u.ilockff = true
}

func (u *Divsr) doGP() {
	if u.coinff { // Gate E50
		if u.ilocksw[u.curprog] == 0 || u.ilockff {
			u.coinff = false
			u.clrff = true
			return
		}
	} else if u.clrff {
		u.progff[u.curprog] = false
		u.progout[u.curprog].Transmit(1)
		if u.ilocksw[u.curprog] == 1 {
			u.ilockff = false
		}
		/*
		 * Implement the PX-4-114 adapters
		 */
		switch u.anssw[u.curprog] {
		case 0:
			u.ans1 = true
			u.su2q |= opA
			if u.divadap == 2 {
				u.su2q |= opClear
			}
		case 1:
			u.ans2 = true
			switch u.divadap {
			case 0:
				u.su2q |= opA | opClear
			case 1:
				u.su2q |= opS
			case 2:
				u.su2q |= opS | opClear
			}
		case 2:
			u.ans3 = true
			u.su3 |= opA
			if u.sradap == 2 {
				u.su3 |= opClear
			}
		case 3:
			u.ans4 = true
			switch u.sradap {
			case 0:
				u.su3 |= opA | opClear
			case 1:
				u.su3 |= opS
			case 2:
				u.su3 |= opS | opClear
			}
		}
		if u.numcl[u.curprog] {
			u.Io.Numerator.Clear()
		}
		if u.dencl[u.curprog] {
			u.Io.Denominator.Clear()
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
		u.su2q &^= opα
	} else if u.nγ { //  Gates L10, G11, H11
		u.nγ = false
		u.sv &^= opγ
		if u.divff {
			u.qα = true
			u.su2q |= opα
			if u.ds {
				u.ds = false
				u.su3 &^= opS
				u.p1 = true
			} else if u.da {
				u.da = false
				u.su3 &^= opA
				u.m1 = true
			}
		} else {
			u.dγ = true
			u.su3 |= opγ
			if u.ds {
				u.ds = false
				u.su3 &^= opS
				u.p2 = true
			} else if u.da {
				u.da = false
				u.su3 &^= opA
				u.m2 = true
			}
		}
	} else if u.npγ { // Gate C9
		u.npγ = false
		u.sv &^= opγ
		u.sac = false
		u.su2s &^= opA | opClear
		u.m1 = false
		u.p1 = false
		u.dpγ = false
		u.su3 &^= opγ
		u.doP()
	} else if u.sα { // Gates K7, L1
		u.sα = false
		u.su2s &^= opα
		u.nac = false
		u.sv &^= opA | opClear
		u.sac = true
		u.su2s |= opA | opClear
		u.npγ = true
		u.sv |= opγ
		u.numrplus, u.numrmin = u.numrmin, u.numrplus
	} else if u.dγ {
		u.dγ = false
		u.su3 &^= opγ
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
		u.sv &^= opα | opβ
		u.dα = false
		u.dβ = false
		u.su3 &^= opα | opβ
		if !u.pringff {
			u.progring++
		}
	case 1: // Gate D6
		s := u.Io.Numerator.Sign()
		if s[0] == 'M' {
			u.numrplus, u.numrmin = u.numrmin, u.numrplus
		}
		s = u.Io.Denominator.Sign()
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
			u.su3 |= opγ
			u.progring++
		}
	case 3:
		u.p1 = false
		u.dγ = false
		u.su3 &^= opγ
		u.doP()
		u.pringff = true
		u.progring = 0
	}
}

func (u *Divsr) doIIIP() {
	if u.npγ { // Gate C9
		u.npγ = false
		u.sv &^= opγ
		u.sac = false
		u.su2s &^= opA | opClear
		u.m1 = false
		u.p1 = false
		u.dpγ = false
		u.su3 &^= opγ
	} else if u.sα {
		u.sα = false
		u.su2s &^= opα
		u.nac = false
		u.sv &^= opA | opClear
		u.sac = true
		u.su2s |= opA | opClear
		u.npγ = true
		u.sv |= opγ
		if u.psrcff {
			u.dpγ = false
			u.su3 &^= opγ
			u.m1 = false
			u.p1 = false
		}
		u.numrplus, u.numrmin = u.numrmin, u.numrplus
	} else if u.qα {
		u.qα = false
		u.su2q &^= opα
		u.m1 = false
		u.p1 = false
	} else if u.dγ {
		u.dγ = false
		u.su3 &^= opγ
		u.m2 = false
		u.p2 = false
	}
	switch u.progring {
	case 1:
		u.doP()
	case 6: // Gate D4
		u.nγ = false
		u.sv &^= opγ
		u.da = false
		u.ds = false
		u.su3 &^= opA | opS
	case 7: // Gate J13
		if !u.overflow() && u.roundoff[u.curprog] == 1 { // Gate K12
			if u.divff {
				u.qα = true
				u.su2q |= opα
				if u.samesign() {
					u.p1 = true
				} else {
					u.m1 = true
				}
			} else {
				u.dγ = true
				u.su3 |= opγ
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

func (u *Divsr) Clock(p Pulse) {
	switch {
	case p&Cpp != 0:
		if u.progring == 0 {
			u.ans1 = false
			u.ans2 = false
			u.ans3 = false
			u.ans4 = false
			u.su2q &^= opA | opS | opClear
			u.su3 &^= opA | opS | opClear
		}
		if u.curprog >= 0 {
			if !u.psrcff { // Gate F4
				u.doGP()
			} else { // Gate F5
				u.doIIIP()
			}
		}
		u.setExternalPrograms()
	case p&Rp != 0:
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
	case p&Onep != 0 && u.p1 || p&Twop != 0 && u.p2:
		if u.placering < 9 {
			u.answer.Transmit(1 << uint(8-u.placering))
		}
	case p&Onep != 0 && u.m2 || p&Twopp != 0 && u.m1:
		u.answer.Transmit(0x7ff)
	case p&Onep != 0 && u.m1 || p&Twopp != 0 && u.m2:
		if u.placering < 9 {
			u.answer.Transmit(0x7ff ^ (1 << uint(8-u.placering)))
		} else {
			u.answer.Transmit(0x7ff)
		}
	case (p&Fourp != 0 || p&Twop != 0) && (u.m1 || u.m2):
		u.answer.Transmit(0x7ff)
	case p&Onepp != 0:
		if u.m1 || u.m2 {
			u.answer.Transmit(1)
		}
		if !u.psrcff && u.sα { // Gate L45
			u.placering++
		}
	}
}

func (u *Divsr) setExternalPrograms() {
	u.Io.Quotient.SetExternalProgram(u.su2q)
	u.Io.Numerator.SetExternalProgram(u.sv)
	u.Io.Denominator.SetExternalProgram(u.su3)
	u.Io.Shift.SetExternalProgram(u.su2s)
}
