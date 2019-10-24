package anonymize

// SetLogFatalf allows us to inject a new log.Fatal to test error cases. It
// returns a function which, when executed, reverts the injection process.
func SetLogFatalf(f func(string, ...interface{})) func() {
	old := logFatalf
	logFatalf = f
	return func() {
		logFatalf = old
	}
}
