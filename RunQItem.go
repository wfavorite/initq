package initq

import "log"

/* ------------------------------------------------------------------------ */

type initQItem struct {
	f     QFunc
	name  string
	state ReqResult
	deps  []string
}

/* ======================================================================== */

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

func (rqi *initQItem) run() ReqResult {

	if rqi == nil {
		log.Fatal("nil item in the InitQ")
	}

	if rqi.state == TryAgain || rqi.state == UnRun {

		rqi.state = rqi.f()

	}

	return rqi.state
}
