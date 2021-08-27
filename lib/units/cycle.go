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
	checkpointInput *Jack

	checkpoint      bool // true if at instruction-level sim checkpoint
	crossValidate   bool // if true, cross-validate with vm, else run ahead

	mode            int
	AddCycle        int64
	targetAddCycle  int64
	phase           int
	stop            bool
	tracer          Tracer
}

// CycleConn defines connections needed for the cycle unit
type CycleConn struct {
	Units           *ClockedUnits // Clocked units

	StepAndVerifyVM func()        // Cross-validate checkpoint with vm
	StepAheadVM     func(n int64) // Step ahead up to cycle n with vm
	SelectiveClear  func() bool   // Clear gate (from initiate unit)
}

// Clock operating modes
const (
	OnePulse   = iota // Run for one pulse
	OneAdd            // Run for one add cycle
	Continuous        // Run continuously
	Test              // Running in test mode (debugger doesn't actually stop)
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
	u := &Cycle{Io: io}
	u.checkpointInput = NewInput("cy.checkpoint", func(*Jack, int) {
		u.checkpoint = true
	})
	return u
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
func (u *Cycle) Stat() string {
	if u.phase >= len(phases) {
		return "0"
	} else {
		return fmt.Sprintf("%d", u.phase)
	}
}

//go:nosplit
func (u *Cycle) sendPulse(pulse Pulse) {
	uu := u.Io.Units
	uu.Initiate.Clock(pulse)
	uu.Mp.Clock(pulse)
	uu.Divsr.Clock(pulse)
	uu.Multiplier.Clock(pulse)
	uu.Constant.Clock(pulse)
	//for i := range uu.Ft {
	//	uu.Ft[i].Clock(pulse)
	//}
	uu.Ft[0].Clock(pulse)
	uu.Ft[1].Clock(pulse)
	uu.Ft[2].Clock(pulse)
	//u.Io.Units.TenStepper.Clock(pulse)
	//for i := range uu.Accumulator {
	//	uu.Accumulator[i].Clock(pulse)
	//}
	uu.Accumulator[0].Clock(pulse)
	uu.Accumulator[1].Clock(pulse)
	uu.Accumulator[2].Clock(pulse)
	uu.Accumulator[3].Clock(pulse)
	uu.Accumulator[4].Clock(pulse)
	uu.Accumulator[5].Clock(pulse)
	uu.Accumulator[6].Clock(pulse)
	uu.Accumulator[7].Clock(pulse)
	uu.Accumulator[8].Clock(pulse)
	uu.Accumulator[9].Clock(pulse)
	uu.Accumulator[10].Clock(pulse)
	uu.Accumulator[11].Clock(pulse)
	uu.Accumulator[12].Clock(pulse)
	uu.Accumulator[13].Clock(pulse)
	uu.Accumulator[14].Clock(pulse)
	uu.Accumulator[15].Clock(pulse)
	uu.Accumulator[16].Clock(pulse)
	uu.Accumulator[17].Clock(pulse)
	uu.Accumulator[18].Clock(pulse)
	uu.Accumulator[19].Clock(pulse)
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
		if u.checkpoint {
			if u.crossValidate {
				u.Io.StepAndVerifyVM()
			} else if u.mode == Continuous {
				u.Io.StepAheadVM(u.targetAddCycle)
			}
			u.checkpoint = false
		}
	}
}

// Returns true if stopped by debugger
func (u *Cycle) StepNAddCycles(n int) bool {
	start := u.AddCycle;
	u.targetAddCycle = start + int64(n)
	for u.AddCycle < u.targetAddCycle {
		u.StepOnePulse()
		if u.stop {
			// If debugger requested a stop, step to start of next add cycle
			for u.phase != 0 {
				u.StepOnePulse()
			}
			// Probably we'll want to single-step add cycles after a breakpoint
			u.mode = OneAdd
			u.stop = false
			return true
		}
	}
	return false
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

func (u *Cycle) SetTestMode() {
	u.mode = Test
}

func (u *Cycle) Stop() {
	if u.mode != Test {
		u.stop = true
	}
}

func modeSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"1p", OnePulse}, {"1P", OnePulse},
		{"1a", OneAdd}, {"1A", OneAdd},
		{"co", Continuous}, {"CO", Continuous},
	}
}

func vmSettings() []BoolSwitchSetting {
	return []BoolSwitchSetting{
		{"check", true},
		{"run", false},
	}
}

func (u *Cycle) FindSwitch(name string) (Switch, error) {
	if name == "op" {
		return &IntSwitch{name, &u.mode, modeSettings()}, nil
	}
	if name == "vm" {
		return &BoolSwitch{name, &u.crossValidate, vmSettings()}, nil
	}
	return nil, fmt.Errorf("unknown switch %s", name)
}

func (u *Cycle) FindJack(name string) (*Jack, error) {
	if name == "checkpoint" {
		return u.checkpointInput, nil
	}
	return nil, fmt.Errorf("unknown jack %s", name)
}
