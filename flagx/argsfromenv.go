// Package flagx extends to capabilities of flags to also be able to read
// from environment variables.  This comes in handy when dockerizing
// applications.
package flagx

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// ArgsFromEnv will expand command-line argument parsing to include setting the
// values of flags from their corresponding environment variables. The
// environment variable for an argument is the upper-case version of the
// command-line flag.
func ArgsFromEnv(flagSet *flag.FlagSet) error {
	// Allow environment variables to be used for unspecified commandline flags.
	// Track what flags were explicitly set so that we won't override those flags.
	specifiedFlags := make(map[string]struct{})
	flagSet.Visit(func(f *flag.Flag) { specifiedFlags[f.Name] = struct{}{} })

	// All flags that were not explicitly set but do have a corresponding evironment variable should be set to that env value.
	// Visit every flag and don't override explicitly set commandline args.
	var err error
	flagSet.VisitAll(func(f *flag.Flag) {
		envVarName := strings.ToUpper(f.Name)
		if val, ok := os.LookupEnv(envVarName); ok {
			if _, specified := specifiedFlags[f.Name]; specified {
				log.Printf("WARNING: Not overriding flag -%s=%q with evironment variable %s=%q\n", f.Name, f.Value, envVarName, val)
			} else {
				if setErr := f.Value.Set(val); setErr != nil {
					err = fmt.Errorf("Could not set argument %s to the value of environment variable %s=%q (err: %s)", f.Name, envVarName, val, setErr)
				}
			}
		}
		log.Printf("Argument %s=%v\n", f.Name, f.Value)
	})
	return err
}
