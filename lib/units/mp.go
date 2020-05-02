package units

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	. "github.com/jeredw/eniacsim/lib"
)

// Mp simulates the ENIAC master programmer unit.
type Mp struct {
	stepper    [10]mpStepper // Steppers (A-K)
	decade     [20]mpDecade  // Decade counters (#20 down to #1)
	associator [8]byte       // Stepper to decade associations

	rewiring           chan int // main thread signals starting/done with rewiring
	waitingForRewiring chan int // Run signals waiting for rewiring

	mu       sync.Mutex
	outputMu sync.Mutex
}

type mpStepper struct {
	stage      int // Stage counter (0..5)
	di, i, cdi chan Pulse
	o          [6]chan Pulse
	csw        int
	inff       int
	kludge     bool
}

func (s *mpStepper) increment() {
	if s.kludge {
		// Don't increment again if re-triggered on "this" Cpp.
		return
	}
	if s.stage >= s.csw {
		s.stage = 0
	} else {
		s.stage++
	}
	s.kludge = true
}

type mpDecade struct {
	val   int        // Counter (one digit)
	carry bool       // Carry out from counter
	di    chan Pulse // Advance counter
	limit [6]int     // Per-stage max value for counter
}

func (d *mpDecade) increment() {
	d.val++
	if d.val == 10 {
		d.val = 0
		d.carry = true
	}
}

var associatorResets [8]byte = [8]byte{'A', 'B', 'C', 'D', 'F', 'G', 'H', 'J'}

func NewMp() *Mp {
	return &Mp{
		rewiring:           make(chan int),
		waitingForRewiring: make(chan int),
		associator:         associatorResets,
	}
}

func (u *Mp) PrinterDecades() string {
	s := ""
	// Printer is wired to decades #14-18 which are at indices 2-6
	for i := 2; i <= 6; i++ {
		s += fmt.Sprintf("%d", u.decade[i].val)
	}
	return s
}

func (u *Mp) Stat() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	var s string
	for i := range u.stepper {
		s += fmt.Sprintf("%d", u.stepper[i].stage)
	}
	s += " "
	for i := range u.decade {
		s += fmt.Sprintf("%d", u.decade[i].val)
	}
	s += " "
	for i := range u.stepper {
		if u.stepper[i].inff < 10 {
			s += fmt.Sprintf("%d", u.stepper[i].inff)
		} else {
			s += "*"
		}
	}
	return s
}

type mpJson struct {
	Stage  [10]int  `json:"stage"`  // A-K
	Inff   [10]bool `json:"inff"`   // A-K
	Decade [20]int  `json:"decade"` // 20 downto 1
}

func (u *Mp) State() json.RawMessage {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := mpJson{}
	for i := range u.stepper {
		s.Stage[i] = u.stepper[i].stage
	}
	for i := range u.decade {
		s.Decade[i] = u.decade[i].val
	}
	for i := range u.stepper {
		s.Inff[i] = u.stepper[i].inff > 0
	}
	result, _ := json.Marshal(s)
	return result
}

func (u *Mp) Reset() {
	u.rewiring <- 1
	<-u.waitingForRewiring
	u.mu.Lock()
	for i := range u.decade {
		u.decade[i].di = nil
		for j := range u.decade[i].limit {
			u.decade[i].limit[j] = 0
		}
	}
	for i := range u.stepper {
		u.stepper[i].di = nil
		u.stepper[i].i = nil
		u.stepper[i].cdi = nil
		for j := range u.stepper[i].o {
			u.stepper[i].o[j] = nil
		}
		u.stepper[i].csw = 0
		u.stepper[i].kludge = false
	}
	for i := range u.associator {
		u.associator[i] = associatorResets[i]
	}
	u.mu.Unlock()
	u.Clear()
	u.rewiring <- 1
}

// Clear resets decades and stepper stage counters.
func (u *Mp) Clear() {
	u.mu.Lock()
	defer u.mu.Unlock()
	for i := range u.decade {
		u.decade[i].val = 0
		u.decade[i].carry = false
	}
	for i := range u.stepper {
		u.stepper[i].stage = 0
		u.stepper[i].inff = 0
	}
}

// Plug connects channel ch to the specified jack.
func (u *Mp) Plug(jack string, ch chan Pulse, output bool) error {
	if len(jack) == 0 {
		return fmt.Errorf("invalid jack")
	}
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	var n int
	name := "p." + jack
	if unicode.IsDigit(rune(jack[0])) {
		fmt.Sscanf(jack, "%ddi", &n)
		if !(n >= 1 && n <= 20) {
			return fmt.Errorf("invalid decade %s", jack)
		}
		SafePlug(name, &u.decade[20-n].di, ch, output)
	} else {
		s := stepperNameToIndex(jack[0])
		if s == -1 {
			return fmt.Errorf("invalid stepper %s", jack)
		}
		if len(jack) < 2 {
			return fmt.Errorf("invalid stepper input %s", jack)
		}
		switch jack[1:] {
		case "di":
			SafePlug(name, &u.stepper[s].di, ch, output)
		case "i":
			SafePlug(name, &u.stepper[s].i, ch, output)
		case "cdi":
			SafePlug(name, &u.stepper[s].cdi, ch, output)
		default:
			if len(jack) < 3 {
				return fmt.Errorf("invalid jack %s", jack)
			}
			fmt.Sscanf(jack[1:], "%do", &n)
			if !(n >= 1 && n <= 6) {
				return fmt.Errorf("invalid output %s", jack)
			}
			u.outputMu.Lock()
			SafePlug(name, &u.stepper[s].o[n-1], ch, output)
			u.outputMu.Unlock()
		}
	}
	return nil
}

type associatorSwitch struct {
	name        string
	left, right string
	data        *byte
}

func (s *associatorSwitch) Get() string {
	return string(*s.data)
}

func (s *associatorSwitch) Set(value string) error {
	ucLeft := strings.ToUpper(s.left)
	ucRight := strings.ToUpper(s.right)
	switch value {
	case s.left, ucLeft:
		*s.data = ucLeft[0]
	case s.right, ucRight:
		*s.data = ucRight[0]
	default:
		return fmt.Errorf("%s associator invalid setting %s", s.name, value)
	}
	return nil
}

type decadeSwitch struct {
	name string
	data *int
}

func (s *decadeSwitch) Set(value string) error {
	n, _ := strconv.Atoi(value)
	if !(n >= 0 && n <= 9) {
		return fmt.Errorf("invalid decade limit %s %s\n", s.name, value)
	}
	*s.data = n
	return nil
}

func (s *decadeSwitch) Get() string {
	return fmt.Sprintf("%d", *s.data)
}

type clearSwitch struct {
	name string
	data *int
}

func (s *clearSwitch) Set(value string) error {
	n, _ := strconv.Atoi(value)
	if !(n >= 1 && n <= 6) {
		return fmt.Errorf("invalid clear stage %s\n", value)
	}
	*s.data = n - 1
	return nil
}

func (s *clearSwitch) Get() string {
	return fmt.Sprintf("%d", 1+*s.data)
}

func (u *Mp) lookupSwitch(name string) (Switch, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("invalid switch")
	}
	var d, s int
	switch name[0] {
	case 'a', 'A':
		switch name {
		case "a20", "A20":
			return &associatorSwitch{name: name, left: "a", right: "b", data: &u.associator[0]}, nil
		case "a18", "A18":
			return &associatorSwitch{name: name, left: "b", right: "c", data: &u.associator[1]}, nil
		case "a14", "A14":
			return &associatorSwitch{name: name, left: "c", right: "d", data: &u.associator[2]}, nil
		case "a12", "A12":
			return &associatorSwitch{name: name, left: "d", right: "e", data: &u.associator[3]}, nil
		case "a10", "A10":
			return &associatorSwitch{name: name, left: "f", right: "g", data: &u.associator[4]}, nil
		case "a8", "A8":
			return &associatorSwitch{name: name, left: "g", right: "h", data: &u.associator[5]}, nil
		case "a4", "A4":
			return &associatorSwitch{name: name, left: "h", right: "j", data: &u.associator[6]}, nil
		case "a2", "A2":
			return &associatorSwitch{name: name, left: "j", right: "k", data: &u.associator[7]}, nil
		default:
			return nil, fmt.Errorf("invalid associator switch %s", name)
		}
	case 'd', 'D':
		fmt.Sscanf(name, "d%ds%d", &d, &s)
		if !(d >= 1 && d <= 20) {
			return nil, fmt.Errorf("invalid decade %s", name)
		}
		if !(s >= 1 && s <= 6) {
			return nil, fmt.Errorf("invalid decade stage %s", name)
		}
		return &decadeSwitch{name: name, data: &u.decade[20-d].limit[s-1]}, nil
	case 'c', 'C':
		if len(name) < 2 {
			return nil, fmt.Errorf("invalid stepper %s\n", name)
		}
		s := stepperNameToIndex(name[1])
		if s == -1 {
			return nil, fmt.Errorf("invalid stepper %s\n", name)
		}
		return &clearSwitch{name: name, data: &u.stepper[s].csw}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

// SetSwitch sets the control switch name to the given value.
func (u *Mp) SetSwitch(name string, value string) error {
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

func (u *Mp) GetSwitch(name string) (string, error) {
	u.rewiring <- 1
	<-u.waitingForRewiring
	defer func() { u.rewiring <- 1 }()
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return "?", err
	}
	return sw.Get(), nil
}

// Return a bitmask of the decades associated with stepper s.
func (u *Mp) getAssociatedDecades(s int) uint {
	var ds uint
	switch s {
	case 0:
		if u.associator[0] == 'A' {
			ds |= 1 << 0
		}
	case 1:
		if u.associator[0] == 'B' {
			ds |= 1 << 0
		}
		ds |= 1 << 1
		if u.associator[1] == 'B' {
			ds |= 1 << 2
		}
	case 2:
		if u.associator[1] == 'C' {
			ds |= 1 << 2
		}
		ds |= 1 << 3
		ds |= 1 << 4
		ds |= 1 << 5
		if u.associator[2] == 'C' {
			ds |= 1 << 6
		}
	case 3:
		if u.associator[2] == 'D' {
			ds |= 1 << 6
		}
		ds |= 1 << 7
		if u.associator[3] == 'D' {
			ds |= 1 << 8
		}
	case 4:
		if u.associator[3] == 'E' {
			ds |= 1 << 8
		}
		ds |= 1 << 9
	case 5:
		if u.associator[4] == 'F' {
			ds |= 1 << 10
		}
	case 6:
		if u.associator[4] == 'G' {
			ds |= 1 << 10
		}
		ds |= 1 << 11
		if u.associator[5] == 'G' {
			ds |= 1 << 12
		}
	case 7:
		if u.associator[5] == 'H' {
			ds |= 1 << 12
		}
		ds |= 1 << 13
		ds |= 1 << 14
		ds |= 1 << 15
		if u.associator[6] == 'H' {
			ds |= 1 << 16
		}
	case 8:
		if u.associator[6] == 'J' {
			ds |= 1 << 16
		}
		ds |= 1 << 17
		if u.associator[7] == 'J' {
			ds |= 1 << 18
		}
	case 9:
		if u.associator[7] == 'K' {
			ds |= 1 << 18
		}
		ds |= 1 << 19
	}
	return ds
}

// Returns true if the decades associated with stepper s have saturated.
func (u *Mp) decadesAtLimit(s int) bool {
	ds := u.getAssociatedDecades(s)
	if ds == 0 {
		return false
	}
	stage := u.stepper[s].stage
	for i := range u.decade {
		if ds&(1<<i) != 0 && u.decade[i].val != u.decade[i].limit[stage] {
			return false
		}
	}
	return true
}

// Clears the associated decades for stepper s.
func (u *Mp) clearDecades(s int) {
	ds := u.getAssociatedDecades(s)
	for i := range u.decade {
		if ds&(1<<i) != 0 {
			u.decade[i].val = 0
			u.decade[i].carry = false
		}
	}
}

// Increments the associated decades for stepper s.
func (u *Mp) incrementDecades(s int) {
	ds := u.getAssociatedDecades(s)
	carryIndex := -1
	for i := 19; i >= 0; i-- {
		if ds&(1<<i) != 0 {
			u.decade[i].increment()
			if carryIndex != -1 {
				u.decade[carryIndex].carry = false
			}
			if !u.decade[i].carry {
				break
			}
			carryIndex = i
		}
	}
}

func (u *Mp) clock(p Pulse, resp chan int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	cyc := p.Val
	if cyc&Cpp != 0 {
		for i := range u.stepper {
			stageBeforeIncrementing := u.stepper[i].stage
			if u.decadesAtLimit(i) {
				u.clearDecades(i)
				u.stepper[i].increment()
			}
			// Unclear what this should be: probably > 3 and < 12
			if u.stepper[i].inff >= 6 {
				u.incrementDecades(i)
				u.stepper[i].inff = 0
				// This stage output could trigger another mp input, so have to release mu.
				// But don't let the main thread rewire output during this time.
				u.outputMu.Lock()
				u.mu.Unlock()
				Handshake(1, u.stepper[i].o[stageBeforeIncrementing], resp)
				u.mu.Lock()
				u.outputMu.Unlock()
			}
		}
	} else if cyc&Tenp != 0 {
		for i := range u.stepper {
			u.stepper[i].kludge = false
		}
	}
	// Simulate "flip-flop...time constant approximately equal to that
	// of the slow buffer output of a transceiver."  Huskey TM II, Ch X
	for i := range u.stepper {
		if u.stepper[i].inff > 0 {
			u.stepper[i].inff++
		}
	}
}

func (u *Mp) MakeClockFunc() ClockFunc {
	resp := make(chan int)
	return func(p Pulse) {
		u.clock(p, resp)
	}
}

func (u *Mp) Run() {
	var p Pulse

	for {
		p.Resp = nil
		select {
		case <-u.rewiring:
			u.waitingForRewiring <- 1
			<-u.rewiring
		case p = <-u.decade[0].di:
			u.mu.Lock()
			u.decade[0].increment()
			u.mu.Unlock()
		case p = <-u.decade[1].di:
			u.mu.Lock()
			u.decade[1].increment()
			u.mu.Unlock()
		case p = <-u.decade[2].di:
			u.mu.Lock()
			u.decade[2].increment()
			u.mu.Unlock()
		case p = <-u.decade[3].di:
			u.mu.Lock()
			u.decade[3].increment()
			u.mu.Unlock()
		case p = <-u.decade[4].di:
			u.mu.Lock()
			u.decade[4].increment()
			u.mu.Unlock()
		case p = <-u.decade[5].di:
			u.mu.Lock()
			u.decade[5].increment()
			u.mu.Unlock()
		case p = <-u.decade[6].di:
			u.mu.Lock()
			u.decade[6].increment()
			u.mu.Unlock()
		case p = <-u.decade[7].di:
			u.mu.Lock()
			u.decade[7].increment()
			u.mu.Unlock()
		case p = <-u.decade[8].di:
			u.mu.Lock()
			u.decade[8].increment()
			u.mu.Unlock()
		case p = <-u.decade[9].di:
			u.mu.Lock()
			u.decade[9].increment()
			u.mu.Unlock()
		case p = <-u.decade[10].di:
			u.mu.Lock()
			u.decade[10].increment()
			u.mu.Unlock()
		case p = <-u.decade[11].di:
			u.mu.Lock()
			u.decade[11].increment()
			u.mu.Unlock()
		case p = <-u.decade[12].di:
			u.mu.Lock()
			u.decade[12].increment()
			u.mu.Unlock()
		case p = <-u.decade[13].di:
			u.mu.Lock()
			u.decade[13].increment()
			u.mu.Unlock()
		case p = <-u.decade[14].di:
			u.mu.Lock()
			u.decade[14].increment()
			u.mu.Unlock()
		case p = <-u.decade[15].di:
			u.mu.Lock()
			u.decade[15].increment()
			u.mu.Unlock()
		case p = <-u.decade[16].di:
			u.mu.Lock()
			u.decade[16].increment()
			u.mu.Unlock()
		case p = <-u.decade[17].di:
			u.mu.Lock()
			u.decade[17].increment()
			u.mu.Unlock()
		case p = <-u.decade[18].di:
			u.mu.Lock()
			u.decade[18].increment()
			u.mu.Unlock()
		case p = <-u.decade[19].di:
			u.mu.Lock()
			u.decade[19].increment()
			u.mu.Unlock()
		case p = <-u.stepper[0].di:
			u.mu.Lock()
			u.stepper[0].increment()
			u.mu.Unlock()
		case p = <-u.stepper[0].i:
			u.mu.Lock()
			u.stepper[0].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[0].cdi:
			u.mu.Lock()
			u.stepper[0].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[1].di:
			u.mu.Lock()
			u.stepper[1].increment()
			u.mu.Unlock()
		case p = <-u.stepper[1].i:
			u.mu.Lock()
			u.stepper[1].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[1].cdi:
			u.mu.Lock()
			u.stepper[1].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[2].di:
			u.mu.Lock()
			u.stepper[2].increment()
			u.mu.Unlock()
		case p = <-u.stepper[2].i:
			u.mu.Lock()
			u.stepper[2].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[2].cdi:
			u.mu.Lock()
			u.stepper[2].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[3].di:
			u.mu.Lock()
			u.stepper[3].increment()
			u.mu.Unlock()
		case p = <-u.stepper[3].i:
			u.mu.Lock()
			u.stepper[3].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[3].cdi:
			u.mu.Lock()
			u.stepper[3].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[4].di:
			u.mu.Lock()
			u.stepper[4].increment()
			u.mu.Unlock()
		case p = <-u.stepper[4].i:
			u.mu.Lock()
			u.stepper[4].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[4].cdi:
			u.mu.Lock()
			u.stepper[4].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[5].di:
			u.mu.Lock()
			u.stepper[5].increment()
			u.mu.Unlock()
		case p = <-u.stepper[5].i:
			u.mu.Lock()
			u.stepper[5].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[5].cdi:
			u.mu.Lock()
			u.stepper[5].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[6].di:
			u.mu.Lock()
			u.stepper[6].increment()
			u.mu.Unlock()
		case p = <-u.stepper[6].i:
			u.mu.Lock()
			u.stepper[6].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[6].cdi:
			u.mu.Lock()
			u.stepper[6].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[7].di:
			u.mu.Lock()
			u.stepper[7].increment()
			u.mu.Unlock()
		case p = <-u.stepper[7].i:
			u.mu.Lock()
			u.stepper[7].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[7].cdi:
			u.mu.Lock()
			u.stepper[7].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[8].di:
			u.mu.Lock()
			u.stepper[8].increment()
			u.mu.Unlock()
		case p = <-u.stepper[8].i:
			u.mu.Lock()
			u.stepper[8].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[8].cdi:
			u.mu.Lock()
			u.stepper[8].stage = 0
			u.mu.Unlock()
		case p = <-u.stepper[9].di:
			u.mu.Lock()
			u.stepper[9].increment()
			u.mu.Unlock()
		case p = <-u.stepper[9].i:
			u.mu.Lock()
			u.stepper[9].inff = 1
			u.mu.Unlock()
		case p = <-u.stepper[9].cdi:
			u.mu.Lock()
			u.stepper[9].stage = 0
			u.mu.Unlock()
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}

func stepperNameToIndex(s byte) int {
	return strings.IndexByte("ABCDEFGHJK", s)
}
