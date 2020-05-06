package main

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
	"strings"
)

type Adapters struct {
	dp      [40]digitProgram
	shift   [40]shifter
	del     [40]deleter
	sd      [40]specialDigit
	permute [40]permuter
}

func NewAdapters() *Adapters {
	return &Adapters{}
}

func (a *Adapters) Reset() {
	for i := 0; i < 40; i++ {
		a.dp[i].in = Wire{}
		for j := 0; j < 11; j++ {
			a.dp[i].out[j] = Wire{}
		}
		a.shift[i].in = Wire{}
		a.shift[i].out = Wire{}
		a.del[i].in = Wire{}
		a.del[i].out = Wire{}
		a.sd[i].in = Wire{}
		a.sd[i].out = Wire{}
	}
}

func (a *Adapters) Switch(ilk, id string, param string) error {
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return fmt.Errorf("invalid id %s", id)
	}
	switch ilk {
	case "dp":
		return fmt.Errorf("digit adapters always specify param in p")
	case "s":
		amount, _ := strconv.Atoi(param)
		if !(amount >= -10 && amount <= 10) {
			return fmt.Errorf("invalid shift amount %s", param)
		}
		a.shift[i].amount = amount
	case "d":
		digit, _ := strconv.Atoi(param)
		if !(digit >= -10 && digit <= 10) {
			return fmt.Errorf("invalid digit %s", param)
		}
		a.del[i].digit = digit
	case "sd":
		digit, _ := strconv.Atoi(param)
		if !(digit >= -10 && digit <= 10) {
			return fmt.Errorf("invalid digit %s", param)
		}
		a.sd[i].digit = uint(digit)
	case "permute":
		order := strings.Split(param, ",")
		if len(order) != 11 {
			return fmt.Errorf("ad.permute usage: ad.permute.1.11,10,9,8,7,6,5,4,3,2,1")
		}
		for j := range order {
			pos, _ := strconv.Atoi(order[j])
			if !(pos >= 0 && pos <= 11) {
				return fmt.Errorf("invalid digit field in permutation")
			}
			a.permute[i].order[j] = pos
		}
	default:
		return fmt.Errorf("invalid type %s", ilk)
	}
	return nil
}

func (a *Adapters) Plug(ilk, id, param string, wire Wire, output bool) error {
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return fmt.Errorf("invalid id %s", id)
	}
	if len(param) > 0 {
		adapters.Switch(ilk, id, param)
	}

	switch ilk {
	case "dp":
		if output {
			// wire is a digit channel such as a data trunk
			a.dp[i].in = wire
			go a.dp[i].run()
		} else {
			// wire is a control channel such as a program trunk
			if len(param) == 0 {
				return fmt.Errorf("p ad.dp.<id> always requires param")
			}
			digit, _ := strconv.Atoi(param)
			if !(digit >= 1 && digit <= 11) {
				return fmt.Errorf("invalid digit %s", param)
			}
			a.dp[i].out[digit-1] = wire
		}
	case "s":
		if output {
			a.shift[i].in = wire
		} else {
			a.shift[i].out = wire
		}
		if a.shift[i].in.Ch != nil && a.shift[i].out.Ch != nil {
			go a.shift[i].run()
		}
	case "d":
		if output {
			a.del[i].in = wire
		} else {
			a.del[i].out = wire
		}
		if a.del[i].in.Ch != nil && a.del[i].out.Ch != nil {
			go a.del[i].run()
		}
	case "sd":
		if output {
			a.sd[i].in = wire
		} else {
			a.sd[i].out = wire
		}
		if a.sd[i].in.Ch != nil && a.sd[i].out.Ch != nil {
			go a.sd[i].run()
		}
	case "permute":
		if output {
			a.permute[i].in = wire
		} else {
			a.permute[i].out = wire
		}
		if a.permute[i].in.Ch != nil && a.permute[i].out.Ch != nil {
			go a.permute[i].run()
		}
	default:
		return fmt.Errorf("invalid type %s", ilk)
	}
	return nil
}

func (a *Adapters) GetPlug(ilk, id, param string) ([]Wire, error) {
	wires := []Wire{}
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return wires, fmt.Errorf("invalid id %s", id)
	}
	switch ilk {
	case "dp":
		wires = append(wires, a.dp[i].in)
		for j := range a.dp[i].out {
			wires = append(wires, a.dp[i].out[j])
		}
	case "s":
		wires = append(wires, a.shift[i].in)
		wires = append(wires, a.shift[i].out)
	case "d":
		wires = append(wires, a.del[i].in)
		wires = append(wires, a.del[i].out)
	case "sd":
		wires = append(wires, a.sd[i].in)
		wires = append(wires, a.sd[i].out)
	case "permute":
		wires = append(wires, a.permute[i].in)
		wires = append(wires, a.permute[i].out)
	default:
		return wires, fmt.Errorf("invalid type %s", ilk)
	}
	return wires, nil
}

type digitProgram struct {
	in  Wire
	out [11]Wire
}

// Emit program pulses when one or more digit positions activate.
func (a *digitProgram) run() {
	resp := make(chan int)
	for {
		d := <-a.in.Ch
		for i := uint(0); i < 11; i++ {
			if d.Val&(1<<i) != 0 {
				Handshake(1, a.out[i], resp)
			}
		}
		if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

type shifter struct {
	in     Wire
	out    Wire
	amount int
}

func (a *shifter) run() {
	for {
		d := <-a.in.Ch
		d.Val = shift(d.Val, a.amount)
		if d.Val != 0 {
			a.out.Ch <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

func shift(value int, amount int) int {
	if amount >= 0 {
		// Shift left by `shift` digits preserving sign
		return (value & 0b1_00000_00000) | ((value << uint(amount)) & 0b0_11111_11111)
	}
	// Shift right by `shift` digits filling new leftmost digits with sign.
	x := value >> uint(-amount)
	if value&0b1_00000_00000 != 0 {
		// 11+amount: sign has already filled top digit
		return x | (0b1_11111_11111 & ^((1 << uint(11+amount)) - 1))
	}
	return x
}

type deleter struct {
	in    Wire
	out   Wire
	digit int
}

func (a *deleter) run() {
	for {
		d := <-a.in.Ch
		if a.digit >= 0 {
			// Keep leftmost `digit` digits (leaving sign alone)
			d.Val &= ^((1 << uint(10-a.digit)) - 1)
		} else {
			// Zero leftmost `digit` digits (as well as sign)
			d.Val &= (1 << uint(10+a.digit)) - 1
		}
		if d.Val != 0 {
			a.out.Ch <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

type specialDigit struct {
	in    Wire
	out   Wire
	digit uint
}

func (a *specialDigit) run() {
	for {
		d := <-a.in.Ch
		x := d.Val >> a.digit
		mask := 0x07fc
		if d.Val&(1<<10) != 0 {
			d.Val = x | mask
		} else {
			d.Val = x & ^mask
		}
		if d.Val != 0 && a.out.Ch != nil {
			a.out.Ch <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

// Permute adapters permute and optionally duplicate or delete digits.
// ad.permute.1.11,10,9,8,7,6,5,4,3,2,1  identity
// ad.permute.1.0,10,9,8,7,6,5,4,3,2,1   delete sign
type permuter struct {
	in    Wire
	out   Wire
	order [11]int
}

func (a *permuter) run() {
	for {
		d := <-a.in.Ch
		permuted := 0
		for i := 0; i < 11; i++ {
			digit := a.order[i]
			if digit != 0 && d.Val&(1<<(digit-1)) != 0 {
				permuted |= 1 << (10 - i)
			}
		}
		d.Val = permuted
		if d.Val != 0 && a.out.Ch != nil {
			a.out.Ch <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}
