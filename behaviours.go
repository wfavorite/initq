package initq

/*
	'Behaviours' allow for modifying 'normal' module flows. These are basically
	const(s) that can be modified - for testing or... whatnot.

	The default/unitialized values (false) are the default behaviours.
*/

/* ------------------------------------------------------------------------ */

// BehaveUnresolvIsErr will cause Process to return (only) an Err
var BehaveUnresolvIsErr bool
