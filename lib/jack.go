package lib

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Plugboard owns one or more jacks and provides a way to find them by name.
type Plugboard interface {
	FindJack(name string) (*Jack, error)
}

// RatsNest tracks all interconnected jacks in the system.
type RatsNest struct {
	jacks map[string]*Jack
}

type JackHandler func(*Jack, int)

type Jack struct {
	Name       string
	OnReceive  JackHandler
	OnTransmit JackHandler
	Receivers  []*Jack
	Disabled   bool // to skip work for inactive accum inputs
	Connected  bool
	OutJack    *Jack

	visited bool
	forward bool // forwarding node (for trays)
}

func newJack(name string, onReceive JackHandler, onTransmit JackHandler) *Jack {
	return &Jack{
		Name:       name,
		OnReceive:  onReceive,
		OnTransmit: onTransmit,
		Receivers:  make([]*Jack, 0, 1),
	}
}

func NewInput(name string, onReceive JackHandler) *Jack {
	return newJack(name, onReceive, nil)
}

func NewOutput(name string, onTransmit JackHandler) *Jack {
	return newJack(name, nil, onTransmit)
}

func NewForwardingJack(name string) *Jack {
	jack := newJack(name, nil, nil)
	jack.forward = true
	return jack
}

// Transmit sends val on jack j, invoking receiver callbacks for each connected
// receiver and afterwards invoking j's transmit callback.
func (j *Jack) Transmit(val int) {
	if j.visited {
		// A previous call to Transmit() on this jack triggered this call.  Break
		// the cycle and return early here.
		//
		// This isn't an error, and can happen legitimately when e.g. two trunks
		// are connected, like p 1 2.  Transmitting on trunk 1 will call transmit
		// on trunk 2, which will attempt to transmit on trunk 1 again.
		return
	}
	j.visited = true
	transmitted := false
	for _, r := range j.Receivers {
		if r.forward {
			transmitted = true
			r.Transmit(val)
		} else if !r.visited && !r.Disabled {
			transmitted = true
			r.OnReceive(r, val)
		}
	}
	if transmitted && j.OnTransmit != nil {
		j.OnTransmit(j, val)
	}
	j.visited = false
}

func (j *Jack) String() string {
	return j.Name
}

func (j *Jack) ConnectionsString() string {
	if len(j.Receivers) == 0 {
		return "unconnected\n"
	}
	var b strings.Builder
	for _, r := range j.Receivers {
		fmt.Fprintf(&b, "%s %s\n", j.String(), r.String())
	}
	return b.String()
}

// Connect connects two jacks, warning about pathological connections.
func Connect(r *RatsNest, j1, j2 *Jack) error {
	if j1 == j2 {
		return fmt.Errorf("%s cannot be connected to itself", j1)
	}
	for i := range j1.Receivers {
		if j1.Receivers[i] == j2 {
			return fmt.Errorf("%s is already connected to %s", j1, j2)
		}
	}
	for i := range j2.Receivers {
		if j2.Receivers[i] == j1 {
			return fmt.Errorf("%s is already connected to %s", j1, j2)
		}
	}
	if j1.OnReceive != nil || j1.forward {
		j2.Receivers = append(j2.Receivers, j1)
		j2.Connected = true
	}
	if j2.OnReceive != nil || j2.forward {
		j1.Receivers = append(j1.Receivers, j2)
		j1.Connected = true
	}
	r.jacks[j1.Name] = j1
	r.jacks[j2.Name] = j2
	return nil
}

func NewRatsNest() *RatsNest {
	return &RatsNest{
		jacks: make(map[string]*Jack, 10),
	}
}

func (r *RatsNest) DumpGraph(w io.Writer) {
	pa := regexp.MustCompile(`pa.(\d+).sa.(\d+)`)
	ct := regexp.MustCompile(`c.(\d+)i`)
	mt := regexp.MustCompile(`m.(\d+)i`)
	at := regexp.MustCompile(`a(\d+).(5|6|7|8|9|10|11|12)i`)
	ft := regexp.MustCompile(`f(\d+).(\d+)i`)
	it := regexp.MustCompile(`i.Ci(\d+)`)
	mp := regexp.MustCompile(`p.([ABCDEFGHJK])i`)
	fmt.Fprintf(w, "digraph eniac {\n")
	fmt.Fprintf(w, "\trankdir=\"LR\";\n")
	for fromName, fromJack := range r.jacks {
		for _, toJack := range fromJack.Receivers {
			fmt.Fprintf(w, "\t\"%s\" -> \"%s\";\n", fromName, toJack.Name)
		}
		if strings.HasPrefix(fromName, "ad.permute.i.") {
			outputJack := strings.Replace(fromName, "ad.permute.i.", "ad.permute.o.", 1)
			fmt.Fprintf(w, "\t\"%s\" -> \"%s\";\n", fromJack, outputJack)
		}
		if strings.HasPrefix(fromName, "ad.s.i.") {
			outputJack := strings.Replace(fromName, "ad.s.i.", "ad.s.o.", 1)
			fmt.Fprintf(w, "\t\"%s\" -> \"%s\";\n", fromJack, outputJack)
		}
		if strings.HasPrefix(fromName, "ad.dp.i") {
			outBase := strings.Replace(fromName, "ad.dp.i.", "ad.dp.o.", 1)
			for k := 0; k < 12; k++ {
				outputJack := fmt.Sprintf("%s.%d", outBase, k)
				if _, ok := r.jacks[outputJack]; ok {
					fmt.Fprintf(w, "\t\"%s\" -> \"%s\";\n", fromJack, outputJack)
				}
			}
		}
		if parts := pa.FindStringSubmatch(fromName); len(parts) != 0 {
			fmt.Fprintf(w, "\t\"%s\" -> \"pa.%s.sb.%s\";\n", fromJack, parts[1], parts[2])
		}
		if parts := ct.FindStringSubmatch(fromName); len(parts) != 0 {
			fmt.Fprintf(w, "\t\"%s\" -> \"c.%so\" [ style=dashed ];\n", fromName, parts[1])
		}
		if parts := mt.FindStringSubmatch(fromName); len(parts) != 0 {
			fmt.Fprintf(w, "\t\"%s\" -> \"m.%so\" [ style=dashed ];\n", fromName, parts[1])
		}
		if parts := at.FindStringSubmatch(fromName); len(parts) != 0 {
			fmt.Fprintf(w, "\t\"%s\" -> \"a%s.%so\" [ style=dashed ];\n", fromName, parts[1], parts[2])
		}
		if parts := ft.FindStringSubmatch(fromName); len(parts) != 0 {
			fmt.Fprintf(w, "\t\"%s\" -> \"f%s.%so\" [ style=dashed ];\n", fromName, parts[1], parts[2])
		}
		if parts := it.FindStringSubmatch(fromName); len(parts) != 0 {
			fmt.Fprintf(w, "\t\"%s\" -> \"i.Co%s\" [ style=dashed ];\n", fromName, parts[1])
		}
		if parts := mp.FindStringSubmatch(fromName); len(parts) != 0 {
			for k := 0; k < 6; k++ {
				outputJack := fmt.Sprintf("p.%s%do", parts[1], k+1)
				if _, ok := r.jacks[outputJack]; ok {
					fmt.Fprintf(w, "\t\"%s\" -> \"%s\" [ style=dashed ];\n", fromJack, outputJack)
				}
			}
		}
	}
	// NOTE this connectivity of multiplier outputs is peculiar to eniac-chess
	fmt.Fprintf(w, "\t\"m.1i\" -> \"m.Rα\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.2i\" -> \"m.Rβ\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.3i\" -> \"m.Rγ\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.4i\" -> \"m.Rδ\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.5i\" -> \"m.Rε\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.6i\" -> \"m.Dα\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.7i\" -> \"m.Dβ\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.8i\" -> \"m.Dγ\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.9i\" -> \"m.Dδ\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.10i\" -> \"m.Dε\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.16i\" -> \"m.A\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.17i\" -> \"m.A\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.18i\" -> \"m.A\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.19i\" -> \"m.A\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.20i\" -> \"m.S\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.21i\" -> \"m.S\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.22i\" -> \"m.S\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.23i\" -> \"m.S\" [ style=dashed ];\n")
	fmt.Fprintf(w, "\t\"m.24i\" -> \"m.S\" [ style=dashed ];\n")
	// NOTE this formatting for MP outputs is peculiar to eniac-chess
	decodeOutputs := []string{
		"p.B1o", "p.B2o", "p.B3o", "p.B4o", "p.B5o", "p.B6o", "p.C1o", "p.C2o", "p.C3o", "p.C4o", "p.C5o", "p.C6o", "p.D1o", "p.D2o", "p.D3o", "p.D4o", "p.D5o", "p.D6o", "p.E1o", "p.E2o", "p.E3o", "p.E4o", "p.E5o", "p.E6o", "p.F1o", "p.F2o", "p.F3o", "p.F4o", "p.F5o", "p.F6o", "p.G4o", "p.G5o", "p.G6o", "p.H1o", "p.H2o", "p.H3o", "p.H4o", "p.H5o", "p.H6o", "p.J1o", "p.J2o", "p.J3o", "p.J4o", "p.J5o", "p.J6o", "p.K1o", "p.K2o", "p.K3o", "p.K4o", "p.K5o", "p.K6o",
	}
	for _, decodeOutput := range decodeOutputs {
		fmt.Fprintf(w, "\t\"%s\" [ style=filled ];\n", decodeOutput)
	}
	fmt.Fprintf(w, "\t{ rank=same \"%s\"}\n", strings.Join(decodeOutputs, `" "`))
	// HACK for eniac-chess memory system
	accumDecoders := []string{
		"a6.S", "a7.S", "a16.S", "a9.S",
	}
	fmt.Fprintf(w, "\t{ rank=same \"%s\"}\n", strings.Join(accumDecoders, `" "`))
	fmt.Fprintf(w, "}\n")
}
