package initq

import (
	"fmt"
	"strings"
)

/* ------------------------------------------------------------------------ */

// QUnresolvable is a specific error type that may be checked for. It is
// returned in the TryProcess case when a Q cannot be satisfied. It affords
// the opportunity to handle what might be a user-related error, or at least
// in a way that does not cause Fatal() assertions.
//
// In addition to the standard Error() method, this includes an
// UnresolvedTasks() method that lists the tasks that could not be completed.
type QUnresolvable struct {
	unsat []string
}

/* ======================================================================== */

// newQUnresolvable creates a new error that has a retrievable list of
// unsatisfied requirements.
func newQUnresolvable(remains []string) (err *QUnresolvable) {
	err = new(QUnresolvable)

	// Just assign the list directly.
	err.unsat = remains

	return err
}

/* ======================================================================== */

// Error returns a single message that satisfies the error interface.
func (qur QUnresolvable) Error() (msg string) {

	if len(qur.unsat) > 0 {
		msg = fmt.Sprintf("run Q cannot be satisfied (%s remain)", strings.Join(qur.unsat, ","))
	} else {
		msg = "run Q cannot be satisfied"
	}
	return
}

/* ======================================================================== */

// UnresolvedTasks returns the tasks that were not satisfied. This eliminates
// the need to parse them out of the Error() output.
func (qur QUnresolvable) UnresolvedTasks() (unsat []string) {
	// Simple assignment of the array.
	unsat = qur.unsat
	return
}
