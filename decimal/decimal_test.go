package decimal

import "testing"

func TestParsePositive(t *testing.T) {
	v, _ := Parse("255.2")
	if int64(v) != 2552000 {
		t.Errorf("Expected %v got %v", Decimal(2552000), v)
	}

	v, err := Parse("25+5.2")
	if err != nil {
		t.Error("Unexpected parsing error")
	} else if int64(v) != 250000 {
		t.Errorf("Expected 25.0000 got %v", v)
	}
}

func TestParseNegative(t *testing.T) {
	v, _ := Parse("-255.2")
	if int64(v) != -2552000 {
		t.Errorf("Expected %v got %v", Decimal(-2552000), v)
	}

	v, err := Parse("25-5.2")
	if err != nil {
		t.Error("Unexpected parsing error")
	} else if int64(v) != 250000 {
		t.Errorf("Expected %v got %v", Decimal(250000), v)
	}
}

func TestParseZero(t *testing.T) {
	v, _ := Parse("0")
	if int64(v) != 0 {
		t.Errorf("Expected %v got %v", Decimal(0), v)
	}
}

func TestParseTruncated(t *testing.T) {
	v, _ := Parse("0.1234567890")
	if int64(v) != 1234 {
		t.Errorf("Expected %v got %v", Decimal(1234), v)
	}
}

func TestParseDotFirst(t *testing.T) {
	v, _ := Parse(".1234")
	if int64(v) != 1234 {
		t.Errorf("Expected %v got %v", Decimal(1234), v)
	}
}

func TestParseDotLast(t *testing.T) {
	v, _ := Parse("1234.")
	if int64(v) != 1234*Divisor {
		t.Errorf("Expected %v got %v", Decimal(1234*Divisor), v)
	}
}

func TestParseMax(t *testing.T) {
	v, _ := Parse("922337203685477.5807")
	if int64(v) != Max {
		t.Errorf("Expected %v got %v", Max, v)
	}

	v, err := Parse("922337203685477.5808")
	if err == nil {
		t.Errorf("Expected error, got %v", v)
	}
}

func TestParseMin(t *testing.T) {
	v, _ := Parse("-922337203685477.5808")
	if int64(v) != Min {
		t.Errorf("Expected %v got %v", Min, v)
	}

	v, err := Parse("-922337203685477.5809")
	if err == nil {
		t.Errorf("Expected error, got %v", v)
	}
}

func TestParseError(t *testing.T) {
	values := []string{"", ".", "blahblah", "hello12"}
	for _, value := range values {
		_, err := Parse(value)
		if err == nil {
			t.Errorf("Parsing of \"%v\" succeded", value)
		}
	}
}

func TestStringRange(t *testing.T) {
	values := []string{
		"125.1200",
		"-1200.2300",
		"0.0000",
		"-922337203685477.5808",
		"922337203685477.5807",
	}
	for _, value := range values {
		v, _ := Parse(value)
		if v.String() != value {
			t.Errorf("Expected %v got %v", value, v.String())
		}
	}
}
