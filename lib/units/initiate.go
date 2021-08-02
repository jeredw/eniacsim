package units

import (
	"bufio"
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
)

// Initiate simulates the ENIAC initiate unit.
type Initiate struct {
	Io InitiateConn

	gate66, gate69                  int
	prff, printPhase1, printPhase2  bool
	lastPrint                       int64
	rdff, rdilock, rdsync, rdfinish bool
	lastCardRead                    int64
	jack                            [18]*Jack
	clrff                           [6]bool

	cardScanner *bufio.Scanner
	punchWriter *bufio.Writer
}

// InitiateConn defines connections needed for the unit
type InitiateConn struct {
	InitButton Button
	Ppunch     chan string
	Units      []Cleared
	ReadCard   func(string)
	Print      func() string

	AddCycle func() int64  // Return the current add cycle
	Stepping func() bool // Return true iff single stepping
}

func NewInitiate(io InitiateConn) *Initiate {
	u := &Initiate{Io: io}
	clearInput := func(prog int) JackHandler {
		return func(*Jack, int) {
			u.clrff[prog] = true
		}
	}
	for i := 0; i < 6; i++ {
		u.jack[2*i] = NewInput(fmt.Sprintf("i.Ci%d", i+1), clearInput(i))
		u.jack[2*i+1] = NewOutput(fmt.Sprintf("i.Co%d", i+1), nil)
	}
	u.jack[12] = NewInput("i.Rl", func(*Jack, int) {
		u.rdilock = true
	})
	u.jack[13] = NewInput("i.Ri", func(*Jack, int) {
		u.rdff = true
	})
	u.jack[14] = NewOutput("i.Ro", nil)
	u.jack[15] = NewInput("i.Pi", func(*Jack, int) {
		if !u.printPhase1 {
			u.prff = true
			if !u.printPhase2 {
				u.printPhase1 = true
				u.lastPrint = u.Io.AddCycle()
			}
		}
	})
	u.jack[16] = NewOutput("i.Po", nil)
	u.jack[17] = NewOutput("i.Io", nil)
	return u
}

func (u *Initiate) SelectiveClear() bool {
	return u.clrff[0] || u.clrff[1] || u.clrff[2] || u.clrff[3] || u.clrff[4] || u.clrff[5]
}

func (u *Initiate) SetCardScanner(cardScanner *bufio.Scanner) {
	u.cardScanner = cardScanner
}

func (u *Initiate) SetPunchWriter(punchWriter *bufio.Writer) {
	u.punchWriter = punchWriter
}

func (u *Initiate) Stat() string {
	s := ""
	for _, f := range u.clrff {
		s += ToBin(f)
	}
	s += ToBin(u.rdff)
	s += ToBin(u.prff)
	s += ToBin(u.rdfinish)
	s += ToBin(u.rdilock)
	s += ToBin(u.rdsync)
	s += "00"
	s += fmt.Sprintf("%d%d", u.gate66, u.gate69)
	return s
}

func (u *Initiate) Reset() {
	u.gate66 = 0
	u.gate69 = 0
	u.prff = false
	u.rdff = false
	u.rdilock = false
	u.rdsync = false
	u.rdfinish = false
	for i := 0; i < 6; i++ {
		u.clrff[i] = false
	}
}

func (u *Initiate) FindJack(jack string) (*Jack, error) {
	if len(jack) == 0 {
		return nil, fmt.Errorf("invalid jack")
	}
	switch jack[0] {
	case 'c', 'C':
		if len(jack) < 3 {
			return nil, fmt.Errorf("invalid jack %s", jack)
		}
		set, _ := strconv.Atoi(jack[2:])
		if !(set >= 1 && set <= 6) {
			return nil, fmt.Errorf("invalid jack %s", jack)
		}
		switch jack[1] {
		case 'i':
			return u.jack[2*(set-1)], nil
		case 'o':
			return u.jack[2*(set-1)+1], nil
		default:
			return nil, fmt.Errorf("invalid jack %s", jack)
		}
	case 'i', 'I':
		return u.jack[17], nil
	case 'p', 'P':
		if len(jack) < 2 {
			return nil, fmt.Errorf("invalid jack %s", jack)
		}
		switch jack[1] {
		case 'i':
			return u.jack[15], nil
		case 'o':
			return u.jack[16], nil
		default:
			return nil, fmt.Errorf("invalid jack %s", jack)
		}
	case 'r', 'R':
		if len(jack) < 2 {
			return nil, fmt.Errorf("invalid jack %s", jack)
		}
		switch jack[1] {
		case 'l':
			return u.jack[12], nil
		case 'i':
			return u.jack[13], nil
		case 'o':
			return u.jack[14], nil
		default:
			return nil, fmt.Errorf("invalid jack %s", jack)
		}
	}
	return nil, fmt.Errorf("invalid jack %s", jack)
}

func (u *Initiate) Clock(cyc Pulse) {
	if cyc&Cpp != 0 {
		if u.gate69 == 1 {
			u.gate66 = 0
			u.gate69 = 0
			u.jack[17].Transmit(1)
		} else if u.gate66 == 1 {
			u.gate69 = 1
		}
		stepping := u.Io.Stepping()
		for i, ff := range u.clrff {
			if ff {
				u.jack[2*i+1].Transmit(1)
				u.clrff[i] = false
			}
		}
		if u.rdsync {
			u.jack[14].Transmit(1)
			u.rdff = false
			u.rdilock = false
			u.rdsync = false
			u.rdfinish = false
		}
		sinceCardRead := u.Io.AddCycle() - u.lastCardRead
		if u.rdff && (stepping || sinceCardRead > MsToAddCycles(375)) {
			if u.cardScanner != nil {
				if u.cardScanner.Scan() {
					card := u.cardScanner.Text()
					u.Io.ReadCard(card)
					u.lastCardRead = u.Io.AddCycle()
					u.rdfinish = true
				} else {
					u.cardScanner = nil
				}
			}
		}
		if u.rdfinish && u.rdilock {
			u.rdsync = true
		}
		sincePrint := u.Io.AddCycle() - u.lastPrint
		if u.printPhase1 && (stepping || sincePrint > MsToAddCycles(150)) {
			s := u.Io.Print()
			if u.punchWriter != nil {
				u.punchWriter.WriteString(s)
				u.punchWriter.WriteByte('\n')
			} else {
				fmt.Println(s)
			}
			if u.Io.Ppunch != nil {
				u.Io.Ppunch <- s
			}
			u.jack[16].Transmit(1)
			u.lastPrint = u.Io.AddCycle()
			u.printPhase1 = false
			u.printPhase2 = true
			u.prff = false
		}
		if u.printPhase2 && (stepping || sincePrint > MsToAddCycles(450)) {
			if u.prff {
				u.lastPrint = u.Io.AddCycle()
				u.printPhase1 = true
			}
			u.printPhase2 = false
		}
	}
}

func (u *Initiate) PushClearButton() {
	for _, c := range u.Io.Units {
		c.Clear()
	}
}

func (u *Initiate) PushReadButton() {
	u.rdff = true
	u.rdilock = true
}

func (u *Initiate) PushInitButton() {
	u.gate66 = 1
}
