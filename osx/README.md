Functions that extend the capability of the os package.

It turns out to be very handy to have the capability to set environment
variables on a temporary basis and subsequently revert them. Also, it turns out
that pretty much no tests check whether the Setenv has actually succeeded.
