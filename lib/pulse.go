package lib

type Pulse struct {
	Val  int
	Resp chan int
}

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
