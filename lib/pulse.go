package lib

type Pulse struct {
	Val  int
	Resp chan int
}

// A ClockFunc responds to a pulse from the cycle unit.
type ClockFunc func(Pulse)

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
