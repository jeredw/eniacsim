package units

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/jeredw/eniacsim/lib"
)

// Simulates the ENIAC printer.
type Printer struct {
	Io PrinterConn

	printing       [16]bool        // Should the field print
	coupling       [16]bool        // Treat the field as part of the next one
	magnets        [80]punchWiring // Plugboard wiring for punch magnets
	usingPlugboard bool            // True if any plugboard switch set
}

type punchWiring struct {
	wasSet     bool // if true, switch was set
	nc         bool // if true, not connected
	fixed      int  // 0: don't punch, 1-13 mean punch row 0-12
	signGroup  int  // 0: no sign, 1-16 indicate sign group
	digitGroup int  // 0: no digit, 1-16 indicate digit group
	digitIndex int  // 0: no digit, 1-5 which digit of group
}

// Connections to printer.
type PrinterConn struct {
	MpPrinterDecades func() string
	Accumulator      [20]StaticWiring
}

// Reset values for coupling switches - by default group digits from the same
// origin unit.
var couplingDefaults = [16]bool{
	false, true, false, true, false, false, true, false,
	true, false, true, false, true, false, true, false,
}

func NewPrinter() *Printer {
	u := &Printer{
		coupling: couplingDefaults,
	}
	for i := 0; i < 80; i++ {
		u.magnets[i] = punchWiring{nc: true}
	}
	return u
}

func (u *Printer) Reset() {
	for i := range u.printing {
		u.printing[i] = false
	}
	for i := range u.coupling {
		u.coupling[i] = couplingDefaults[i]
	}
	for i := 0; i < 80; i++ {
		u.magnets[i] = punchWiring{nc: true}
	}
}

// Print an 80-column punched card from groups of 5-digit fields.
//
// Groups are converted from signed tens' complement to signed magnitude.
func (u *Printer) Print() string {
	mpd := u.Io.MpPrinterDecades()
	a13 := string(u.Io.Accumulator[13-1].Value())
	a14 := string(u.Io.Accumulator[14-1].Value())
	a15 := string(u.Io.Accumulator[15-1].Value())
	a16 := string(u.Io.Accumulator[16-1].Value())
	a17 := string(u.Io.Accumulator[17-1].Value())
	a18 := string(u.Io.Accumulator[18-1].Value())
	a19 := string(u.Io.Accumulator[19-1].Value())
	a20 := string(u.Io.Accumulator[20-1].Value())

	// digits will contain 80 digits (5 MP decades, a13, a14, a15[lo], a16-a20)
	// and signs will contain 16 signs for the corresponding 5-digit fields.
	digits := strings.Join([]string{
		mpd, a13[2:12], a14[2:12], a15[7:12], a16[2:12],
		a17[2:12], a18[2:12], a19[2:12], a20[2:12],
	}, "")
	signs := []byte{
		'P', a13[0], a13[0], a14[0], a14[0], a15[0], a16[0], a16[0],
		a17[0], a17[0], a18[0], a18[0], a19[0], a19[0], a20[0], a20[0],
	}

	// Using plugboard if any magnet switch is set
	if !u.usingPlugboard {
		for _, p := range u.magnets {
			if p.wasSet {
				u.usingPlugboard = true
				break
			}
		}
	}

	// Group digit fields and convert to IBM card format.
	ibmDigits := ""
	groupStart := 0
	groupEnd := 0
	for i := 0; i < 16; i++ {
		if !u.coupling[i] || i == 15 {
			groupEnd = (i + 1) * 5
			if u.usingPlugboard {
				ibmDigits += TensComplementToIBMCard(signs[i], digits[groupStart:groupEnd])
			} else {
				ibmDigits += TensComplementToIBMCardDigits(signs[i], digits[groupStart:groupEnd])
			}
			groupStart = groupEnd
		}
	}
	card := ""
	if !u.usingPlugboard {
		for i := 0; i < 16; i++ {
			if !u.printing[i] {
				card += "     "
			} else {
				card += ibmDigits[5*i : 5*(i+1)]
			}
		}
	} else {
		for _, p := range u.magnets {
			if p.nc {
				card += " "
			} else if p.fixed > 0 {
				card += string("0123456789-&"[p.fixed-1])
			} else {
				digit := ibmDigits[5*(p.digitGroup-1)+(p.digitIndex-1)]
				if p.signGroup > 0 && signs[p.signGroup-1] == 'M' {
					if digit == '0' {
						digit = '-'
					} else {
						digit = 'I' + (digit - '0')
					}
				}
				card += string(digit)
			}
		}
	}
	return card
}

type punchMagnetSwitch struct {
	Name string
	Data *punchWiring
}

func (s *punchMagnetSwitch) Get() string {
	if s.Data.nc {
		return "nc"
	}
	if s.Data.fixed > 0 {
		return fmt.Sprintf("%d", s.Data.fixed-1)
	}
	f := make([]string, 0, 3)
	f = append(f, fmt.Sprintf("%d", s.Data.digitGroup))
	f = append(f, fmt.Sprintf("%d", s.Data.digitIndex))
	if s.Data.signGroup > 0 {
		f = append(f, fmt.Sprintf("m%d", s.Data.signGroup))
	}
	return strings.Join(f, ",")
}

func (s *punchMagnetSwitch) Set(value string) error {
	if value == "nc" {
		*s.Data = punchWiring{wasSet: true, nc: true}
		return nil
	}

	if !strings.ContainsRune(value, ',') {
		if fixed, err := strconv.Atoi(value); err == nil {
			if !(fixed >= 0 && fixed <= 12) {
				return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
			}
			*s.Data = punchWiring{wasSet: true, fixed: fixed + 1}
			return nil
		}
		return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
	}

	f := strings.Split(value, ",")
	if len(f) < 2 || len(f) > 3 {
		return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
	}
	wiring := punchWiring{wasSet: true}

	group, err := strconv.Atoi(f[0])
	if err != nil || !(group >= 1 && group <= 16) {
		return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
	}
	index, err := strconv.Atoi(f[1])
	if err != nil || !(index >= 1 && index <= 5) {
		return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
	}
	wiring.digitGroup = group
	wiring.digitIndex = index

	if len(f) == 3 {
		if len(f[2]) < 2 || f[2][0] != 'm' {
			return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
		}
		signGroup, err := strconv.Atoi(f[2][1:])
		if err != nil || !(signGroup >= 1 && signGroup <= 16) {
			return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
		}
		wiring.signGroup = signGroup
	}

	*s.Data = wiring
	return nil
}

func (u *Printer) FindSwitch(name string) (Switch, error) {
	if len(name) > 2 && name[:2] == "pm" {
		col, _ := strconv.Atoi(name[2:])
		if !(col >= 1 && col <= 80) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &punchMagnetSwitch{name, &u.magnets[col-1]}, nil
	}
	if !strings.ContainsRune(name, '-') {
		field, _ := strconv.Atoi(name)
		if !(field >= 1 && field <= 16) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &BoolSwitch{name, &u.printing[field-1], printSettings()}, nil
	}

	f := strings.Split(name, "-")
	if len(f) != 2 {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	field1, _ := strconv.Atoi(f[0])
	field2, _ := strconv.Atoi(f[1])
	if !(field1 >= 1 && field1 <= 16) {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	if field1 == 16 && field2 == 1 {
		return nil, fmt.Errorf("16-1 switch is not implemented")
	}
	if field2 != field1+1 {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	return &BoolSwitch{name, &u.coupling[field1-1], couplingSettings()}, nil
}
