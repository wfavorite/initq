# initq

## Overview

Starting up complex applications with lots of moving parts can be *messy*. It starts with "I need to do this", then "Well that is dependent on this" and "This thing should be done first"...

For simple startup dependencies, this is not a big deal, but as an application grows (i found that) the startup code tends to get messy - even when things are broken into functions/methods.

``initq`` is a very simple "Initialization Queue" of tasks that has a means to *process* dependencies.

The setup and design of an InitQ resembles that of the [flag](https://pkg.go.dev/flag) package. An ``InitQ`` is likened to a ``flag.FlagSet``. Both are used to avoid lots of unnecessary boiler-plate code, and make a noisy process read easier.

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
- If a *bad thing* happened (like a command line typo, or a missing/corrupted config file), set an error message and return ``initq.Stop``. (initq does not handle error messages. The expectation is that the setup methods would set that in the core structure they are called on. See notes on "Internal Errors".)
- If the component was properly setup, then return ``initq.Satisfied``. This signifies that the requirement need not be tried again. (If the 'semaphore' dependency method is used, then this signifies completion of the requirement.)

## Internal Errors

There are three kinds of errors in this process:

1. initq has a bug. This has test coverage, so hopefully it is not a thing. But internal errors are ``log.Fatal()`` assertions.
2. The defined Q has a 'bug'. This is when the caller creates circular dependencies or methods that never complete. These also cause ``log.Fatal()`` assertions.
3. Some part of the initialization failed - such as the user specified a wrong command-line option. This is by-far the most typical case of failure.

The *design intent* of the ``log.Fatal()`` assertions is that these things should be caught in test. They are *calling* (or perhaps internal) errors that should be uncovered immediately and not present as edge cases in production.

The typical error case is an *application thing* and should be handled by the application code/logic.

The ``TryProcess()`` method is used to handle what is typically seen as an internal error - that *might* intermittently happen as a "user error". (Meaning: Don't log.Fatal() to the user.) This method may return a ``QUnsatisfied`` type that can be queried for the remaining / unsatisfied tasks in the Q.

> __NOTE:__
>> Satisfaction of a required task should __not__ hinge on anything a user passed, or steps that stem from task failures. Each task should handle failures with a ``Stop`` return and specific error message. Allowing for ``TryProcess()`` means that the caller *could* force a condition where the Q is not satisfied at 'runtime'. ``TryProcess()`` means that (at least) a caller error of this type could leak into a production / untested release - but be handled like a normal error to the user.

## Semaphore mode

Dependent requirement labels may be added as optional parameters to the ``Add()`` method. These are used to detect task completion when there is no other evidence of such.

> __NOTE:__
>> The desired pattern is to *sense* if a dependency has completed. For example: When the config has been parsed, a reference to a struct containing config options may be an indicator that it was parsed. If the reference is ``nil``, then the config was never parsed. (In this case) The *pointer* is the semaphore.

Here is an example of a function that has no clear evidence of success (other than it saying so):

```go
	iq.Add("settime", SyncTimeClock)
	iq.Add("scheduler", cd.StartScheduler, "settime")
```

In the above example:

- ``SyncTimeClock()`` either works or it does not. It does not modify the environment such that it can be known to have successfully completed. If it is successful it returns ``initq.Satisfied``, if not, then it returns a ``initq.Stop`` that will stop processing of the Q.
- ``SyncTimeClock()`` would *ideally* be a method (so it can easily set an error message), but was specified as a function because it suggests it has 'undetectable' state. A method is typically used for this reason, but a function that never fails would be a possible use case.
- ``StartScheduler()`` sets a "semaphore requirement" on the "settime" task. This means that the ``StartScheduler()`` method will not be called until ``SyncTimeClock()`` has returned ``initq.Satisfied``.
- All task and dependent labels are case-sensitive and must match exactly. I have used raw strings in these examples where ``const`` labels may be a more appropriate means of avoiding mis-matches on dependencies to tasks.

## Design notes

This was originally written (within my company) as "startq". That code belongs to my previous employer - so i wrote a entirely new and better solution. I encourage all users of the previous to consider the newer, better module here.

The original version used a [pig](https://en.wikipedia.org/wiki/Pigging) function that was used to determine how many times the Q (implemented as a ``chan``) was iterated. This was totally unnecessary and a complication on a really simple thing. Now more functionality (to include the 'semaphore' option) has been placed into the Q processing loops.

The original version had a *bolt on* concept of a semaphore. Usage throughout dozens of projects determined that the entire semaphore concept was not the *preferred method* of handling dependencies. In practical experience, the goal was to *not* use explicit semaphores, but use indications of success rather than flags. And yes, indications of success very well can be semaphore-ish. An initialized pointer/structure is indication that that dependent requirement was successfully completed - it acts as a semaphore. I did choose to include a much cleaner and better integrated semaphore approach - mainly for appeal to those who 'need' it for some reason.

This Q method really is a simple thing. It doesn't work on magic, it is not rocket science, it just makes tons of application initialization code more approachable. The idea is that instead of scattering tons of dependency checks and requirements in a (series of) function(s), the dependencies are encapsulated in each of the initialization calls. There are lots of ways - i am sure - to work this problem, but this approach works really well in practice.

As to why *i* use it:

1. I found that it was (slightly) challenging to keep lots of do-this-then-that initialization code clean. Inserting new requirements became an ordering exercise - even when the requirement had no dependencies. Then there is the boiler-plate error / dependency handling (that causes some localized code sprawl). A few months later one looks back on their code and asks: "Why did i choose to initialize this first? Do i need to put things before or after it?". If the dependencies are coded in the init methods, then it tends to be more localized and clear.
2. Initialization code is mostly *static* and *boring*. Perhaps hyperbole, but in my experience mostly true. When you *do* visit it, i find that inserting new initialization requirements into the code is much simpler when it is highly localized to a specific function or method.
3. The block of ``Add()`` methods becomes a convenient jump-off location to go to the methods. If this sits in ``main()`` then it becomes an 'index' of requirements rather than lots of boiler-plate code that must be parsed to find a thing. If i am curious about command line options that might be of issue, i just right click on the function reference in the ``Add()`` call and jump to the specific requirement.
4. When adding new startup requirements one needs to parse a stack of initialization functions to determine *where* the new initialization task may be inserted. Using initq is is only necessary to describe (either in code or the semaphore dependency method) what must be completed first. The Process function will know if a invalid or circular dependency was added.
