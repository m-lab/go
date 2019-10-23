package flagx_test

import (
	"flag"
	"log"
	"testing"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/osx"
)

func TestArgsFromEnvDefaults(t *testing.T) {
	flagSet := flag.NewFlagSet("test_flags", flag.ContinueOnError)
	flagVal := flagSet.String("pusher_util_test_var", "default", "")
	flagSet.Parse([]string{})
	if err := flagx.ArgsFromEnv(flagSet); err != nil {
		t.Error(err)
	}
	if *flagVal != "default" {
		t.Error("Bad flag value", *flagVal)
	}
}

func TestArgsFromEnvSpecifiedNoEnv(t *testing.T) {
	flagSet := flag.NewFlagSet("test_flags", flag.ContinueOnError)
	flagVal := flagSet.String("pusher_util_test_var", "default", "")
	flagSet.Parse([]string{"-pusher_util_test_var=value_from_cmdline"})
	if err := flagx.ArgsFromEnv(flagSet); err != nil {
		t.Error(err)
	}
	if *flagVal != "value_from_cmdline" {
		t.Error("Bad flag value", *flagVal)
	}
}

func TestArgsFromEnvNotSpecifiedYesEnv(t *testing.T) {
	flagSet := flag.NewFlagSet("test_flags", flag.ContinueOnError)
	flagVal := flagSet.String("pusher_util_test_var", "default", "")
	revert := osx.MustSetenv("PUSHER_UTIL_TEST_VAR", "value_from_env")
	defer revert()
	flagSet.Parse([]string{})
	if err := flagx.ArgsFromEnv(flagSet); err != nil {
		t.Error(err)
	}
	if *flagVal != "value_from_env" {
		t.Error("Bad flag value", *flagVal)
	}
}

func TestArgsFromEnvWontOverride(t *testing.T) {
	flagSet := flag.NewFlagSet("test_flags", flag.ContinueOnError)
	flagVal := flagSet.String("pusher_util_test_var", "default", "")
	revert := osx.MustSetenv("PUSHER_UTIL_TEST_VAR", "value_from_env")
	defer revert()
	flagSet.Parse([]string{"-pusher_util_test_var=value_from_cmdline"})
	if err := flagx.ArgsFromEnv(flagSet); err != nil {
		t.Error(err)
	}
	if *flagVal != "value_from_cmdline" {
		t.Error("Bad flag value", *flagVal)
	}
}

func TestArgsFromEnvWithBadEnv(t *testing.T) {
	flagSet := flag.NewFlagSet("test_flags", flag.ContinueOnError)
	flagVal := flagSet.Int("pusher_util_test_var", 1, "")
	revert := osx.MustSetenv("PUSHER_UTIL_TEST_VAR", "bad_value_from_env")
	defer revert()
	flagSet.Parse([]string{""})
	err := flagx.ArgsFromEnv(flagSet)
	if err == nil {
		t.Error("Should have had an error")
	} else {
		log.Printf("After an invalid Set() (err=%q), the flag is %d\n", err, *flagVal)
	}
}

func TestArgsFromEnvIllegalCharsInFlag(t *testing.T) {
	flagSet := flag.NewFlagSet("test_flags", flag.ContinueOnError)
	flagVal := flagSet.String("2pusher:util-test.var1", "default", "")
	revert := osx.MustSetenv("_2PUSHER_UTIL_TEST_VAR1", "value_from_env")
	defer revert()
	flagSet.Parse([]string{})
	if err := flagx.ArgsFromEnv(flagSet); err != nil {
		t.Error(err)
	}
	if *flagVal != "value_from_env" {
		t.Error("Bad flag value", *flagVal)
	}
}
