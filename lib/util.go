package lib

func ToBin(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func MsToAddCycles(ms int) int {
	return ms * 5000 / 1000
}
