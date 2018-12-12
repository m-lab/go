ListenAndServe in the Go http package is impossible to use safely in an
asynchronous way.  What we frequently want is to start the http server and have
it run, asynchronously, in perpetuity, but not have the call return until the
server is able to receive requests.

httpx provides `ListenAndServeAsync`, which starts an http server in a
background goroutine, but does not return until the listening socket is
established. Therefore, once this function returns, the server can have an HTTP
GET run against it and the GET should succeed (or at least call the appropriate
server code).
