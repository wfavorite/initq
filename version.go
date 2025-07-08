package initq

/*
	Version History:

	0.0.0   25-7-1 - Initial (re)creation of "startq".
	               - No "pig" in this implementation. The Q marker is not part
				     of a struct, not a function (pig) return value.
					 (https://en.wikipedia.org/wiki/Pigging)
				   - Semaphores are intentionally omitted (as explicit items).
				     Non-detectable dependencies can be specified with a
					 reference to the (silent) dependency in the Add() method.
                   - Basic documentation.
	0.1.0   25-7-2 - Module rename.
	               - Fixed typo in behaviours.go filename. Yea, i *now* think
				     it 'correct'.
	0.2.0   25-7-2 - Aligned struct, method, and other naming to match the
	                 module name. (Removed RunQ references.)
				   - Added markdown. (Several commits are sure to follow to
				     get the content right.)
	0.3.0   25-7-2 - Markdown cleanup.
	               - Code documentation / comment cleanup.
				   - Spellchecked. (Oh my gosh - my spelling!)
				   - No code changes, no version changes (until module name
				     re-write).
				   - Official module path was changed.
				   - Versioned 0.3.0.
	0.3.1   25-7-8 - Minor documentation cleanup.
*/

// VersionString is the version of the project.
const VersionString = "0.3.1"

/*
	ToDos:
	[ ] Add should sanity check to see if the name matches any of the deps.
	[ ] Consider jumping to v1. I use "GNU-style" versioning that may remain
	    at 0.something for a million years. (0 is a valid version just as it
		is a valid number - IMHO.) On the other hand, pkg.go.dev only flags
		"Stable" when >= v1.
	[ ] Clearer documentation is always appreciated. This is a low-priority ask
	    and is not tied to a specific deliverable. That said, ALL types, funcs,
		etc... should have clear and complete documentation.

	Done:
	[X] Write a 'dummy' app to require/import this module.
	[X] Proper module name.
	[X] Improve documentation.
	[X] Clean up remaining previous module name references. To include file
	    renames.
	[Q] Author. (Don't do this if you intend to publish.)
	[X] Consider a rename. "runq" suggests other things. While avoiding startq,
	    as a name... that leaves... "initq"? "upq"?
	[X] Determine a way to check the 'builtin semaphore'.
	[X] Implement the RunQItem name in some sort of (possibly optional)
	    diagnostic output that can be used to debug Qs that cannot possibly be
		satisfied.

*/
