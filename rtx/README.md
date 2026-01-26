# rtx

Functions that maybe should be part of the standard go runtime.

Contains a function `Must` that calls `log.Fatal` with a nice error message if
its first argument is not a nil error.  Using `Must` aids in the pursuit of 100%
code coverage, because it means that the error pathway of `log.Fatal` is in this
package instead of inline with the code being tested.

`PanicOnError` is like `Must`, but it panics instead of logging and exiting.

It also contains `ValueOrDie`, which is like `Must`, but for functions that return
a value and `ValueOrPanic` which panics instead of dying.
