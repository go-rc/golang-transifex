package testutil

import (
	"testing"
)

func AssertEquals(msg, expected, actual string, t *testing.T) {
	if actual != expected {
		t.Errorf("%s: Expected/Actual \n%q\n%q", msg, expected, actual)
	}
}
func AssertEqualsInt(msg string, expected, actual int, t *testing.T) {
	if actual != expected {
		t.Errorf("%s: Expected/Actual \n%q\n%q", msg, expected, actual)
	}
}
