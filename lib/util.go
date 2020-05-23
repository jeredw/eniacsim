package lib

func MsToAddCycles(ms int) int {
	return ms * 5000 / 1000
}

// Cleared units receive clear signals from the initiate unit.
type Cleared interface {
	Clear()
}
