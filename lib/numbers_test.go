package lib

import (
	"testing"
)

func TestTensComplementToIBMCard_P(t *testing.T) {
	got := TensComplementToIBMCard('P', "42")
	expected := "42"
	if got != expected {
		t.Errorf("expected %s got %s", expected, got)
	}
}

func TestTensComplementToIBMCard_N0(t *testing.T) {
	got := TensComplementToIBMCard('M', "0")
	expected := "0"
	if got != expected {
		t.Errorf("expected %s got %s", expected, got)
	}
}

func TestTensComplementToIBMCard_N10(t *testing.T) {
	got := TensComplementToIBMCard('M', "10")
	expected := "R0"
	if got != expected {
		t.Errorf("expected %s got %s", expected, got)
	}
}

func TestTensComplementToIBMCard_N123000(t *testing.T) {
	got := TensComplementToIBMCard('M', "123000")
	expected := "Q77000" // 877000 i.e. 1000000 - 123000
	if got != expected {
		t.Errorf("expected %s got %s", expected, got)
	}
}

func TestTensComplementToIBMCard_N123(t *testing.T) {
	got := TensComplementToIBMCard('M', "321")
	expected := "O79" // 679 i.e. 1000 - 321
	if got != expected {
		t.Errorf("expected %s got %s", expected, got)
	}
}

func TestTensComplementToIBMCard_N1(t *testing.T) {
	got := TensComplementToIBMCard('M', "9999999999")
	expected := "-000000001" // -1
	if got != expected {
		t.Errorf("expected %s got %s", expected, got)
	}
}

func sliceEquals(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestIBMCardToNinesComplement_P(t *testing.T) {
	gotSign, gotDigits := IBMCardToNinesComplement("42")
	expectedSign, expectedDigits := false, []int{2, 4}
	if gotSign != expectedSign {
		t.Errorf("expected sign %v got %v", expectedSign, gotSign)
	}
	if !sliceEquals(gotDigits, expectedDigits) {
		t.Errorf("expected digits %v got %v", expectedDigits, gotDigits)
	}
}

func TestIBMCardToNinesComplement_N0(t *testing.T) {
	gotSign, gotDigits := IBMCardToNinesComplement("0")
	expectedSign, expectedDigits := false, []int{0}
	if gotSign != expectedSign {
		t.Errorf("expected sign %v got %v", expectedSign, gotSign)
	}
	if !sliceEquals(gotDigits, expectedDigits) {
		t.Errorf("expected digits %v got %v", expectedDigits, gotDigits)
	}
}

func TestIBMCardToNinesComplement_N10(t *testing.T) {
	gotSign, gotDigits := IBMCardToNinesComplement("R0")
	expectedSign, expectedDigits := true, []int{9, 0}
	if gotSign != expectedSign {
		t.Errorf("expected sign %v got %v", expectedSign, gotSign)
	}
	if !sliceEquals(gotDigits, expectedDigits) {
		t.Errorf("expected digits %v got %v", expectedDigits, gotDigits)
	}
}

func TestIBMCardToNinesComplement_N123000(t *testing.T) {
	gotSign, gotDigits := IBMCardToNinesComplement("Q77000")
	expectedSign, expectedDigits := true, []int{9, 9, 9, 2, 2, 1}
	if gotSign != expectedSign {
		t.Errorf("expected sign %v got %v", expectedSign, gotSign)
	}
	if !sliceEquals(gotDigits, expectedDigits) {
		t.Errorf("expected digits %v got %v", expectedDigits, gotDigits)
	}
}

func TestIBMCardToNinesComplement_N123(t *testing.T) {
	gotSign, gotDigits := IBMCardToNinesComplement("O79")
	expectedSign, expectedDigits := true, []int{0, 2, 3}
	if gotSign != expectedSign {
		t.Errorf("expected sign %v got %v", expectedSign, gotSign)
	}
	if !sliceEquals(gotDigits, expectedDigits) {
		t.Errorf("expected digits %v got %v", expectedDigits, gotDigits)
	}
}

func TestIBMCardToNinesComplement_N1(t *testing.T) {
	gotSign, gotDigits := IBMCardToNinesComplement("-000000001")
	expectedSign, expectedDigits := true, []int{8, 9, 9, 9, 9, 9, 9, 9, 9, 9}
	if gotSign != expectedSign {
		t.Errorf("expected sign %v got %v", expectedSign, gotSign)
	}
	if !sliceEquals(gotDigits, expectedDigits) {
		t.Errorf("expected digits %v got %v", expectedDigits, gotDigits)
	}
}
