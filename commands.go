package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	. "github.com/jeredw/eniacsim/lib"
	"github.com/jeredw/eniacsim/lib/units"
)

func doCommand(command string) int {
	f := strings.Fields(command)
	for i, s := range f {
		if s[0] == '#' {
			f = f[:i]
			break
		}
	}
	if len(f) == 0 {
		return 0
	}
	switch f[0] {
	case "b":
		doButton(f)
	case "d":
		doDump(f)
	case "D":
		doDumpAll()
	case "f":
		doFile(f)
	case "l":
		doLoad(f)
	case "n":
		cycle.Io.CycleButton.Push <- 1
		<-cycle.Io.CycleButton.Done
		doDumpAll()
	case "p":
		doPlug(command, f)
	case "q":
		return -1
	case "r":
		doReset(f)
	case "R":
		doResetAll()
	case "s":
		doSwitch(command, f)
	case "set":
		doSet(f)
	case "u":
	case "dt":
	case "pt":
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
	return 0
}

func doButton(f []string) {
	if len(f) != 2 {
		fmt.Println("button syntax: b button")
		return
	}
	switch f[1] {
	case "c":
		initiate.Io.InitButton.Push <- 5
		<-initiate.Io.InitButton.Done
	case "i":
		initiate.Io.InitButton.Push <- 4
		<-initiate.Io.InitButton.Done
	case "p":
		cycle.Io.CycleButton.Push <- 1
		<-cycle.Io.CycleButton.Done
	case "r":
		initiate.Io.InitButton.Push <- 3
		<-initiate.Io.InitButton.Done
	}
}

func doDump(f []string) {
	if len(f) != 2 {
		fmt.Println("Status syntax: d unit")
		return
	}
	switch f[1][0] {
	case 'a':
		unit, _ := strconv.Atoi(f[1][1:])
		fmt.Println(units.Accstat(unit - 1))
	case 'b':
		fmt.Println(debugstat())
	case 'c':
		fmt.Println(constant.Stat())
	case 'd':
		fmt.Println(divsr.Stat2())
	case 'f':
		unit, _ := strconv.Atoi(f[1][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Println("Invalid function table")
			return
		}
		fmt.Println(ft[unit-1].Stat())
	case 'i':
		fmt.Println(initiate.Stat())
	case 'm':
		fmt.Println(multiplier.Stat())
	case 'p':
		fmt.Println(mp.Stat())
	}
}

func doDumpAll() {
	fmt.Println()
	fmt.Println(initiate.Stat())
	fmt.Println(mp.Stat())
	acchdr := "      9876543210 9876543210 r 123456789012"
	fmt.Printf("%s   %s\n", acchdr, acchdr)
	for i := 0; i < 20; i += 2 {
		fmt.Print(units.Accstat(i))
		fmt.Print("   ")
		fmt.Println(units.Accstat(i + 1))
	}
	fmt.Println(divsr.Stat2())
	fmt.Println(multiplier.Stat())
	for i := 0; i < 3; i++ {
		fmt.Println(ft[i].Stat())
	}
	fmt.Println(constant.Stat())
	fmt.Println()
}

func doFile(f []string) {
	if len(f) != 3 {
		fmt.Println("file syntax: f (r|p) filename")
		return
	}
	switch f[1] {
	case "r":
		fp, err := os.Open(f[2])
		if err != nil {
			fmt.Printf("Card reader open: %s\n", err)
			return
		}
		initiate.SetCardScanner(bufio.NewScanner(fp))
	case "p":
		fp, err := os.Create(f[2])
		if err != nil {
			fmt.Printf("Card punch open: %s\n", err)
			return
		}
		initiate.SetPunchWriter(bufio.NewWriter(fp))
	}
}

func doLoad(f []string) {
	if len(f) != 2 {
		fmt.Println("Load syntax: l file")
		return
	}
	fd, err := os.Open(f[1])
	if err != nil {
		fd, err = os.Open("programs/" + f[1])
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		if doCommand(sc.Text()) < 0 {
			return
		}
	}
	fd.Close()
}

func doPlug(command string, f []string) {
	if len(f) != 3 {
		fmt.Println("Invalid jumper spec", command)
		return
	}
	p1 := strings.Split(f[1], ".")
	p2 := strings.Split(f[2], ".")
	/*
	 * Ugly special case of 20 digit interconnects
	 */
	if len(p1) == 2 && p1[0][0] == 'a' && len(p1[1]) >= 2 &&
		(p1[1][:2] == "st" || p1[1][:2] == "su" ||
			p1[1][:2] == "il" || p1[1][:2] == "ir") {
		units.Accinterconnect(p1, p2)
		return
	}
	ch := make(chan Pulse)
	doPlugSide(0, command, f, p1, ch)
	doPlugSide(1, command, f, p2, ch)
}

func doPlugSide(side int, command string, f []string, p []string, ch chan Pulse) {
	switch {
	case p[0] == "ad":
		if len(p) != 4 {
			fmt.Println("Adapter jumper syntax: ad.ilk.unit.param")
			return
		}
		unit, _ := strconv.Atoi(p[2])
		param, _ := strconv.Atoi(p[3])
		adplug(p[1], 1-side, unit-1, param, ch)
	case p[0][0] == 'a':
		if len(p) != 2 {
			fmt.Println("Accumulator jumper syntax: aunit.terminal")
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		units.Accplug(unit-1, p[1], ch)
	case p[0] == "c":
		if len(p) != 2 {
			fmt.Println("Invalid constant jumper:", command)
			return
		}
		err := constant.Plug(p[1], ch)
		if err != nil {
			fmt.Printf("Constant: %s\n", err)
		}
	case p[0] == "d":
		if len(p) != 2 {
			fmt.Println("Divider jumper syntax: d.terminal")
			return
		}
		err := divsr.Plug(p[1], ch)
		if err != nil {
			fmt.Printf("Divider: %s\n", err)
		}
	case p[0] == "debug":
		if side == 1 {
			if len(p) != 2 {
				fmt.Println("Debugger jumper syntax: debug.bpn")
				return
			}
			unit, _ := strconv.Atoi(p[1][2:])
			debugplug(unit, ch, f[1])
		}
	case p[0][0] == 'f':
		if len(p) != 2 {
			fmt.Println("Function table jumper syntax: funit.terminal")
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Println("Invalid function table")
			return
		}
		err := ft[unit-1].Plug(p[1], ch)
		if err != nil {
			fmt.Printf("Function table %d: %s", unit, err)
		}
	case p[0] == "i":
		if len(p) != 2 {
			fmt.Println("Initiator jumper syntax: i.terminal")
			return
		}
		initiate.Plug(p[1], ch)
	case p[0] == "m":
		if len(p) != 2 {
			fmt.Println("Multiplier jumper syntax: m.terminal")
			return
		}
		err := multiplier.Plug(p[1], ch)
		if err != nil {
			fmt.Printf("Multiplier: %s\n", err)
		}
	case p[0] == "p":
		err := mp.Plug(p[1], ch)
		if err != nil {
			fmt.Printf("Programmer: %s\n", err)
		}
	case unicode.IsDigit(rune(p[0][0])):
		hpos := strings.IndexByte(p[0], '-')
		if hpos == -1 {
			tray, _ := strconv.Atoi(p[0])
			if tray < 1 {
				fmt.Println("Invalid data trunk", p[0])
				return
			}
			if side == 1 {
				trunkxmit(0, tray-1, ch)
			} else {
				trunkrecv(0, tray-1, ch)
			}
		} else {
			tray, _ := strconv.Atoi(p[0][:hpos])
			line, _ := strconv.Atoi(p[0][hpos+1:])
			if side == 1 {
				trunkxmit(1, (tray-1)*11+line-1, ch)
			} else {
				trunkrecv(1, (tray-1)*11+line-1, ch)
			}
		}
	default:
		fmt.Println("Invalid jack spec: ", p)
	}
}

func doReset(f []string) {
	if len(f) != 2 {
		fmt.Println("Status syntax: r unit")
		return
	}
	p := strings.Split(f[1], ".")
	switch p[0] {
	case "a":
		if len(p) != 2 {
			fmt.Println("Accumulator reset syntax: r a.unit")
		} else {
			unit, _ := strconv.Atoi(p[1])
			units.Accreset(unit)
		}
	case "b":
		debugreset()
	case "c":
		constant.Reset()
	case "d":
		divsr.Reset()
	case "f":
		if len(p) != 2 {
			fmt.Println("Function table reset syntax: r f.unit")
		} else {
			unit, _ := strconv.Atoi(p[1])
			if !(unit >= 1 && unit <= 3) {
				fmt.Println("Invalid function table")
				return
			}
			ft[unit-1].Reset()
		}
	case "i":
		initiate.Reset()
	case "m":
		multiplier.Reset()
	case "p":
		mp.Reset()
	}
}

func doResetAll() {
	initiate.Reset()
	cycle.Io.Reset <- 1
	debugreset()
	mp.Reset()
	ft[0].Reset()
	ft[1].Reset()
	ft[2].Reset()
	for i := 0; i < 20; i++ {
		units.Accreset(i)
	}
	divsr.Reset()
	multiplier.Reset()
	constant.Reset()
	printer.Reset()
	adreset()
	trayreset()
}

func doSwitch(command string, f []string) {
	if len(f) < 3 {
		fmt.Println("No switch setting")
		return
	}
	p := strings.Split(f[1], ".")
	switch {
	case p[0][0] == 'a':
		if len(p) != 2 {
			fmt.Println("Invalid accumulator switch:", command)
		} else {
			unit, _ := strconv.Atoi(p[0][1:])
			accsw[unit-1] <- [2]string{p[1], f[2]}
		}
	case p[0] == "c":
		if len(p) != 2 {
			fmt.Println("Constant switch syntax: s c.switch value")
			break
		}
		err := constant.Switch(p[1], f[2])
		if err != nil {
			fmt.Printf("Constant: %s\n", err)
		}
	case p[0] == "cy":
		if len(p) != 2 {
			fmt.Println("Cycling switch syntax: s cy.switch value")
		} else {
			cycle.Io.Switches <- [2]string{p[1], f[2]}
		}
	case p[0] == "d" || p[0] == "ds":
		if len(p) != 2 {
			fmt.Println("Divider switch syntax: s d.switch value")
			break
		}
		err := divsr.Switch(p[1], f[2])
		if err != nil {
			fmt.Printf("Divider: %s\n", err)
		}
	case p[0][0] == 'f':
		if len(p) != 2 {
			fmt.Println("Function table switch syntax: s funit.switch value", command)
			break
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Println("Invalid function table")
			break
		}
		err := ft[unit-1].Switch(p[1], f[2])
		if err != nil {
			fmt.Printf("Function table %d: %s", unit, err)
		}
	case p[0] == "m":
		if len(p) != 2 {
			fmt.Println("Multiplier switch syntax: s m.switch value")
			break
		}
		err := multiplier.Switch(p[1], f[2])
		if err != nil {
			fmt.Printf("Multiplier: %s\n", err)
		}
	case p[0] == "p":
		if len(p) != 2 {
			fmt.Println("Programmer switch syntax: s p.switch value")
			break
		}
		err := mp.Switch(p[1], f[2])
		if err != nil {
			fmt.Printf("Programmer: %s\n", err)
		}
	case p[0] == "pr":
		if len(p) != 2 {
			fmt.Println("Printer switch syntax: s pr.switch value")
			break
		}
		err := printer.Switch(p[1], f[2])
		if err != nil {
			fmt.Printf("Printer: %s\n", err)
		}
	default:
		fmt.Printf("unknown unit for switch: %s\n", p[0])
	}
}

func doSet(f []string) {
	if len(f) != 3 {
		fmt.Println("set syntax: set a13 -9876543210")
		return
	}
	unit, _ := strconv.Atoi(f[1][1:])
	value, _ := strconv.ParseInt(f[2], 10, 64)
	units.Accset(unit-1, value)
}
