package lib

import (
	"fmt"
	"strings"
	"sync"
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
	Connections []*Jack
	Disabled bool  // to skip work for inactive accum inputs

	visited bool
	mu      sync.Mutex
}

func NewJack(name string, onReceive JackHandler, onTransmit JackHandler) *Jack {
	return &Jack{
		Name:        name,
		OnReceive:   onReceive,
		OnTransmit:  onTransmit,
		Connections: make([]*Jack, 0, 1),
	}
}

func NewInput(name string, onReceive JackHandler) *Jack {
	return NewJack(name, onReceive, nil)
}

func NewOutput(name string, onTransmit JackHandler) *Jack {
	return NewJack(name, nil, onTransmit)
}

// Transmit sends val on jack j, invoking receiver callbacks for each connected
// receiver and afterwards invoking j's transmit callback.
func (j *Jack) Transmit(val int) {
	j.mu.Lock()
	defer j.mu.Unlock()
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
	for i := range j.Connections {
		r := j.Connections[i]
		if !r.Disabled && !r.visited && r.OnReceive != nil {
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
	j.mu.Lock()
	defer j.mu.Unlock()
	if len(j.Connections) == 0 {
		return "unconnected\n"
	}
	var b strings.Builder
	for i := range j.Connections {
		fmt.Fprintf(&b, "%s %s\n", j.String(), j.Connections[i].String())
	}
	return b.String()
}

func (j *Jack) Connected() bool {
	return len(j.Connections) > 0
}

func (j *Jack) Disconnect() {
	j.mu.Lock()
	defer j.mu.Unlock()
	for i := range j.Connections {
		j.Connections[i].removeConnection(j)
	}
}

func (j *Jack) removeConnection(other *Jack) {
	j.mu.Lock()
	defer j.mu.Unlock()
	for i := range j.Connections {
		if j.Connections[i] == other {
			j.Connections[i] = j.Connections[len(j.Connections)-1]
			j.Connections = j.Connections[:len(j.Connections)-1]
			return
		}
	}
}

// Connect connects two jacks, warning about pathological connections.
func Connect(j1, j2 *Jack) error {
	if j1 == j2 {
		return fmt.Errorf("%s cannot be connected to itself", j1)
	}
	j1.mu.Lock()
	defer j1.mu.Unlock()
	j2.mu.Lock()
	defer j2.mu.Unlock()
	for i := range j1.Connections {
		if j1.Connections[i] == j2 {
			return fmt.Errorf("%s is already connected to %s", j1, j2)
		}
	}
	for i := range j2.Connections {
		if j2.Connections[i] == j1 {
			return fmt.Errorf("%s is already connected to %s", j1, j2)
		}
	}
	j1.Connections = append(j1.Connections, j2)
	j2.Connections = append(j2.Connections, j1)
	return nil
}
