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

func illegalToUnderscore(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return r
	case r >= '0' && r <= '9':
		return r
	default:
		return '_'
	}
}

// MakeShellVariableName makes every passed-in string match the regexp
// [A-Z_][A-Z0-9_]* Characters in shell variable names should be part of the
// portable character set (defined to be [A-Z0-9_] in
// https://pubs.opengroup.org/onlinepubs/000095399/basedefs/xbd_chap08.html) and
// may not begin with a digit.
func MakeShellVariableName(flagName string) string {
	newName := strings.Map(illegalToUnderscore, strings.ToUpper(flagName))
	if len(newName) > 0 && newName[0] >= '0' && newName[0] <= '9' {
		newName = "_" + newName
	}
	return newName
}

// AssignedFlags returns a map[string]struct{} where every key in the map is the
// name of a flag in the passed-in FlagSet that was explicitly set on the
// command-line.
func AssignedFlags(flagSet *flag.FlagSet) map[string]struct{} {
	specifiedFlags := make(map[string]struct{})
	flagSet.Visit(func(f *flag.Flag) { specifiedFlags[f.Name] = struct{}{} })
	return specifiedFlags
}

// ArgsFromEnv will expand command-line argument parsing to include setting the
// values of flags from their corresponding environment variables. The
// environment variable for an argument is the upper-case version of the
// command-line flag.
//
// If you have two flags that map to the same environment variable string (like
// "-a" and "-A"), then the behavior of this function is unspecified and should
// not be relied upon. Also, your flags should have more descriptive names.
func ArgsFromEnv(flagSet *flag.FlagSet) error {
	return ArgsFromEnvWithLog(flagSet, true)
}

// ArgsFromEnvWithLog operates as ArgsFromEnv with an additional option to
// disable logging of all flag values. This is helpful for command line
// applications that wish to disable extra argument logging.
func ArgsFromEnvWithLog(flagSet *flag.FlagSet, logArgs bool) error {
	// Allow environment variables to be used for unspecified commandline flags.
	// Track what flags were explicitly set so that we won't override those flags.
	specifiedFlags := AssignedFlags(flagSet)

	// All flags that were not explicitly set but do have a corresponding evironment variable should be set to that env value.
	// Visit every flag and don't override explicitly set commandline args.
	var err error
	flagSet.VisitAll(func(f *flag.Flag) {
		envVarName := MakeShellVariableName(f.Name)
		if val, ok := os.LookupEnv(envVarName); ok {
			if _, specified := specifiedFlags[f.Name]; specified {
				log.Printf("WARNING: Not overriding flag -%s=%q with evironment variable %s=%q\n", f.Name, f.Value, envVarName, val)
			} else {
				if setErr := f.Value.Set(val); setErr != nil {
					err = fmt.Errorf("Could not set argument %s to the value of environment variable %s=%q (err: %s)", f.Name, envVarName, val, setErr)
				}
			}
		}
		if logArgs {
			log.Printf("Argument %s=%v\n", f.Name, f.Value)
		}
	})
	return err
}
