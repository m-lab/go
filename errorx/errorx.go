package errorx

import "errors"

// Suppress filters out errors of the passed-in type. It allows you to easily
// convert code that uses the "standard error" return-type pattern into code
// that returns null under normal circumstances. If the incoming error is or is
// derived from any of the errors to suppress, then that error is filtered out.
func Suppress(incomingError error, errorsToSuppress ...error) error {
	for _, err := range errorsToSuppress {
		if errors.Is(incomingError, err) {
			return nil
		}
	}
	return incomingError
}
