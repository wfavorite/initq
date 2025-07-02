package initq

/*
	'Behaviours' allow for modifying 'normal' module flows. These are basically
	const(s) that can be modified - for testing or... whatnot.

	The default/uninitialized values (false) are the default behaviours.
	Setting a behaviour to true is a *change* from the default.
*/

/* ------------------------------------------------------------------------ */

// BehaveUnresolvIsErr will cause Process to return (only) an Err instead of
// a log.Fatal(). Ideally Fatal() is preferred as it is clearly differentiated
// from natural application up issues and is designed to expose the error
// immediately - as opposed to appearing to be a 'runtime' issue.
//
// This behaviour is used in the test code.
var BehaveUnresolvIsErr bool
