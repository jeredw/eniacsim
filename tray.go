package main

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/jeredw/eniacsim/lib"
)

// Trays models ENIAC's data and program control lines.
type Trays struct {
	data    [20]*Jack
	program [11][11]*Jack
}

func NewTrays() *Trays {
	t := &Trays{}
	forward := func(j *Jack, val int) {
		j.Transmit(val)
	}
	for i := range t.data {
		t.data[i] = NewJack(fmt.Sprintf("%d", i+1), forward, nil)
	}
	for i := range t.program {
		for j := range t.program[0] {
			t.program[i][j] = NewJack(fmt.Sprintf("%d-%d", i+1, j+1), forward, nil)
		}
	}
	return t
}

func (t *Trays) Reset() {
	for i := range t.data {
		t.data[i].Disconnect()
	}
	for i := range t.program {
		for j := range t.program[0] {
			t.program[i][j].Disconnect()
		}
	}
}

func (t *Trays) FindJack(name string) (*Jack, error) {
	dash := strings.IndexByte(name, '-')
	if dash == -1 {
		tray, _ := strconv.Atoi(name)
		if !(tray >= 1 && tray <= 20) {
			return nil, fmt.Errorf("invalid data trunk %s", name)
		}
		return t.data[tray-1], nil
	}
	tray, _ := strconv.Atoi(name[:dash])
	if !(tray >= 1 && tray <= 11) {
		return nil, fmt.Errorf("invalid program trunk %s", name)
	}
	if len(name) <= dash+1 {
		return nil, fmt.Errorf("invalid program trunk %s", name)
	}
	line, _ := strconv.Atoi(name[dash+1:])
	if !(line >= 1 && line <= 11) {
		return nil, fmt.Errorf("invalid program trunk %s", name)
	}
	return t.program[tray-1][line-1], nil
}
