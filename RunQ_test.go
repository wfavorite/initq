package initq

import "testing"

/* ======================================================================== */
/*
	This is the CoreData structure that sits at the heart of the application.
	The initialiazation methods are called on it, to initialize data in it,
	and processes that use it.

	Things that tend to exist here:

	- Read config file
	- Start an internal scheduler
	- Set up an event Q
	- Connect to a DB
	- Define/setup/start a service
*/

type coredata struct {
	// These are 'normally' pointers to objects that are initialized.
	//
	// Here i just use booleans for test.

	Conf bool // Config read from disk (after command line specifys where)
	Cmdl bool // Cmdline is parsed first (a requirement to find the config)
	Data bool // Data is the database connector (requires config)
	Serv bool // Service is the service setup which requires
	Schd bool // Scheduler is a thing that (when introduced) will not complete
}

func (cd *coredata) ReadConfigFile() ReqResult {
	if !cd.Cmdl {
		return TryAgain
	}

	// "Read the config file"
	cd.Conf = true

	// Cmdline has been satisfied. Use it to load the (correct) config.
	return Satisfied
}

func (cd *coredata) ParseCommandLIne() ReqResult {

	// No dependencies, Consider this satisfied on first attempt.

	// "Parse the command line"
	cd.Cmdl = true

	return Satisfied
}

func (cd *coredata) SetupDBConnection() ReqResult {

	// The DB cannot be connected until we know the connect strings from the
	// config file.
	if !cd.Conf {
		return TryAgain
	}

	// "Setup the DB connection"
	cd.Data = true

	// No dependencies, Consider this satisfied on first attempt.
	return Satisfied
}

func (cd *coredata) SetupServer() ReqResult {

	// The server can be setup/started once the DB connection is done.
	if !cd.Data {
		return TryAgain
	}

	// "Setup (or perhaps start) the server"
	// TYPO ERROR: (on purpose) Someone forgot to set state!
	//cd.Serv = true <----- Intentionally skipped!

	// Server is setup/running
	return Satisfied
}

func (cd *coredata) StartScheduler() ReqResult {

	// TYPO ERROR: (on purpose) See SetupServer note.
	if !cd.Serv {
		return TryAgain
	}

	// "Start the scheduler"
	cd.Schd = true

	// Scheduler setup/running.
	return Satisfied
}

/* ======================================================================== */

func TestRunQ(t *testing.T) {

	var rq *RunQ
	var cd *coredata

	// ----------
	// Simple Q, easily satisfied.

	rq = NewRunQ()

	rq.Add("one", func() ReqResult { return Satisfied })
	rq.Add("two", func() ReqResult { return Satisfied })
	rq.Add("threee", func() ReqResult { return Satisfied })

	if err := rq.Process(); err != nil {
		t.Errorf("Q did not finish - %s", err.Error())
	}

	// NOTE: Taking advantage of the Behaviour here to facilitate
	// error reporting in test.
	BehaveUnresolvIsErr = true

	// ----------
	// A Q that cannot be satisfied (unsat never gets Satisfied).

	rq = NewRunQ()

	rq.Add("good1", func() ReqResult { return Satisfied })
	rq.Add("unsat", func() ReqResult { return TryAgain })
	rq.Add("good2", func() ReqResult { return Satisfied })

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	}

	// ----------
	// A circular condition where each requires the other.
	// Both are capable of returning Satisfied, but neither is capable of
	// running because of the internal 'semaphore' requirements prevent
	// them from running.

	rq = NewRunQ()

	rq.Add("black", func() ReqResult { return Satisfied }, "white")
	rq.Add("white", func() ReqResult { return Satisfied }, "black")

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	}

	// ----------
	// A more realistic case - that should succeed.
	// This does *not* include the "scheduler" (that will fail).
	// This is *mostly* in correct order.

	rq = NewRunQ()
	cd = new(coredata)

	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("server", cd.SetupServer)

	if err := rq.Process(); err != nil {
		t.Errorf("Q did not finish - %s", err.Error())
	}

	// ----------
	// A more realistic case - that should succeed.
	// This does *not* include the "scheduler" (that will fail).
	// This in backwards (worst case scenario) order.

	rq = NewRunQ()
	cd = new(coredata)

	rq.Add("server", cd.SetupServer)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)

	if err := rq.Process(); err != nil {
		t.Errorf("Q did not finish - %s", err.Error())
	}

	// ----------
	// Another realistic case - that cannot succeed.
	// This include thes "scheduler" (that will fail).

	rq = NewRunQ()
	cd = new(coredata)

	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("server", cd.SetupServer)
	rq.Add("scheduler", cd.StartScheduler)

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	}

	// Explicitly clear the Behaviour on exit.
	BehaveUnresolvIsErr = false

	// ----------
	// One item stops early

	rq = NewRunQ()

	rq.Add("one", func() ReqResult { return Satisfied })
	rq.Add("two", func() ReqResult { return Satisfied })
	rq.Add("stopper", func() ReqResult { return Stop })
	rq.Add("three", func() ReqResult { return Satisfied })
	rq.Add("four", func() ReqResult { return Satisfied })

	if err := rq.Process(); err == nil {
		t.Errorf("An early-term Q managed to finish.")
	} else {
		if err != ErrQStopped {
			t.Errorf("Expected the Q to be err/stopped")
		}
	}

}
