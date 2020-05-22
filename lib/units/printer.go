package units

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
	"strings"
	"sync"
)

// Simulates the ENIAC printer.
type Printer struct {
	Io PrinterConn

	printing [16]bool // Should the field print
	coupling [16]bool // Treat the field as part of the next one

	mu sync.Mutex
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
	return &Printer{
		coupling: couplingDefaults,
	}
}

func (u *Printer) Reset() {
	for i := range u.printing {
		u.printing[i] = false
	}
	for i := range u.coupling {
		u.coupling[i] = couplingDefaults[i]
	}
}

// Print an 80-column punched card from groups of 5-digit fields.
//
// Groups are converted from signed magnitude to signed tens' complement for 80
// col/12 row IBM cards.
func (u *Printer) Print() string {
	u.mu.Lock()
	defer u.mu.Unlock()

	mpd := u.Io.MpPrinterDecades()
	a13 := u.Io.Accumulator[13-1].Value()
	a14 := u.Io.Accumulator[14-1].Value()
	a15 := u.Io.Accumulator[15-1].Value()
	a16 := u.Io.Accumulator[16-1].Value()
	a17 := u.Io.Accumulator[17-1].Value()
	a18 := u.Io.Accumulator[18-1].Value()
	a19 := u.Io.Accumulator[19-1].Value()
	a20 := u.Io.Accumulator[20-1].Value()

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

	// Group digit fields and convert to IBM card format.
	ibmDigits := ""
	groupStart := 0
	groupEnd := 0
	for i := 0; i < 16; i++ {
		if !u.coupling[i] || i == 15 {
			groupEnd = (i + 1) * 5
			ibmDigits += ToIBMCard(signs[i], digits[groupStart:groupEnd])
			groupStart = groupEnd
		}
	}
	card := ""
	for i := 0; i < 16; i++ {
		if !u.printing[i] {
			card += "     "
		} else {
			card += ibmDigits[5*i : 5*(i+1)]
		}
	}
	return card
}

func (u *Printer) FindSwitch(name string) (Switch, error) {
	if !strings.ContainsRune(name, '-') {
		field, _ := strconv.Atoi(name)
		if !(field >= 1 && field <= 16) {
			return nil, fmt.Errorf("invalid switch %s", name)
		}
		return &BoolSwitch{&u.mu, name, &u.printing[field-1], printSettings()}, nil
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
	return &BoolSwitch{&u.mu, name, &u.coupling[field1-1], couplingSettings()}, nil
}
