package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	. "github.com/jeredw/eniacsim/lib"
	"github.com/jeredw/eniacsim/lib/units"
)

func doCommand(w io.Writer, command string) int {
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
		doButton(w, f)
	case "d":
		doDump(w, f)
	case "D":
		doDumpAll(w)
	case "f":
		doFile(w, f)
	case "l":
		doLoad(w, f)
	case "n":
		cycle.Io.CycleButton.Push <- 1
		<-cycle.Io.CycleButton.Done
		doDumpAll(w)
	case "p":
		doPlug(w, command, f)
	case "p?":
		doGetPlug(w, command, f)
	case "q":
		return -1
	case "r":
		doReset(w, f)
	case "R":
		doResetAll(w)
	case "s":
		doSetSwitch(w, command, f)
	case "s?":
		doGetSwitch(w, command, f)
	case "set":
		doSet(w, f)
	case "ts":
		doTraceStart(w, f)
	case "te":
		doTraceEnd(w, f)
	case "u":
	case "dt":
	case "pt":
	default:
		fmt.Fprintf(w, "Unknown command: %s\n", command)
	}
	return 0
}

func doButton(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "button syntax: b button")
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

func doDump(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "Status syntax: d unit")
		return
	}
	switch f[1][0] {
	case 'a':
		unit, _ := strconv.Atoi(f[1][1:])
		if !(unit >= 1 && unit <= 20) {
			fmt.Fprintf(w, "Invalid accumulator %s\n", f[1][1:])
			return
		}
		fmt.Fprintln(w, accumulator[unit-1].Stat())
	case 'b':
		fmt.Fprintln(w, debugger.Stat())
	case 'c':
		fmt.Fprintln(w, constant.Stat())
	case 'd':
		fmt.Fprintln(w, divsr.Stat2())
	case 'f':
		unit, _ := strconv.Atoi(f[1][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Fprintf(w, "Invalid function table %s\n", f[1][1:])
			return
		}
		fmt.Fprintln(w, ft[unit-1].Stat())
	case 'i':
		fmt.Fprintln(w, initiate.Stat())
	case 'm':
		fmt.Fprintln(w, multiplier.Stat())
	case 'p':
		fmt.Fprintln(w, mp.Stat())
	}
}

func doDumpAll(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, initiate.Stat())
	fmt.Fprintln(w, mp.Stat())
	header := "      9876543210 9876543210 r 123456789012"
	fmt.Fprintf(w, "%s   %s\n", header, header)
	for i := 0; i < 20; i += 2 {
		ai := accumulator[i].Stat()
		ai1 := accumulator[i+1].Stat()
		fmt.Fprintf(w, "a%-2d %s   a%-2d %s\n", i+1, ai, i+2, ai1)
	}
	fmt.Fprintln(w, divsr.Stat2())
	fmt.Fprintln(w, multiplier.Stat())
	for i := 0; i < 3; i++ {
		fmt.Fprintln(w, ft[i].Stat())
	}
	fmt.Fprintln(w, constant.Stat())
	fmt.Fprintln(w)
}

func doFile(w io.Writer, f []string) {
	if len(f) != 3 {
		fmt.Fprintln(w, "file syntax: f (r|p) filename")
		return
	}
	switch f[1] {
	case "r":
		fp, err := os.Open(f[2])
		if err != nil {
			fmt.Fprintf(w, "Card reader open: %s\n", err)
			return
		}
		initiate.SetCardScanner(bufio.NewScanner(fp))
	case "p":
		fp, err := os.Create(f[2])
		if err != nil {
			fmt.Fprintf(w, "Card punch open: %s\n", err)
			return
		}
		initiate.SetPunchWriter(bufio.NewWriter(fp))
	}
}

func doLoad(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "Load syntax: l file")
		return
	}
	fd, err := os.Open(f[1])
	if err != nil {
		fd, err = os.Open("programs/" + f[1])
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
	}
	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		if doCommand(os.Stdout, sc.Text()) < 0 {
			return
		}
	}
	fd.Close()
}

func doPlug(w io.Writer, command string, f []string) {
	if len(f) != 3 {
		fmt.Fprintln(w, "Invalid jumper spec", command)
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
		// Handle commands like p aXX.{st,su,il,ir} *
		err := units.Interconnect(accumulator, p1, p2)
		if err != nil {
			fmt.Fprintf(w, "Interconnect: %s\n", err)
		}
		return
	}
	wire := NewWire(f[1], f[2])
	doPlugSide(w, 0, command, f, p1, *wire)
	doPlugSide(w, 1, command, f, p2, *wire)
}

func doPlugSide(w io.Writer, side int, command string, f []string, p []string, wire Wire) {
	output := side == 1
	switch {
	case p[0] == "ad":
		if len(p) == 3 {
			err := adapters.Plug(p[1], p[2], "", wire, output)
			if err != nil {
				fmt.Fprintf(w, "Adapter: %s\n", err)
			}
			return
		}
		if len(p) != 4 {
			fmt.Fprintln(w, "Adapter jumper syntax: ad.<type>.<id>.param")
			return
		}
		err := adapters.Plug(p[1], p[2], p[3], wire, output)
		if err != nil {
			fmt.Fprintf(w, "Adapter: %s\n", err)
		}
	case p[0][0] == 'a':
		if len(p) != 2 {
			fmt.Fprintln(w, "Accumulator jumper syntax: aunit.terminal")
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 20) {
			fmt.Fprintf(w, "Invalid accumulator %s\n", p[0][1:])
			return
		}
		err := accumulator[unit-1].Plug(p[1], wire)
		if err != nil {
			fmt.Fprintf(w, "Accumulator %d: %s\n", unit, err)
		}
	case p[0] == "c":
		if len(p) != 2 {
			fmt.Fprintln(w, "Invalid constant jumper:", command)
			return
		}
		err := constant.Plug(p[1], wire)
		if err != nil {
			fmt.Fprintf(w, "Constant: %s\n", err)
		}
	case p[0] == "d":
		if len(p) != 2 {
			fmt.Fprintln(w, "Divider jumper syntax: d.terminal")
			return
		}
		err := divsr.Plug(p[1], wire)
		if err != nil {
			fmt.Fprintf(w, "Divider: %s\n", err)
		}
	case p[0] == "debug":
		if side == 1 {
			if len(p) != 2 {
				fmt.Fprintln(w, "Debugger jumper syntax: debug.bpn")
				return
			}
			err := debugger.Plug(p[1], wire, f[1])
			if err != nil {
				fmt.Fprintf(w, "Debugger: %s\n", err)
			}
		}
	case p[0][0] == 'f':
		if len(p) != 2 {
			fmt.Fprintln(w, "Function table jumper syntax: funit.terminal")
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Fprintf(w, "Invalid function table %s\n", p[0][1:])
			return
		}
		err := ft[unit-1].Plug(p[1], wire)
		if err != nil {
			fmt.Fprintf(w, "Function table %d: %s\n", unit, err)
		}
	case p[0] == "i":
		if len(p) != 2 {
			fmt.Fprintln(w, "Initiator jumper syntax: i.terminal")
			return
		}
		err := initiate.Plug(p[1], wire)
		if err != nil {
			fmt.Fprintf(w, "Initiate: %s\n", err)
		}
	case p[0] == "m":
		if len(p) != 2 {
			fmt.Fprintln(w, "Multiplier jumper syntax: m.terminal")
			return
		}
		err := multiplier.Plug(p[1], wire)
		if err != nil {
			fmt.Fprintf(w, "Multiplier: %s\n", err)
		}
	case p[0] == "p":
		if len(p) != 2 {
			fmt.Fprintln(w, "Programmer jumper syntax: p.terminal")
			return
		}
		err := mp.Plug(p[1], wire)
		if err != nil {
			fmt.Fprintf(w, "Programmer: %s\n", err)
		}
	case unicode.IsDigit(rune(p[0][0])):
		err := trays.Plug(p[0], wire, output)
		if err != nil {
			fmt.Fprintf(w, "Trays: %s\n", err)
		}
	default:
		fmt.Fprintln(w, "Invalid jack spec: ", p)
	}
}

func doGetPlug(w io.Writer, command string, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "Invalid jumper spec", command)
		return
	}
	p := strings.Split(f[1], ".")
	switch {
	case p[0] == "ad":
		if len(p) == 3 {
			wires, err := adapters.GetPlug(p[1], p[2], "")
			if err != nil {
				fmt.Fprintf(w, "Adapter: %s\n", err)
			} else {
				printWires(w, wires)
			}
			return
		}
		if len(p) != 4 {
			fmt.Fprintln(w, "Adapter jumper syntax: ad.<type>.<id>.param")
			return
		}
		wires, err := adapters.GetPlug(p[1], p[2], p[3])
		if err != nil {
			fmt.Fprintf(w, "Adapter: %s\n", err)
		} else {
			printWires(w, wires)
		}
	case p[0][0] == 'a':
		if len(p) != 2 {
			fmt.Fprintln(w, "Accumulator jumper syntax: aunit.terminal")
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 20) {
			fmt.Fprintf(w, "Invalid accumulator %s\n", p[0][1:])
			return
		}
		wire, err := accumulator[unit-1].GetPlug(p[1])
		if err != nil {
			fmt.Fprintf(w, "Accumulator %d: %s\n", unit, err)
		} else {
			fmt.Fprintln(w, wire.ToString())
		}
	case p[0] == "c":
		if len(p) != 2 {
			fmt.Fprintln(w, "Invalid constant jumper:", command)
			return
		}
		wire, err := constant.GetPlug(p[1])
		if err != nil {
			fmt.Fprintf(w, "Constant: %s\n", err)
		} else {
			fmt.Fprintln(w, wire.ToString())
		}
	case p[0] == "d":
		if len(p) != 2 {
			fmt.Fprintln(w, "Divider jumper syntax: d.terminal")
			return
		}
		wire, err := divsr.GetPlug(p[1])
		if err != nil {
			fmt.Fprintf(w, "Divider: %s\n", err)
		} else {
			fmt.Fprintln(w, wire.ToString())
		}
	case p[0][0] == 'f':
		if len(p) != 2 {
			fmt.Fprintln(w, "Function table jumper syntax: funit.terminal")
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Fprintf(w, "Invalid function table %s\n", p[0][1:])
			return
		}
		wire, err := ft[unit-1].GetPlug(p[1])
		if err != nil {
			fmt.Fprintf(w, "Function table %d: %s\n", unit, err)
		} else {
			fmt.Fprintln(w, wire.ToString())
		}
	case p[0] == "i":
		if len(p) != 2 {
			fmt.Fprintln(w, "Initiator jumper syntax: i.terminal")
			return
		}
		wire, err := initiate.GetPlug(p[1])
		if err != nil {
			fmt.Fprintf(w, "Initiate: %s\n", err)
		} else {
			fmt.Fprintln(w, wire.ToString())
		}
	case p[0] == "m":
		if len(p) != 2 {
			fmt.Fprintln(w, "Multiplier jumper syntax: m.terminal")
			return
		}
		wire, err := multiplier.GetPlug(p[1])
		if err != nil {
			fmt.Fprintf(w, "Multiplier: %s\n", err)
		} else {
			fmt.Fprintln(w, wire.ToString())
		}
	case p[0] == "p":
		if len(p) != 2 {
			fmt.Fprintln(w, "Programmer jumper syntax: p.terminal")
			return
		}
		wire, err := mp.GetPlug(p[1])
		if err != nil {
			fmt.Fprintf(w, "Programmer: %s\n", err)
		} else {
			fmt.Fprintln(w, wire.ToString())
		}
	case unicode.IsDigit(rune(p[0][0])):
		wires, err := trays.GetPlug(p[0])
		if err != nil {
			fmt.Fprintf(w, "Trays: %s\n", err)
		} else {
			printWires(w, wires)
		}
	default:
		fmt.Fprintln(w, "Invalid jack spec: ", p)
	}
}

func doReset(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "Status syntax: r unit")
		return
	}
	p := strings.Split(f[1], ".")
	switch p[0] {
	case "a":
		if len(p) != 2 {
			fmt.Fprintln(w, "Accumulator reset syntax: r a.unit")
			return
		}
		unit, _ := strconv.Atoi(p[1])
		if !(unit >= 1 && unit <= 20) {
			fmt.Fprintf(w, "Invalid accumulator %s", p[1])
			return
		}
		accumulator[unit-1].Reset()
	case "b":
		debugger.Reset()
	case "c":
		constant.Reset()
	case "d":
		divsr.Reset()
	case "f":
		if len(p) != 2 {
			fmt.Fprintln(w, "Function table reset syntax: r f.unit")
			return
		}
		unit, _ := strconv.Atoi(p[1])
		if !(unit >= 1 && unit <= 3) {
			fmt.Fprintln(w, "Invalid function table")
			return
		}
		ft[unit-1].Reset()
	case "i":
		initiate.Reset()
	case "m":
		multiplier.Reset()
	case "p":
		mp.Reset()
	}
}

func doResetAll(w io.Writer) {
	initiate.Reset()
	cycle.Io.Reset <- 1
	debugger.Reset()
	mp.Reset()
	ft[0].Reset()
	ft[1].Reset()
	ft[2].Reset()
	for i := 0; i < 20; i++ {
		accumulator[i].Reset()
	}
	divsr.Reset()
	multiplier.Reset()
	constant.Reset()
	printer.Reset()
	adapters.Reset()
	trays.Reset()
}

func doGetSwitch(w io.Writer, command string, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "expected s? u.switch")
		return
	}
	p := strings.Split(f[1], ".")
	switch {
	case p[0][0] == 'a':
		if len(p) != 2 {
			fmt.Fprintln(w, "Invalid accumulator switch:", command)
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 20) {
			fmt.Fprintf(w, "Invalid accumulator %s\n", p[0][1:])
			return
		}
		value, err := accumulator[unit-1].GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "Accumulator %d: %s\n", unit, err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	case p[0] == "c":
		if len(p) != 2 {
			fmt.Fprintln(w, "Constant switch syntax: s? c.switch")
			return
		}
		value, err := constant.GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "Constant: %s\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	case p[0] == "cy":
		if len(p) != 2 {
			fmt.Fprintln(w, "Cycling switch syntax: s? cy.switch")
			return
		}
		value, err := cycle.GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "Cycling: %s\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	case p[0] == "d" || p[0] == "ds":
		if len(p) != 2 {
			fmt.Fprintln(w, "Divider switch syntax: s? d.switch")
			return
		}
		value, err := divsr.GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "Divider: %s\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	case p[0][0] == 'f':
		if len(p) != 2 {
			fmt.Fprintln(w, "Function table switch syntax: s? funit.switch", command)
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Fprintln(w, "Invalid function table")
			return
		}
		value, err := ft[unit-1].GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "Function table %d: %s\n", unit, err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	case p[0] == "m":
		if len(p) != 2 {
			fmt.Fprintln(w, "Multiplier switch syntax: s? m.switch")
			return
		}
		value, err := multiplier.GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "error: %s\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	case p[0] == "p":
		if len(p) != 2 {
			fmt.Fprintln(w, "Programmer switch syntax: s? p.switch")
			break
		}
		value, err := mp.GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "Programmer: %s\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	case p[0] == "pr":
		if len(p) != 2 {
			fmt.Fprintln(w, "Printer switch syntax: s? pr.switch")
			return
		}
		value, err := printer.GetSwitch(p[1])
		if err != nil {
			fmt.Fprintf(w, "Printer: %s\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", value)
		}
	default:
		fmt.Fprintf(w, "unknown unit for switch: %s\n", p[0])
	}
}

func doSetSwitch(w io.Writer, command string, f []string) {
	if len(f) < 3 {
		fmt.Fprintln(w, "No switch setting")
		return
	}
	p := strings.Split(f[1], ".")
	switch {
	case p[0] == "ad":
		if len(p) != 3 {
			fmt.Fprintln(w, "Invalid adapter switch:", command)
			return
		}
		err := adapters.Switch(p[1], p[2], f[2])
		if err != nil {
			fmt.Fprintf(w, "Adapter: %s\n", err)
		}
	case p[0][0] == 'a':
		if len(p) != 2 {
			fmt.Fprintln(w, "Invalid accumulator switch:", command)
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 20) {
			fmt.Fprintf(w, "Invalid accumulator %s\n", p[0][1:])
			return
		}
		err := accumulator[unit-1].SetSwitch(p[1], f[2])
		if err != nil {
			fmt.Fprintf(w, "Accumulator %d: %s\n", unit, err)
		}
	case p[0] == "c":
		if len(p) != 2 {
			fmt.Fprintln(w, "Constant switch syntax: s c.switch value")
			return
		}
		err := constant.SetSwitch(p[1], f[2])
		if err != nil {
			fmt.Fprintf(w, "Constant: %s\n", err)
		}
	case p[0] == "cy":
		if len(p) != 2 {
			fmt.Fprintln(w, "Cycling switch syntax: s cy.switch value")
			return
		}
		cycle.Io.Switches <- [2]string{p[1], f[2]}
	case p[0] == "d" || p[0] == "ds":
		if len(p) != 2 {
			fmt.Fprintln(w, "Divider switch syntax: s d.switch value")
			return
		}
		err := divsr.SetSwitch(p[1], f[2])
		if err != nil {
			fmt.Fprintf(w, "Divider: %s\n", err)
		}
	case p[0][0] == 'f':
		if len(p) != 2 {
			fmt.Fprintln(w, "Function table switch syntax: s funit.switch value", command)
			return
		}
		unit, _ := strconv.Atoi(p[0][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Fprintln(w, "Invalid function table")
			return
		}
		err := ft[unit-1].SetSwitch(p[1], f[2])
		if err != nil {
			fmt.Fprintf(w, "Function table %d: %s", unit, err)
		}
	case p[0] == "m":
		if len(p) != 2 {
			fmt.Fprintln(w, "Multiplier switch syntax: s m.switch value")
			return
		}
		err := multiplier.SetSwitch(p[1], f[2])
		if err != nil {
			fmt.Fprintf(w, "Multiplier: %s\n", err)
		}
	case p[0] == "p":
		if len(p) != 2 {
			fmt.Fprintln(w, "Programmer switch syntax: s p.switch value")
			break
		}
		err := mp.SetSwitch(p[1], f[2])
		if err != nil {
			fmt.Fprintf(w, "Programmer: %s\n", err)
		}
	case p[0] == "pr":
		if len(p) != 2 {
			fmt.Fprintln(w, "Printer switch syntax: s pr.switch value")
			return
		}
		err := printer.SetSwitch(p[1], f[2])
		if err != nil {
			fmt.Fprintf(w, "Printer: %s\n", err)
		}
	default:
		fmt.Fprintf(w, "unknown unit for switch: %s\n", p[0])
	}
}

func doSet(w io.Writer, f []string) {
	if len(f) != 3 {
		fmt.Fprintln(w, "set syntax: set a13 -9876543210")
		return
	}
	unit, _ := strconv.Atoi(f[1][1:])
	if !(unit >= 1 && unit <= 20) {
		fmt.Fprintf(w, "Invalid accumulator %s\n", f[1][1:])
		return
	}
	value, err := strconv.ParseInt(f[2], 10, 64)
	if err != nil {
		fmt.Fprintf(w, "Invalid accumulator value %s\n", err)
		return
	}
	accumulator[unit-1].Set(value)
}

func doTraceStart(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "trace start syntax: ts p|f|pf")
		return
	}
	tracePulses := strings.IndexByte(f[1], 'p') != -1
	traceRegs := strings.IndexByte(f[1], 'f') != -1
	if !tracePulses && !traceRegs {
		fmt.Fprintln(w, "trace start: expecting p for pulses, f for regs")
		return
	}
	log = NewTrace(tracePulses, traceRegs)
	for i := range accumulator {
		log.Register(accumulator[i].AttachTrace(log.tracePulse))
	}
	cycle.Io.TraceAddCycle = func() {
		log.RunCallbacks()
	}
	cycle.Io.TracePulse = func() {
		log.Tick()
	}
}

func doTraceEnd(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "trace end syntax: te file")
		return
	}
	if log == nil {
		fmt.Fprintln(w, "not tracing; missing ts?")
		return
	}
	fd, err := os.Create(f[1])
	if err != nil {
		fmt.Fprintf(w, "trace end create: %s\n", err)
		return
	}
	log.WriteVcd(fd, time.Now())
}

func printWires(w io.Writer, wires []Wire) {
	connected := 0
	for i := range wires {
		if wires[i].Ch != nil {
			fmt.Fprintln(w, wires[i].ToString())
			connected++
		}
	}
	if connected == 0 {
		fmt.Fprintln(w, "unconnected")
	}
}
