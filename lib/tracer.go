package lib

type Tracer interface {
	// AdvanceTimestep ticks the trace time.
	AdvanceTimestep()
	// UpdateValues polls and dumps all registered values.
	UpdateValues()

	// RegisterValueCallback registers a function to log a value when
	// UpdateValues is called.
	RegisterValueCallback(update func())

	// LogValue records the instantaneous value of a register.
	LogValue(signalName string, bits int, value int64)
	// LogPulse records a momentary pulse on a signal which implicitly returns to
	// 0 at the next time step.
	LogPulse(signalName string, bits int, value int64)
}
