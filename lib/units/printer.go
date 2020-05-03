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
	AccValue         [20]func() string
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
// col/12 row IBM cards.  Negative values are indicated by an 11 punch in any
// column associated with a group.
//     ______________________________________________
//    /&-0123456789ABCDEFGHIJKLMNOPQR/STUVWXYZ
//12|  x           xxxxxxxxx
//11|   x                   xxxxxxxxx
// 0|    x                           xxxxxxxxx
// 1|     x        x        x        x
// 2|      x        x        x        x
// 3|       x        x        x        x
// 4|        x        x        x        x
// 5|         x        x        x        x
// 6|          x        x        x        x
// 7|           x        x        x        x
// 8|            x        x        x        x
// 9|             x        x        x        x
//  |________________________________________________
func (u *Printer) Print() string {
	u.mu.Lock()
	defer u.mu.Unlock()

	mpd := u.Io.MpPrinterDecades()
	a13 := u.Io.AccValue[13-1]()
	a14 := u.Io.AccValue[14-1]()
	a15 := u.Io.AccValue[15-1]()
	a16 := u.Io.AccValue[16-1]()
	a17 := u.Io.AccValue[17-1]()
	a18 := u.Io.AccValue[18-1]()
	a19 := u.Io.AccValue[19-1]()
	a20 := u.Io.AccValue[20-1]()

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
			ibmDigits += toIBMCard(signs[i], digits[groupStart:groupEnd])
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

func toIBMCard(sign byte, digits string) string {
	if sign == 'P' {
		return digits
	}

	// The number is negative, so convert it to tens' complement with an 11-punch
	// in the leftmost digit (see the diagram for Print).  In ASCII, a digit with
	// an 11-punch is 'J' + (1 through 9).
	var nz int // rightmost nonzero digit
	for nz = len(digits) - 1; nz >= 0 && digits[nz] == '0'; nz-- {
	}
	if nz < 0 {
		// negative 0 is still 0
		return digits
	} else if nz == 0 {
		// special case for 10's comp and 11-punch
		// -[123456789]000000... -> [RQPONMLKJ]000000...
		return string('J'+'9'-digits[0]) + digits[1:]
	} else {
		sc := string('J' + '9' - digits[0] - 1)
		if sc == "I" {
			// 0 + 11-punch is an illegal encoding, so just use "-" which is 11-punch
			// on its own.
			sc = "-"
		}
		// 10^k - n = ((10^k-1) - n) + 1
		for i := 1; i < nz; i++ {
			sc += string('0' + '9' - digits[i])
		}
		sc += string('0' + '9' - digits[nz] + 1)
		sc += digits[nz+1:]
		return sc
	}
}

func (u *Printer) lookupSwitch(name string) (Switch, error) {
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
		return nil, fmt.Errorf("16-1 switch is not implemented", name)
	}
	if field2 != field1+1 {
		return nil, fmt.Errorf("invalid switch %s", name)
	}
	return &BoolSwitch{name, &u.coupling[field1-1], couplingSettings()}, nil
}

func (u *Printer) SetSwitch(name, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return err
	}
	return sw.Set(value)
}

func (u *Printer) GetSwitch(name string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	sw, err := u.lookupSwitch(name)
	if err != nil {
		return "", err
	}
	return sw.Get(), nil
}
