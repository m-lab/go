Generate a globally unique ID for any TCP socket.  When we say globally, we
really mean globally - it should be impossible to have two machines generate the
same UUID.

The only case the uniqueness of the UUID could be violated is if two machines
have the same hostname and booted up at the exact same second in time, but it is
bad practice to give machines the same hostname (so don't).
