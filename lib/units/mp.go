package units

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	. "github.com/jeredw/eniacsim/lib"
)

// Mp simulates the ENIAC master programmer unit.
type Mp struct {
	stepper    [10]mpStepper // Steppers (A-K)
	decade     [20]mpDecade  // Decade counters (#20 down to #1)
	associator [8]byte       // Stepper to decade associations
	unplugDecades bool       // Disassociate all decades from steppers
}

type mpStepper struct {
	stage           int // Stage counter (0..5)
	di, i, cdi      *Jack
	o               [6]*Jack
	csw             int
	inff            int
	waitForNextTenp bool
}

func (s *mpStepper) increment() {
	if s.waitForNextTenp {
		// Only count once per 10P (ignore 9Ps and non-digit pulses).
		return
	}
	if s.stage >= s.csw {
		s.stage = 0
	} else {
		s.stage++
	}
	s.waitForNextTenp = true
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
			// Increment decade and carry into any associated, more significant decades.
			ds := u.getAssociatedDecadesForDecade(d)
			for i := d; i >= 0; i-- {
				if ds&(1<<i) != 0 {
					u.decade[i].increment()
					if !u.decade[i].carry {
						break
					}
					u.decade[i].carry = false
				}
			}
		}
	}
	for i := 0; i < 20; i++ {
		u.decade[i].di = NewInput(fmt.Sprintf("p.%ddi", i+1), decadeIncrement(i))
	}
	stepperIncrement := func(s int) JackHandler {
		return func(*Jack, int) {
			u.stepper[s].increment()
		}
	}
	stepperInput := func(s int) JackHandler {
		return func(*Jack, int) {
			u.stepper[s].inff = 1
		}
	}
	stepperClear := func(s int) JackHandler {
		return func(*Jack, int) {
			u.stepper[s].stage = 0
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
	for i := range u.decade {
		for j := range u.decade[i].limit {
			u.decade[i].limit[j] = 0
		}
	}
	for i := range u.stepper {
		u.stepper[i].csw = 0
		u.stepper[i].waitForNextTenp = false
	}
	for i := range u.associator {
		u.associator[i] = associatorResets[i]
	}
	u.Clear()
}

// Clear resets decades and stepper stage counters.
func (u *Mp) Clear() {
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

func (u *Mp) FindSwitch(name string) (Switch, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("invalid switch")
	}
	if name == "gate63" {
		return &BoolSwitch{name, &u.unplugDecades, unplugDecadesSettings()}, nil
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

// Return a bitmask of the decades associated with stepper s.
func (u *Mp) getAssociatedDecadesForStepper(s int) uint {
	if u.unplugDecades {
		return 0
	}
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

// Returns a bitmask of associated decades which include decade d.
func (u *Mp) getAssociatedDecadesForDecade(d int) uint {
	for s := range u.stepper {
		ds := u.getAssociatedDecadesForStepper(s)
		if ds&(1<<d) != 0 {
			return ds
		}
	}
	return 0
}

// Returns true if the decades associated with stepper s have saturated.
func (u *Mp) decadesAtLimit(s int) bool {
	ds := u.getAssociatedDecadesForStepper(s)
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
	ds := u.getAssociatedDecadesForStepper(s)
	for i := range u.decade {
		if ds&(1<<i) != 0 {
			u.decade[i].val = 0
			u.decade[i].carry = false
		}
	}
}

// Increments the associated decades for stepper s.
func (u *Mp) incrementDecades(s int) {
	ds := u.getAssociatedDecadesForStepper(s)
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

func (u *Mp) Clock(cyc Pulse) {
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
				u.stepper[i].o[stageBeforeIncrementing].Transmit(1)
			}
		}
	} else if cyc&Tenp != 0 {
		for i := range u.stepper {
			u.stepper[i].waitForNextTenp = false
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

func stepperNameToIndex(s byte) int {
	return strings.IndexByte("ABCDEFGHJK", s)
}

func stepperIndexToName(i int) byte {
	return "ABCDEFGHJK"[i]
}
