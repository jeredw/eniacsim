package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/jeredw/eniacsim/lib"
	. "github.com/jeredw/eniacsim/lib/units"
)

type Debugger struct {
	Io DebuggerConn

	assert     [40]*assertion
	breakpoint [40]*Jack
	dump       [40]*dump
	quit       *Jack
}

type DebuggerConn struct {
	Accumulator [20]StaticWiring
}

func NewDebugger() *Debugger {
	u := &Debugger{}
	for i := range u.breakpoint {
		num := i + 1
		u.breakpoint[i] = NewInput(fmt.Sprintf("debug.bp.%d", num), func(j *Jack, val int) {
			fmt.Printf("[debug.bp.%d] break", num)
			cycle.Stop()
		})
	}
	for i := range u.dump {
		dump := &dump{}
		num := i + 1
		dump.trigger = NewInput(fmt.Sprintf("debug.dump.%d", num), func(j *Jack, val int) {
			value := u.Io.Accumulator[dump.accum-1].Value()
			fmt.Printf("[debug.dump.%d] a%d = %s\n", num, dump.accum, value)
		})
		u.dump[i] = dump
	}
	for i := range u.assert {
		assert := &assertion{}
		num := i + 1
		assert.trigger = NewInput(fmt.Sprintf("debug.assert.%d", num), func(j *Jack, val int) {
			if !u.testAssertion(assert) {
				value := u.Io.Accumulator[assert.accum-1].Value()
				fmt.Printf("[debug.assert.%d] a%d = %s !~ %s\n", num, assert.accum, value, assert.expectedDigits)
				cycle.Stop()
			}
		})
		u.assert[i] = assert
	}
	u.quit = NewInput("debug.quit", func(*Jack, int) {
		os.Exit(0)
	})
	return u
}

func (u *Debugger) FindJack(name string) (*Jack, error) {
	if name == "quit" {
		return u.quit, nil
	}
	p := strings.Split(name, ".")
	if len(p) != 2 {
		return nil, fmt.Errorf("invalid debugger connection %s", name)
	}
	kind := p[0]
	id := p[1]
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return nil, fmt.Errorf("invalid id %s", id)
	}
	i--
	switch kind {
	case "assert":
		return u.assert[i].trigger, nil
	case "bp":
		return u.breakpoint[i], nil
	case "dump":
		return u.dump[i].trigger, nil
	}
	return nil, fmt.Errorf("invalid debugger connection %s", name)
}

func (u *Debugger) FindSwitch(name string) (Switch, error) {
	p := strings.Split(name, ".")
	if len(p) != 2 {
		return nil, fmt.Errorf("invalid debugger switch %s", name)
	}
	kind := p[0]
	id := p[1]
	i, _ := strconv.Atoi(id)
	if !(i >= 1 && i <= 40) {
		return nil, fmt.Errorf("invalid id %s", id)
	}
	i--
	switch kind {
	case "assert":
		return u.assert[i], nil
	case "dump":
		return u.dump[i], nil
	}
	return nil, fmt.Errorf("invalid debugger switch %s", name)
}

// assertion checks an accumulator value when triggered and stops if false:
//   p 1-1 debug.assert.0
//   s debug.assert.0 a5~Mxxxxxxxxxx
//   #s debug.assert.0 a5~Pxxxxxxxxxx
//   #s debug.assert.0 a5~x54xxxxxxxx
type assertion struct {
	trigger        *Jack
	accum          int
	expectedDigits string // 'x' means don't care
}

func (u *Debugger) testAssertion(assert *assertion) bool {
	// value is a string like "M 9876543210"
	value := u.Io.Accumulator[assert.accum-1].Value()
	if assert.expectedDigits[0] != 'x' && value[0] != assert.expectedDigits[0] {
		return false
	}
	for i := 1; i <= 10; i++ {
		if assert.expectedDigits[i] != 'x' && value[1+i] != assert.expectedDigits[i] {
			return false
		}
	}
	return true
}

func (s *assertion) Set(value string) error {
	p := strings.Split(value, "~")
	if len(p) != 2 {
		return fmt.Errorf("invalid assertion %s", value)
	}
	accum := p[0]
	digits := p[1]
	if len(accum) < 2 || accum[0] != 'a' {
		return fmt.Errorf("invalid accumulator in assertion")
	}
	n, _ := strconv.Atoi(accum[1:])
	if !(n >= 1 && n <= 20) {
		return fmt.Errorf("invalid accumulator in assertion")
	}
	s.accum = n

	if len(digits) != 11 {
		return fmt.Errorf("invalid digit string in assertion '%s'", digits)
	}
	for i := 0; i < 11; i++ {
		if digits[i] == 'x' {
			continue
		}
		if i == 0 && !(digits[i] == 'M' || digits[i] == 'P') {
			return fmt.Errorf("invalid digit string in assertion")
		}
		if i > 0 && !(digits[i] >= '0' && digits[i] <= '9') {
			return fmt.Errorf("invalid digit string in assertion")
		}
	}
	s.expectedDigits = digits

	return nil
}

func (s *assertion) Get() string {
	return fmt.Sprintf("a%d = %s", s.accum, s.expectedDigits)
}

// dump prints an accumulator to stdout when triggered:
//   p 1-1 debugger.dump.0
//   s debugger.dump.0 a20
type dump struct {
	trigger *Jack
	accum   int
}

func (s *dump) Set(value string) error {
	if len(value) < 2 || value[0] != 'a' {
		return fmt.Errorf("invalid accumulator in dump")
	}
	n, _ := strconv.Atoi(value[1:])
	if !(n >= 1 && n <= 20) {
		return fmt.Errorf("invalid accumulator in dump")
	}
	s.accum = n

	return nil
}

func (s *dump) Get() string {
	return fmt.Sprintf("a%d", s.accum)
}
