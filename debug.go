package main

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"strconv"
)

type Debugger struct {
	bps [10]bp
}

type bp struct {
	n      int
	ch     chan Pulse
	what   string
	update chan int
}

func NewDebugger() *Debugger {
	return &Debugger{}
}

func (u *Debugger) Stat() string {
	var s string
	for n, bp := range u.bps {
		if bp.ch != nil {
			s += fmt.Sprintf("bp%d: %s\n", n, bp.what)
		} else {
			s += fmt.Sprintf("bp%d: -\n", n)
		}
	}
	return s
}

func (u *Debugger) Plug(name string, ch chan Pulse, what string) error {
	if len(name) < 3 {
		return fmt.Errorf("invalid connection %s", name)
	}
	if name[:2] != "bp" {
		return fmt.Errorf("invalid connection %s", name)
	}
	n, _ := strconv.Atoi(name[2:])
	if !(n >= 0 && n <= 9) {
		return fmt.Errorf("invalid connection %s", name)
	}
	if u.bps[n].update != nil {
		u.bps[n].update <- 1
	}
	u.bps[n] = bp{n, ch, what, make(chan int)}
	go awaitBreakpoint(&u.bps[n])
	return nil
}

func (u *Debugger) Reset() {
	for n, b := range u.bps {
		if b.update != nil {
			b.update <- 1
		}
		u.bps[n] = bp{n, nil, "", nil}
	}
}

func awaitBreakpoint(b *bp) {
	for {
		var p Pulse
		p.Resp = nil
		select {
		case <-b.update:
			return
		case p = <-b.ch:
		}
		if p.Val != 0 {
			fmt.Printf("triggered bp%d %s\n", b.n, b.what)
			cycle.Io.Stop <- 1
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}
