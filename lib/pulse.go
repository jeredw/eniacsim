package lib

// Pulse is a signal sent on the synchronizing trunk.
type Pulse int

// Kinds of pulses (most descriptions from Technical Manual Part 2, Table 3-1.)
const (
	// A program pulse used to control tbe activity of the various units of the
	// ENIAC.
	Cpp = 1 << iota
	// Used to cycle the decades of accumulator during the process of
	// transmission of the number (or its complement) registered in the
	// accumulator.
	Tenp
	// A coded system in the multiplier, function tables, and the constant
	// transmitter makes use of combinations of these pulses to represent the
	// digits zero to nine.
	Onep
	Twop
	Twopp
	Fourp
	// Some of these pulses are used to represent the digits zero to nine.
	Ninep
	// In the process of taking the complement this pulse is used to obtain the
	// complement with respect to 10^n instead of 10^n-1.
	Onepp
	// This pulse is used to reset flip-flops in the accumulator decade units and
	// to provide a carry-over pulse in the process of addition.
	Rp
	// This gate controls the carry over process when adding in an accumulator
	// and produces the clearing action when so desired.
	Ccg
	// Sent by the initiate unit to clear any of the twenty accumulators
	// depending upon the settings of their selective clear switches.
	Scg
)
