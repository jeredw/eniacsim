package lib

// Pulse is a type of signal sent on the cycle trunk.
type Pulse int

// Clocked things receive a clock pulse on the cycle trunk.
type Clocked interface {
	Clock(Pulse)
}

// Cleared things receive clear signals from the initiate unit.
type Cleared interface {
	Clear()
}

// A TraceFunc records the value of a signal at the current simulation timestep.
type TraceFunc func(string, int, int64)

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
