package main

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/jeredw/eniacsim/lib"
)

// Trays models ENIAC's data and program control lines.
//
// Data trunks carry pulses for 11 digits in a single integer with bit N set
// when there is a pulse for that digit.  Program trunks use the same
// representation but just transmit 1s.
type Trays struct {
	data    [20]trunk
	program [11][11]trunk
}

// trunk relays messages from N sender channels to M receiver channels.
type trunk struct {
	sender   [16]Wire
	receiver []Wire
	started  bool

	rewiring           chan int
	waitingForRewiring chan int
}

func NewTrays() *Trays {
	return &Trays{}
}

func (t *Trays) Reset() {
	for i := range t.data {
		t.data[i].reset()
	}
	for i := range t.program {
		for j := range t.program[0] {
			t.program[i][j].reset()
		}
	}
}

func (t *Trays) Plug(name string, wire Wire, output bool) error {
	dash := strings.IndexByte(name, '-')
	if dash == -1 {
		tray, _ := strconv.Atoi(name)
		if !(tray >= 1 && tray <= 20) {
			return fmt.Errorf("invalid data trunk %s", name)
		}
		trunk := &t.data[tray-1]
		if output {
			return trunk.addSender(wire)
		} else {
			return trunk.addReceiver(wire)
		}
	} else {
		tray, _ := strconv.Atoi(name[:dash])
		if !(tray >= 1 && tray <= 11) {
			return fmt.Errorf("invalid program trunk %s", name)
		}
		if len(name) <= dash+1 {
			return fmt.Errorf("invalid program trunk %s", name)
		}
		line, _ := strconv.Atoi(name[dash+1:])
		if !(line >= 1 && line <= 11) {
			return fmt.Errorf("invalid program trunk %s", name)
		}
		trunk := &t.program[tray-1][line-1]
		if output {
			return trunk.addSender(wire)
		} else {
			return trunk.addReceiver(wire)
		}
	}
	return nil
}

func (t *trunk) reset() {
	for i, _ := range t.sender {
		t.sender[i] = Wire{}
	}
	t.receiver = nil
	t.started = false
	if t.rewiring != nil {
		t.rewiring <- -1
		t.rewiring = nil
	}
	t.waitingForRewiring = nil
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
		case x = <-t.sender[0].Ch:
		case x = <-t.sender[1].Ch:
		case x = <-t.sender[2].Ch:
		case x = <-t.sender[3].Ch:
		case x = <-t.sender[4].Ch:
		case x = <-t.sender[5].Ch:
		case x = <-t.sender[6].Ch:
		case x = <-t.sender[7].Ch:
		case x = <-t.sender[8].Ch:
		case x = <-t.sender[9].Ch:
		case x = <-t.sender[10].Ch:
		case x = <-t.sender[11].Ch:
		case x = <-t.sender[12].Ch:
		case x = <-t.sender[13].Ch:
		case x = <-t.sender[14].Ch:
		case x = <-t.sender[15].Ch:
		}
		p.Val = x.Val
		if x.Val != 0 {
			needresp := 0
			for _, wire := range t.receiver {
				if wire.Ch != nil {
				pulseloop:
					for {
						select {
						case wire.Ch <- p:
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

func (t *trunk) addSender(wire Wire) error {
	if !t.started {
		t.rewiring = make(chan int)
		t.waitingForRewiring = make(chan int)
		go t.run()
		t.started = true
	}
	for i, w := range t.sender {
		if w.Ch == nil {
			t.rewiring <- 1
			<-t.waitingForRewiring
			t.sender[i] = wire
			t.rewiring <- 1
			return nil
		}
	}
	return fmt.Errorf("too many senders")
}

func (t *trunk) addReceiver(wire Wire) error {
	if t.receiver == nil {
		t.receiver = make([]Wire, 0, 20)
	}
	for i, w := range t.receiver {
		if w.Ch == nil {
			t.receiver[i] = wire
			return nil
		}
	}
	if len(t.receiver) != cap(t.receiver) {
		t.receiver = append(t.receiver, wire)
		return nil
	}
	return fmt.Errorf("too many receivers")
}
