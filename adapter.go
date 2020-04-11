package main

import (
	. "github.com/jeredw/eniacsim/lib"
)

var dpin [40]chan Pulse
var dpout [40][11]chan Pulse
var shiftin [40]chan Pulse
var shiftout [40]chan Pulse
var delin [40]chan Pulse
var delout [40]chan Pulse
var sdin [40]chan Pulse
var sdout [40]chan Pulse

func adreset() {
	for i := 0; i < 40; i++ {
		dpin[i] = nil
		for j := 0; j < 11;j ++ {
			dpout[i][j] = nil
		}
		shiftin[i] = nil
		shiftout[i] = nil
		delin[i] = nil
		delout[i] = nil
		sdin[i] = nil
		sdout[i] = nil
	}
}

func adplug(ilk string, inout, which, param int, ch chan Pulse) {
	switch ilk {
	case "dp":
		if inout == 0 {
			dpin[which] = ch
			go digitprog(dpin[which], which)
		} else {
			dpout[which][param-1] = ch
		}
	case "s":
		if inout == 0 {
			shiftin[which] = ch
		} else {
			shiftout[which] = ch
		}
		if shiftin[which] != nil && shiftout[which] != nil {
			go shifter(shiftin[which], shiftout[which], param)
		}
	case "d":
		if inout == 0 {
			delin[which] = ch
		} else {
			delout[which] = ch
		}
		if delin[which] != nil && delout[which] != nil {
			go deleter(delin[which], delout[which], param)
		}
	case "sd":
		if inout == 0 {
			sdin[which] = ch
		} else {
			sdout[which] = ch
		}
		if sdin[which] != nil && sdout[which] != nil {
			go specdig(sdin[which], sdout[which], uint(param))
		}
	}
}

func digitprog(in chan Pulse, which int) {
	resp := make(chan int)
	for {
		d :=<- in
		for i := uint(0); i < 11; i++ {
			if d.Val & (1 << i) != 0 &&  dpout[which][i] != nil {
				dpout[which][i] <- Pulse{1, resp}
				<- resp
			}
		}
		if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

func shifter(in, out chan Pulse, shift int) {
	for {
		d :=<- in
		if shift >= 0 {
			d.Val = (d.Val & (1 << 10)) | ((d.Val  << uint(shift)) & ((1 << 10) - 1))
		} else {
			x := d.Val >> uint(-shift)
			if d.Val & (1 << 10) != 0 {
				d.Val = x | (((1 << 11) - 1) & ^((1 << uint(11 + shift)) - 1))
			} else {
				d.Val = x
			}
		}
		if d.Val != 0 {
			out <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

func deleter(in, out chan Pulse, which int) {
	for {
		d :=<- in
		if which >= 0 {
			d.Val &= ^((1 << uint(10 - which)) - 1)
		} else {
			d.Val &= (1 << uint(10 + which)) - 1
		}
		if d.Val != 0 {
			out <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}

func specdig(in, out chan Pulse, which uint) {
	for {
		d :=<- in
		x := d.Val >> which
		mask := 0x07fc
		if d.Val & (1 << 10) != 0 {
			d.Val = x | mask
		} else {
			d.Val = x & ^mask
		}
		if d.Val != 0 && out != nil {
			out <- d
		} else if d.Resp != nil {
			d.Resp <- 1
		}
	}
}
