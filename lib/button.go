package lib

// Button is a synchronized control input
// Usage: b.Push <- command, <-b.Done
type Button struct {
	Push chan int
	Done chan int
}

func NewButton() Button {
	return Button{Push: make(chan int), Done: make(chan int)}
}
