package main

import (
	"fmt"

	. "github.com/jeredw/eniacsim/lib"
)

type bp struct {
	n      int
	ch     chan Pulse
	what   string
	update chan int
}

var bps [10]bp

func debugplug(n int, ch chan Pulse, what string) {
	if bps[n].update != nil {
		bps[n].update <- 1
	}
	bps[n] = bp{n, ch, what, make(chan int)}
	go dobp(&bps[n])
}

func debugreset() {
	for n, b := range bps {
		if b.update != nil {
			b.update <- 1
		}
		bps[n] = bp{n, nil, "", nil}
	}
}

func dobp(b *bp) {
	for {
		var p Pulse
		select {
		case <-b.update:
			return
		case p = <-b.ch:
		}
		if p.Val != 0 {
			fmt.Printf("triggered bp%d %s\n", b.n, b.what)
			stopmu.Lock()
			stop = true
			stopmu.Unlock()
		}
		if p.Resp != nil {
			p.Resp <- 1
		}
	}
}

func debugstat() string {
	var s string
	for n, bp := range bps {
		if bp.ch != nil {
			s += fmt.Sprintf("bp%d: %s\n", n, bp.what)
		} else {
			s += fmt.Sprintf("bp%d: -\n", n)
		}
	}
	return s
}
