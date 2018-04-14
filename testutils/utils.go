package testutils

import "testing"

func AssertStringEqual(t *testing.T, actual, expected string) {
	if actual != expected {
		t.Errorf("Expected value (%v) != Actual value (%v)", expected, actual)
	}
}

func AssertStringNotEqual(t *testing.T, actual, notexpected string) {
	if actual == notexpected {
		t.Errorf("Expected value (%v) == Actual value (%v)", notexpected, actual)
	}
}

func AssertBoolEqual(t *testing.T, actual, expected bool) {
	if actual != expected {
		t.Errorf("Expected value (%v) != Actual value (%v)", expected, actual)
	}
}

func AssertBoolNotEqual(t *testing.T, actual, notexpected bool) {
	if actual == notexpected {
		t.Errorf("Expected value (%v) == Actual value (%v)", notexpected, actual)
	}
}
