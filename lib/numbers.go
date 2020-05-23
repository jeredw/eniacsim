package lib

import (
	"unicode"
)

// Digits represented as combinations of 1P/2P/2P'/4P.
var BCD = []Pulse{
	0,                           // 0
	Onep,                        // 1
	Twop,                        // 2
	Onep | Twop,                 // 3
	Fourp,                       // 4
	Onep | Fourp,                // 5
	Twop | Fourp,                // 6
	Onep | Twop | Fourp,         // 7
	Twop | Twopp | Fourp,        // 8
	Onep | Twop | Twopp | Fourp, // 9
}

func TenDigitsToInt64BCD(digits [10]int) int64 {
	var n int64
	for i := range digits {
		n = (n << 4) + int64(digits[i])
	}
	return n
}

func DigitsToInt64BCD(digits []int) int64 {
	var n int64
	for i := range digits {
		n = (n << 4) + int64(digits[i])
	}
	return n
}

func BoolToInt64(b bool) int64 {
	if b {
		return int64(1)
	}
	return int64(0)
}

func ToBin(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// ENIAC's IBM card punch and reader used 80 column cards.  This program
// represents cards as ASCII strings where characters corresponding to IBM code
// points indicate what rows are punched.
//
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
//
// ENIAC stores negative numbers using ten's complement but cards use signed
// magnitude.  Digits on cards are represented as a punch in the corresponding
// row.  Negative numbers are indicated by an 11 punch in any digit position,
// e.g. the first digit.

// TensComplementToIBMCard converts a sign and ten's complement digits to a
// signed magnitude number in IBM card code.
func TensComplementToIBMCard(sign byte, digits string) string {
	if sign == 'P' {
		return digits
	}

	// Compute magnitude of digits and store the accompanying sign with an
	// 11-punch in the leftmost digit (see diagram).
	nz := findFirstNonZero(digits)
	if nz < 0 {
		// negative 0 is still 0
		return digits
	}
	if nz == 0 {
		// special case for 10's comp and 11-punch
		// -[123456789]000000... -> [RQPONMLKJ]000000...
		return string('J'+'9'-digits[0]) + digits[1:]
	}
	// nz > 0
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

// IBMCardToNinesComplement converts a signed magnitude IBM card field to nines
// complement digits.  Tens' complement correction is added back in the
// constant transmitter unit.
func IBMCardToNinesComplement(field string) (sign bool, digits []int) {
	sign = false
	for _, c := range field {
		if c == '-' || c == ']' || c == '}' || c >= 'J' && c <= 'R' {
			sign = true
			break
		}
	}
	numDigits := len(field)
	digits = make([]int, numDigits)
	if sign == false {
		for i, c := range field {
			if unicode.IsDigit(c) {
				digits[i] = runeToDigit(c)
			}
		}
		return
	}
	for i := 0; i < numDigits; i++ {
		digits[i] = 9 - runeToDigit(rune(field[i]))
	}
	return
}

func findFirstNonZero(digits string) int {
	for i := len(digits) - 1; i >= 0; i-- {
		if digits[i] != '0' {
			return i
		}
	}
	return -1
}

func runeToDigit(c rune) int {
	if c == '-' || c == ']' || c == '}' {
		return 0
	}
	if c >= 'J' && c <= 'R' {
		return int(c - 'J' + 1)
	}
	return int(c - '0')
}
