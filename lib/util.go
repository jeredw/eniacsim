package lib

import (
	"fmt"
)

func ToBin(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func MsToAddCycles(ms int) int {
	return ms * 5000 / 1000
}

func SafePlug(jackName string, jack *chan Pulse, ch chan Pulse) {
	if *jack != nil {
		fmt.Printf("Duplicate connection on %s\n", jackName)
	}
	*jack = ch
}
