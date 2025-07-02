// Package initq implements an application startup 'run' queue of requirements
// for the application to run. A run Q is filled with functions / methods
// that are responsible for identifying requirements and handling individual
// initialization tasks.
//
// Application startup is reduced to a series of tasks that know about the
// required dependencies to start.
//
// For example:
//   - Command line is first read to find config file location.
//   - Config file is read to find service configuration.
//   - Service is configured from the specified config.
//
// As long as each "Run Q requirement function" knows about dependencies (or
// the dependencies are expressed as 'built-in semaphores') then the items
// can be placed into the Run Q in any order.
//
// The example as code:
//
//	cd := new(CoreData)
//	iq := initq.NewInitQ()
//
//	iq.Add("cmdline", cd.ParseCommandLine) // No dependencies
//	iq.Add("config", cd.ReadConfig)        // Command line must have been read
//	iq.Add("service", cd.StartService)     //
//
//	if err := iq.Process() ; err != nil {
//		fmt.Fprintln(os.Stderr, "ERROR:", cd.ErrorMsg)
//		os.Exit(1)
//	}
//
// Requirement functions are generally expected to *know* their dependencies
// and are capable of determining if they were satisfied (before they can
// conplete). In the above example all three init functions may allocate
// new structs (of config, command-line, and service options) that serve as
// indication that a dependency was completed. It is also to explicitly
// specify dependent modules.
//
// For example:
//
//		rq.Add("settime", SetSystemClock)            // No clear indicator of success
//	 rq.Add("runsvc" , cd.RunService, "settime")  // Explicitly requres settime success
package initq

import (
	"log"
	"strings"
)

/* ------------------------------------------------------------------------ */

// QFunc is the prototype for a InitQ requirement.
type QFunc func() ReqResult

/* ------------------------------------------------------------------------ */

// InitQ is the primary / core structure for the module. All public methods
// are based on this structure.
type InitQ struct {
	// q is the to-run list.
	q []*initQItem
}

/* ======================================================================== */

// NewInitQ creates a new initialized / empty InitQ.
func NewInitQ() (rq *InitQ) {

	rq = new(InitQ)

	return
}

/* ======================================================================== */

// Add puts InitQ requirements on the initialization Q. Invalid input is fatal.
//
// The name is a convenient label that may be used in messaging.
func (rq *InitQ) Add(name string, f QFunc, deps ...string) {

	// Fatal on misuse is appropriate.
	// This is better than letting the user think things went ok when they
	// did not. This is not a random runtime fatal error, but one that is
	// designed to be caught early / in test.
	if rq == nil {
		log.Fatal("Add called on a nil InitQ.")
	}

	// Check inputs...
	// ...but we only care about the funciton input.
	if f == nil {
		log.Fatal("Method Add called on a nil function.")
	}

	// Initialize and append to the Q.
	rqi := newInitQItem(name, f, deps...)
	rq.q = append(rq.q, rqi)

}

/* ======================================================================== */

// Process is used to iteratively work all items in the Q until they are
// satisfied. If the Q cannot be processed to completion in an expected number
// of itemrations, then an error is returned.
func (rq *InitQ) Process() (err error) {

	// Fatal is appropriate.
	// Discussion on *why* is in the Add method.
	if rq == nil {
		log.Fatal("Method Add called on a nil function.")
	}

	passes := 0
	qlen := len(rq.q)

	// The top loop drops us out when we have exceeded the maximum possible
	// passes.
	for passes <= qlen {

		satisfied := true

		// The next loop is a pass of the InitQ.
		for _, rqi := range rq.q {

			// Check for dependencies.
			allDepsGood := true
			for _, dep := range rqi.deps {
				if rq.satisfied(dep) == false {
					allDepsGood = false
				}
			}

			if allDepsGood == false {
				rqi.state = TryAgain
				satisfied = false
				continue
			}

			// "run" each item. If previously satisfied, the run will be
			// skipped.
			switch rqi.run() {
			case UnRun:
				// This case really should not need to be handled here. I am
				// leaving this here in the event design changes such that
				satisfied = false
			case TryAgain:
				satisfied = false
			case Stop:
				// This returns the ONLY error in this method. All others
				// are asserts.
				return ErrQStopped
			}
		}

		passes++

		if satisfied {
			return
		}
	}

	// The Q has now run as many times as there are items in the Q. Assuming a
	// worst case ordering of requirements, it *should* be satified by now.
	//
	// Reaching here is an assert/fatal condiditon. It is not a transient
	// runtime thing, but a failure in setup of the module. (Eg: defining a Q
	// with circular / unresolvable requirements.)
	//
	// The goal here is to help the user understand what items were unable to
	// be satisfied.
	//
	// The return / exit type can be modified with the BehaveUnresolvIsErr
	// behaviour 'toggle'.

	if BehaveUnresolvIsErr == false {

		remaining := make([]string, 0)
		for _, rqi := range rq.q {
			if rqi.state == TryAgain {
				remaining = append(remaining, rqi.name)
			}
		}
		log.Fatalf("run Q cannot be satisfied (%s remain)", strings.Join(remaining, ","))
	}

	// This is unreachable under normal circumstances.
	return ErrQUnsolvable
}

/* ======================================================================== */

// satisfied reports if a named requirement has been satisfied. This is used
// to check required dependencies of a requiriement.
func (rq *InitQ) satisfied(name string) bool {

	for _, rqi := range rq.q {

		// This is a dep we care about.
		if rqi.name == name {
			if rqi.state == Satisfied {
				return true
			}
		}
	}

	return false
}
