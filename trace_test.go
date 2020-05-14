package main

import (
	"bytes"
	"fmt"
	. "github.com/jeredw/eniacsim/lib"
	"testing"
	"time"
)

func TestIdCode(t *testing.T) {
	code := "!"
	want := "\""
	got := incrementIdCode(code)
	if got != want {
		t.Errorf("incrementIdCode(%s) = %s; want %s", code, got, want)
	}
}

func TestIdCode2(t *testing.T) {
	code := "~"
	want := "!!"
	got := incrementIdCode(code)
	if got != want {
		t.Errorf("incrementIdCode(%s) = %s; want %s", code, got, want)
	}
}

func TestIdCode3(t *testing.T) {
	code := "!!"
	want := "!\""
	got := incrementIdCode(code)
	if got != want {
		t.Errorf("incrementIdCode(%s) = %s; want %s", code, got, want)
	}
}

func TestIdCode4(t *testing.T) {
	code := "!~"
	want := "\"!"
	got := incrementIdCode(code)
	if got != want {
		t.Errorf("incrementIdCode(%s) = %s; want %s", code, got, want)
	}
}

func TestWriteVcdValue_MultiBit(t *testing.T) {
	var buf bytes.Buffer
	writeVcdValue(&buf, 42, 6)
	want := "b101010 "
	got := buf.String()
	if got != want {
		t.Errorf("writeVcdValue(42) = %s; want %s", got, want)
	}
}

func TestWriteVcdValue_OneBit(t *testing.T) {
	var buf bytes.Buffer
	writeVcdValue(&buf, 0, 1)
	want := "0"
	got := buf.String()
	if got != want {
		t.Errorf("writeVcdValue(0) = %s; want %s", got, want)
	}
}

func TestWriteVcd(t *testing.T) {
	tr := trace{
		signals: map[string]*waveform{
			"A.foo": &waveform{
				kind:   "wire",
				name:   "A.foo",
				bits:   1,
				values: []datapoint{{0, 0}, {10, 1}, {20, 0}},
			},
			"A.bar_α": &waveform{
				kind:   "reg",
				name:   "A.bar_α",
				bits:   6,
				values: []datapoint{{0, 0}, {10, 42}},
			},
			"B.baz": &waveform{
				kind:   "wire",
				name:   "B.baz",
				bits:   1,
				values: []datapoint{{0, 0}, {15, 1}, {20, 0}},
			},
		},
		curTime: 21,
	}
	var buf bytes.Buffer
	ts := time.Unix(int64(1589469793), int64(0))
	tr.WriteVcd(&buf, ts)
	got := buf.String()
	want := fmt.Sprintf(`$version Generated by eniacsim $end
$date %s $end
$timescale 10us $end
$scope module A $end
$var reg 6 ! A.bar_alpha[5:0] $end
$var wire 1 " A.foo $end
$upscope $end
$scope module B $end
$var wire 1 # B.baz $end
$upscope $end
$enddefinitions $end
$dumpvars
b000000 !
0"
0#
$end
#0
b000000 !
0"
0#
#10
b101010 !
1"
#15
1#
#20
0"
0#
`, ts.Format(time.UnixDate))
	if got != want {
		t.Errorf("WriteVcd mismatch; got %s", got)
	}
}

func assertTracesAreEqual(t *testing.T, got, want *trace) {
	if got.curTime != want.curTime {
		t.Fatalf("got.curTime = %d; want %d", got.curTime, want.curTime)
	}
	if len(got.signals) != len(want.signals) {
		t.Fatalf("len(got.signals) = %d; want %d", len(got.signals), len(want.signals))
	}
	for name, gotSignal := range got.signals {
		wantSignal, ok := want.signals[name]
		if !ok {
			t.Fatalf("unexpected signal %s", name)
		}
		if gotSignal.kind != wantSignal.kind {
			t.Fatalf("signals[%s].kind = %s; want %s", name, gotSignal.kind, wantSignal.kind)
		}
		if gotSignal.name != wantSignal.name {
			t.Fatalf("signals[%s].name = %s; want %s", name, gotSignal.name, wantSignal.name)
		}
		if gotSignal.bits != wantSignal.bits {
			t.Fatalf("signals[%s].bits = %d; want %d", name, gotSignal.bits, wantSignal.bits)
		}
		if len(gotSignal.values) != len(wantSignal.values) {
			t.Fatalf("signals[%s].values = %v; want %v", name, gotSignal.values, wantSignal.values)
		}
		for i := range gotSignal.values {
			if gotSignal.values[i] != wantSignal.values[i] {
				t.Fatalf("signals[%s].values = %v; want %v", name, gotSignal.values, wantSignal.values)
			}
		}
	}
}

func TestPulseTracing(t *testing.T) {
	tr := NewTrace(true, false)
	if tr.tracePulse == nil {
		t.Fatalf("expecting tracePulse callback")
	}
	tr.tracePulse("Unit.signal", 1, 1)
	tr.Tick()
	tr.tracePulse("Unit.signal2", 6, 42)
	tr.Tick()
	tr.tracePulse("Unit.signal", 1, 1)
	want := trace{
		signals: map[string]*waveform{
			"Unit.signal": &waveform{
				kind:   "wire",
				name:   "Unit.signal",
				bits:   1,
				values: []datapoint{{0, 1}, {1, 0}, {2, 1}, {3, 0}},
			},
			"Unit.signal2": &waveform{
				kind:   "wire",
				name:   "Unit.signal2",
				bits:   6,
				values: []datapoint{{1, 42}, {2, 0}},
			},
		},
		curTime: 2,
	}
	assertTracesAreEqual(t, tr, &want)
}

func TestRegTracing(t *testing.T) {
	tr := NewTrace(false, true)
	if tr.traceReg == nil {
		t.Fatalf("expecting traceReg callback")
	}
	reg1Calls := 0
	reg2Calls := 0
	tr.Register([]func(TraceFunc){
		func(traceReg TraceFunc) {
			reg1Calls++
			if reg1Calls == 1 {
				traceReg("Unit.reg1", 6, int64(42))
			} else {
				traceReg("Unit.reg1", 6, int64(0))
			}
		},
		func(traceReg TraceFunc) {
			reg2Calls++
			if reg2Calls <= 2 {
				traceReg("Unit.reg2", 1, 1)
			} else {
				traceReg("Unit.reg2", 1, 0)
			}
		},
	})
	tr.RunCallbacks()
	tr.Tick()
	tr.RunCallbacks()
	tr.Tick()
	tr.RunCallbacks()
	tr.Tick()
	if reg1Calls != 3 {
		t.Fatalf("expecting 3 calls to reg1, got %d", reg1Calls)
	}
	if reg2Calls != 3 {
		t.Fatalf("expecting 3 calls to reg2, got %d", reg2Calls)
	}
	want := trace{
		signals: map[string]*waveform{
			"Unit.reg1": &waveform{
				kind:   "reg",
				name:   "Unit.reg1",
				bits:   6,
				values: []datapoint{{0, 42}, {1, 0}},
			},
			"Unit.reg2": &waveform{
				kind:   "reg",
				name:   "Unit.reg2",
				bits:   1,
				values: []datapoint{{0, 1}, {2, 0}},
			},
		},
		curTime: 3,
	}
	assertTracesAreEqual(t, tr, &want)
}
