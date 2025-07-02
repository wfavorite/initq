package initq

/*
	Version History:

	0.0.0   25-7-1 - Initial (re)creation of "startq".
	               - No "pig" in this implementation. The Q marker is not part
				     of a struct, not a function (pig) return value.
					 (https://en.wikipedia.org/wiki/Pigging)
				   - Semaphores are intentionally omitted (as explicit items).
				     Non-detectable dependences can be specified with a
					 reference to the (silent) dependency in the Add() method.
                   - Basic documentation.
	0.1.0   25-7-2 - Module rename.
	               - Fixed typo in behaviours.go filename. Yea, i *now* think
				     it 'correct'.
*/

// VersionString is the version of the project.
const VersionString = "0.1.0"

/*
	ToDos:
    [ ] Improve documentation.

	Done:
	[Q] Author. (Don't do this if you intend to publish.)
	[X] Consider a rename. "runq" suggests other things. While avoiding startq,
	    as a name... that leaves... "initq"? "upq"?
	[X] Determine a way to check the 'builtin semaphore'.
	[X] Implement the RunQItem name in some sort of (possibly optional)
	    diagnostic output that can be used to debug Qs that cannot possibly be
		satisfied.

*/
