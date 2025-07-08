package initq

import (
	"strings"
	"testing"
)

/* ======================================================================== */
/*
	This is the CoreData structure that sits at the heart of the 'application'.
	The initialization methods are called on it, to initialize data in it,
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

	Conf bool // Config read from disk (after command line specifies where)
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

// StartScheduler will never be Satisfied due to an omission in the SetupServer
// method.
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

func TestInitQ(t *testing.T) {

	var rq *InitQ
	var cd *coredata

	// ----------
	// Simple Q, easily satisfied.

	rq = NewInitQ()

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

	rq = NewInitQ()

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

	rq = NewInitQ()

	rq.Add("black", func() ReqResult { return Satisfied }, "white")
	rq.Add("white", func() ReqResult { return Satisfied }, "black")

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	}

	// ----------
	// A more realistic case - that should succeed.
	// This does *not* include the "scheduler" (that will fail).
	// This is *mostly* in correct order.

	rq = NewInitQ()
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

	rq = NewInitQ()
	cd = new(coredata)

	rq.Add("server", cd.SetupServer)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)

	if err := rq.Process(); err != nil {
		t.Errorf("Q did not finish - %s", err.Error())
	}

	// ----------
	// A standard-ish case that should pass.
	// ...but with TryProcess()
	// BehaveUnresolvIsErr is true

	rq = NewInitQ()
	cd = new(coredata)

	rq.Add("server", cd.SetupServer)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)

	if err := rq.TryProcess(); err != nil {
		t.Errorf("Q did not finish - %s", err.Error())
	}

	// ----------
	// Another realistic case - that cannot succeed.
	// This include the "scheduler" (that will fail).

	rq = NewInitQ()
	cd = new(coredata)

	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("server", cd.SetupServer)
	rq.Add("scheduler", cd.StartScheduler)

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	}

	// ----------
	// A dependency was typo'd.

	rq = NewInitQ()

	rq.Add("black", func() ReqResult { return Satisfied }, "white")
	rq.Add("white", func() ReqResult { return Satisfied }, "blue")

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	}

	// ----------
	// An Add call self-references in the dependencies.

	rq = NewInitQ()

	rq.Add("selfref", func() ReqResult { return Satisfied }, "selfref")
	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	} else {
		if !strings.Contains(err.Error(), "Add(selfref)") {
			t.Errorf("Expected a specific error - got %s", err.Error())
		}
	}

	// ----------
	// An Add call with a blank label.

	rq = NewInitQ()

	rq.Add("", func() ReqResult { return Satisfied })
	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	} else {
		if !strings.Contains(err.Error(), "Add") {
			t.Errorf("Expected a specific error - got %s", err.Error())
		}
	}

	// ----------
	// An Add call with a nil function reference.

	rq = NewInitQ()

	rq.Add("nilfunc", nil)
	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	} else {
		if !strings.Contains(err.Error(), "Add(nilfunc)") {
			t.Errorf("Expected a specific error - got %s", err.Error())
		}
	}

	// ----------
	// An Add call with invalid dependency name.

	rq = NewInitQ()

	rq.Add("cmdline", func() ReqResult { return Satisfied })
	rq.Add("config", func() ReqResult { return Satisfied }, "CmdLine")

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	} else {
		if !strings.Contains(err.Error(), "CmdLine") {
			t.Errorf("Expected a specific error - got %s", err.Error())
		}
	}

	// ----------
	// An Add call with redundant/typo name label.

	rq = NewInitQ()

	rq.Add("cmdline", func() ReqResult { return Satisfied })
	rq.Add("cmdline", func() ReqResult { return Satisfied })

	if err := rq.Process(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	} else {
		if !strings.Contains(err.Error(), "cmdline") {
			t.Errorf("Expected a specific error - got %s", err.Error())
		}
	}

	// Explicitly clear the Behaviour on completion of 'internal'
	// error detection.
	BehaveUnresolvIsErr = false

	// ----------
	// One item stops early

	rq = NewInitQ()

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

	// ----------
	// A standard-ish case that should pass.
	// ...but with TryProcess()
	// BehaveUnresolvIsErr is false

	rq = NewInitQ()
	cd = new(coredata)

	rq.Add("server", cd.SetupServer)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)

	if err := rq.TryProcess(); err != nil {
		t.Errorf("Q did not finish - %s", err.Error())
	}

	// ----------
	// TryProcess - the fail case
	// This include the "scheduler" (that will fail).

	rq = NewInitQ()
	cd = new(coredata)

	rq.Add("config", cd.ReadConfigFile)
	rq.Add("cmdline", cd.ParseCommandLIne)
	rq.Add("dbconn", cd.SetupDBConnection)
	rq.Add("server", cd.SetupServer)
	rq.Add("scheduler", cd.StartScheduler)

	if err := rq.TryProcess(); err == nil {
		t.Errorf("An unresolvable Q managed to finish.")
	} else {
		if uqe, ok := err.(*QUnresolvable); ok {

			tasks := uqe.UnresolvedTasks()

			if len(tasks) != 1 {
				t.Errorf("Unexpected number of remaining tasks. Expected 1; found %d.", len(tasks))
			} else {
				if tasks[0] != "scheduler" {
					t.Errorf("Unexpected unresolved task. Expected scheduler, got %s", tasks[0])
				}
			}
		} else {
			t.Errorf("Failed to match against *QUnresolvable type. Got %T", err)
		}
	}

}
