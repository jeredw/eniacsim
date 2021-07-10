package units

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
)

type ClockedUnits struct {
	Initiate *Initiate
	Mp *Mp
	Divsr *Divsr
	Multiplier *Multiplier
	Constant *Constant
	Ft [3]*Ft
	Accumulator [20]*Accumulator
	TenStepper *AuxStepper
	FtSelector *AuxStepper
	PmDiscriminator [2]*AuxStepper
	JkSelector [2]*AuxStepper
	OrderSelector *OrderSelector
}

// Cycle simulates ENIAC's clock generation circuits
type Cycle struct {
	Io CycleConn // Connections to other units

	mode       int

	AddCycle   int
	phase      int
	tracer     Tracer
}

// CycleConn defines connections needed for the cycle unit
type CycleConn struct {
	Units          *ClockedUnits // Clocked units
	SelectiveClear func() bool // Clear gate (from initiate unit)
}

// Clock operating modes
const (
	OnePulse   = iota // Run for one pulse, then wait for CycleButton
	OneAdd            // Run for one add cycle, then wait for CycleButton
	Continuous        // Run continuously
	Test              // Running in test mode (debugger doesn't actually stop)
	Stopped           // Stopped by debugger
)

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
		Io: io,
	}
}

// Stepping returns whether the clock is being single stepped
func (u *Cycle) Stepping() bool {
	return u.mode == OneAdd || u.mode == OnePulse
}

// Stopped returns whether the clock is stopped (debugger)
func (u *Cycle) Stopped() bool {
  return u.mode == Stopped
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
func (u *Cycle) Stat() string {
	if u.phase >= len(phases) {
		return "0"
	} else {
		return fmt.Sprintf("%d", u.phase)
	}
}

func (u *Cycle) sendPulse(pulse Pulse) {
	u.Io.Units.Initiate.Clock(pulse)
	u.Io.Units.Mp.Clock(pulse)
	u.Io.Units.Divsr.Clock(pulse)
	u.Io.Units.Multiplier.Clock(pulse)
	u.Io.Units.Constant.Clock(pulse)
	for i := range u.Io.Units.Ft {
		u.Io.Units.Ft[i].Clock(pulse)
	}
	//u.Io.Units.TenStepper.Clock(pulse)
	for i := range u.Io.Units.Accumulator {
		u.Io.Units.Accumulator[i].Clock(pulse)
	}
	//u.Io.Units.FtSelector.Clock(pulse)
	//for i := range u.Io.Units.PmDiscriminator {
	//	u.Io.Units.PmDiscriminator[i].Clock(pulse)
	//}
	//for i := range u.Io.Units.JkSelector {
	//	u.Io.Units.JkSelector[i].Clock(pulse)
	//}
	//u.Io.Units.OrderSelector.Clock(pulse)
}

func (u *Cycle) StepOnePulse() {
	if u.mode == Stopped {
		return
	}
	if u.tracer != nil {
		u.tracer.AdvanceTimestep()
		//u.tracer.LogValue("cyc.pulse", 6, int64(u.phase/2))
	}
	if u.phase == 32 && u.Io.SelectiveClear() {
		u.sendPulse(Scg)
	} else if phases[u.phase] != 0 {
		u.sendPulse(phases[u.phase])
	}
	u.phase++
	if phases[u.phase] != 0 {
		u.sendPulse(phases[u.phase])
	}
	u.phase++
	// Count if add cycle boundary
	if u.phase == len(phases) {
		u.phase = 0
		if u.tracer != nil {
			u.tracer.UpdateValues()
		}
		u.AddCycle++
	}
}

func (u *Cycle) StepNAddCycles(n int) {
	start := u.AddCycle;
	for u.AddCycle < start + n && u.mode != Stopped {
		u.StepOnePulse()
	}
}

func (u *Cycle) StepOneAddCycle() {
	u.StepNAddCycles(1)
}

func (u *Cycle) Step() {
	if u.mode == OnePulse {
		u.StepOnePulse()
	} else if u.mode == OneAdd {
		u.StepOneAddCycle()
	}
}

func (u* Cycle) SetTestMode() {
	u.mode = Test
}

func (u *Cycle) Stop() {
	if u.mode != Test {
		u.mode = Stopped
	}
}

func modeSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"1p", OnePulse}, {"1P", OnePulse},
		{"1a", OneAdd}, {"1A", OneAdd},
		{"co", Continuous}, {"CO", Continuous},
	}
}

func (u *Cycle) FindSwitch(name string) (Switch, error) {
	if name == "op" {
		return &IntSwitch{name, &u.mode, modeSettings()}, nil
	}
	return nil, fmt.Errorf("unknown switch %s", name)
}
