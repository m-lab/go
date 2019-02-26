Functions that maybe should be part of the standard go runtime.

Contains a function `Must` that calls `log.Fatal` with a nice error message if
its first argument is not a nil error.  Using `Must` aids in the pursuit of 100%
code coverage, because it means that the error pathway of `log.Fatal` is in this
package instead of inline with the code being tested.
