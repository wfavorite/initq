package initq

import "log"

/* ------------------------------------------------------------------------ */

type runQItem struct {
	f     QFunc
	name  string
	state ReqResult
	deps  []string
}

/* ======================================================================== */

func newRunQItem(name string, f QFunc, deps ...string) (rqi *runQItem) {

	rqi = new(runQItem)

	rqi.f = f
	rqi.name = name
	rqi.state = UnRun
	for _, d := range deps {
		rqi.deps = append(rqi.deps, d)
	}

	return
}

/* ======================================================================== */

func (rqi *runQItem) run() ReqResult {

	if rqi == nil {
		log.Fatal("nil item in the RunQ")
	}

	if rqi.state == TryAgain || rqi.state == UnRun {

		rqi.state = rqi.f()

	}

	return rqi.state
}
