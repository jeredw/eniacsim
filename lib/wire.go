package lib

import (
	"fmt"
	"strings"
)

// Wire sends pulses from unit to unit.  Unlike actual wires, Wires are
// directed, so that information only travels from source to sink.
type Wire struct {
	Source string
	Sink   string
	Ch     chan Pulse
}

func NewWire(source, sink string) *Wire {
	return &Wire{
		Source: source,
		Sink:   sink,
		Ch:     make(chan Pulse),
	}
}

func (w Wire) ToString() string {
	return fmt.Sprintf("[%s->%s]", w.Source, w.Sink)
}

// Plug connects wire to jack, warning first if this replaces another
// connection.
func Plug(jack *Wire, wire Wire) {
	if jack.Ch != nil {
		fmt.Printf("warning: connection %s replacing %s\n", wire, *jack)
	}
	*jack = wire
}

// Handshake sends val on wire and then waits for an acknowledgement on resp.
func Handshake(val int, wire Wire, resp chan int) {
	if wire.Ch != nil {
		wire.Ch <- Pulse{val, resp}
		<-resp
	}
}

// Tee returns a Wire, t, that merges output from wires a and b.  Note that
// outputs on a are forwarded to t but not b.  So that Tee can be used at both
// unit inputs and outputs, inputs to "t" are also reflected at a and b.
func Tee(a, b Wire) Wire {
	sources := []string{}
	sinks := []string{}
	if len(a.Source) != 0 {
		sources = append(sources, a.Source)
	}
	if len(b.Source) != 0 && b.Source != a.Source {
		sources = append(sources, b.Source)
	}
	if len(a.Sink) != 0 {
		sinks = append(sinks, a.Sink)
	}
	if len(b.Sink) != 0 && b.Sink != a.Sink {
		sinks = append(sinks, b.Sink)
	}
	t := NewWire(strings.Join(sources, ","), strings.Join(sinks, ","))
	go func() {
		for {
			select {
			case pa := <-a.Ch:
				if pa.Val != 0 {
					t.Ch <- pa
				}
			case pb := <-b.Ch:
				if pb.Val != 0 {
					t.Ch <- pb
				}
			case pt := <-t.Ch:
				if pt.Val != 0 {
					var pt2 Pulse
					if a.Ch != nil {
						pt2.Resp = make(chan int)
						pt2.Val = pt.Val
						a.Ch <- pt2
						<-pt2.Resp
					}
					if b.Ch != nil {
						pt2.Resp = make(chan int)
						pt2.Val = pt.Val
						b.Ch <- pt2
						<-pt2.Resp
					}
					pt.Resp <- 1
				}
			}
		}
	}()
	return *t
}
