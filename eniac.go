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

var prsw chan [2]string
var accsw [20]chan [2]string
var ftsw [3]chan [2]string
var width, height int
var demomode, tkkludge, usecontrol *bool

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [configuration file]\n", os.Args[0])
		flag.PrintDefaults()
	}
	usecontrol = flag.Bool("c", false, "use a portable control station connected to GPIO pins")
	demomode = flag.Bool("D", false, "automatically cycle among perspectives")
	nogui := flag.Bool("g", false, "run without GUI")
	tkkludge = flag.Bool("K", false, "work around wish memory leaks")
	wp := flag.Int("w", 0, "`width` of the simulation window in pixels")
	testCycles := flag.Int("t", 0, "run for n add cycles and dump state")
	flag.Parse()

	var ppunch chan string
	width = *wp
	if !*nogui {
		go gui()
		ppunch = make(chan string)
	}
	if *usecontrol {
		go ctlstation()
	}

	initiate = units.NewInitiate(units.InitiateConn{
		InitButton: NewButton(),
		Ppunch:     ppunch,
	})
	mp = units.NewMp()
	divsr = units.NewDivsr()
	multiplier = units.NewMultiplier()
	constant = units.NewConstant()
	clockFuncs := []ClockFunc{
		initiate.MakeClockFunc(),
		mp.MakeClockFunc(),
		divsr.MakeClockFunc(),
		multiplier.MakeClockFunc(),
		constant.MakeClockFunc(),
	}
	clearFuncs := []func(){
		func() { mp.Clear() },
	}
	prsw = make(chan [2]string)
	for i := 0; i < 20; i++ {
		accsw[i] = make(chan [2]string)
		clockFuncs = append(clockFuncs, units.Makeaccpulse(i))
		clear := func(i int) func() { return func() { units.Accclear(i) } }(i)
		clearFuncs = append(clearFuncs, clear)
	}
	clearFuncs = append(clearFuncs, func() { divsr.Clear() })
	for i := 0; i < 3; i++ {
		ftsw[i] = make(chan [2]string)
		clockFuncs = append(clockFuncs, units.Makeftpulse(i))
	}

	cycle = units.NewCycle(units.CycleConn{
		Units:       clockFuncs,
		Clear:       func() bool { return initiate.ShouldClear() },
		CycleButton: NewButton(),
		Switches:    make(chan [2]string),
		Reset:       make(chan int),
		Stop:        make(chan int),
		TestButton:  NewButton(),
		TestCycles:  *testCycles,
	})
	initiate.Io.ClearUnits = clearFuncs
	initiate.Io.AddCycle = func() int { return cycle.AddCycle() }
	initiate.Io.Stepping = func() bool { return cycle.Stepping() }
	initiate.Io.ReadCard = func(s string) { constant.ReadCard(s) }
	initiate.Io.Printer = units.PrConn{
		MpStat: func() string { return mp.Stat() },
	}
	divsr.Io.Acc2Sign = func() string { return units.Accsign(2) }
	divsr.Io.Acc2Clear = func() { units.Accclear(2) }
	divsr.Io.Acc4Sign = func() string { return units.Accsign(4) }
	divsr.Io.Acc4Clear = func() { units.Accclear(4) }
	multiplier.Io.Acc8Clear = func() { units.Accclear(8) }
	multiplier.Io.Acc8Value = func() string { return units.Accvalue(8) }
	multiplier.Io.Acc9Clear = func() { units.Accclear(9) }
	multiplier.Io.Acc9Value = func() string { return units.Accvalue(9) }

	go units.Prctl(prsw)

	go initiate.Run()
	go mp.Run()
	go cycle.Run()
	go divsr.Run()
	go multiplier.Run()
	go constant.Run()
	for i := 0; i < 20; i++ {
		go units.Accctl(i, accsw[i])
		go units.Accunit(i, units.AccumulatorConn{
			Sv:    func() int { return divsr.Sv() },
			Su2:   func() int { return divsr.Su2() },
			Su3:   func() int { return divsr.Su3() },
			Multl: func() bool { return multiplier.Multl() },
			Multr: func() bool { return multiplier.Multr() },
		})
	}
	for i := 0; i < 3; i++ {
		go units.Ftctl(i, ftsw[i])
		go units.Ftunit(i)
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
