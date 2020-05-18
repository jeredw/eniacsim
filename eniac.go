package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	//	"net/http"
	//	_ "net/http/pprof"

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

var log *trace

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
	useWebGui := flag.Bool("W", false, "run web GUI")
	flag.Parse()

	var ppunch chan string
	if *useWebGui {
		go webGui()
	} else if !*noGui {
		go gui(*demoMode, *tkKludge, *useControl, *width)
		ppunch = make(chan string)
	}
	if *useControl {
		go ctlstation()
	}

	//	go func() {
	//		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	//	}()

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

	clockedUnits := []Clocked{initiate, mp, divsr, multiplier, constant}
	clearedUnits := []Cleared{mp, divsr}
	for i := 0; i < 20; i++ {
		clockedUnits = append(clockedUnits, accumulator[i])
		clearedUnits = append(clearedUnits, accumulator[i])
	}
	for i := 0; i < 3; i++ {
		clockedUnits = append(clockedUnits, ft[i])
	}

	cycle.Io.Units = clockedUnits
	cycle.Io.Clear = func() bool { return initiate.ShouldClear() }
	initiate.Io.Units = clearedUnits
	initiate.Io.AddCycle = func() int { return cycle.AddCycle() }
	initiate.Io.Stepping = func() bool { return cycle.Stepping() }
	initiate.Io.ReadCard = func(s string) { constant.ReadCard(s) }
	initiate.Io.Print = func() string { return printer.Print() }
	divsr.Io.A2 = accumulator[2]
	divsr.Io.A4 = accumulator[4]
	multiplier.Io.A8 = accumulator[8]
	multiplier.Io.A9 = accumulator[9]
	printer.Io.MpPrinterDecades = func() string { return mp.PrinterDecades() }
	for i := 0; i < 20; i++ {
		printer.Io.Acc[i] = accumulator[i]
	}

	go initiate.Run()
	go cycle.Run()

	if flag.NArg() >= 1 {
		// Seriously ugly hack to give other goprocs time to get initialized
		time.Sleep(100 * time.Millisecond)
		doCommand(os.Stdout, "l "+flag.Arg(0))
	}

	if *testCycles > 0 {
		doTraceStart(os.Stdout, []string{"ts", "pf"})
		cycle.Io.TestButton.Push <- 1
		<-cycle.Io.TestButton.Done
		doDumpAll(os.Stdout)
		doTraceEnd(os.Stdout, []string{"te", "/tmp/test.vcd"})
		return
	}

	sc := bufio.NewScanner(os.Stdin)
	var prompt = func() {
		fmt.Printf("%04d> ", cycle.AddCycle()%10000)
	}
	prompt()
	for sc.Scan() {
		if doCommand(os.Stdout, sc.Text()) < 0 {
			break
		}
		prompt()
	}
}
