# initq

## Overview

Starting up complex applications with lots of moving parts can be *messy*. It starts with "I need to do this", then "Well that is dependent on this" and "This thing should be done first"...

For simple startup dependencies, this is not a big deal, but as an application grows (i found that) the startup code tends to get messy - even when things are broken into functions/methods.

``initq`` is a very simple "Initialization Queue" of tasks that has a means to *process* dependencies.

Here is a sample use of this module:

```go
	cd := new(CoreData) // Core data structure for the application.
	iq := NewInitQ()

    // Add requirements to the 

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
```
