package uniformnames_test

import (
	"testing"

	"github.com/m-lab/go/uniformnames"
)

func TestNames(t *testing.T) {
	for _, goodname := range []string{
		"good",
		"good123",
		"good123abc",
		"finealso",
	} {
		if uniformnames.Check(goodname) != nil {
			t.Errorf("%q was a good name but did not pass the check", goodname)
		}
	}
	for _, badname := range []string{
		"1bad",
		"bad-123",
		"bad123ABC",
		"AlsoBad",
		"not_okay",
		"!nope",
		"",
		":(",
	} {
		if uniformnames.Check(badname) == nil {
			t.Errorf("%q was a bad name but passed the check", badname)
		}
	}
}
