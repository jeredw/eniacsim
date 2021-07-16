package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
	"unicode"

	. "github.com/jeredw/eniacsim/lib"
	"github.com/jeredw/eniacsim/lib/units"
)

var perfCycles int64
var perfTime time.Duration

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
	case "g":
		doRun(w, f)
	case "l":
		doLoad(w, f)
	case "n":
		cycle.Step()
		doDumpAll(w)
	case "perf":
		rate := float64(perfCycles) / perfTime.Seconds()
		speedup := rate / 5000.0
		fmt.Printf("%.2f MHz (%d cycles, %v simulated, %v realtime [%.2fx])\n", rate/1e6, perfCycles, perfTime, time.Duration(speedup)*perfTime, speedup)
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
		u.Initiate.PushClearButton()
	case "i":
		u.Initiate.PushInitButton()
	case "p":
		cycle.Step()
	case "r":
		u.Initiate.PushReadButton()
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
		fmt.Fprintln(w, u.Accumulator[unit-1].Stat())
	case 'b':
		fmt.Fprintln(w, debugger.Stat())
	case 'c':
		fmt.Fprintln(w, u.Constant.Stat())
	case 'd':
		fmt.Fprintln(w, u.Divsr.Stat2())
	case 'f':
		unit, _ := strconv.Atoi(f[1][1:])
		if !(unit >= 1 && unit <= 3) {
			fmt.Fprintf(w, "Invalid function table %s\n", f[1][1:])
			return
		}
		fmt.Fprintln(w, u.Ft[unit-1].Stat())
	case 'i':
		fmt.Fprintln(w, u.Initiate.Stat())
	case 'm':
		fmt.Fprintln(w, u.Multiplier.Stat())
	case 'p':
		fmt.Fprintln(w, u.Mp.Stat())
	}
}

func doDumpAll(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, u.Initiate.Stat())
	fmt.Fprintln(w, u.Mp.Stat())
	header := "      9876543210 9876543210 r 123456789012"
	fmt.Fprintf(w, "%s   %s\n", header, header)
	for i := 0; i < 20; i += 2 {
		ai := u.Accumulator[i].Stat()
		ai1 := u.Accumulator[i+1].Stat()
		fmt.Fprintf(w, "a%-2d %s   a%-2d %s\n", i+1, ai, i+2, ai1)
	}
	fmt.Fprintln(w, u.Divsr.Stat2())
	fmt.Fprintln(w, u.Multiplier.Stat())
	for i := 0; i < 3; i++ {
		fmt.Fprintln(w, u.Ft[i].Stat())
	}
	fmt.Fprintln(w, u.Constant.Stat())
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
		u.Initiate.SetCardScanner(bufio.NewScanner(fp))
	case "p":
		fp, err := os.Create(f[2])
		if err != nil {
			fmt.Fprintf(w, "Card punch open: %s\n", err)
			return
		}
		u.Initiate.SetPunchWriter(bufio.NewWriter(fp))
	}
}

func doRun(w io.Writer, f []string) {
	doSetSwitch(w, "s cy.op co", []string{"s", "cy.op", "co"})
	interrupt := make(chan os.Signal, 1)
	done := make(chan int)
	signal.Notify(interrupt, os.Interrupt)
	var elapsedTime time.Duration
	var elapsedCycles int64
	var stoppedByDebugger bool
	go func() {
		startTime := time.Now()
		startCycle := cycle.AddCycle
	loop:
		for {
			select {
			case <-interrupt:
				break loop
			default:
				if cycle.StepNAddCycles(10000) {
					stoppedByDebugger = true
					break loop
				}
			}
		}
		elapsedTime = time.Since(startTime)
		elapsedCycles = cycle.AddCycle - startCycle
		done <- 1
	}()
	<-done
	perfCycles += elapsedCycles
	perfTime += elapsedTime
	if stoppedByDebugger {
		doDumpAll(w)
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
	if handled, err := doInterconnect(f[1], f[2], p1, p2); handled {
		if err != nil {
			fmt.Fprintf(w, "Interconnect: %s\n", err)
		}
		return
	}

	jack1, pb1, err := findJack(f[1], 0)
	if err != nil {
		fmt.Fprintf(w, "Plug error: %s\n", err)
		return
	}
	err = setAdapterSwitchFromJack(pb1, p1)
	if err != nil {
		fmt.Fprintf(w, "Adapter: %s\n", err)
		return
	}
	jack2, pb2, err := findJack(f[2], 1)
	if err != nil {
		fmt.Fprintf(w, "Plug error: %s\n", err)
		fmt.Fprintln(w, command)
		return
	}
	err = setAdapterSwitchFromJack(pb2, p2)
	if err != nil {
		fmt.Fprintf(w, "Adapter: %s\n", err)
		return
	}
	err = Connect(jack1, jack2)
	if err != nil {
		fmt.Fprintf(w, "Plug error: %s\n", err)
		return
	}
}

func findJack(name string, pos int) (*Jack, Plugboard, error) {
	p := strings.SplitN(name, ".", 2)
	pb, err := findPlugboard(p[0])
	if err != nil {
		return nil, nil, err
	}
	jackName := name
	if pb != trays {
		if len(p) != 2 {
			return nil, nil, fmt.Errorf("bad jack name %s", name)
		}
		jackName = p[1]
	}
	if pb == adapters {
		jackName = rewriteAdapterJackName(jackName, pos)
	}
	jack, err := pb.FindJack(jackName)
	return jack, pb, err
}

func rewriteAdapterJackName(s string, pos int) string {
	dir := "o."
	if pos == 1 {
		dir = "i."
	}
	return dir + s
}

func setAdapterSwitchFromJack(pb Plugboard, p []string) error {
	if pb == adapters && len(p) == 4 && p[1] != "dp" {
		sw, err := adapters.FindSwitch(fmt.Sprintf("%s.%s", p[1], p[2]))
		if err != nil {
			return err
		}
		return sw.Set(p[3])
	}
	return nil
}

func findPlugboard(name string) (Plugboard, error) {
	switch {
	case name == "ad":
		return adapters, nil
	case len(name) > 1 && name[0] == 'a':
		n, _ := strconv.Atoi(name[1:])
		if !(n >= 1 && n <= 20) {
			return nil, fmt.Errorf("invalid accumulator %s", name[1:])
		}
		return u.Accumulator[n-1], nil
	case name == "c":
		return u.Constant, nil
	case name == "d":
		return u.Divsr, nil
	case name == "debug":
		return debugger, nil
	case len(name) > 1 && name[0] == 'f':
		n, _ := strconv.Atoi(name[1:])
		if !(n >= 1 && n <= 3) {
			return nil, fmt.Errorf("invalid function table %s", name[1:])
		}
		return u.Ft[n-1], nil
	case name == "i":
		return u.Initiate, nil
	case name == "m":
		return u.Multiplier, nil
	case name == "os":
		return u.OrderSelector, nil
	case name == "p":
		return u.Mp, nil
	case name == "pa":
		return pulseAmps, nil
	case name == "sft":
		return u.FtSelector, nil
	case name == "sjk1":
		return u.JkSelector[0], nil
	case name == "sjk2":
		return u.JkSelector[1], nil
	case name == "st":
		return u.TenStepper, nil
	case name == "pm1":
		return u.PmDiscriminator[0], nil
	case name == "pm2":
		return u.PmDiscriminator[1], nil
	case isTrayName(name):
		return trays, nil
	}
	return nil, fmt.Errorf("invalid unit name %s", name)
}

func isTrayName(name string) bool {
	if len(name) >= 1 && unicode.IsDigit(rune(name[0])) {
		return true
	}
	if len(name) >= 2 && (name[0] >= 'A' && name[0] <= 'Z') && name[1] == '-' {
		return true
	}
	return false
}

func doGetPlug(w io.Writer, command string, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "Invalid jumper spec", command)
		return
	}
	jack, _, err := findJack(f[1], 0)
	if err != nil {
		fmt.Fprintf(w, "Get plug: %s\n", err)
		return
	}
	fmt.Fprintf(w, "%s", jack.ConnectionsString())
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
		u.Accumulator[unit-1].Reset()
	case "b":
		debugger.Reset()
	case "c":
		u.Constant.Reset()
	case "d":
		u.Divsr.Reset()
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
		u.Ft[unit-1].Reset()
	case "i":
		u.Initiate.Reset()
	case "m":
		u.Multiplier.Reset()
	case "os":
		u.OrderSelector.Reset()
	case "p":
		u.Mp.Reset()
	case "pm1":
		u.PmDiscriminator[0].Reset()
	case "pm2":
		u.PmDiscriminator[1].Reset()
	case "st":
		u.TenStepper.Reset()
	case "sft":
		u.FtSelector.Reset()
	case "sjk1":
		u.JkSelector[0].Reset()
	case "sjk2":
		u.JkSelector[1].Reset()
	}
}

func doResetAll(w io.Writer) {
	u.Initiate.Reset()
	debugger.Reset()
	u.Mp.Reset()
	u.Ft[0].Reset()
	u.Ft[1].Reset()
	u.Ft[2].Reset()
	for i := 0; i < 20; i++ {
		u.Accumulator[i].Reset()
	}
	u.Divsr.Reset()
	u.Multiplier.Reset()
	u.Constant.Reset()
	printer.Reset()
	adapters.Reset()
	trays.Reset()
	u.TenStepper.Reset()
	u.OrderSelector.Reset()
}

func findSwitch(name string) (Switch, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("missing switch name")
	}
	p := strings.SplitN(name, ".", 2)
	if len(p) != 2 {
		return nil, fmt.Errorf("invalid switch: %s", name)
	}
	sb, err := findSwitchboard(p[0])
	if err != nil {
		return nil, fmt.Errorf("invalid switch: %s", err)
	}
	sw, err := sb.FindSwitch(p[1])
	if err != nil {
		return nil, err
	}
	return sw, nil
}

func findSwitchboard(name string) (Switchboard, error) {
	switch {
	case name == "ad":
		return adapters, nil
	case len(name) > 1 && name[0] == 'a':
		n, _ := strconv.Atoi(name[1:])
		if !(n >= 1 && n <= 20) {
			return nil, fmt.Errorf("invalid accumulator %s", name[1:])
		}
		return u.Accumulator[n-1], nil
	case name == "c":
		return u.Constant, nil
	case name == "cy":
		return cycle, nil
	case name == "d" || name == "ds":
		return u.Divsr, nil
	case name == "debug":
		return debugger, nil
	case len(name) > 1 && name[0] == 'f':
		n, _ := strconv.Atoi(name[1:])
		if !(n >= 1 && n <= 3) {
			return nil, fmt.Errorf("invalid function table %s", name[1:])
		}
		return u.Ft[n-1], nil
	case name == "m":
		return u.Multiplier, nil
	case name == "p":
		return u.Mp, nil
	case name == "pr":
		return printer, nil
	}
	return nil, fmt.Errorf("invalid unit name %s", name)
}

func doGetSwitch(w io.Writer, command string, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "expected s? u.switch")
		return
	}
	sw, err := findSwitch(f[1])
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	fmt.Fprintf(w, "%s\n", sw.Get())
}

func doSetSwitch(w io.Writer, command string, f []string) {
	if len(f) < 3 {
		fmt.Fprintln(w, "No switch setting")
		return
	}
	sw, err := findSwitch(f[1])
	if err != nil {
		fmt.Fprintf(w, "error finding switch: %s\n", err)
		return
	}
	err = sw.Set(f[2])
	if err != nil {
		fmt.Fprintf(w, "error setting switch: %s\n", err)
		return
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
	u.Accumulator[unit-1].Set(value)
}

func doTraceStart(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "trace start syntax: ts p|f|pf")
		return
	}
	pulses := strings.IndexByte(f[1], 'p') != -1
	regs := strings.IndexByte(f[1], 'f') != -1
	if !pulses && !regs {
		fmt.Fprintln(w, "trace start: expecting p for pulses, f for regs")
		return
	}
	waves = NewWavedump(pulses, regs)
	for i := range u.Accumulator {
		u.Accumulator[i].AttachTracer(waves)
	}
	u.Multiplier.AttachTracer(waves)
	u.Constant.AttachTracer(waves)
	u.Divsr.AttachTracer(waves)
	cycle.AttachTracer(waves)
}

func doTraceEnd(w io.Writer, f []string) {
	if len(f) != 2 {
		fmt.Fprintln(w, "trace end syntax: te file")
		return
	}
	if waves == nil {
		fmt.Fprintln(w, "not tracing; missing ts?")
		return
	}
	fd, err := os.Create(f[1])
	if err != nil {
		fmt.Fprintf(w, "trace end create: %s\n", err)
		return
	}
	bw := bufio.NewWriter(fd)
	waves.WriteVcd(bw, time.Now())
	bw.Flush()
}

func doInterconnect(f1 string, f2 string, p1 []string, p2 []string) (bool, error) {
	// Handle commands like p aXX.{st,su,il,ir} *
	if len(p1) == 2 && p1[0][0] == 'a' && len(p1[1]) >= 2 &&
		len(p2) == 2 && p2[0][0] == 'a' && len(p2[1]) >= 2 &&
		(p1[1][:2] == "il" || p1[1][:2] == "ir") &&
		(p2[1][:2] == "il" || p2[1][:2] == "ir") {
		return true, units.Interconnect(u.Accumulator, p1, p2)
	}
	if handled, err := doMultiplierInterconnect(f1, f2); handled {
		return true, err
	}
	if handled, err := doDivsrInterconnect(f1, f2); handled {
		return true, err
	}
	return false, nil
}

func doMultiplierInterconnect(f1 string, f2 string) (bool, error) {
	// Handle p m.[LR] aXX
	if strings.HasPrefix(f2, "m.") {
		f1, f2 = f2, f1
	}
	var conn *units.StaticWiring
	switch f1 {
	case "m.l", "m.L":
		conn = &u.Multiplier.Io.Lhpp
	case "m.r", "m.R":
		conn = &u.Multiplier.Io.Rhpp
	case "m.ier":
		conn = &u.Multiplier.Io.Ier
	case "m.icand":
		conn = &u.Multiplier.Io.Icand
	}
	if conn != nil {
		if len(f2) < 2 || !strings.HasPrefix(f2, "a") {
			return true, fmt.Errorf("multiplier interconnect must be connected to accum")
		}
		unit, _ := strconv.Atoi(f2[1:])
		if !(unit >= 1 && unit <= 20) {
			return true, fmt.Errorf("invalid accumulator")
		}
		*conn = u.Accumulator[unit-1]
		return true, nil
	}
	return false, nil
}

func doDivsrInterconnect(f1 string, f2 string) (bool, error) {
	// Handle p d.{} aXX
	if strings.HasPrefix(f2, "d.") {
		f1, f2 = f2, f1
	}
	var conn *units.StaticWiring
	switch f1 {
	case "d.quotient":
		conn = &u.Divsr.Io.Quotient
	case "d.numerator":
		conn = &u.Divsr.Io.Numerator
	case "d.denominator":
		conn = &u.Divsr.Io.Denominator
	case "d.shift":
		conn = &u.Divsr.Io.Shift
	}
	if conn != nil {
		if len(f2) < 2 || !strings.HasPrefix(f2, "a") {
			return true, fmt.Errorf("divsr interconnect must be connected to accum")
		}
		unit, _ := strconv.Atoi(f2[1:])
		if !(unit >= 1 && unit <= 20) {
			return true, fmt.Errorf("invalid accumulator")
		}
		*conn = u.Accumulator[unit-1]
		return true, nil
	}
	return false, nil
}
