package initq

import (
	"fmt"
	"slices"
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

	// VOIR: This was originally a direct assignment. Technically, a direct
	// VOIR: assignment is incorrect as you can not be sure that the backing
	// VOIR: array that is referenced does not change. In this case (of an
	// VOIR: internal error) the backing data does not change. I chose to do
	// VOIR: the proper way when the simple / incorrect would have worked.
	// VOIR:
	// VOIR: There are multiple methods for the copy... a for-range loop,
	// VOIR: a make-copy sequence, or slices.Clone(). I chose slices.Clone().
	err.unsat = slices.Clone(remains)

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
	// Simple assignment of the array. Unlike the creation of the error where
	// the input is copied, this is just a reference to the original in the
	// error struct.
	unsat = qur.unsat
	return
}
