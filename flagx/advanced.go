package flagx

import "flag"

var (
	// Advanced is a *flag.FlagSet for advanced flags. Packages should add flags
	// to Advanced when those flags should NOT be included in the default
	// flag.CommandLine flag set. Advanced flags may be enabled by calling
	// EnableAdvancedFlags _before_ calling flag.Parse().
	Advanced = flag.NewFlagSet("advanced", flag.ExitOnError)
)

// EnableAdvancedFlags adds all flags registered with the Advanced flag set to
// the default flag.CommandLine flag set. EnableAdvancedFlags should be called
// before flag.Parse().
func EnableAdvancedFlags() {
	Advanced.VisitAll(func(f *flag.Flag) {
		flag.Var(f.Value, f.Name, f.Usage)
	})
}
