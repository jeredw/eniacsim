package main

import (
	"fmt"
	"sync"

	. "github.com/jeredw/eniacsim/lib"
)

const (
	Cpp = 1 << iota
	Onep
	Ninep
	Tenp
	Scg
	Rp
	Onepp
	Ccg
	Twop
	Twopp
	Fourp
)

const (
	PulseMode = iota
	AddMode
	ContMode
)

var clocks = []int{
	0, Tenp, // 0
	Onep | Ninep, Tenp, // 1
	Twop | Ninep, Tenp, // 2
	Twop | Ninep, Tenp, // 3
	Twopp | Ninep, Tenp, // 4
	Twopp | Ninep, Tenp, // 5
	Fourp | Ninep, Tenp, // 6
	Fourp | Ninep, Tenp, // 7
	Fourp | Ninep, Tenp, // 8
	Fourp | Ninep, Tenp, // 9
	Onepp, 0, // 10
	Ccg, 0, // 11
	0, 0, // 12
	Rp, 0, // 13
	0, 0, // 14
	0, 0, // 15
	0, 0, // 16
	Cpp, 0, // 17
	0, 0, // 18
	Rp, 0, // 19
}

var intbch chan int
var cycbutdone chan int
var cmodemu sync.Mutex
var cmode = ContMode
var acycmu sync.Mutex
var acyc = 0
var stopmu sync.Mutex
var stop = false
var cyc = 0

func cycstat() string {
	// race: written by cycleunit()
	if cyc >= len(clocks) {
		return "0"
	} else {
		return fmt.Sprintf("%d", cyc)
	}
}

func cycsetmode(newmode int) {
	if *testcycles > 0 && newmode != ContMode {
		return
	}
	cmodemu.Lock()
	waiting_for_button := intbch != nil && (cmode == AddMode || cmode == PulseMode)
	cmode = newmode
	cmodemu.Unlock()
	if waiting_for_button {
		intbch <- 1
		<-cycbutdone
	}
}

func cycreset() {
	cycsetmode(ContMode)
}

func cyclectl(cch chan [2]string) {
	for {
		x := <-cch
		switch x[0] {
		case "op":
			switch x[1] {
			case "1p", "1P":
				cycsetmode(PulseMode)
			case "1a", "1A":
				cycsetmode(AddMode)
				acycmu.Lock()
				acyc = 0
				acycmu.Unlock()
			case "co", "CO":
				cycsetmode(ContMode)
			case "cy", "CY":
				cmodemu.Lock()
				waiting_for_button := cmode == AddMode || cmode == PulseMode
				cmode = (cmode + 1) % 3
				cmodemu.Unlock()
				if waiting_for_button {
					intbch <- 1
					<-cycbutdone
				}
			default:
				fmt.Println("cycle unit op swtch value: one of 1p, 1a, co, cy")
			}
		default:
			fmt.Println("cycle unit switch: s cy.op.val")
		}
	}
}

func cycleunit(fns []ClockFunc, bch chan int) {
	var p Pulse

	if *testcycles > 0 {
		<-teststart
	}

	intbch = make(chan int)
	go func() {
		for {
			b := <-bch
			cmodemu.Lock()
			waiting_for_button := cmode == AddMode || cmode == PulseMode
			cmodemu.Unlock()
			if waiting_for_button {
				intbch <- b
			} else {
				cycbutdone <- 1
			}
		}
	}()

	p.Resp = make(chan int)
	for {
		stopmu.Lock()
		stop = false
		stopmu.Unlock()
		cmodemu.Lock()
		wait_for_add := cmode == AddMode
		cmodemu.Unlock()
		if wait_for_add {
			<-intbch
		}
		for cyc = 0; cyc < len(clocks); cyc++ {
			cmodemu.Lock()
			wait_for_pulse := cmode == PulseMode
			cmodemu.Unlock()
			if wait_for_pulse {
				<-intbch
			}
			if cyc == 32 && (initclrff[0] || initclrff[1] || initclrff[2] ||
				initclrff[3] || initclrff[4] || initclrff[5]) {
				p.Val = Scg
				for _, f := range fns {
					f(p)
				}
			} else if clocks[cyc] != 0 {
				p.Val = clocks[cyc]
				for _, f := range fns {
					f(p)
				}
			}
			cyc++
			if clocks[cyc] != 0 {
				p.Val = clocks[cyc]
				for _, f := range fns {
					f(p)
				}
			}
			if wait_for_pulse {
				cycbutdone <- 1
			}
		}
		acycmu.Lock()
		acyc++
		if *testcycles > 0 && acyc >= *testcycles {
			acycmu.Unlock()
			testdone <- 1
			break
		}
		acycmu.Unlock()
		if wait_for_add {
			cycbutdone <- 1
		}
		stopmu.Lock()
		if stop {
			cmodemu.Lock()
			cmode = AddMode
			cmodemu.Unlock()
		}
		stopmu.Unlock()
	}
}
