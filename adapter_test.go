package main

import (
	. "github.com/jeredw/eniacsim/lib"
	"testing"
)

func send(digits [11]int, ch chan Pulse, resp chan int) {
	for j := 1; j <= 9; j++ {
		val := 0
		for i := 0; i < 11; i++ {
			if digits[10-i] >= j {
				val |= 1 << i
			}
		}
		if val != 0 {
			ch <- Pulse{val, nil}
		}
	}
	ch <- Pulse{0, resp}
}

func receive(digits *[11]int, ch chan Pulse, resp chan int, done chan int) {
	for {
		select {
		case p := <-ch:
			for i := 0; i < 11; i++ {
				if p.Val&(1<<i) != 0 {
					digits[10-i]++
				}
			}
		case <-resp:
			done <- 1
			return
		}
	}
}

func testShifter(digits [11]int, amount int) [11]int {
	sent := make(chan int)
	done := make(chan int)
	s := shifter{
		in:     make(chan Pulse),
		out:    make(chan Pulse),
		amount: amount,
	}
	go s.run()
	go send(digits, s.in, sent)
	result := [11]int{}
	go receive(&result, s.out, sent, done)
	<-done
	return result
}

func testDeleter(digits [11]int, digit int) [11]int {
	sent := make(chan int)
	done := make(chan int)
	d := deleter{
		in:    make(chan Pulse),
		out:   make(chan Pulse),
		digit: digit,
	}
	go d.run()
	go send(digits, d.in, sent)
	result := [11]int{}
	go receive(&result, d.out, sent, done)
	<-done
	return result
}

func testPermuter(digits [11]int, order [11]int) [11]int {
	sent := make(chan int)
	done := make(chan int)
	p := permuter{
		in:    make(chan Pulse),
		out:   make(chan Pulse),
		order: order,
	}
	go p.run()
	go send(digits, p.in, sent)
	result := [11]int{}
	go receive(&result, p.out, sent, done)
	<-done
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
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9},
		[11]int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
	want := [11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9}
	if result != want {
		t.Errorf("permute(x,id) = %v; want %v", result, want)
	}
}

func TestPermuterReverse(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9},
		[11]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	want := [11]int{9, 0, 3, 5, 7, 6, 8, 5, 5, 5, 9}
	if result != want {
		t.Errorf("permute(x,rev) = %v; want %v", result, want)
	}
}

func TestPermuterDel(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9},
		[11]int{0, 10, 0, 8, 0, 6, 0, 4, 0, 2, 0})
	want := [11]int{0, 5, 0, 5, 0, 6, 0, 5, 0, 0, 0}
	if result != want {
		t.Errorf("permute(x,rev) = %v; want %v", result, want)
	}
}

func TestPermuterSwap2(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9},
		[11]int{11, 2, 1, 4, 3, 6, 5, 8, 7, 10, 9})
	want := [11]int{9, 0, 9, 5, 3, 6, 7, 5, 8, 5, 5}
	if result != want {
		t.Errorf("permute(x,swap2s) = %v; want %v", result, want)
	}
}

func TestPermuterDup2(t *testing.T) {
	result := testPermuter([11]int{9, 5, 5, 5, 8, 6, 7, 5, 3, 0, 9},
		[11]int{11, 10, 10, 9, 9, 8, 8, 7, 7, 6, 6})
	want := [11]int{9, 5, 5, 5, 5, 5, 5, 8, 8, 6, 6}
	if result != want {
		t.Errorf("permute(x,dup2) = %v; want %v", result, want)
	}
}
