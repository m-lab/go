Functions that maybe should be part of the standard go runtime.

Contains a function `Must` that calls `log.Fatal` with a nice error message if
its first argument is not a nil error.  Using `Must` aids in the pursuit of 100%
code coverage, because it means that the error pathway of `log.Fatal` is in this
package instead of inline with the code being tested.

It also contains a function `ErrorLoggingCloser` that wraps any implementation
of `io.Closer` into an `errorLoggingCloser`, which makes sure errors happening
when calling Close() are logged. It can be used in a `defer` statement, e.g.:
```
defer ErrorLoggingCloser(resource).Close()
```
