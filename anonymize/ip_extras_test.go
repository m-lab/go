package anonymize

// SetFlag sets the value of the ipAnonymization method flag. Because it is
// defined in a file named _test.go, it is only available during testing thanks to
// Go's filename compilation rules. This allows the other _test.go files to
// blackbox test the package.
func SetFlag(newFlagValue string) {
	*ipAnonymization = newFlagValue
}
