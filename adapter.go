package main

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
	"strings"
)

type Adapters struct {
	dp      [80]digitProgram
	shift   [80]shifter
	del     [80]deleter
	sd      [80]specialDigit
	permute [80]permuter
}

func NewAdapters() *Adapters {
	a := &Adapters{}
	dpInput := func(i int) JackHandler {
		return func(j *Jack, val int) {
			a.dp[i].adapt(val)
		}
	}
	shiftInput := func(i int) JackHandler {
		return func(j *Jack, val int) {
			a.shift[i].adapt(val)
		}
	}
	delInput := func(i int) JackHandler {
		return func(j *Jack, val int) {
			a.del[i].adapt(val)
		}
	}
	sdInput := func(i int) JackHandler {
		return func(j *Jack, val int) {
			a.sd[i].adapt(val)
		}
	}
	permuteInput := func(i int) JackHandler {
		return func(j *Jack, val int) {
			a.permute[i].adapt(val)
		}
	}
	for i := 0; i < 40; i++ {
		a.dp[i].in = NewInput(fmt.Sprintf("ad.dp.i.%d", i+1), dpInput(i))
		for j := 0; j < 11; j++ {
			a.dp[i].out[j] = NewOutput(fmt.Sprintf("ad.dp.o.%d.%d", i+1, j+1), nil)
		}
		a.shift[i].in = NewInput(fmt.Sprintf("ad.s.i.%d", i+1), shiftInput(i))
		a.shift[i].out = NewOutput(fmt.Sprintf("ad.s.o.%d", i+1), nil)
		a.del[i].in = NewInput(fmt.Sprintf("ad.d.i.%d", i+1), delInput(i))
		a.del[i].out = NewOutput(fmt.Sprintf("ad.d.o.%d", i+1), nil)
		a.sd[i].in = NewInput(fmt.Sprintf("ad.sd.i.%d", i+1), sdInput(i))
		a.sd[i].out = NewOutput(fmt.Sprintf("ad.sd.o.%d", i+1), nil)
		a.permute[i].in = NewInput(fmt.Sprintf("ad.permute.i.%d", i+1), permuteInput(i))
		a.permute[i].out = NewOutput(fmt.Sprintf("ad.permute.o.%d", i+1), nil)
	}
	return a
}

func (a *Adapters) Reset() {
	for i := 0; i < 80; i++ {
		a.dp[i].in.Disconnect()
		for j := 0; j < 11; j++ {
			a.dp[i].out[j].Disconnect()
		}
		a.shift[i].in.Disconnect()
		a.shift[i].out.Disconnect()
		a.del[i].in.Disconnect()
		a.del[i].out.Disconnect()
		a.sd[i].in.Disconnect()
		a.sd[i].out.Disconnect()
		a.permute[i].in.Disconnect()
		a.permute[i].out.Disconnect()
	}
}

type adParamSwitch struct {
	minValue int
	maxValue int
	data     *int
}

func (s *adParamSwitch) Set(value string) error {
	param, _ := strconv.Atoi(value)
	if !(param >= s.minValue && param <= s.maxValue) {
		return fmt.Errorf("invalid parameter %s", value)
	}
	*s.data = param
	return nil
}

func (s *adParamSwitch) Get() string {
	return fmt.Sprintf("%d", s.data)
}

type permuteSwitch struct {
	ad *permuter
}

func (s *permuteSwitch) Set(value string) error {
	order := strings.Split(value, ",")
	if len(order) != 11 {
		return fmt.Errorf("ad.permute usage: ad.permute.1.11,10,9,8,7,6,5,4,3,2,1")
	}
	for j := range order {
		pos, _ := strconv.Atoi(order[j])
		if !(pos >= 0 && pos <= 11) {
			return fmt.Errorf("invalid digit field in permutation")
		}
		s.ad.order[j] = pos
	}
	return nil
}

func (s *permuteSwitch) Get() string {
	return fmt.Sprintf("%v", s.ad.order)
}

func (a *Adapters) FindSwitch(name string) (Switch, error) {
	p := strings.Split(name, ".")
	if len(p) != 2 {
		return nil, fmt.Errorf("invalid adapter switch %s", name)
	}
	kind := p[0]
	id := p[1]
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return nil, fmt.Errorf("invalid id %s", id)
	}
	i--
	switch kind {
	case "dp":
		return nil, fmt.Errorf("dp param must be specified with p")
	case "s":
		return &adParamSwitch{-10, 10, &a.shift[i].amount}, nil
	case "d":
		return &adParamSwitch{-10, 10, &a.del[i].digit}, nil
	case "sd":
		return &adParamSwitch{0, 10, &a.sd[i].digit}, nil
	case "permute":
		return &permuteSwitch{&a.permute[i]}, nil
	}
	return nil, fmt.Errorf("invalid type %s", kind)
}

func (a *Adapters) FindJack(name string) (*Jack, error) {
	p := strings.Split(name, ".")
	if len(p) < 3 {
		return nil, fmt.Errorf("invalid jack %s", name)
	}
	dir := p[0]
	kind := p[1]
	id := p[2]
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return nil, fmt.Errorf("invalid id %s", id)
	}
	i--
	switch {
	case kind == "dp" && dir == "i":
		return a.dp[i].in, nil
	case kind == "dp" && dir == "o":
		if len(p) < 4 {
			return nil, fmt.Errorf("invalid jack %s", name)
		}
		digit, _ := strconv.Atoi(p[3])
		if !(digit >= 1 && digit <= 11) {
			return nil, fmt.Errorf("invalid digit %s in %s", p[3], name)
		}
		return a.dp[i].out[digit-1], nil
	case kind == "s" && dir == "i":
		return a.shift[i].in, nil
	case kind == "s" && dir == "o":
		return a.shift[i].out, nil
	case kind == "d" && dir == "i":
		return a.del[i].in, nil
	case kind == "d" && dir == "o":
		return a.del[i].out, nil
	case kind == "sd" && dir == "i":
		return a.sd[i].in, nil
	case kind == "sd" && dir == "o":
		return a.sd[i].out, nil
	case kind == "permute" && dir == "i":
		return a.permute[i].in, nil
	case kind == "permute" && dir == "o":
		return a.permute[i].out, nil
	}
	return nil, fmt.Errorf("invalid type %s", kind)
}

type digitProgram struct {
	in  *Jack
	out [11]*Jack
}

// Emit program pulses when one or more digit positions activate.
func (a *digitProgram) adapt(val int) {
	for i := uint(0); i < 11; i++ {
		if val&(1<<i) != 0 {
			a.out[i].Transmit(1)
		}
	}
}

type shifter struct {
	in     *Jack
	out    *Jack
	amount int
}

func (a *shifter) adapt(val int) {
	val = shift(val, a.amount)
	if val != 0 {
		a.out.Transmit(val)
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
	in    *Jack
	out   *Jack
	digit int
}

func (a *deleter) adapt(val int) {
	if a.digit >= 0 {
		// Keep leftmost `digit` digits (leaving sign alone)
		val &= ^((1 << uint(10-a.digit)) - 1)
	} else {
		// Zero leftmost `digit` digits (as well as sign)
		val &= (1 << uint(10+a.digit)) - 1
	}
	if val != 0 {
		a.out.Transmit(val)
	}
}

type specialDigit struct {
	in    *Jack
	out   *Jack
	digit int
}

func (a *specialDigit) adapt(val int) {
	x := val >> uint(a.digit)
	mask := 0x07fc
	if val&(1<<10) != 0 {
		val = x | mask
	} else {
		val = x & ^mask
	}
	if val != 0 {
		a.out.Transmit(val)
	}
}

// Permute adapters permute and optionally duplicate or delete digits.
// ad.permute.1.11,10,9,8,7,6,5,4,3,2,1  identity
// ad.permute.1.0,10,9,8,7,6,5,4,3,2,1   delete sign
type permuter struct {
	in    *Jack
	out   *Jack
	order [11]int
}

func (a *permuter) adapt(val int) {
	permuted := 0
	for i := 0; i < 11; i++ {
		digit := a.order[i]
		if digit != 0 && val&(1<<(digit-1)) != 0 {
			permuted |= 1 << (10 - i)
		}
	}
	val = permuted
	if val != 0 {
		a.out.Transmit(val)
	}
}
