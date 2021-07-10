package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"net/http"
	_ "net/http/pprof"

	. "github.com/jeredw/eniacsim/lib"
	"github.com/jeredw/eniacsim/lib/units"
)

var cycle *units.Cycle
var u *units.ClockedUnits
var printer *units.Printer
var debugger *Debugger
var trays *Trays
var adapters *Adapters
var pulseAmps *PulseAmps

var waves *wavedump

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [configuration file]\n", os.Args[0])
		flag.PrintDefaults()
	}
	useControl := flag.Bool("c", false, "use a portable control station connected to GPIO pins")
	demoMode := flag.Bool("D", false, "automatically cycle among perspectives")
	_ = flag.Bool("g", false, "deprecated")
	tkKludge := flag.Bool("K", false, "work around wish memory leaks")
	width := flag.Int("w", 0, "`width` of the simulation window in pixels")
	testCycles := flag.Int("t", 0, "run for n add cycles and dump state")
	useWebGui := flag.Bool("W", false, "run web GUI")
	useTkGui := flag.Bool("T", false, "run tk GUI")
	quiet := flag.Bool("q", false, "don't print a prompt")
	flag.Parse()

	var ppunch chan string
	if *useWebGui {
		go webGui()
	} else if *useTkGui {
		go gui(*demoMode, *tkKludge, *useControl, *width)
		ppunch = make(chan string)
	}
	if *useControl {
		go ctlstation()
	}

	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	trays = NewTrays()
	adapters = NewAdapters()
	pulseAmps = NewPulseAmps()
	debugger = NewDebugger()
	cycle = units.NewCycle(units.CycleConn{})
	u = &units.ClockedUnits{}
	u.Initiate = units.NewInitiate(units.InitiateConn{
		Ppunch: ppunch,
	})
	u.Mp = units.NewMp()
	u.Divsr = units.NewDivsr()
	u.Multiplier = units.NewMultiplier()
	u.Constant = units.NewConstant()
	printer = units.NewPrinter()
	for i := 0; i < 3; i++ {
		u.Ft[i] = units.NewFt(i)
	}
	for i := 0; i < 20; i++ {
		u.Accumulator[i] = units.NewAccumulator(i)
	}
	u.TenStepper = units.NewAuxStepper("st", 10)
	u.FtSelector = units.NewAuxStepper("sft", 6)
	u.OrderSelector = units.NewOrderSelector()
	for i := 0; i < 2; i++ {
		u.PmDiscriminator[i] = units.NewAuxStepper(fmt.Sprintf("pm%d", i+1), 2)
	}
	for i := 0; i < 2; i++ {
		u.JkSelector[i] = units.NewAuxStepper(fmt.Sprintf("sjk%d", i+1), 6)
	}

	clearedUnits := []Cleared{u.Mp, u.Divsr}
	for i := 0; i < 20; i++ {
		clearedUnits = append(clearedUnits, u.Accumulator[i])
	}

	cycle.Io.Units = u
	cycle.Io.SelectiveClear = func() bool { return u.Initiate.SelectiveClear() }
	u.Initiate.Io.Units = clearedUnits
	u.Initiate.Io.AddCycle = func() int64 { return cycle.AddCycle }
	u.Initiate.Io.Stepping = func() bool { return cycle.Stepping() }
	u.Initiate.Io.ReadCard = func(s string) { u.Constant.ReadCard(s) }
	u.Initiate.Io.Print = func() string { return printer.Print() }
	u.Divsr.Io.Quotient = u.Accumulator[2-1]
	u.Divsr.Io.Numerator = u.Accumulator[3-1]
	u.Divsr.Io.Denominator = u.Accumulator[5-1]
	u.Divsr.Io.Shift = u.Accumulator[7-1]
	u.Multiplier.Io.Ier = u.Accumulator[9-1]
	u.Multiplier.Io.Icand = u.Accumulator[10-1]
	u.Multiplier.Io.Lhpp = u.Accumulator[11-1]
	u.Multiplier.Io.Rhpp = u.Accumulator[13-1]
	printer.Io.MpPrinterDecades = func() string { return u.Mp.PrinterDecades() }
	for i := 0; i < 20; i++ {
		printer.Io.Accumulator[i] = u.Accumulator[i]
		debugger.Io.Accumulator[i] = u.Accumulator[i]
	}

	if flag.NArg() >= 1 {
		doCommand(os.Stdout, "l "+flag.Arg(0))
	}

	if *testCycles > 0 {
		doTraceStart(os.Stdout, []string{"ts", "pf"})
		cycle.SetTestMode()
		cycle.StepNAddCycles(*testCycles)
		doDumpAll(os.Stdout)
		doTraceEnd(os.Stdout, []string{"te", "/tmp/test.vcd"})
		return
	}

	sc := bufio.NewScanner(os.Stdin)
	var prompt = func() {
		if !*quiet {
			fmt.Printf("%d> ", cycle.AddCycle)
		}
	}
	prompt()
	for sc.Scan() {
		if doCommand(os.Stdout, sc.Text()) < 0 {
			break
		}
		prompt()
	}
}
