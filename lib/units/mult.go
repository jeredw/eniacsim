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

	multin, multout                                       [24]*Jack
	R, D                                                  [5]*Jack
	A, S, AS, AC, SC, ASC, RS, DS, F                      *Jack
	lhppI, lhppII, rhppI, rhppII                          *Jack
	stage                                                 int
	multff                                                [24]bool
	iersw, iercl, icandsw, icandcl, sigsw, placsw, prodsw [24]int
	reset1ff, reset3ff                                    bool
	buffer61, f44                                         bool
	ier, icand                                            string
	sigfig                                                int
	multl, multr                                          bool

	tracer Tracer
	mu     sync.Mutex
}

// Connections to other units.
type MultiplierConn struct {
	Ier   StaticWiring
	Icand StaticWiring
	Lhpp  StaticWiring
	Rhpp  StaticWiring
}

func NewMultiplier() *Multiplier {
	u := &Multiplier{}
	for i := 0; i < 24; i++ {
		u.multin[i] = u.newProgramInput(i)
		u.multout[i] = u.newOutput(fmt.Sprintf("m.%do", i+1), 1)
	}
	outs := []rune("αβγδε")
	for i := 0; i < 5; i++ {
		u.R[i] = u.newOutput(fmt.Sprintf("m.R%c", outs[i]), 1)
		u.D[i] = u.newOutput(fmt.Sprintf("m.D%c", outs[i]), 1)
	}
	u.A = u.newOutput("m.A", 1)
	u.S = u.newOutput("m.S", 1)
	u.AS = u.newOutput("m.AS", 1)
	u.AC = u.newOutput("m.AC", 1)
	u.SC = u.newOutput("m.SC", 1)
	u.ASC = u.newOutput("m.ASC", 1)
	u.RS = u.newOutput("m.RS", 1)
	u.DS = u.newOutput("m.DS", 1)
	u.F = u.newOutput("m.F", 1)
	u.lhppI = u.newOutput("m.lhppI", 10)
	u.lhppII = u.newOutput("m.lhppII", 10)
	u.rhppI = u.newOutput("m.rhppI", 10)
	u.rhppII = u.newOutput("m.rhppII", 10)
	return u
}

func (u *Multiplier) newProgramInput(program int) *Jack {
	return NewInput(fmt.Sprintf("m.%di", program+1), func(j *Jack, val int) {
		u.multargs(program)
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, 1, int64(val))
		}
	})
}

func (u *Multiplier) newOutput(name string, width int) *Jack {
	return NewOutput(name, func(j *Jack, val int) {
		if u.tracer != nil {
			u.tracer.LogPulse(j.Name, width, int64(val))
		}
	})
}

func (u *Multiplier) AttachTracer(tracer Tracer) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.tracer = tracer
	u.tracer.RegisterValueCallback(func() {
		ierSign, ier := StringToSignAndDigits(u.ier)
		icandSign, icand := StringToSignAndDigits(u.icand)
		u.tracer.LogValue("m.ierSign", 1, BoolToInt64(ierSign))
		u.tracer.LogValue("m.ier", 40, DigitsToInt64BCD(ier))
		u.tracer.LogValue("m.icandSign", 1, BoolToInt64(icandSign))
		u.tracer.LogValue("m.icand", 40, DigitsToInt64BCD(icand))
		u.tracer.LogValue("m.stage", 4, int64(u.stage))
	})
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
	for i := 0; i < 24; i++ {
		u.multin[i].Disconnect()
		u.multout[i].Disconnect()
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
		u.R[i].Disconnect()
		u.D[i].Disconnect()
	}
	u.A.Disconnect()
	u.S.Disconnect()
	u.AS.Disconnect()
	u.AC.Disconnect()
	u.SC.Disconnect()
	u.ASC.Disconnect()
	u.RS.Disconnect()
	u.DS.Disconnect()
	u.F.Disconnect()
	u.lhppI.Disconnect()
	u.lhppII.Disconnect()
	u.rhppI.Disconnect()
	u.rhppII.Disconnect()
	u.stage = 0
	u.reset1ff = false
	u.reset3ff = false
	u.multl = false
	u.multr = false
	u.buffer61 = false
	u.f44 = false
}

func (u *Multiplier) FindJack(jack string) (*Jack, error) {
	if len(jack) == 0 {
		return nil, fmt.Errorf("invalid jack")
	}
	switch jack {
	case "Rα", "Ra", "rα", "ra":
		return u.R[0], nil
	case "Rβ", "Rb", "rβ", "rb":
		return u.R[1], nil
	case "Rγ", "Rg", "rγ", "rg":
		return u.R[2], nil
	case "Rδ", "Rd", "rδ", "rd":
		return u.R[3], nil
	case "Rε", "Re", "rε", "re":
		return u.R[4], nil
	case "Dα", "Da", "dα", "da":
		return u.D[0], nil
	case "Dβ", "Db", "dβ", "db":
		return u.D[1], nil
	case "Dγ", "Dg", "dγ", "dg":
		return u.D[2], nil
	case "Dδ", "Dd", "dδ", "dd":
		return u.D[3], nil
	case "Dε", "De", "dε", "de":
		return u.D[4], nil
	case "A", "a":
		return u.A, nil
	case "S", "s":
		return u.S, nil
	case "AS", "as":
		return u.AS, nil
	case "AC", "ac":
		return u.AC, nil
	case "SC", "sc":
		return u.SC, nil
	case "ASC", "asc":
		return u.ASC, nil
	case "RS", "rs":
		return u.RS, nil
	case "DS", "ds":
		return u.DS, nil
	case "F", "f":
		return u.F, nil
	case "LHPPI", "lhppi", "lhppI":
		return u.lhppI, nil
	case "LHPPII", "lhppii", "lhppII":
		return u.lhppII, nil
	case "RHPPI", "rhppi", "rhppI":
		return u.rhppI, nil
	case "RHPPII", "rhppii", "rhppII":
		return u.rhppII, nil
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
		return u.multin[prog-1], nil
	case 'o':
		return u.multout[prog-1], nil
	}
	return nil, fmt.Errorf("invalid jack %s", jack)
}

func (u *Multiplier) FindSwitch(name string) (Switch, error) {
	switch {
	case len(name) > 6 && name[:6] == "ieracc":
		prog, _ := strconv.Atoi(name[6:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.iersw[prog-1], recvSettings()}, nil
	case len(name) > 5 && name[:5] == "iercl":
		prog, _ := strconv.Atoi(name[5:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.iercl[prog-1], mclSettings()}, nil
	case len(name) > 8 && name[:8] == "icandacc":
		prog, _ := strconv.Atoi(name[8:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.icandsw[prog-1], recvSettings()}, nil
	case len(name) > 7 && name[:7] == "icandcl":
		prog, _ := strconv.Atoi(name[7:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.icandcl[prog-1], mclSettings()}, nil
	case len(name) > 2 && name[:2] == "sf":
		prog, _ := strconv.Atoi(name[2:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.sigsw[prog-1], msfSettings()}, nil
	case len(name) > 5 && name[:5] == "place":
		prog, _ := strconv.Atoi(name[5:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.placsw[prog-1], mplSettings()}, nil
	case len(name) > 4 && name[:4] == "prod":
		prog, _ := strconv.Atoi(name[4:])
		if !(prog >= 1 && prog <= 24) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &IntSwitch{&u.mu, name, &u.prodsw[prog-1], prodSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

func (u *Multiplier) partialProducts(p Pulse) (lhpp, rhpp int) {
	lhpp, rhpp = 0, 0
	for i := 0; i < 10; i++ {
		tensDigit := timesTens[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
		onesDigit := timesOnes[u.ier[u.stage]-'0'][u.icand[i+2]-'0']
		if tensDigit&p != 0 {
			lhpp |= 1 << uint(9-i)
		}
		if onesDigit&p != 0 {
			rhpp |= 1 << uint(9-i)
		}
	}
	return
}

func (u *Multiplier) shiftProducts(lhpp, rhpp int) {
	if lhpp != 0 {
		u.lhppI.Transmit(lhpp >> uint(u.stage-2))
	}
	if lhpp != 0 {
		u.lhppII.Transmit((lhpp << uint(12-u.stage)) & 0x3ff)
	}
	if rhpp != 0 {
		u.rhppI.Transmit(rhpp >> uint(u.stage-1))
	}
	if rhpp != 0 {
		u.rhppII.Transmit((rhpp << uint(11-u.stage)) & 0x3ff)
	}
}

var timesTens [10][10]Pulse = [10][10]Pulse{
	{BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0]},
	{BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0]},
	{BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[1], BCD[1], BCD[1], BCD[1], BCD[1]},
	{BCD[0], BCD[0], BCD[0], BCD[0], BCD[1], BCD[1], BCD[1], BCD[2], BCD[2], BCD[2]},
	{BCD[0], BCD[0], BCD[0], BCD[1], BCD[1], BCD[2], BCD[2], BCD[2], BCD[3], BCD[3]},
	{BCD[0], BCD[0], BCD[1], BCD[1], BCD[2], BCD[2], BCD[3], BCD[3], BCD[4], BCD[4]},
	{BCD[0], BCD[0], BCD[1], BCD[1], BCD[2], BCD[3], BCD[3], BCD[4], BCD[4], BCD[5]},
	{BCD[0], BCD[0], BCD[1], BCD[2], BCD[2], BCD[3], BCD[4], BCD[4], BCD[5], BCD[6]},
	{BCD[0], BCD[0], BCD[1], BCD[2], BCD[3], BCD[4], BCD[4], BCD[5], BCD[6], BCD[7]},
	{BCD[0], BCD[0], BCD[1], BCD[2], BCD[3], BCD[4], BCD[5], BCD[6], BCD[7], BCD[8]},
}

var timesOnes [10][10]Pulse = [10][10]Pulse{
	{BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0], BCD[0]},
	{BCD[0], BCD[1], BCD[2], BCD[3], BCD[4], BCD[5], BCD[6], BCD[7], BCD[8], BCD[9]},
	{BCD[0], BCD[2], BCD[4], BCD[6], BCD[8], BCD[0], BCD[2], BCD[4], BCD[6], BCD[8]},
	{BCD[0], BCD[3], BCD[6], BCD[9], BCD[2], BCD[5], BCD[8], BCD[1], BCD[4], BCD[7]},
	{BCD[0], BCD[4], BCD[8], BCD[2], BCD[6], BCD[0], BCD[4], BCD[8], BCD[2], BCD[6]},
	{BCD[0], BCD[5], BCD[0], BCD[5], BCD[0], BCD[5], BCD[0], BCD[5], BCD[0], BCD[5]},
	{BCD[0], BCD[6], BCD[2], BCD[8], BCD[4], BCD[0], BCD[6], BCD[2], BCD[8], BCD[4]},
	{BCD[0], BCD[7], BCD[4], BCD[1], BCD[8], BCD[5], BCD[2], BCD[9], BCD[6], BCD[3]},
	{BCD[0], BCD[8], BCD[6], BCD[4], BCD[2], BCD[0], BCD[8], BCD[6], BCD[4], BCD[2]},
	{BCD[0], BCD[9], BCD[8], BCD[7], BCD[6], BCD[5], BCD[4], BCD[3], BCD[2], BCD[1]},
}

func (u *Multiplier) activeProgram() int {
	for i := range u.multff {
		if u.multff[i] {
			return i
		}
	}
	return -1
}

func (u *Multiplier) places() int {
	if i := u.activeProgram(); i != -1 {
		if u.placsw[i]+2 < 10 {
			return u.placsw[i] + 2
		}
	}
	return 10
}

func (u *Multiplier) Clock(c Pulse) {
	switch {
	case c&Cpp != 0:
		u.doCpp()
	case c&Ccg != 0 && u.stage == 13:
		if i := u.activeProgram(); i != -1 {
			if u.iercl[i] == 1 {
				u.Io.Ier.Clear()
			}
			if u.icandcl[i] == 1 {
				u.Io.Icand.Clear()
			}
		}
	case c&(Onep|Fourp) != 0 && u.stage == 1:
		if c&Onep != 0 {
			u.multl = true
			u.multr = true
			u.Io.Lhpp.SetExternalProgram(opα)
			u.Io.Rhpp.SetExternalProgram(opα)
			u.sigfig = -1
			if i := u.activeProgram(); i != -1 {
				u.sigfig = u.sigsw[i]
			}
		}
		// Transmit +5 in the roundoff place.
		if u.sigfig == 0 && u.lhppII.Connected() {
			u.lhppII.Transmit(1 << 10)
		} else if u.sigfig > 0 && u.sigfig < 9 {
			u.lhppI.Transmit(1 << uint(u.sigfig-1))
		}
	case c&(Onep|Twop|Twopp|Fourp) != 0 && u.stage >= 2 && u.stage < 12:
		if c&Onep != 0 {
			u.ier = u.Io.Ier.Value()
			u.icand = u.Io.Icand.Value()
		}
		lhpp, rhpp := u.partialProducts(c)
		u.shiftProducts(lhpp, rhpp)
	case c&Onepp != 0 && u.stage >= 2 && u.stage < 12:
		if u.stage == u.places()+1 && u.ier[0] == 'M' && u.icand[0] == 'M' {
			u.rhppI.Transmit(1 << 10)
		}
	case c&Rp != 0 && u.buffer61:
		u.buffer61 = false
		u.f44 = true
	}
}

func (u *Multiplier) doCpp() {
	if u.f44 {
		u.stage = 1
		u.f44 = false
	} else if u.stage == 12 {
		u.reset1ff = true
		u.reset3ff = true
		u.F.Transmit(1)
		u.stage++
	} else if u.stage == 13 {
		if i := u.activeProgram(); i != -1 {
			u.multout[i].Transmit(1)
			u.multff[i] = false
			switch u.prodsw[i] {
			case 0:
				u.A.Transmit(1)
			case 1:
				u.S.Transmit(1)
			case 2:
				u.AS.Transmit(1)
			case 4:
				u.AC.Transmit(1)
			case 5:
				u.SC.Transmit(1)
			case 6:
				u.ASC.Transmit(1)
			}
		}
		u.reset1ff = false
		u.reset3ff = false
		u.stage = 0
	} else if u.stage != 0 {
		if u.stage == u.places()+1 {
			if u.ier[0] == 'M' {
				u.DS.Transmit(1)
			}
			if u.icand[0] == 'M' {
				u.RS.Transmit(1)
			}
			u.multl = false
			u.multr = false
			u.Io.Lhpp.SetExternalProgram(0)
			u.Io.Rhpp.SetExternalProgram(0)
			u.stage = 12
		} else {
			u.stage++
		}
	}
}

func (u *Multiplier) multargs(prog int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	ier := u.iersw[prog]
	icand := u.icandsw[prog]
	if ier < 5 {
		u.R[ier].Transmit(1)
	}
	if icand < 5 {
		u.D[icand].Transmit(1)
	}
	u.multff[prog] = true
	u.buffer61 = true
}
