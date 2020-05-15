package main

import (
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"io"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type datapoint struct {
	time  int
	value int64
}

type waveform struct {
	kind   string
	name   string
	bits   int
	values []datapoint

	// Temporaries used while writing vcd files.
	idCode   string
	curIndex int
}

func newWaveform(kind string, name string, bits int) *waveform {
	return &waveform{
		kind:   kind,
		name:   name,
		bits:   bits,
		values: make([]datapoint, 0, 100),
	}
}

type trace struct {
	signals map[string]*waveform
	curTime int

	tracePulse   TraceFunc
	traceReg     TraceFunc
	regCallbacks []func(TraceFunc)

	mu sync.Mutex
}

// NewTrace returns a new empty trace.
//
// If tracePulses is true, sets a function pointer for logging pulses, and if
// traceRegs is true, sets a function pointer for logging registers (if these
// pointers are nil then the corresponding type of logging is disabled).
func NewTrace(tracePulses bool, traceRegs bool) *trace {
	t := &trace{
		signals:      make(map[string]*waveform),
		regCallbacks: make([]func(TraceFunc), 0, 50),
		curTime:      0,
	}
	if tracePulses {
		t.tracePulse = func(name string, bits int, value int64) {
			t.AppendWireValue(name, bits, value)
		}
	}
	if traceRegs {
		t.traceReg = func(name string, bits int, value int64) {
			t.AppendRegValue(name, bits, value)
		}
	}
	return t
}

// Register enqueues callbacks to run periodically to poll register values.
func (t *trace) Register(callbacks []func(TraceFunc)) {
	if t.traceReg == nil {
		return
	}
	for i := range callbacks {
		t.regCallbacks = append(t.regCallbacks, callbacks[i])
	}
}

// Tick advances the current signal time.
func (t *trace) Tick() {
	t.curTime++
}

// RunCallbacks runs registered callbacks to update register values.
func (t *trace) RunCallbacks() {
	if t.traceReg == nil {
		return
	}
	for i := range t.regCallbacks {
		t.regCallbacks[i](t.traceReg)
	}
}

// AppendWireValue logs a value sent on the wire name.
func (t *trace) AppendWireValue(name string, bits int, value int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(name) == 0 {
		panic("empty name")
	}
	s, ok := t.signals[name]
	if !ok {
		s = newWaveform("wire", name, bits)
		t.signals[name] = s
	}
	numValues := len(s.values)
	// Pulses implicitly go back to 0 at the next time step.
	// In case there is some actual different value at the next time step,
	// replace the implicitly added zero.
	if numValues > 0 && s.values[numValues-1].time == t.curTime {
		s.values[numValues-1] = datapoint{t.curTime, value}
	} else {
		s.values = append(s.values, datapoint{t.curTime, value})
	}
	s.values = append(s.values, datapoint{t.curTime + 1, 0})
}

// AppendRegValue stores the new value of reg name if it has changed.
func (t *trace) AppendRegValue(name string, bits int, value int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s, ok := t.signals[name]
	if !ok {
		s = newWaveform("reg", name, bits)
		t.signals[name] = s
	}
	numValues := len(s.values)
	if numValues > 0 && s.values[numValues-1].value == value {
		return
	}
	s.values = append(s.values, datapoint{t.curTime, value})
}

func (t *trace) WriteVcd(w io.Writer, ts time.Time) {
	fmt.Fprintf(w, "$version Generated by eniacsim $end\n")
	fmt.Fprintf(w, "$date %s $end\n", ts.Format(time.UnixDate))
	fmt.Fprintf(w, "$timescale 10us $end\n")
	nextId := "!"
	scopes := t.groupSignals()
	for i := range scopes {
		fmt.Fprintf(w, "$scope module %s $end\n", scopes[i].name)
		for _, signalName := range scopes[i].signals {
			s := t.signals[signalName]
			varName := signalName
			if s.bits > 1 {
				varName += fmt.Sprintf("[%d:0]", s.bits-1)
			}
			fmt.Fprintf(w, "$var %s %d %s %s $end\n", s.kind, s.bits, nextId, greekToAscii(varName))
			s.idCode = nextId
			nextId = incrementIdCode(nextId)
		}
		fmt.Fprintf(w, "$upscope $end\n")
	}
	fmt.Fprintf(w, "$enddefinitions $end\n")
	fmt.Fprintf(w, "$dumpvars\n")
	for i := range scopes {
		for _, name := range scopes[i].signals {
			s := t.signals[name]
			writeVcdValue(w, 0, s.bits)
			fmt.Fprintf(w, "%s\n", s.idCode)
		}
	}
	fmt.Fprintf(w, "$end\n")
	now := 0
	for now <= t.curTime+1 {
		printedTimeMarker := false
		nextTime := math.MaxInt32
		for i := range scopes {
			// Iterate over name slice, not map keys, for stable output order.
			for _, name := range scopes[i].signals {
			redo:
				s := t.signals[name]
				if s.curIndex >= len(s.values) {
					continue
				}
				p := s.values[s.curIndex]
				if p.time == now {
					if !printedTimeMarker {
						fmt.Fprintf(w, "#%d\n", now)
						printedTimeMarker = true
					}
					writeVcdValue(w, p.value, s.bits)
					fmt.Fprintf(w, "%s\n", s.idCode)
					s.curIndex++
					// The next event time might be for this signal.
					goto redo
				}
				if p.time < nextTime {
					nextTime = p.time
				}
			}
		}
		now = nextTime
	}
}

type scope struct {
	name    string
	signals []string
}

func (t *trace) groupSignals() []scope {
	scopes := make(map[string]*scope)
	for signalName, _ := range t.signals {
		p := strings.IndexByte(signalName, '.')
		if p == -1 {
			panic(fmt.Sprintf("signal name not like foo.bar: '%s'", signalName))
		}
		scopeName := signalName[:p]
		s, ok := scopes[scopeName]
		if !ok {
			s = &scope{name: scopeName, signals: make([]string, 0, 10)}
			scopes[scopeName] = s
		}
		s.signals = append(s.signals, signalName)
	}
	// Sort the signal names within each scope and the scopes themselves
	// alphabetically so that trace output order is stable.
	keys := make([]string, 0, len(scopes))
	for name, _ := range scopes {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	result := make([]scope, 0, len(keys))
	for i := range keys {
		s := scopes[keys[i]]
		sort.Strings(s.signals)
		result = append(result, *s)
	}
	return result
}

func writeVcdValue(w io.Writer, value int64, bits int) {
	if bits != 1 {
		fmt.Fprintf(w, "b")
	}
	for i := bits - 1; i >= 0; i-- {
		if value&(1<<i) != 0 {
			fmt.Fprintf(w, "1")
		} else {
			fmt.Fprintf(w, "0")
		}
	}
	if bits != 1 {
		fmt.Fprintf(w, " ")
	}
}

func incrementIdCode(idCode string) string {
	idBytes := []byte(idCode)
	for i := len(idBytes) - 1; i >= 0; i-- {
		idBytes[i]++
		if idBytes[i] < 127 {
			return string(idBytes)
		}
		idBytes[i] = 33
	}
	return "!" + string(idBytes)
}

func greekToAscii(s string) string {
	remap := map[rune]string{
		'α': "alpha",
		'β': "beta",
		'γ': "gamma",
		'δ': "delta",
		'ε': "epsilon",
	}
	var b strings.Builder
	for _, char := range s {
		ascii, found := remap[char]
		if found {
			fmt.Fprintf(&b, "%s", ascii)
		} else {
			fmt.Fprintf(&b, "%c", char)
		}
	}
	return b.String()
}