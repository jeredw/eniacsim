// Package cycle simulates ENIAC's clock generation circuits.
//
// This is the main control thread for the simulator, where a for loop
// generates a repeating sequence of control pulses and calls a "ClockFunc" for
// each clocked unit for each pulse.
//
// As with the real ENIAC, clocks can be single stepped for debugging.  This
// code also supports a "test mode" that runs the clocks for a specified number
// of add cycles and then halts, to support regression tests.
package cycle

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"sync"
)

// Conn defines connections needed for the unit
type Conn struct {
	Units       []ClockFunc    // Funcs per clocked unit
	Clear       func() bool    // Clear gate (from initiate unit)
	Switches    chan [2]string // Sim controls to change mode
	Reset       chan int       // Sim control to reset unit
	Stop        chan int       // Triggered by debug breakpoints
	CycleButton Button         // Sim control to step clock
	TestButton  Button         // Interlock for test mode
	TestCycles  int            // How many cycles to run in test mode
}

type controlState struct {
	newMode        int // New operating mode requested
	buttonsPending int // How many cycle buttons to ack

	mu sync.Mutex
}

// Clock operating modes
const (
	OnePulse   = iota // Run for one pulse, then wait for CycleButton
	OneAdd            // Run for one add cycle, then wait for CycleButton
	Continuous        // Run continuously
	Test              // Run for conn.TestCycles then stop
)

// Basic sequence of pulses to generate
// Note due to the phase shift of 9P there are 40 distinct phases per add cycle
// though only 20 pulse times.
var phases = []int{
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

var mode = Continuous
var state controlState = controlState{newMode: Continuous, buttonsPending: 0}
var addCycleMu sync.Mutex
var addCycle = 0
var phase = 0

// AddCycle returns the current add cycle number, reset on s.cy.op 1a
func AddCycle() int {
	// Called by main thread
	addCycleMu.Lock()
	defer addCycleMu.Unlock()
	return addCycle
}

// Stepping returns whether the clock is being single stepped
func Stepping() bool {
	return mode == OneAdd || mode == OnePulse
}

// Mode exposes the clock mode enum for the tk gui
func Mode() int {
	return mode
}

// Stat returns the current phase of the pulse train
// FIXME: Do we really need this?  Delete if possible
func Stat() string {
	// data race: written by Run()
	if phase >= len(phases) {
		return "0"
	} else {
		return fmt.Sprintf("%d", phase)
	}
}

// Run forwards the basic ENIAC pulse train to each unit.  It runs forever as
// its own goroutine.
//
// Originally, this used a fanout of channels to distribute pulses to 50-60
// clock goroutines, but that incurred enormous context switch overhead;
// keeping to one thread and function calls is 15-20x faster.
//
// All control inputs are applied at the start of the nearest pulse.
func Run(io Conn) {
	if io.TestCycles > 0 {
		mode = Test
		<-io.TestButton.Push // wait for determinism
	}

	update := make(chan int)
	go readControls(io, update)

	var p Pulse
	p.Resp = make(chan int)
	for {
		for phase = 0; phase < len(phases); phase++ {
			applyControls(io, update)
			if phase == 32 && io.Clear() {
				p.Val = Scg
				for _, f := range io.Units {
					f(p)
				}
			} else if phases[phase] != 0 {
				p.Val = phases[phase]
				for _, f := range io.Units {
					f(p)
				}
			}
			phase++
			if phases[phase] != 0 {
				p.Val = phases[phase]
				for _, f := range io.Units {
					f(p)
				}
			}
			if mode == OnePulse {
				state.mu.Lock()
				ackButtons(io)
				state.mu.Unlock()
			}
		}
		addCycleMu.Lock()
		addCycle++
		if mode == Test && addCycle >= io.TestCycles {
			addCycleMu.Unlock()
			io.TestButton.Done <- 1
			break
		}
		addCycleMu.Unlock()
		if mode == OneAdd {
			state.mu.Lock()
			ackButtons(io)
			state.mu.Unlock()
		}
	}
}

func readControls(io Conn, update chan int) {
	for {
		select {
		case <-io.Reset:
			state.mu.Lock()
			state.newMode = Continuous
			state.mu.Unlock()
		case <-io.Stop:
			state.mu.Lock()
			state.newMode = OneAdd
			state.mu.Unlock()
		case x := <-io.Switches:
			state.mu.Lock()
			state.newMode = parseOp(x, state.newMode)
			state.mu.Unlock()
		case <-io.CycleButton.Push:
			state.mu.Lock()
			state.buttonsPending++
			state.mu.Unlock()
		}
		update <- 1
	}
}

func applyControls(io Conn, update chan int) {
	for i := 0; i < 10; i++ {
		if mode == OnePulse || (mode == OneAdd && phase == 0) {
			<-update
		} else {
			drain(update)
		}
		state.mu.Lock()
		if mode == Continuous || mode == Test {
			ackButtons(io)
		}
		if mode == Test || mode == state.newMode {
			state.mu.Unlock()
			return
		}
		switch state.newMode {
		case Continuous:
			mode = state.newMode
			ackButtons(io)
			state.mu.Unlock()
			return
		case OnePulse:
			// 1a (paused) 1p should wait for b p
			if !(mode == OneAdd && phase == 0 && state.buttonsPending == 0) {
				mode = state.newMode
				state.mu.Unlock()
				return
			}
		case OneAdd:
			// 1p (paused on phase 0) 1a should wait for b p
			if !(mode == OnePulse && phase == 0 && state.buttonsPending == 0) {
				addCycleMu.Lock()
				addCycle = 0
				addCycleMu.Unlock()
				mode = state.newMode
				state.mu.Unlock()
				return
			}
		}
		state.mu.Unlock()
	}
	panic("infinite loop")
}

func parseOp(x [2]string, defaultValue int) int {
	switch x[0] {
	case "op":
		switch x[1] {
		case "1p", "1P":
			return OnePulse
		case "1a", "1A":
			return OneAdd
		case "co", "CO":
			return Continuous
		case "cy", "CY":
			return (mode + 1) % 3
		default:
			fmt.Println("cycle unit op switch value: one of 1p, 1a, co, cy")
		}
	default:
		fmt.Println("cycle unit switch: s cy.op.val")
	}
	return defaultValue
}

func ackButtons(io Conn) {
	for state.buttonsPending > 0 {
		io.CycleButton.Done <- 1
		state.buttonsPending--
	}
}

func drain(c chan int) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
