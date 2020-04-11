package lib

type Pulse struct {
	Val int
	Resp chan int
}

type ClockFunc func(Pulse)
