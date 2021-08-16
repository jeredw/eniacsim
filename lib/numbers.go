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
	for i := 9; i >= 0; i-- {
		n = (n << 4) + int64(digits[i])
	}
	return n
}

func DigitsToInt64BCD(digits []int) int64 {
	var n int64
	for i := len(digits) - 1; i >= 0; i-- {
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

func StringToSignAndDigits(s string) (sign bool, digits []int) {
	if len(s) < 3 {
		return false, []int{}
	}
	sign = false
	if s[0] == 'M' {
		sign = true
	}
	numDigits := len(s) - 2
	digits = make([]int, numDigits)
	for i := 0; i < numDigits; i++ {
		digits[i] = int(s[2+(numDigits-1-i)] - '0')
	}
	return
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
// signed magnitude number in IBM card code with the sign indicated in the
// leftmost digit position.
func TensComplementToIBMCard(sign byte, digits string) string {
	if sign == 'P' {
		return digits
	}

	// Sign is M so have to convert from 10's complement to signed magnitude.
	nz := findRightmostNonZero(digits)
	if nz < 0 {
		// M00000...=-10^k requires an extra digit in signed magnitude
		// The conversion process described in ENIAC Technical Manual IX-12
		// suggests M0 would punch as 11-0, which we represent as '-'.
		return string('-') + digits[1:]
	}
	if nz == 0 {
		// Subtract leading digit from 10 i.e. M9 -> 11-1 (J)
		// -[123456789]000000... -> [RQPONMLKJ]000000...
		return string('J'+'9'-digits[0]) + digits[1:]
	}
	// nz > 0
	// Subtract leading digit from 9 and indicate sign.
	sc := string('I' + '9' - digits[0])
	if sc == "I" {
		sc = "-"
	}
	// Use 9s complement of digits up to rightmost nz.
	// 10^k - n = ((10^k-1) - n) + 1
	for i := 1; i < nz; i++ {
		sc += string('0' + '9' - digits[i])
	}
	// Subtract rightmost nz digit from 10 instead of 9.
	sc += string('1' + '9' - digits[nz])
	// Preserve digits to right of rightmost nz.
	sc += digits[nz+1:]
	return sc
}

// TensComplementToIBMCardDigits converts a sign and ten's complement digits
// to a signed magnitude number in IBM card code, returning just the digits.
func TensComplementToIBMCardDigits(sign byte, digits string) string {
	if sign == 'P' {
		return digits
	}

	// Sign is M so have to convert from 10's complement to signed magnitude.
	nz := findRightmostNonZero(digits)
	if nz < 0 {
		// M00000...=-10^k will punch the first 0
		return digits
	}
	if nz == 0 {
		// Subtract leading digit from 10 i.e. M9 -> 11-1 (J)
		// -[123456789]000000... -> [987654321]000000...
		return string('1'+'9'-digits[0]) + digits[1:]
	}
	sc := ""
	// nz > 0 (first digit is not rightmost zero)
	// Use 9s complement of digits up to rightmost nz.
	// 10^k - n = ((10^k-1) - n) + 1
	for i := 0; i < nz; i++ {
		sc += string('0' + '9' - digits[i])
	}
	// Subtract rightmost nz digit from 10 instead of 9.
	sc += string('1' + '9' - digits[nz])
	// Preserve digits to right of rightmost nz.
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
				digits[numDigits-1-i] = runeToDigit(c)
			}
		}
		return
	}
	for i := 0; i < numDigits; i++ {
		digits[numDigits-1-i] = 9 - runeToDigit(rune(field[i]))
	}
	return
}

func findRightmostNonZero(digits string) int {
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
