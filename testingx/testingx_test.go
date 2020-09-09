package testingx

import (
	"errors"
	"testing"
)

type fakeReporter struct {
	called int
}

func (f *fakeReporter) Helper() {}
func (f *fakeReporter) Fatal(args ...interface{}) {
	f.called++
}

func TestMust(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := &fakeReporter{}
		Must(f, nil, "print nothing")
		if f.called != 0 {
			t.Fatal("t.Fatal called with nil error!")
		}
		err := errors.New("fake error")
		Must(f, err, "print nothing: %s", "custom args")
		if f.called != 1 {
			t.Fatal("t.Fatal NOT called with non-nil error!")
		}
	})
}
