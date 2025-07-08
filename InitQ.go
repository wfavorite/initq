// Package initq implements an application startup 'run' queue of requirements
// for the application to run. An init Q is filled with functions / methods
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
// As long as each "Init Q requirement function" knows about dependencies (or
// the dependencies are expressed as 'built-in semaphores') then the items
// can be placed into the Init Q in any order.
//
// The example as code:
//
//	cd := new(CoreData)
//	iq := initq.NewInitQ()
//
//	iq.Add("cmdline", cd.ParseCommandLine) // No dependencies
//	iq.Add("config", cd.ReadConfig)        // Command line must be parsed
//	iq.Add("service", cd.StartService)     // Config must be read
//
//	if err := iq.Process() ; err != nil {
//		fmt.Fprintln(os.Stderr, "ERROR:", cd.ErrorMsg)
//		os.Exit(1)
//	}
//
// Requirement functions are generally expected to *know* their dependencies
// and are capable of determining if they were satisfied (before they can
// complete). In the above example all three init functions may allocate
// new structs (of config, command-line, and service options) that serve as
// indication that a dependency was completed. It is also to explicitly
// specify dependent modules.
//
// For example:
//
//	rq.Add("settime", SetSystemClock)            // No clear indicator of success
//	rq.Add("runsvc" , cd.RunService, "settime")  // Explicitly requires settime success
//
// It is possible that an init queue could be constructed that cannot be
// satisfied. This happens when circular dependencies are created or the
// task fails to detect dependent tasks and/or never returns a Satisfied value.
// These conditions are considered 'build time' problems, and will trigger
// a log.Fatal() assertion - such that the problem is likely to be discovered
// in test rather than regular use.
package initq

import (
	"fmt"
	"log"
	"slices"
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

	// addErr captures errors in the Add call so that they may be handled in
	// the Process call. The intent is to keep Add calls 'clean', yet still
	// capture failures in a testable manner.
	addErr string
}

/* ======================================================================== */

// NewInitQ creates a new initialized / empty InitQ.
//
// Use the Add method to add required tasks to the queue, and Process to
// run the queue to completion.
func NewInitQ() (rq *InitQ) {

	rq = new(InitQ)

	return
}

/* ======================================================================== */

// Add puts InitQ requirements on the initialization Q. Invalid input is fatal.
//
// The name is a convenient label that may be used in messaging. The label is
// case-sensitive, so any dependency requirements must match exactly. (A const
// label is appropriate here.)
//
// The second parameter is the function reference. This function is responsible
// for understanding dependent attributes (that may be derived from environment
// state), and returns the status of the initialization attempt.
//
// The final (optional) parameters are a means of expressing dependent
// required tasks if completion cannot be derived from the environment.
func (rq *InitQ) Add(name string, f QFunc, deps ...string) {

	// Fatal on misuse is appropriate.
	// This is better than letting the user think things went ok when they
	// did not. This is not a random runtime fatal error, but one that is
	// designed to be caught early / in test.
	if rq == nil {
		log.Fatal("Add called on a nil InitQ.")
	}

	// Check inputs.
	if len(name) == 0 {
		rq.addErr = "Add called with an empty name label."
		if BehaveUnresolvIsErr {
			return
		}
		log.Fatal(rq.addErr)
	}

	// A function reference must be passed.
	if f == nil {
		rq.addErr = fmt.Sprintf("Add(%s) called with a nil function.", name)
		if BehaveUnresolvIsErr {
			return
		}
		log.Fatal(rq.addErr)
	}

	// None of the deps should self-reference.
	for _, d := range deps {
		if d == name {
			rq.addErr = fmt.Sprintf("Add(%s) called with a self-referencing dependency.", name)
			if BehaveUnresolvIsErr {
				return
			}
			log.Fatal("Unable to add a self-referencing dependency.")
		}
	}

	// Initialize and append to the Q.
	rqi := newInitQItem(name, f, deps...)
	rq.q = append(rq.q, rqi)

}

/* ======================================================================== */

// Process is used to iteratively work all items in the Q until they are
// satisfied. If the Q cannot be processed to completion in an expected number
// of iterations, then a log.Fatal() is asserted.
//
// Under normal conditions, the only error returned from this method is the
// ErrQStopped error. This is returned when a requirement function (sets an
// error and) returns the Stop value.
func (rq *InitQ) Process() (err error) {
	// The default behaviour.
	return rq.process(false)
}

/* ======================================================================== */

// TryProcess is used to iteratively work all items in the Q until they are
// satisfied. If the Q cannot be processed to completion in an expected number
// of iterations, then an error of type ErrQUnsolvable is returned.
//
// Under normal conditions, the only error returned from this method is the
// ErrQStopped error. This is returned when a requirement function (sets an
// error and) returns the Stop value. If user input causes an unresolvable Q
// and the desire is to handle this as an error (much easier to pass to the
// user than a Fatal() call) then this function is more appropriate than the
// Process() variant.
func (rq *InitQ) TryProcess() (err error) {
	// The modified behaviour.
	return rq.process(true)
}

/* ======================================================================== */

// process is the common implementation of both Process and TryProcess. It
// takes a boolean to enable (true) the return of a dedicated error, rather
// than a log.Fatal(). The error is comparable, so will not have distinct
// messaging about why the Q could not be satisfied, and the caller will need
// to handle that
func (rq *InitQ) process(unsatIsError bool) (err error) {

	// Fatal is appropriate.
	// Discussion on *why* is in the Add method.
	if rq == nil {
		log.Fatal("Method Process called on a nil function.")
	}

	// Handle any errors that may have been created. There is no need to test
	// the behaviour as that is the only way this internal error message is
	// set.
	if len(rq.addErr) > 0 {
		// The error message is kind of useless as it is not explicitly
		// returned. The value is the ability to test for the specific error
		// type.
		return fmt.Errorf("%s", rq.addErr)
	}

	// Check to see if any dependencies are 'dangling'. This is the case
	// where a 'semaphore' dependency references a task that does not exist.
	// This cannot be checked in the Add calls.
	// First build a simpler lookup list.
	validLabels := make([]string, 0)
	for _, task := range rq.q {

		// This part *could* be done in Add - but easier here.
		if slices.Contains(validLabels, task.name) {

			fatalMsg := fmt.Sprintf("The %s task label was used more than once.", task.name)
			if BehaveUnresolvIsErr {
				// This is unreachable under normal circumstances.
				return fmt.Errorf("%s", fatalMsg)
			}
			log.Fatalf("%s", fatalMsg)
		}

		validLabels = append(validLabels, task.name)
	}
	// Now walk all dependencies looking for solid matches.
	for _, task := range rq.q {
		for _, dep := range task.deps {
			if !slices.Contains(validLabels, dep) {
				fatalMsg := fmt.Sprintf("Task %s has dependency %s that does not match any existing task.", task.name, dep)
				if BehaveUnresolvIsErr {
					// This is unreachable under normal circumstances.
					return fmt.Errorf("%s", fatalMsg)
				}
				log.Fatalf("%s", fatalMsg)
			}
		}
	}
	// End of dependency / label sanity checks.

	passes := 0
	qlen := len(rq.q)

	// The top loop drops us out when we have exceeded the maximum possible
	// passes.
	for passes <= qlen {

		// Assume the Q has been satisfied - unless shown otherwise.
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
			// skipped. We only care about the 'unsatisfied' cases (that prove
			// the Q unsatisfied) - which means we go around again.
			switch rqi.run() {
			case UnRun:
				// This case really should not need to be handled here. I am
				// leaving this here in the event design changes such that it
				// comes to be. Testing for it will be difficult without some
				// sort of complication / interface on the run method. It is
				// at least captured and handled.
				fatalMsg := fmt.Sprintf("Failed to process task %s.", rqi.name)
				if BehaveUnresolvIsErr {
					return fmt.Errorf("%s", fatalMsg)
				}
				log.Fatalf("%s", fatalMsg)
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
	// worst case ordering of requirements, it *should* be satisfied by now.
	//
	// Reaching here is (typically) an assert/fatal condition. It is not a
	// transient runtime thing, but a failure in setup of the module. (Eg:
	// defining a Q with circular / unresolvable requirements.)
	//
	// The goal here is to help the caller understand what items were unable to
	// be satisfied / not cause a Fatal() assertion when undesirable.
	//
	// The return / exit type can be modified with the BehaveUnresolvIsErr
	// behaviour 'toggle' or the unsatIsError method parameter.

	// Generate the error message content (even if it is not used).
	remaining := make([]string, 0)
	for _, rqi := range rq.q {
		if rqi.state == TryAgain {
			remaining = append(remaining, rqi.name)
		}
	}

	// The explicit / priority case: The caller wants a meaningful message.
	if unsatIsError {
		err = newQUnresolvable(remaining)
		return
	}

	// This *excludes* the testable case - with a standard / comparable error.
	// The value is inverted (== false) so that the method ends with a return.
	if BehaveUnresolvIsErr == false {
		log.Fatalf("run Q cannot be satisfied (%s remain)", strings.Join(remaining, ","))
	}

	// The original / designed for test case.
	return ErrQUnsolvable

}

/* ======================================================================== */

// satisfied reports if a named requirement has been satisfied. This is used
// to check required dependencies of a requirement.
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
