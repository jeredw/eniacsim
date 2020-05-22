package units

import (
	"bufio"
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
	"sync"
)

// Initiate simulates the ENIAC initiate unit.
type Initiate struct {
	Io InitiateConn

	gate66, gate69                  int
	prff, printPhase1, printPhase2  bool
	lastPrint                       int
	rdff, rdilock, rdsync, rdfinish bool
	lastCardRead                    int
	jack                            [18]*Jack
	clrff                           [6]bool

	cardScanner *bufio.Scanner
	punchWriter *bufio.Writer

	mu sync.Mutex
}

// InitiateConn defines connections needed for the unit
type InitiateConn struct {
	InitButton Button
	Ppunch     chan string
	Units      []Cleared
	ReadCard   func(string)
	Print      func() string

	AddCycle func() int  // Return the current add cycle
	Stepping func() bool // Return true iff single stepping
}

func NewInitiate(io InitiateConn) *Initiate {
	u := &Initiate{Io: io}
	clearInput := func(prog int) JackHandler {
		return func(*Jack, int) {
			u.mu.Lock()
			u.clrff[prog] = true
			u.mu.Unlock()
		}
	}
	for i := 0; i < 6; i++ {
		u.jack[2*i] = NewInput(fmt.Sprintf("i.Ci%d", i+1), clearInput(i))
		u.jack[2*i+1] = NewOutput(fmt.Sprintf("i.Co%d", i+1), nil)
	}
	u.jack[12] = NewInput("i.Rl", func(*Jack, int) {
		u.mu.Lock()
		u.rdilock = true
		u.mu.Unlock()
	})
	u.jack[13] = NewInput("i.Ri", func(*Jack, int) {
		u.mu.Lock()
		u.rdff = true
		u.mu.Unlock()
	})
	u.jack[14] = NewOutput("i.Ro", nil)
	u.jack[15] = NewInput("i.Pi", func(*Jack, int) {
		u.mu.Lock()
		if !u.printPhase1 {
			u.prff = true
			if !u.printPhase2 {
				u.printPhase1 = true
				u.lastPrint = u.Io.AddCycle()
			}
		}
		u.mu.Unlock()
	})
	u.jack[16] = NewOutput("i.Po", nil)
	u.jack[17] = NewOutput("i.Io", nil)
	return u
}

func (u *Initiate) ShouldClear() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.clrff[0] || u.clrff[1] || u.clrff[2] || u.clrff[3] || u.clrff[4] || u.clrff[5]
}

func (u *Initiate) SetCardScanner(cardScanner *bufio.Scanner) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.cardScanner = cardScanner
}

func (u *Initiate) SetPunchWriter(punchWriter *bufio.Writer) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.punchWriter = punchWriter
}

func (u *Initiate) Stat() string {
	u.mu.Lock()
	defer u.mu.Unlock()
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
	u.mu.Lock()
	u.gate66 = 0
	u.gate69 = 0
	u.prff = false
	u.rdff = false
	u.rdilock = false
	u.rdsync = false
	u.rdfinish = false
	for i := 0; i < 18; i++ {
		u.jack[i].Disconnect()
	}
	for i := 0; i < 6; i++ {
		u.clrff[i] = false
	}
	u.mu.Unlock()
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
		u.mu.Lock()
		defer u.mu.Unlock()
		if u.gate69 == 1 {
			u.gate66 = 0
			u.gate69 = 0
			u.mu.Unlock()
			u.jack[17].Transmit(1)
			u.mu.Lock()
		} else if u.gate66 == 1 {
			u.gate69 = 1
		}
		stepping := u.Io.Stepping()
		for i, ff := range u.clrff {
			if ff {
				u.mu.Unlock()
				u.jack[2*i+1].Transmit(1)
				u.mu.Lock()
				u.clrff[i] = false
			}
		}
		if u.rdsync {
			u.mu.Unlock()
			u.jack[14].Transmit(1)
			u.mu.Lock()
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
			u.mu.Unlock()
			u.jack[16].Transmit(1)
			u.mu.Lock()
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

func (u *Initiate) Run() {
	for {
		b := <-u.Io.InitButton.Push
		u.mu.Lock()
		switch b {
		case 4:
			u.gate66 = 1
		case 5:
			for _, c := range u.Io.Units {
				c.Clear()
			}
		case 3:
			u.rdff = true
			u.rdilock = true
		}
		u.mu.Unlock()
		u.Io.InitButton.Done <- 1
	}
}
