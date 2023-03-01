package testingx

import (
	"bytes"
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

func TestMustReadFile_Success(t *testing.T) {
	f := &fakeReporter{}
	got := MustReadFile(f, "./testdata/valid-file.txt")
	if f.called != 0 {
		t.Fatal("MustReadFile() t.Fatal called for valid file")
	}

	want := []byte("foo")
	if !bytes.Equal(got, want) {
		t.Fatalf("MustReadFile() got = %s, want = %s", string(got), string(want))
	}
}

func TestMustReadFile_Error(t *testing.T) {
	f := &fakeReporter{}
	MustReadFile(f, "./testdata/invalid-file.txt")
	if f.called != 1 {
		t.Fatal("MustReadFile() t.Fatal NOT called for invalid file")
	}
}
