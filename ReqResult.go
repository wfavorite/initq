package initq

/* ------------------------------------------------------------------------ */

// ReqResult is the type returned by a requirement function. It is the type
// backing the various enum values.
type ReqResult int

/* ------------------------------------------------------------------------ */

const (
	// UnRun is the initialized state when the requirement has not run.
	UnRun ReqResult = iota

	// Satisfied means that the requirement was completed, and does not need
	// to run again.
	Satisfied

	// TryAgain is returned when a requirement has yet to be satisfied.
	// A requirement cannot complete if a dependent requirement has not yet
	// been handled.
	TryAgain

	// Stop is returned when the Q should be stopped early, without
	// error.
	Stop
)
