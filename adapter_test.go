package main

import (
	"testing"

	. "github.com/jeredw/eniacsim/lib"
)

func send(digits [11]int, jack *Jack) {
	for j := 1; j <= 9; j++ {
		val := 0
		for i := 0; i < 11; i++ {
			if digits[10-i] >= j {
				val |= 1 << i
			}
		}
		if val != 0 {
			jack.Transmit(val)
		}
	}
}

func receive(digits *[11]int, val int) {
	for i := 0; i < 11; i++ {
		if val&(1<<i) != 0 {
			digits[10-i]++
		}
	}
}

func testShifter(digits [11]int, amount int) [11]int {
	result := [11]int{}
	s := &shifter{amount: amount}
	s.in = NewInput("i", func(j *Jack, val int) {
		s.adapt(val)
	})
	s.out = NewOutput("o", nil)
	testSource := NewOutput("to", nil)
	testSink := NewInput("ti", func(j *Jack, val int) {
		receive(&result, val)
	})
	Connect(NewRatsNest(), testSource, s.in)
	Connect(NewRatsNest(), s.out, testSink)
	send(digits, testSource)
	return result
}

func testDeleter(digits [11]int, digit int) [11]int {
	result := [11]int{}
	d := &deleter{digit: digit}
	d.in = NewInput("i", func(j *Jack, val int) {
		d.adapt(val)
	})
	d.out = NewOutput("o", nil)
	testSource := NewOutput("to", nil)
	testSink := NewInput("ti", func(j *Jack, val int) {
		receive(&result, val)
	})
	Connect(NewRatsNest(), testSource, d.in)
	Connect(NewRatsNest(), d.out, testSink)
	send(digits, testSource)
	return result
}

func testPermuter(digits [11]int, order string) [11]int {
	result := [11]int{}
	p := &permuter{}
	s := &permuteSwitch{ad: p}
	s.Set(order)
	p.in = NewInput("i", func(j *Jack, val int) {
		p.adapt(p, val)
	})
	p.out = NewOutput("o", nil)
	testSource := NewOutput("to", nil)
	testSink := NewInput("ti", func(j *Jack, val int) {
		receive(&result, val)
	})
	Connect(NewRatsNest(), testSource, p.in)
	Connect(NewRatsNest(), p.out, testSink)
	send(digits, testSource)
	return result
}

func TestShiftM3AgXI4(t *testing.T) {
	// "M 4 823 000 000 is received through a -3 shifter as "M 9 994 823 000".
	result := testShifter([11]int{9, 4, 8, 2, 3, 0, 0, 0, 0, 0, 0}, -3)
	want := [11]int{9, 9, 9, 9, 4, 8, 2, 3, 0, 0, 0}
	if result != want {
		t.Errorf("shift(x,-3) = %v; want %v", result, want)
	}
}

func TestShiftP5(t *testing.T) {
	result := testShifter([11]int{9, 0, 0, 0, 0, 0, 0, 0, 0, 4, 2}, +5)
	want := [11]int{9, 0, 0, 0, 4, 2, 0, 0, 0, 0, 0}
	if result != want {
		t.Errorf("shift(x,+5) = %v; want %v", result, want)
	}
}

func TestDeleterP2(t *testing.T) {
	result := testDeleter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, +2)
	want := [11]int{9, 5, 5, 0, 0, 0, 0, 0, 0, 0, 0}
	if result != want {
		t.Errorf("delete(x,+2) = %v; want %v", result, want)
	}
}

func TestDeleterP1(t *testing.T) {
	result := testDeleter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, +1)
	want := [11]int{9, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if result != want {
		t.Errorf("delete(x,+1) = %v; want %v", result, want)
	}
}

func TestDeleter0(t *testing.T) {
	result := testDeleter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, 0)
	want := [11]int{9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if result != want {
		t.Errorf("delete(x,0) = %v; want %v", result, want)
	}
}

func TestDeleterM1(t *testing.T) {
	result := testDeleter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, -1)
	want := [11]int{0, 0, 5, 5, 8, 6, 7, 5, 3, 0, 9}
	if result != want {
		t.Errorf("delete(x,-1) = %v; want %v", result, want)
	}
}

func TestPermuterIdentity(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, "11,10,9,8,7,6,5,4,3,2,1")
	want := [11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}
	if result != want {
		t.Errorf("permute(x,id) = %v; want %v", result, want)
	}
}

func TestPermuterReverse(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, "1,2,3,4,5,6,7,8,9,10,11")
	want := [11]int{9, 0, 3, 5, 7, 6, 8, 5, 5, 5, 9}
	if result != want {
		t.Errorf("permute(x,rev) = %v; want %v", result, want)
	}
}

func TestPermuterDel(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, "0,10,0,8,0,6,0,4,0,2,0")
	want := [11]int{0, 5, 0, 5, 0, 6, 0, 5, 0, 0, 0}
	if result != want {
		t.Errorf("permute(x,rev) = %v; want %v", result, want)
	}
}

func TestPermuterSwap2(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, "11,2,1,4,3,6,5,8,7,10,9")
	want := [11]int{9, 0, 9, 5, 3, 6, 7, 5, 8, 5, 5}
	if result != want {
		t.Errorf("permute(x,swap2s) = %v; want %v", result, want)
	}
}

func TestPermuterDup2(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}, "11,10,10,9,9,8,8,7,7,6,6")
	want := [11]int{9, 5, 5, 5, 5, 5, 5, 8, 8, 6, 6}
	if result != want {
		t.Errorf("permute(x,dup2) = %v; want %v", result, want)
	}
}

var testShift = [...]uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
var benchResult int

func BenchmarkShift0(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		s := i << 11
		for i := 0; i < 11; i++ {
			r |= (s >> testShift[i]) & (1 << i)
		}
	}
	benchResult = r
}

func BenchmarkShift1(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		s := i << 11
		r = ((s >> testShift[0]) & (1 << 0)) |
			((s >> testShift[1]) & (1 << 1)) |
			((s >> testShift[2]) & (1 << 2)) |
			((s >> testShift[3]) & (1 << 3)) |
			((s >> testShift[4]) & (1 << 4)) |
			((s >> testShift[5]) & (1 << 5)) |
			((s >> testShift[6]) & (1 << 6)) |
			((s >> testShift[7]) & (1 << 7)) |
			((s >> testShift[8]) & (1 << 8)) |
			((s >> testShift[9]) & (1 << 9)) |
			((s >> testShift[10]) & (1 << 10))
	}
	benchResult = r
}

func BenchmarkShift2(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		s := i << 11
		r = ((s >> (testShift[0] & 63)) & (1 << 0)) |
			((s >> (testShift[1] & 63)) & (1 << 1)) |
			((s >> (testShift[2] & 63)) & (1 << 2)) |
			((s >> (testShift[3] & 63)) & (1 << 3)) |
			((s >> (testShift[4] & 63)) & (1 << 4)) |
			((s >> (testShift[5] & 63)) & (1 << 5)) |
			((s >> (testShift[6] & 63)) & (1 << 6)) |
			((s >> (testShift[7] & 63)) & (1 << 7)) |
			((s >> (testShift[8] & 63)) & (1 << 8)) |
			((s >> (testShift[9] & 63)) & (1 << 9)) |
			((s >> (testShift[10] & 63)) & (1 << 10))
	}
	benchResult = r
}
