Functions that extend the io package.

`ErrorLoggingCloser` wraps any implementation of `io.Closer` into an
`errorLoggingCloser`, which makes sure errors happening when calling
Close() are logged. It can be used in a `defer` statement, e.g.:
```
defer ErrorLoggingCloser(resource).Close()
```
