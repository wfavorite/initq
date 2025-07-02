# initq

## Overview

Starting up complex applications with lots of moving parts can be *messy*. It starts with "I need to do this", then "Well that is dependent on this" and "This thing should be done first"...

For simple startup dependencies, this is not a big deal, but as an application grows (i found that) the startup code tends to get messy - even when things are broken into functions/methods.

``initq`` is a very simple "Initialization Queue" of tasks that has a means to *process* dependencies.

Here is a sample use of this module:

```go
	cd := new(CoreData) // Core data structure for the application.
	iq := initq.NewInitQ()

	// Add requirements to the InitQ. They do not need to be specifically
	// ordered - as each should be aware of, and respond to (missing)
	// dependencies.
	//
	// Each Add() contains:
	// - A friendly name that may be used in failure messages or 'semaphore'
	//   dependencies.
	// - A function/method pointer to a func that can initialize that part, and
	//   is aware of dependencies.

	iq.Add("cmdline", cd.ParseCommandLine)   // Knows config file location
	iq.Add("config", cd.ReadConfigFile)      // Reads config file
	iq.Add("listener", cd.ConfigureListener) // Uses config to setup listener

	if err := iq.Process(); err != nil {
		// It is possible to check for the specific error, but in this case
		// we know that the only way to exit early is if there was a command-
		// line error, the config file was missing/broken, or some sort of
		// initialization error. In every case, we brought our own error.
		fmt.Fprintln(os.Stderr, "ERROR:", cd.Error)
		os.Exit(1)
	}

	// Having reached here - all components have been properly initialized.
```

## Setup functions / methods

Setup functions are responsible for:

- Being able to determine if dependent components have completed. For example; The ``ReadConfigFile()`` method should know if the ``ParseCommandLine()`` completed.
- If a dependent component has yet to initialize, return ``initq.TryAgain``.
- If a *bad thing* happened (like a command line typo, or a missing/curropted config file), set an error message and return ``initq.Stop``. (initq does not handle error messages. The expectation is that the setup methods would set that in the core structure they are called on. See notes on "Internal Errors".)
- If the component was properly setup, then return ``initq.Satisfied``. This signifies that the requirement need not be tried again. (If the 'semaphore' dependency method is used, then this signfiies completion of the requriement.)

## Internal Errors

There are three kinds of errors in this process:

1. initq has a bug. This has test coverage, so hopefully it is not a thing. But internal errors are ``log.Fatal()`` assertions.
2. The defined Q has a 'bug'. This is when the user creates circular dependencies or methods that never complete. These also cause ``log.Fatal()`` assertions.
3. Some part of the initialization failed - such as the user specified a wrong command-line option. This is by-far the most typical case of failure.

The *design intent* of the ``log.Fatal()`` assertions is that these things should be caught in test. They are *usage* (or perhaps internal) errors that should be uncovered immediately and not present as edge cases in production.

The typical error case is an *application thing* and should be handled by the application code/logic.

## Semaphore mode

Dependent requirement labels may be added to the ``Add()`` method. These are used to detect task completion when there is no other evidence of such.

> __NOTE:__
>> The desired pattern is to *sense* if a dependency has completed. For example: When the config has been parsed, a reference to a struct containing config options may be an indicator that it was parsed. If the reference is ``nil``, then the config was never parsed. (In this case) The *pointer* is the semaphore.

Here is an example of a function that has no clear evidence of success (other than it saying so):

```go
	iq.Add("settime", SyncTimeClock)
	iq.Add("scheduler", cd.StartScheduler, "settime")
```

In the above example:

- ``SyncTimeClock()`` either works or it does not. It does not modify the environment such that it can be known to have successfully completed. If it is successful it return ``initq.Success``, if not, then it returns a ``initq.Stop`` that will stop processing of the Q.
- ``SyncTimeClock()`` would *ideally* be a method (so it can easily set an error message), but was specified as a function because it suggests it has 'undetectable' state. A method is typically used for this reason, but a function that never fails would be a possible use case.
- ``StartScheduler()`` sets a "semaphore requirement" on the "settime" task. This means that the ``StartScheduler()`` method will not be called until ``SyncTimeClock()`` has returned ``initq.Success``.
