package lib

import (
	"fmt"
	"strings"
)

// Plugboard owns one or more jacks and provides a way to find them by name.
type Plugboard interface {
	FindJack(name string) (*Jack, error)
}

type JackHandler func(*Jack, int)

type Jack struct {
	Name        string
	OnReceive   JackHandler
	OnTransmit  JackHandler
	Receivers   []*Jack
	Disabled    bool  // to skip work for inactive accum inputs
	Connected   bool
	OutJack     *Jack

	visited     bool
	forward     bool  // forwarding node (for trays)
}

func newJack(name string, onReceive JackHandler, onTransmit JackHandler) *Jack {
	return &Jack{
		Name:        name,
		OnReceive:   onReceive,
		OnTransmit:  onTransmit,
		Receivers:   make([]*Jack, 0, 1),
	}
}

func NewInput(name string, onReceive JackHandler) *Jack {
	return newJack(name, onReceive, nil)
}

func NewOutput(name string, onTransmit JackHandler) *Jack {
	return newJack(name, nil, onTransmit)
}

func NewForwardingJack(name string) *Jack {
	jack := newJack(name, nil, nil)
	jack.forward = true
	return jack
}

// Transmit sends val on jack j, invoking receiver callbacks for each connected
// receiver and afterwards invoking j's transmit callback.
func (j *Jack) Transmit(val int) {
	if j.visited {
		// A previous call to Transmit() on this jack triggered this call.  Break
		// the cycle and return early here.
		//
		// This isn't an error, and can happen legitimately when e.g. two trunks
		// are connected, like p 1 2.  Transmitting on trunk 1 will call transmit
		// on trunk 2, which will attempt to transmit on trunk 1 again.
		return
	}
	j.visited = true
	transmitted := false
	for _, r := range j.Receivers {
		if r.forward {
			transmitted = true
			r.Transmit(val)
		} else if !r.visited && !r.Disabled {
			transmitted = true
			r.OnReceive(r, val)
		}
	}
	if transmitted && j.OnTransmit != nil {
		j.OnTransmit(j, val)
	}
	j.visited = false
}

func (j *Jack) String() string {
	return j.Name
}

func (j *Jack) ConnectionsString() string {
	if len(j.Receivers) == 0 {
		return "unconnected\n"
	}
	var b strings.Builder
	for _, r := range j.Receivers {
		fmt.Fprintf(&b, "%s %s\n", j.String(), r.String())
	}
	return b.String()
}

// Connect connects two jacks, warning about pathological connections.
func Connect(j1, j2 *Jack) error {
	if j1 == j2 {
		return fmt.Errorf("%s cannot be connected to itself", j1)
	}
	for i := range j1.Receivers {
		if j1.Receivers[i] == j2 {
			return fmt.Errorf("%s is already connected to %s", j1, j2)
		}
	}
	for i := range j2.Receivers {
		if j2.Receivers[i] == j1 {
			return fmt.Errorf("%s is already connected to %s", j1, j2)
		}
	}
	if j1.OnReceive != nil || j1.forward {
		j2.Receivers = append(j2.Receivers, j1)
		j2.Connected = true
	}
	if j2.OnReceive != nil || j2.forward {
		j1.Receivers = append(j1.Receivers, j2)
		j1.Connected = true
	}
	return nil
}
