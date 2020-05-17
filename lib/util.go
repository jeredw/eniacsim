package lib

func ToBin(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func MsToAddCycles(ms int) int {
	return ms * 5000 / 1000
}

// A TraceFunc records the value of a signal at the current simulation timestep.
type TraceFunc func(string, int, int64)

// Cleared units receive clear signals from the initiate unit.
type Cleared interface {
	Clear()
}
