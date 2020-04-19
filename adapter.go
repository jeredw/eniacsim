package main

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
)

type Adapters struct {
	dp    [40]digitProgram
	shift [40]shifter
	del   [40]deleter
	sd    [40]specialDigit
}

func NewAdapters() *Adapters {
	return &Adapters{}
}

func (a *Adapters) Reset() {
	for i := 0; i < 40; i++ {
		a.dp[i].in = nil
		for j := 0; j < 11; j++ {
			a.dp[i].out[j] = nil
		}
		a.shift[i].in = nil
		a.shift[i].out = nil
		a.del[i].in = nil
		a.del[i].out = nil
		a.sd[i].in = nil
		a.sd[i].out = nil
	}
}

func (a *Adapters) Plug(ilk, id, param string, ch chan Pulse, output bool) error {
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return fmt.Errorf("invalid id %s", id)
	}

	switch ilk {
	case "dp":
		if output {
			// ch is a digit channel such as a data trunk
			a.dp[i].in = ch
			go a.dp[i].run()
		} else {
			// ch is a control channel such as a program trunk
			digit, _ := strconv.Atoi(param)
			if !(digit >= 1 && digit <= 11) {
				return fmt.Errorf("invalid digit %s", param)
			}
			a.dp[i].out[digit-1] = ch
		}
	case "s":
		if output {
			a.shift[i].in = ch
		} else {
			a.shift[i].out = ch
		}
		amount, _ := strconv.Atoi(param)
		if !(amount >= -10 && amount <= 10) {
			return fmt.Errorf("invalid shift amount %s", param)
		}
		a.shift[i].amount = amount
		if a.shift[i].in != nil && a.shift[i].out != nil {
			go a.shift[i].run()
		}
	case "d":
		if output {
			a.del[i].in = ch
		} else {
			a.del[i].out = ch
		}
		digit, _ := strconv.Atoi(param)
		if !(digit >= -10 && digit <= 10) {
			return fmt.Errorf("invalid digit %s", param)
		}
		a.del[i].digit = digit
		if a.del[i].in != nil && a.del[i].out != nil {
			go a.del[i].run()
		}
	case "sd":
		if output {
			a.sd[i].in = ch
		} else {
			a.sd[i].out = ch
		}
		digit, _ := strconv.Atoi(param)
		if !(digit >= -10 && digit <= 10) {
			return fmt.Errorf("invalid digit %s", param)
		}
		a.sd[i].digit = uint(digit)
		if a.sd[i].in != nil && a.sd[i].out != nil {
			go a.sd[i].run()
		}
	default:
		return fmt.Errorf("invalid type %s", ilk)
	}
	return nil
}

type digitProgram struct {
	in  chan Pulse
	out [11]chan Pulse
}

// Emit program pulses when one or more digit positions activate.
func (a *digitProgram) run() {
	resp := make(chan int)
	for {
		d := <-a.in
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
	in     chan Pulse
	out    chan Pulse
	amount int
}

func (a *shifter) run() {
	for {
		d := <-a.in
		d.Val = shift(d.Val, a.amount)
		if d.Val != 0 {
			a.out <- d
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
	in    chan Pulse
	out   chan Pulse
	digit int
}

func (a *deleter) run() {
	for {
		d := <-a.in
		if a.digit >= 0 {
			// Keep leftmost `digit` digits (leaving sign alone)
			d.Val &= ^((1 << uint(10-a.digit)) - 1)
		} else {
			// Zero leftmost `digit` digits (as well as sign)
			d.Val &= (1 << uint(10+a.digit)) - 1
		}
		if d.Val != 0 {
			a.out <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

type specialDigit struct {
	in    chan Pulse
	out   chan Pulse
	digit uint
}

func (a *specialDigit) run() {
	for {
		d := <-a.in
		x := d.Val >> a.digit
		mask := 0x07fc
		if d.Val&(1<<10) != 0 {
			d.Val = x | mask
		} else {
			d.Val = x & ^mask
		}
		if d.Val != 0 && a.out != nil {
			a.out <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}
