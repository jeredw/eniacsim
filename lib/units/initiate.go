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
	jack                            [18]chan Pulse
	clrff                           [6]bool

	cardScanner *bufio.Scanner
	punchWriter *bufio.Writer

	update           chan int
	waitingForUpdate chan int

	mu sync.Mutex
}

// InitiateConn defines connections needed for the unit
type InitiateConn struct {
	Printer PrConn

	InitButton Button
	Ppunch     chan string
	ClearUnits []func()

	AddCycle func() int  // Return the current add cycle
	Stepping func() bool // Return true iff single stepping
}

func NewInitiate(io InitiateConn) *Initiate {
	return &Initiate{
		Io:               io,
		update:           make(chan int),
		waitingForUpdate: make(chan int),
	}
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
	u.update <- 1
	<-u.waitingForUpdate
	u.mu.Lock()
	u.gate66 = 0
	u.gate69 = 0
	u.prff = false
	u.rdff = false
	u.rdilock = false
	u.rdsync = false
	u.rdfinish = false
	for i := 0; i < 18; i++ {
		u.jack[i] = nil
	}
	for i := 0; i < 6; i++ {
		u.clrff[i] = false
	}
	u.mu.Unlock()
	u.update <- 1
}

func (u *Initiate) Plug(jackName string, ch chan Pulse) {
	fullName := "i." + jackName
	u.update <- 1
	<-u.waitingForUpdate
	u.mu.Lock()
	switch jackName[0] {
	case 'c', 'C':
		set, _ := strconv.Atoi(jackName[2:])
		if set >= 1 && set <= 6 {
			switch jackName[1] {
			case 'i':
				SafePlug(fullName, &u.jack[2*(set-1)], ch)
			case 'o':
				SafePlug(fullName, &u.jack[2*(set-1)+1], ch)
			default:
				goto badJack
			}
		} else {
			goto badJack
		}
	case 'i', 'I':
		SafePlug(fullName, &u.jack[17], ch)
	case 'p', 'P':
		switch jackName[1] {
		case 'i':
			SafePlug(fullName, &u.jack[15], ch)
		case 'o':
			SafePlug(fullName, &u.jack[16], ch)
		default:
			goto badJack
		}
	case 'r', 'R':
		switch jackName[1] {
		case 'l':
			SafePlug(fullName, &u.jack[12], ch)
		case 'i':
			SafePlug(fullName, &u.jack[13], ch)
		case 'o':
			SafePlug(fullName, &u.jack[14], ch)
		default:
			goto badJack
		}
	default:
		goto badJack
	}
	u.mu.Unlock()
	u.update <- 1
	return

badJack:
	u.mu.Unlock()
	u.update <- 1
	fmt.Printf("Invalid initiate jack: %s\n", jackName)
}

func (u *Initiate) MakeClockFunc() ClockFunc {
	resp := make(chan int)
	return func(p Pulse) {
		u.clock(p, resp)
	}
}

func (u *Initiate) clock(p Pulse, resp chan int) {
	cyc := p.Val
	if cyc&Cpp != 0 {
		u.mu.Lock()
		defer u.mu.Unlock()
		if u.gate69 == 1 {
			u.gate66 = 0
			u.gate69 = 0
			Handshake(1, u.jack[17], resp)
		} else if u.gate66 == 1 {
			u.gate69 = 1
		}
		stepping := u.Io.Stepping()
		for i, ff := range u.clrff {
			if ff {
				Handshake(1, u.jack[2*i+1], resp)
				u.clrff[i] = false
			}
		}
		if u.rdsync {
			Handshake(1, u.jack[14], resp)
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
					proccard(card)
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
			s := doprint(u.Io.Printer)
			if u.punchWriter != nil {
				u.punchWriter.WriteString(s)
				u.punchWriter.WriteByte('\n')
			} else {
				fmt.Println(s)
			}
			if u.Io.Ppunch != nil {
				u.Io.Ppunch <- s
			}
			Handshake(1, u.jack[16], resp)
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
	go u.readInputs()
	for {
		select {
		case b := <-u.Io.InitButton.Push:
			u.mu.Lock()
			switch b {
			case 4:
				u.gate66 = 1
			case 5:
				for _, f := range u.Io.ClearUnits {
					f()
				}
			case 3:
				u.rdff = true
				u.rdilock = true
			}
			u.mu.Unlock()
			u.Io.InitButton.Done <- 1
		}
	}
}

func (u *Initiate) readInputs() {
	for {
		select {
		case <-u.update:
			u.waitingForUpdate <- 1
			<-u.update
		case p := <-u.jack[12]:
			u.mu.Lock()
			u.rdilock = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[13]:
			u.mu.Lock()
			u.rdff = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[15]:
			u.mu.Lock()
			if !u.printPhase1 {
				u.prff = true
				if !u.printPhase2 {
					u.printPhase1 = true
					u.lastPrint = u.Io.AddCycle()
				}
			}
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[0]:
			u.mu.Lock()
			u.clrff[0] = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[2]:
			u.mu.Lock()
			u.clrff[1] = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[4]:
			u.mu.Lock()
			u.clrff[2] = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[6]:
			u.mu.Lock()
			u.clrff[3] = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[8]:
			u.mu.Lock()
			u.clrff[4] = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		case p := <-u.jack[10]:
			u.mu.Lock()
			u.clrff[5] = true
			u.mu.Unlock()
			if p.Resp != nil {
				p.Resp <- 1
			}
		}
	}
}
