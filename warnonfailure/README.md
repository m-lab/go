Package containing custom versions of common functions/methods that always
log a warning in case of error.

`Close` wraps a `resource.Close()` call and logs the error, if any, adding a
custom message.

This function is intended to be used as a better alternative to completely
ignoring an error in cases where there isn't any obvious reason to handle it
explicitly. Also, adding a custom message around the error allows to easily
find any instances of this in our log files.

It can be used in a `defer` statement, e.g.:
```
defer Close(resource, "Warning: ignoring error")
```
