package pretty

import (
	"strings"
	"testing"
)

func TestPretty(t *testing.T) {
	// Try to format a struct with supported types.
	s := struct {
		I       int
		S       string
		M       map[string]string
		ignored func() error
	}{
		I: 100,
		S: "test",
		M: map[string]string{
			"a": "b",
		},
		// Private fields are ignored.
		ignored: func() error { return nil },
	}
	expected := `{
  "I": 100,
  "S": "test",
  "M": {
    "a": "b"
  }
}`
	Print(s)
	v := Sprint(s)
	if v != expected {
		t.Errorf("Sprint() generated unexpected output; got %q, want %q", v, expected)
	}

	// Try to format a struct with an unsupported type.
	s2 := struct {
		F func() error
		S string
	}{
		// Public fields with incompatible formats generate errors.
		F: func() error { return nil },
		S: "test2",
	}
	// Both are expected to fail due to the function pointer.
	Print(s2)
	v = Sprint(s2)
	if !strings.Contains(v, "unsupported type") {
		t.Errorf("Sprint() did not return unsupported type error: got %q", v)
	}
}
