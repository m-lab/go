package anonymize

// SetFlag sets the value of the ipAnonymization method flag. Because it is
// defined in a file named _test.go, it is only available during testing thanks to
// Go's filename compilation rules. This allows the other _test.go files to
// blackbox test the package.
func SetFlag(newFlagValue string) {
	*ipAnonymization = newFlagValue
}

// SetLogFatalf allows us to inject a new log.Fatal to test error cases. It
// returns a function which, when executed, reverts the injection process.
func SetLogFatalf(f func(string, ...interface{})) func() {
	old := logFatalf
	logFatalf = f
	return func() {
		logFatalf = old
	}
}
