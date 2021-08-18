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

	OutputConnected bool // to skip sending 1pp when S outputs not connected
	Disabled        bool // to skip work for inactive accum inputs

	finalReceivers []*Jack // receivers after routing

	visited  bool
	forward  bool // jack forwards inputs
	polarity int  // polarity (0=unspecified, 1=input, 2=output, 3=both)

	OtherSide *Jack // jack for other side of adapter
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

func NewRoutingJack(name string, polarity int) *Jack {
	jack := newJack(name, nil, nil)
	jack.polarity = polarity
	jack.forward = true
	return jack
}

// Transmit sends val on jack j, invoking receiver callbacks for each connected
// receiver and afterwards invoking j's transmit callback.
func (j *Jack) Transmit(val int) {
	transmitted := false
	for _, r := range j.finalReceivers {
		if !r.Disabled {
			transmitted = true
			r.OnReceive(r, val)
		}
	}
	if transmitted && j.OnTransmit != nil {
		j.OnTransmit(j, val)
	}
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
	if j2.isOutput() && j1.isInput() {
		j2.Receivers = append(j2.Receivers, j1)
		j2.OutputConnected = true
	}
	if j1.isOutput() && j2.isInput() {
		j1.Receivers = append(j1.Receivers, j2)
		j1.OutputConnected = true
	}
	r.jacks[j1.Name] = j1
	r.jacks[j2.Name] = j2
	r.updateFinalReceivers()
	return nil
}

func (j *Jack) isInput() bool {
	switch j.polarity {
	case 0:
		return j.OnReceive != nil
	case 1:
		return true
	case 2:
		return false
	case 3:
		return true
	}
	return false
}

func (j *Jack) isOutput() bool {
	switch j.polarity {
	case 0:
		return j.OnReceive == nil
	case 1:
		return false
	case 2:
		return true
	case 3:
		return true
	}
	return false
}

func NewRatsNest() *RatsNest {
	return &RatsNest{
		jacks: make(map[string]*Jack, 10),
	}
}

func (r *RatsNest) updateFinalReceivers() {
	for _, jack := range r.jacks {
		if jack.isOutput() && !jack.forward {
			jack.finalReceivers = findFinalReceivers(jack, make([]*Jack, 0, 4))
			if jack.OtherSide != nil {
				// When adapters have only one connected receiver, hardwire it
				if len(jack.finalReceivers) == 1 {
					(jack.OtherSide).OtherSide = jack.finalReceivers[0]
				} else {
					(jack.OtherSide).OtherSide = jack
				}
			}
		}
	}
}

func findFinalReceivers(j *Jack, receivers []*Jack) []*Jack {
	if j.visited {
		return receivers
	}
	j.visited = true
	for _, r := range j.Receivers {
		if !r.forward {
			// r is a non-routing receiver
			receivers = append(receivers, r)
		} else {
			// search through forwarder's receivers to find non-routing connections
			receivers = findFinalReceivers(r, receivers)
		}
	}
	j.visited = false
	return receivers
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
		if strings.HasPrefix(fromName, "pa.") {
			// pas are converted to edges
			continue
		}
		if matched, _ := regexp.MatchString(`^(\d+)|(\d+-\d+)$`, fromName); matched {
			fmt.Fprintf(w, "\t\"%s\" [ shape=diamond label=\"\" style=filled fillcolor=darkgray ];\n", fromName)
		}
		for _, toJack := range fromJack.finalReceivers {
			if parts := pa.FindStringSubmatch(toJack.Name); len(parts) != 0 {
				bSideName := fmt.Sprintf("pa.%s.sb.%s", parts[1], parts[2])
				if bSide, ok := r.jacks[bSideName]; ok {
					for _, toJack := range bSide.Receivers {
						fmt.Fprintf(w, "\t\"%s\" -> \"%s\" [ color=\"gray:invis:gray\" ];\n", fromName, toJack.Name)
					}
				} else {
					panic("missing pa output")
				}
				continue
			}
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
		if strings.HasPrefix(fromName, "ad.d.i.") {
			outputJack := strings.Replace(fromName, "ad.d.i.", "ad.d.o.", 1)
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
	fmt.Fprintf(w, "\t\"i.Pi\" -> \"i.Po\" [ style=dashed ];\n")
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
		fmt.Fprintf(w, "\t\"%s\" [ style=filled shape=box ];\n", decodeOutput)
	}
	fmt.Fprintf(w, "\t{ rank=same \"%s\"}\n", strings.Join(decodeOutputs, `" "`))
	// Making digit sources "sources" in the graph just generally makes sense.
	sources := []string{
		"f1.A", "f1.B", "f2.A", "f2.B", "f3.A", "f3.B",
		"c.o",
		"a1.A", "a1.S", "a2.A", "a2.S", "a3.A", "a3.S", "a4.A", "a4.S", "a5.A", "a5.S", "a6.A", "a6.S", "a7.A", "a7.S", "a8.A", "a8.S", "a9.A", "a9.S", "a10.A", "a10.S", "a11.A", "a11.S", "a12.A", "a12.S", "a13.A", "a13.S", "a14.A", "a14.S", "a15.A", "a15.S", "a16.A", "a16.S", "a17.A", "a17.S", "a18.A", "a18.S", "a19.A", "a19.S", "a20.A", "a20.S",
	}
	for _, source := range sources {
		fmt.Fprintf(w, "\t\"%s\" [ style=bold ];\n", source)
	}
	fmt.Fprintf(w, "\t{ rank=source \"%s\"}\n", strings.Join(sources, `" "`))
	fmt.Fprintf(w, "}\n")
}
