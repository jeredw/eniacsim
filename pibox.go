package main

import (
	"fmt"
	"os"
	"time"
)

//
//  This code assumes that we're running on a Raspberry Pi
// with Linux.  We also assume that the necessary exports
// have already been done.
//
func ctlstation() {
	fd5, err := os.Open("/sys/class/gpio/gpio5/value")
	if err != nil {
		return
	}
	fd6, _ := os.Open("/sys/class/gpio/gpio6/value")
	fd13, _ := os.Open("/sys/class/gpio/gpio13/value")
	fd19, _ := os.Open("/sys/class/gpio/gpio19/value")
	fd26, _ := os.Open("/sys/class/gpio/gpio26/value")
	fd21, _ := os.Open("/sys/class/gpio/gpio21/value")
	fd20, _ := os.Open("/sys/class/gpio/gpio20/value")

	buf := make([]byte, 1)

	curstate := 0
	filterset := 0
	filtercnt := 0
	// Seriously ugly hack to give other goprocs time to get initialized
	time.Sleep(100 * time.Millisecond)
	for {
		time.Sleep(10 * time.Millisecond)
		newstate := 0
		n, err := fd5.ReadAt(buf, 0)
		if n != 1 {
			fmt.Println(err)
		}
		if buf[0] == '0' {
			newstate |= 0x02
		}
		fd6.ReadAt(buf, 0)
		if buf[0] == '0' {
			newstate |= 0x01
		}
		fd13.ReadAt(buf, 0)
		if buf[0] == '0' {
			newstate |= 0x40
		}
		fd19.ReadAt(buf, 0)
		if buf[0] == '0' {
			newstate |= 0x20
		}
		fd26.ReadAt(buf, 0)
		if buf[0] == '0' {
			newstate |= 0x10
		}
		fd21.ReadAt(buf, 0)
		if buf[0] == '0' {
			newstate |= 0x04
		}
		fd20.ReadAt(buf, 0)
		if buf[0] == '0' {
			newstate |= 0x08
		}
		if newstate != filterset || newstate&0x70 == 0 {
			filtercnt = 0
			filterset = newstate
		} else {
			filtercnt++
		}
		if filtercnt == 4 {
			if newstate != curstate {
				diff := newstate ^ curstate
				if diff&0x70 != 0 {
					switch newstate & 0x70 {
					case 0x10:
						doCommand(os.Stdout, "s cy.op 1a")
					case 0x20:
						doCommand(os.Stdout, "s cy.op 1p")
					case 0x60:
						doCommand(os.Stdout, "s cy.op co")
					}
				}
				if diff&0x01 != 0 && newstate&0x01 != 0 {
					doCommand(os.Stdout, "b c")
				}
				if diff&0x02 != 0 && newstate&0x02 != 0 {
					doCommand(os.Stdout, "b r")
				}
				if diff&0x04 != 0 && newstate&0x04 != 0 {
					doCommand(os.Stdout, "b i")
				}
				if diff&0x08 != 0 && newstate&0x08 != 0 {
					doCommand(os.Stdout, "b p")
				}
				curstate = newstate
			}
		}
	}
}
