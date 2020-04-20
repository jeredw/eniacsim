package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	. "github.com/jeredw/eniacsim/lib"
	"github.com/jeredw/eniacsim/lib/units"
)

var cycle *units.Cycle
var initiate *units.Initiate
var mp *units.Mp
var divsr *units.Divsr
var multiplier *units.Multiplier
var constant *units.Constant
var printer *units.Printer
var ft [3]*units.Ft
var accumulator [20]*units.Accumulator
var debugger *Debugger
var trays *Trays
var adapters *Adapters

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [configuration file]\n", os.Args[0])
		flag.PrintDefaults()
	}
	useControl := flag.Bool("c", false, "use a portable control station connected to GPIO pins")
	demoMode := flag.Bool("D", false, "automatically cycle among perspectives")
	noGui := flag.Bool("g", false, "run without GUI")
	tkKludge := flag.Bool("K", false, "work around wish memory leaks")
	width := flag.Int("w", 0, "`width` of the simulation window in pixels")
	testCycles := flag.Int("t", 0, "run for n add cycles and dump state")
	flag.Parse()

	var ppunch chan string
	if !*noGui {
		go gui(*demoMode, *tkKludge, *useControl, *width)
		ppunch = make(chan string)
	}
	if *useControl {
		go ctlstation()
	}

	trays = NewTrays()
	adapters = NewAdapters()
	debugger = NewDebugger()
	cycle = units.NewCycle(units.CycleConn{
		CycleButton: NewButton(),
		Switches:    make(chan [2]string),
		Reset:       make(chan int),
		Stop:        make(chan int),
		TestButton:  NewButton(),
		TestCycles:  *testCycles,
	})
	initiate = units.NewInitiate(units.InitiateConn{
		InitButton: NewButton(),
		Ppunch:     ppunch,
	})
	mp = units.NewMp()
	divsr = units.NewDivsr()
	multiplier = units.NewMultiplier()
	constant = units.NewConstant()
	printer = units.NewPrinter()
	for i := 0; i < 3; i++ {
		ft[i] = units.NewFt(i)
	}
	for i := 0; i < 20; i++ {
		accumulator[i] = units.NewAccumulator(i)
	}

	clockFuncs := []ClockFunc{
		initiate.MakeClockFunc(),
		mp.MakeClockFunc(),
		divsr.MakeClockFunc(),
		multiplier.MakeClockFunc(),
		constant.MakeClockFunc(),
	}
	for i := 0; i < 20; i++ {
		clockFuncs = append(clockFuncs, accumulator[i].MakeClockFunc())
	}
	for i := 0; i < 3; i++ {
		clockFuncs = append(clockFuncs, ft[i].MakeClockFunc())
	}
	clearFuncs := []func(){
		func() { mp.Clear() },
	}
	for i := 0; i < 20; i++ {
		clearFuncs = append(clearFuncs, func(i int) func() {
			return func() {
				accumulator[i].Clear()
			}
		}(i))
	}
	clearFuncs = append(clearFuncs, func() { divsr.Clear() })

	cycle.Io.Units = clockFuncs
	cycle.Io.Clear = func() bool { return initiate.ShouldClear() }
	initiate.Io.ClearUnits = clearFuncs
	initiate.Io.AddCycle = func() int { return cycle.AddCycle() }
	initiate.Io.Stepping = func() bool { return cycle.Stepping() }
	initiate.Io.ReadCard = func(s string) { constant.ReadCard(s) }
	initiate.Io.Print = func() string { return printer.Print() }
	divsr.Io.A2Sign = func() string { return accumulator[2].Sign() }
	divsr.Io.A2Clear = func() { accumulator[2].Clear() }
	divsr.Io.A4Sign = func() string { return accumulator[4].Sign() }
	divsr.Io.A4Clear = func() { accumulator[4].Clear() }
	multiplier.Io.A8Clear = func() { accumulator[8].Clear() }
	multiplier.Io.A8Value = func() string { return accumulator[8].Value() }
	multiplier.Io.A9Clear = func() { accumulator[9].Clear() }
	multiplier.Io.A9Value = func() string { return accumulator[9].Value() }
	printer.Io.MpPrinterDecades = func() string { return mp.PrinterDecades() }
	for i := 0; i < 20; i++ {
		printer.Io.AccValue[i] = func(i int) func() string {
			return func() string {
				return accumulator[i].Value()
			}
		}(i)
		accumulator[i].Io.Sv = func() int { return divsr.Sv() }
		accumulator[i].Io.Su2 = func() int { return divsr.Su2() }
		accumulator[i].Io.Su3 = func() int { return divsr.Su3() }
		accumulator[i].Io.Multl = func() bool { return multiplier.Multl() }
		accumulator[i].Io.Multr = func() bool { return multiplier.Multr() }
	}

	go initiate.Run()
	go mp.Run()
	go cycle.Run()
	go divsr.Run()
	go multiplier.Run()
	go constant.Run()
	for i := 0; i < 20; i++ {
		go accumulator[i].Run()
	}
	for i := 0; i < 3; i++ {
		go ft[i].Run()
	}

	if flag.NArg() >= 1 {
		// Seriously ugly hack to give other goprocs time to get initialized
		time.Sleep(100 * time.Millisecond)
		doCommand("l " + flag.Arg(0))
	}

	if *testCycles > 0 {
		cycle.Io.TestButton.Push <- 1
		<-cycle.Io.TestButton.Done
		doDumpAll()
		return
	}

	sc := bufio.NewScanner(os.Stdin)
	var prompt = func() {
		fmt.Printf("%04d> ", cycle.AddCycle()%10000)
	}
	prompt()
	for sc.Scan() {
		if doCommand(sc.Text()) < 0 {
			break
		}
		prompt()
	}
}
