Functions that maybe should be part of the standard go runtime.

Contains a function `Must` that calls `log.Fatal` with a nice error message if
its first argument is not a nil error.  Using `Must` aids in the pursuit of 100%
code coverage, because it means that the error pathway of `log.Fatal` is in this
package instead of inline with the code being tested.

A similar function `Should` is provided, with the difference that the program
won't exit after the error message is printed, and it's meant to be used for
non-critical errors. A possible usage is to simplify deferred calls to close
resources:

`defer Should(resources.Close(), "Helpful message")`
