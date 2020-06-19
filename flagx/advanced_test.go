package flagx

import (
	"flag"
	"testing"

	"github.com/go-test/deep"
)

func TestEnableAdvancedFlags(t *testing.T) {
	t.Run("example", func(t *testing.T) {
		// Add advanced flags.
		e := Enum{
			Options: []string{"a", "b"},
			Value:   "b",
		}
		s := "advanced"
		Advanced.Var(&e, "advanced-enum-flag", "advanced-usage")
		Advanced.StringVar(&s, "advanced-string-flag", "", "advanced-usage")

		// Add default flags.
		// Reset the default command line flag to simplilfy checking tests.
		flag.CommandLine = flag.NewFlagSet("default", flag.ContinueOnError)
		s2 := "default"
		flag.StringVar(&s2, "default-flag", "", "default-usage")

		// Add advanced flags to default set.
		EnableAdvancedFlags()

		found := map[string]bool{}
		expected := map[string]bool{
			"advanced-enum-flag":   true,
			"advanced-string-flag": true,
			"default-flag":         true,
		}
		// Verify that they are present in the default set now.
		flag.CommandLine.VisitAll(
			func(f *flag.Flag) {
				found[f.Name] = true
				t.Logf("Found: %q", f.Name)
			},
		)
		if diff := deep.Equal(found, expected); diff != nil {
			t.Errorf("EnableAdvancedFlags() found the wrong advanced flags: %#v", diff)
		}
	})
}
