package main

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
)

type Debugger struct {
	breakpoint [10]*Jack
}

func NewDebugger() *Debugger {
	u := &Debugger{}
	for i := range u.breakpoint {
		u.breakpoint[i] = NewInput(fmt.Sprintf("debug.bp%d", i), func(j *Jack, val int) {
			fmt.Printf("break on %s", j.ConnectionsString())
			cycle.Io.Stop <- 1
		})
	}
	return u
}

func (u *Debugger) Stat() string {
	var s string
	for i := range u.breakpoint {
		s += fmt.Sprintf("bp%d: %s", i, u.breakpoint[i].ConnectionsString())
	}
	return s
}

func (u *Debugger) FindJack(name string) (*Jack, error) {
	if len(name) < 3 {
		return nil, fmt.Errorf("invalid connection %s", name)
	}
	if name[:2] != "bp" {
		return nil, fmt.Errorf("invalid connection %s", name)
	}
	n, _ := strconv.Atoi(name[2:])
	if !(n >= 0 && n <= 9) {
		return nil, fmt.Errorf("invalid connection %s", name)
	}
	return u.breakpoint[n], nil
}

func (u *Debugger) Reset() {
	for i := range u.breakpoint {
		u.breakpoint[i].Disconnect()
	}
}
