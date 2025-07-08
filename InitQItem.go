package initq

import "log"

/* ------------------------------------------------------------------------ */

// initQItem contains all items necessary to define a required task, as well
// as the optional 'semaphore' expression of requirements.
type initQItem struct {
	// f is the init function pointer/reference.
	f QFunc

	// name is the "name" of the requirement. It may be used for Fatal() error
	// messaging or 'dependent semaphore' checks. The name is case sensitive.
	name string

	// state is the current state of the initialization. It may have never run,
	// have skipped (TryAgain), or have completed (Satisfied).
	state ReqResult

	// deps are optional dependent tasks (matching name) that must be Satisfied
	// before this item can attempt to run. These are used when there is no other
	// indication of success of dependent tasks.
	deps []string
}

/* ======================================================================== */

// newInitQItem is the preferred constructor for new Q items.
func newInitQItem(name string, f QFunc, deps ...string) (rqi *initQItem) {

	rqi = new(initQItem)

	rqi.f = f
	rqi.name = name
	rqi.state = UnRun
	for _, d := range deps {
		rqi.deps = append(rqi.deps, d)
	}

	return
}

/* ======================================================================== */

// run will run the required task function if it should be run. Once a task
// function returns Satisfied, then it will not be run again.
func (rqi *initQItem) run() ReqResult {

	if rqi == nil {
		log.Fatal("nil item in the InitQ")
	}

	// Only run if one should.
	if rqi.state == TryAgain || rqi.state == UnRun {
		rqi.state = rqi.f()
	}

	return rqi.state
}
