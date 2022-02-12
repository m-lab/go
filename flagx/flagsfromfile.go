package flagx

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
)

// FlagsFromFile parses flags read from a named file. Flags in the file are
// effectively inserted in place into the original command line. So, this flag
// uses the same semantics as the flag package. Namely, Set is called once, in
// order, for each flag present. CommandLine must be defined before Set is called.
type FlagsFromFile struct {
	CommandLine *flag.FlagSet
	name        string
}

// Get returns the named argument file passed to Set. Before Set is called, the
// return value is undefined.
func (ff *FlagsFromFile) Get() string {
	return ff.name
}

// Set reads flags from the named file and parses them using flags taken from
// the registered CommandLine flagset.
func (ff *FlagsFromFile) Set(name string) error {
	// Record name of file argument.
	ff.name = name

	// Create a temporary flagset and copy default flags. This is necessary so
	// Parse does not reset the original command line arguments.
	nfs := flag.NewFlagSet("", flag.ContinueOnError)
	nfs.SetOutput(ioutil.Discard) // silence errors.

	// Copy original flags to new flagset. Because the Value is always a
	// reference type, the Parse will change the original flags.
	ff.CommandLine.VisitAll(func(f *flag.Flag) {
		nfs.Var(f.Value, f.Name, f.Usage)
	})

	b, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	// Collapse newlines lines and parse flags.
	file := strings.TrimSpace(string(b))
	line := strings.ReplaceAll(file, "\n", " ")
	err = nfs.Parse(strings.Fields(line))
	if err != nil {
		return fmt.Errorf("while parsing %q: %w", name, err)
	}
	return nil
}

// String returns the named argument file passed to Set. Before Set is called,
// the return value is undefined.
func (ff *FlagsFromFile) String() string {
	return ff.name
}
