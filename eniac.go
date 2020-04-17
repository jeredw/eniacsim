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

var conssw, multsw, divsw, prsw chan [2]string
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
	divsw = make(chan [2]string)
	multsw = make(chan [2]string)
	conssw = make(chan [2]string)
	clockFuncs := []ClockFunc{
		initiate.MakeClockFunc(),
		mp.MakeClockFunc(),
		units.Makedivpulse(),
		units.Makemultpulse(),
		units.Makeconspulse(),
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
	clearFuncs = append(clearFuncs, units.Divclear)
	clearFuncs = append(clearFuncs, units.Multclear)
	initiate.Io.ClearUnits = clearFuncs
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
	initiate.Io.AddCycle = func() int { return cycle.AddCycle() }
	initiate.Io.Stepping = func() bool { return cycle.Stepping() }
	initiate.Io.Printer = units.PrConn{
		MpStat: func() string { return mp.Stat() },
	}

	go units.Consctl(conssw)
	go units.Divsrctl(divsw)
	go units.Multctl(multsw)
	go units.Prctl(prsw)

	go initiate.Run()
	go mp.Run()
	go cycle.Run()
	go units.Divunit()
	go units.Multunit()
	go units.Consunit()
	for i := 0; i < 20; i++ {
		go units.Accctl(i, accsw[i])
		go units.Accunit(i)
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
