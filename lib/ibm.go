package lib

import (
  "unicode"
)

// ENIAC had a printer and card reader that used 80-column/12 row IBM cards.
// In this program cards are represented as ASCII strings that indicate what
// holes are punched in correspondence with IBM code points.
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
// In ENIAC cards, digits are represented as themselves, and negative values
// are indicated by an 11 punch in any column associated with a number.

// ToIBMCard converts ENIAC signed digits to IBM punched card encoding.
func ToIBMCard(sign byte, digits string) string {
	if sign == 'P' {
		return digits
	}

  // The number is negative, so convert it to tens' complement with an 11-punch
  // in the leftmost digit (see the diagram above).  In ASCII, a digit with an
  // 11-punch is 'J' + (1 through 9).
	var nz int // rightmost nonzero digit
	for nz = len(digits) - 1; nz >= 0 && digits[nz] == '0'; nz-- {
	}
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

// FromIBMCard converts an IBM card field to a sign and digits.
func FromIBMCard(field string) (sign bool, digits []int) {
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
		for j, c := range field {
			if unicode.IsDigit(c) {
        digits[(numDigits-1)-j] = runeToDigit(c)
			}
    }
		return
	}
	var nz int // rightmost nonzero digit
	for nz = numDigits - 1; nz >= 0 && field[nz] == '0'; nz-- {
	}
	for ; nz >= 0; nz-- {
		digits[(numDigits-1)-nz] = 9 - runeToDigit(rune(field[nz]))
	}
  return
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
