package main

import (
	"fmt"

	. "github.com/jeredw/eniacsim/lib"
)

type trunk struct {
	xmit    [16]chan Pulse
	recv    []chan Pulse
	started bool

	rewiring           chan int
	waitingForRewiring chan int
}

var dtrays [20]trunk
var ctrays [121]trunk

func (t *trunk) reset() {
	for i, _ := range t.xmit {
		t.xmit[i] = nil
	}
	t.recv = nil
	t.started = false
	if t.rewiring != nil {
		t.rewiring <- -1
		t.rewiring = nil
	}
	t.waitingForRewiring = nil
}

func trayreset() {
	for i := range dtrays {
		dtrays[i].reset()
	}
	for i := range ctrays {
		ctrays[i].reset()
	}
}

func (t *trunk) run() {
	var x, p Pulse

	p.Resp = make(chan int)
	for {
		select {
		case q := <-t.rewiring:
			if q == -1 {
				return
			}
			t.waitingForRewiring <- 1
			<-t.rewiring
			continue
		case x = <-t.xmit[0]:
		case x = <-t.xmit[1]:
		case x = <-t.xmit[2]:
		case x = <-t.xmit[3]:
		case x = <-t.xmit[4]:
		case x = <-t.xmit[5]:
		case x = <-t.xmit[6]:
		case x = <-t.xmit[7]:
		case x = <-t.xmit[8]:
		case x = <-t.xmit[9]:
		case x = <-t.xmit[10]:
		case x = <-t.xmit[11]:
		case x = <-t.xmit[12]:
		case x = <-t.xmit[13]:
		case x = <-t.xmit[14]:
		case x = <-t.xmit[15]:
		}
		p.Val = x.Val
		if x.Val != 0 {
			needresp := 0
			for _, c := range t.recv {
				if c != nil {
				pulseloop:
					for {
						select {
						case c <- p:
							needresp++
							break pulseloop
						case <-p.Resp:
							needresp--
						}
					}
				}
			}
			for needresp > 0 {
				<-p.Resp
				needresp--
			}
		}
		if x.Resp != nil {
			x.Resp <- 1
		}
	}
}

func trunkxmit(ilk, n int, ch chan Pulse) {
	var t *trunk

	if ilk == 0 {
		t = &dtrays[n]
	} else {
		t = &ctrays[n]
	}
	if !t.started {
		t.rewiring = make(chan int)
		t.waitingForRewiring = make(chan int)
		go t.run()
		t.started = true
	}
	for i, c := range t.xmit {
		if c == nil {
			t.rewiring <- 1
			<-t.waitingForRewiring
			t.xmit[i] = ch
			t.rewiring <- 1
			return
		}
	}
	fmt.Printf("Too many transmitters on %d:%d\n", ilk, n)
}

func trunkrecv(ilk, n int, ch chan Pulse) {
	var t *trunk

	if ilk == 0 {
		t = &dtrays[n]
	} else {
		t = &ctrays[n]
	}
	if t.recv == nil {
		t.recv = make([]chan Pulse, 0, 20)
	}
	for i, c := range t.recv {
		if c == nil {
			t.recv[i] = ch
			return
		}
	}
	if len(t.recv) != cap(t.recv) {
		t.recv = append(t.recv, ch)
	}
}
