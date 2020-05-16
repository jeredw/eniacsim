package lib

// Pulse is a type of signal sent on the cycle trunk.
type Pulse int

// A ClockFunc responds to a pulse from the cycle unit.
type ClockFunc func(Pulse)

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
