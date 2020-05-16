package units

import (
	"encoding/json"
	"fmt"
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

	mu sync.Mutex
}

type mpStepper struct {
	stage      int // Stage counter (0..5)
	di, i, cdi *Jack
	o          [6]*Jack
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
	val   int    // Counter (one digit)
	carry bool   // Carry out from counter
	di    *Jack  // Advance counter
	limit [6]int // Per-stage max value for counter
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
	u := &Mp{associator: associatorResets}
	decadeIncrement := func(d int) JackHandler {
		return func(*Jack, int) {
			u.mu.Lock()
			u.decade[d].increment()
			u.mu.Unlock()
		}
	}
	for i := 0; i < 20; i++ {
		u.decade[i].di = NewInput(fmt.Sprintf("p.%ddi", i+1), decadeIncrement(i))
	}
	stepperIncrement := func(s int) JackHandler {
		return func(*Jack, int) {
			u.mu.Lock()
			u.stepper[s].increment()
			u.mu.Unlock()
		}
	}
	stepperInput := func(s int) JackHandler {
		return func(*Jack, int) {
			u.mu.Lock()
			u.stepper[s].inff = 1
			u.mu.Unlock()
		}
	}
	stepperClear := func(s int) JackHandler {
		return func(*Jack, int) {
			u.mu.Lock()
			u.stepper[s].stage = 0
			u.mu.Unlock()
		}
	}
	for i := 0; i < 10; i++ {
		stepper := stepperIndexToName(i)
		u.stepper[i].di = NewInput(fmt.Sprintf("p.%cdi", stepper), stepperIncrement(i))
		u.stepper[i].i = NewInput(fmt.Sprintf("p.%ci", stepper), stepperInput(i))
		u.stepper[i].cdi = NewInput(fmt.Sprintf("p.%ccdi", stepper), stepperClear(i))
		for j := 0; j < 6; j++ {
			u.stepper[i].o[j] = NewOutput(fmt.Sprintf("p.%c%do", stepper, j+1), nil)
		}
	}
	return u
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
	u.mu.Lock()
	for i := range u.decade {
		u.decade[i].di.Disconnect()
		for j := range u.decade[i].limit {
			u.decade[i].limit[j] = 0
		}
	}
	for i := range u.stepper {
		u.stepper[i].di.Disconnect()
		u.stepper[i].i.Disconnect()
		u.stepper[i].cdi.Disconnect()
		for j := range u.stepper[i].o {
			u.stepper[i].o[j].Disconnect()
		}
		u.stepper[i].csw = 0
		u.stepper[i].kludge = false
	}
	for i := range u.associator {
		u.associator[i] = associatorResets[i]
	}
	u.mu.Unlock()
	u.Clear()
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

func (u *Mp) FindJack(jack string) (*Jack, error) {
	if len(jack) == 0 {
		return nil, fmt.Errorf("invalid jack")
	}
	var n int
	if unicode.IsDigit(rune(jack[0])) {
		fmt.Sscanf(jack, "%ddi", &n)
		if !(n >= 1 && n <= 20) {
			return nil, fmt.Errorf("invalid decade %s", jack)
		}
		return u.decade[20-n].di, nil
	}

	s := stepperNameToIndex(jack[0])
	if s == -1 {
		return nil, fmt.Errorf("invalid stepper %s", jack)
	}
	if len(jack) < 2 {
		return nil, fmt.Errorf("invalid stepper input %s", jack)
	}
	switch jack[1:] {
	case "di":
		return u.stepper[s].di, nil
	case "i":
		return u.stepper[s].i, nil
	case "cdi":
		return u.stepper[s].cdi, nil
	}
	if len(jack) < 3 {
		return nil, fmt.Errorf("invalid jack %s", jack)
	}
	fmt.Sscanf(jack[1:], "%do", &n)
	if !(n >= 1 && n <= 6) {
		return nil, fmt.Errorf("invalid output %s", jack)
	}
	return u.stepper[s].o[n-1], nil
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

func (u *Mp) lookupSwitch(name string) (Switch, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("invalid switch")
	}
	var d, s int
	switch name[0] {
	case 'a', 'A':
		switch name {
		case "a20", "A20":
			return &associatorSwitch{name, "a", "b", &u.associator[0]}, nil
		case "a18", "A18":
			return &associatorSwitch{name, "b", "c", &u.associator[1]}, nil
		case "a14", "A14":
			return &associatorSwitch{name, "c", "d", &u.associator[2]}, nil
		case "a12", "A12":
			return &associatorSwitch{name, "d", "e", &u.associator[3]}, nil
		case "a10", "A10":
			return &associatorSwitch{name, "f", "g", &u.associator[4]}, nil
		case "a8", "A8":
			return &associatorSwitch{name, "g", "h", &u.associator[5]}, nil
		case "a4", "A4":
			return &associatorSwitch{name, "h", "j", &u.associator[6]}, nil
		case "a2", "A2":
			return &associatorSwitch{name, "j", "k", &u.associator[7]}, nil
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
		return &IntSwitch{name, &u.decade[20-d].limit[s-1], mpDecadeSettings()}, nil
	case 'c', 'C':
		if len(name) < 2 {
			return nil, fmt.Errorf("invalid stepper %s\n", name)
		}
		s := stepperNameToIndex(name[1])
		if s == -1 {
			return nil, fmt.Errorf("invalid stepper %s\n", name)
		}
		return &IntSwitch{name, &u.stepper[s].csw, mpClearSettings()}, nil
	}
	return nil, fmt.Errorf("invalid switch %s", name)
}

// SetSwitch sets the control switch name to the given value.
func (u *Mp) SetSwitch(name string, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return err
	}
	return sw.Set(value)
}

func (u *Mp) GetSwitch(name string) (string, error) {
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

func (u *Mp) clock(cyc Pulse) {
	u.mu.Lock()
	defer u.mu.Unlock()
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
				// Live with a potential data race from rewiring stepper here.
				u.mu.Unlock()
				u.stepper[i].o[stageBeforeIncrementing].Transmit(1)
				u.mu.Lock()
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
	return func(p Pulse) {
		u.clock(p)
	}
}

func stepperNameToIndex(s byte) int {
	return strings.IndexByte("ABCDEFGHJK", s)
}

func stepperIndexToName(i int) byte {
	return "ABCDEFGHJK"[i]
}
