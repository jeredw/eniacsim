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

// Handshake sends val on channel ch and then waits for an acknowledgement on
// channel resp
func Handshake(val int, ch chan Pulse, resp chan int) {
	if ch != nil {
		ch <- Pulse{val, resp}
		<-resp
	}
}

// Tee returns a new channel which merges output from channels a and b.
// If it is so wired, values sent into the output channel are also reflected
// back to channels a and b.  Tee is used to simulate the wired or behavior of
// ENIAC program and data trunks.
func Tee(a, b chan Pulse) chan Pulse {
	var t = make(chan Pulse)
	go func() {
		for {
			select {
			case pa := <-a:
				if pa.Val != 0 {
					t <- pa
				}
			case pb := <-b:
				if pb.Val != 0 {
					t <- pb
				}
			case pt := <-t:
				if pt.Val != 0 {
					var pt2 Pulse
					if a != nil {
						pt2.Resp = make(chan int)
						pt2.Val = pt.Val
						a <- pt2
						<-pt2.Resp
					}
					if b != nil {
						pt2.Resp = make(chan int)
						pt2.Val = pt.Val
						b <- pt2
						<-pt2.Resp
					}
					pt.Resp <- 1
				}
			}
		}
	}()
	return t
}
