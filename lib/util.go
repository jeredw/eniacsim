package lib

func MsToAddCycles(ms int64) int64 {
	return ms * 5000 / 1000
}

// Cleared units receive clear signals from the initiate unit.
type Cleared interface {
	Clear()
}
