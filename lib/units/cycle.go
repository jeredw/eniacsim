package units

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"sync"
)

// Cycle simulates ENIAC's clock generation circuits
type Cycle struct {
	Io CycleConn // Connections to simulator control and other units

	mode       int          // The current clock operating mode
	control    controlState // Simulator control updates
	addCycleMu sync.Mutex
	addCycle   int
	phase      int

	switches chan [2]string // Sim controls to change mode
	tracer   Tracer
}

// CycleConn defines connections needed for the cycle unit
type CycleConn struct {
	Units          []Clocked   // Clocked units
	SelectiveClear func() bool // Clear gate (from initiate unit)
	Reset          chan int    // Sim control to reset unit
	Stop           chan int    // Triggered by debug breakpoints
	CycleButton    Button      // Sim control to step clock
	TestButton     Button      // Interlock for test mode
	TestCycles     int         // How many cycles to run in test mode
}

// Clock operating modes
const (
	OnePulse   = iota // Run for one pulse, then wait for CycleButton
	OneAdd            // Run for one add cycle, then wait for CycleButton
	Continuous        // Run continuously
	Test              // Run for conn.TestCycles then stop
)

type controlState struct {
	newMode        int // New operating mode requested
	buttonsPending int // How many cycle buttons to ack

	mu sync.Mutex
}

// Basic sequence of pulses to generate
// Note due to the phase shift of 10P there are 40 distinct phases per add
// cycle though only 20 pulse times.
var phases = []Pulse{
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
	Ccg, 0, // 11  (Ccg is really asserted for ~7 pulse times)
	0, 0, // 12
	Rp, 0, // 13
	0, 0, // 14
	0, 0, // 15
	0, 0, // 16
	Cpp, 0, // 17
	0, 0, // 18
	Rp, 0, // 19
}

// NewCycle constructs a new Cycle instance.
func NewCycle(io CycleConn) *Cycle {
	return &Cycle{
		Io:       io,
		mode:     Continuous,
		control:  controlState{newMode: Continuous, buttonsPending: 0},
		switches: make(chan [2]string),
	}
}

// AddCycle returns the current add cycle number, reset on s.cy.op 1a
func (u *Cycle) AddCycle() int {
	// Called by main thread
	u.addCycleMu.Lock()
	defer u.addCycleMu.Unlock()
	return u.addCycle
}

// Stepping returns whether the clock is being single stepped
func (u *Cycle) Stepping() bool {
	return u.mode == OneAdd || u.mode == OnePulse
}

// Mode exposes the clock mode enum for the tk gui
func (u *Cycle) Mode() int {
	return u.mode
}

// AttachTracer connects a trace logger.
func (u *Cycle) AttachTracer(tracer Tracer) {
	u.tracer = tracer
}

// Stat returns the current phase of the pulse train
// FIXME: Do we really need this?  Delete if possible
func (u *Cycle) Stat() string {
	// data race: written by Run()
	if u.phase >= len(phases) {
		return "0"
	} else {
		return fmt.Sprintf("%d", u.phase)
	}
}

// Run forwards the basic ENIAC pulse train to each unit.  It runs forever as
// its own goroutine.
//
// This is the main control thread for the simulator, where a for loop
// generates a repeating sequence of control pulses and calls "Clock" for each
// clocked unit for each pulse.
//
// As with the real ENIAC, clocks can be single stepped for debugging.  This
// code also supports a "test mode" that runs the clocks for a specified number
// of add cycles and then halts, to support regression tests.
//
// Originally, this used a fanout of channels to distribute pulses to 50-60
// clock goroutines, but that incurred enormous context switch overhead;
// keeping to one thread and function calls is 15-20x faster.
//
// Simulator control changes are synchronized and applied at the start of the
// nearest pulse.
func (u *Cycle) Run() {
	if u.Io.TestCycles > 0 {
		u.mode = Test
		<-u.Io.TestButton.Push // wait for determinism
	}

	// readControls polls simulator controls, updates the control struct and
	// signals updates on the update channel.
	update := make(chan int)
	go u.readControls(update)

	for {
		for u.phase = 0; u.phase < len(phases); u.phase++ {
			if u.tracer != nil {
				u.tracer.AdvanceTimestep()
				//u.tracer.LogValue("cyc.pulse", 6, int64(u.phase/2))
			}
			u.applyControls(update)
			if u.phase == 32 && u.Io.SelectiveClear() {
				for _, c := range u.Io.Units {
					c.Clock(Scg)
				}
			} else if phases[u.phase] != 0 {
				for _, c := range u.Io.Units {
					c.Clock(phases[u.phase])
				}
			}
			u.phase++
			if phases[u.phase] != 0 {
				for _, c := range u.Io.Units {
					c.Clock(phases[u.phase])
				}
			}
			if u.mode == OnePulse {
				u.control.mu.Lock()
				u.ackButtons()
				u.control.mu.Unlock()
			}
		}
		if u.tracer != nil {
			u.tracer.UpdateValues()
		}
		u.addCycleMu.Lock()
		u.addCycle++
		if u.mode == Test && u.addCycle >= u.Io.TestCycles {
			u.addCycleMu.Unlock()
			u.Io.TestButton.Done <- 1
			break
		}
		u.addCycleMu.Unlock()
		if u.mode == OneAdd {
			u.control.mu.Lock()
			u.ackButtons()
			u.control.mu.Unlock()
		}
	}
}

func (u *Cycle) readControls(update chan int) {
	for {
		select {
		case <-u.Io.Reset:
			u.control.mu.Lock()
			u.control.newMode = Continuous
			u.control.mu.Unlock()
		case <-u.Io.Stop:
			u.control.mu.Lock()
			u.control.newMode = OneAdd
			u.control.mu.Unlock()
		case x := <-u.switches:
			u.control.mu.Lock()
			u.control.newMode = parseOp(x, u.control.newMode)
			u.control.mu.Unlock()
		case <-u.Io.CycleButton.Push:
			u.control.mu.Lock()
			u.control.buttonsPending++
			u.control.mu.Unlock()
		}
		update <- 1
	}
}

func (u *Cycle) applyControls(update chan int) {
	for i := 0; i < 10; i++ {
		// If stepping, wait for b p or a mode change
		if u.mode == OnePulse || (u.mode == OneAdd && u.phase == 0) {
			<-update
		} else {
			drain(update)
		}
		u.control.mu.Lock()
		if u.mode == Continuous || u.mode == Test {
			// Ignore b p when cycling continuously so main thread doesn't block
			u.ackButtons()
		}
		if u.mode == Test || u.mode == u.control.newMode {
			u.control.mu.Unlock()
			return
		}
		switch u.control.newMode {
		case Continuous:
			u.mode = u.control.newMode
			u.ackButtons()
			u.control.mu.Unlock()
			return
		case OnePulse:
			// 1a (paused) 1p should wait for b p
			if !(u.mode == OneAdd && u.phase == 0 && u.control.buttonsPending == 0) {
				u.mode = u.control.newMode
				u.control.mu.Unlock()
				return
			}
		case OneAdd:
			// 1p (paused on phase 0) 1a should wait for b p
			if !(u.mode == OnePulse && u.phase == 0 && u.control.buttonsPending == 0) {
				u.addCycleMu.Lock()
				u.addCycle = 0
				u.addCycleMu.Unlock()
				u.mode = u.control.newMode
				u.control.mu.Unlock()
				return
			}
		}
		u.control.mu.Unlock()
	}
	panic("infinite loop")
}

func (u *Cycle) ackButtons() {
	for u.control.buttonsPending > 0 {
		u.Io.CycleButton.Done <- 1
		u.control.buttonsPending--
	}
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
		default:
			fmt.Println("cycle unit op switch value: one of 1p, 1a, co")
		}
	default:
		fmt.Println("cycle unit switch: s cy.op.val")
	}
	return defaultValue
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

type opSwitch struct {
	cycle *Cycle
}

func (s *opSwitch) Get() string {
	s.cycle.control.mu.Lock()
	defer s.cycle.control.mu.Unlock()
	switch s.cycle.control.newMode {
	case OnePulse:
		return "1p"
	case OneAdd:
		return "1a"
	case Continuous:
		return "co"
	}
	return "?"
}

func (s *opSwitch) Set(value string) error {
	s.cycle.switches <- [2]string{"op", value}
	return nil
}

func (u *Cycle) FindSwitch(name string) (Switch, error) {
	if name == "op" {
		return &opSwitch{u}, nil
	}
	return nil, fmt.Errorf("unknown switch %s", name)
}
