package initq

import "fmt"

/* ------------------------------------------------------------------------ */

// ErrQUnsolvable is returned when the run Q cannot be satisfied in a number
// of iterations equal to the count of the Q.
//
// Assuming a worst case scenario, where the Q was defined perfectly in the
// wrong order, it should be solvable in a number of iterations equal to the
// items in the Q. This would be by satifying the last (on the first run), then
// second to last (on the next), and so on until the top of the Q is complete.
//
// This error is not enabled / returned unless the BehaveUnresolvIsError
// behaviour is set to true.
var ErrQUnsolvable = fmt.Errorf("unsolvable run Q")

/* ------------------------------------------------------------------------ */

// ErrQStopped is returned by Process() when a requirement function returns the
// Stop RunQResult. This is the one condition that the Process() method errors
// on - so it can be checked for, but is not a hard requirement to do so.
var ErrQStopped = fmt.Errorf("run Q early termination")
