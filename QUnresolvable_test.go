package initq

import (
	"strings"
	"testing"
)

/* ======================================================================== */

func TestQUnresolvable(t *testing.T) {

	// Things that may be reused
	var err error
	var msg string

	// -------------
	// Standard / expected / contracted behaviours

	err = newQUnresolvable([]string{"manny", "moe", "jack"})

	msg = err.Error()

	if !strings.Contains(msg, "cannot be satisfied") {
		t.Errorf("Missing the error preamble")
	}

	if !strings.Contains(msg, "remain)") {
		t.Errorf("Missing the error postfix messaging")
	}

	if !strings.Contains(msg, "manny") {
		// This case shows the error message
		t.Errorf("Missing manny.")
		t.Logf("Error is: %s", msg)
	}

	if !strings.Contains(msg, "moe") {
		t.Errorf("Missing moe")
	}

	if !strings.Contains(msg, "jack") {
		t.Errorf("Missing jack")
	}

	if unr, ok := err.(*QUnresolvable); ok {
		tasks := unr.UnresolvedTasks()

		// I could repeatedly use slices.Contains(). Instead i use counts.
		if len(tasks) != 3 {
			t.Errorf("Unexpected number of unresolved tasks found. Expected 3, found %d", len(tasks))
		} else {
			foundExpected := 0
			for _, t := range tasks {
				if t == "manny" {
					foundExpected++
				}
				if t == "moe" {
					foundExpected++
				}
				if t == "jack" {
					foundExpected++
				}
			}

			if foundExpected != 3 {
				t.Errorf("Missing expected set of unsatisfied tasks")
			}
		}
	} else {
		t.Errorf("QUnresolvable type not matched")
	}

	// -------------
	// Misuse / edge case
	// This should never happen. If it does - it is contracted to behave in a way.

	err = newQUnresolvable([]string{})
	msg = err.Error()

	if !strings.Contains(msg, "cannot be satisfied") {
		t.Errorf("Missing the error preamble")
	}

	if strings.Contains(msg, "remain") {
		t.Errorf("The error postfix messaging assumes a list - but none exists")
	}

}
