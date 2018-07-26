package flagext_test

import (
	"flag"
	"log"
	"os"
	"testing"

	flagx "github.com/m-lab/go/flagext"
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
	oldVal, ok := os.LookupEnv("PUSHER_UTIL_TEST_VAR")
	os.Setenv("PUSHER_UTIL_TEST_VAR", "value_from_env")
	defer func() {
		if ok {
			os.Setenv("PUSHER_UTIL_TEST_VAR", oldVal)
		} else {
			os.Unsetenv("PUSHER_UTIL_TEST_VAR")
		}
	}()
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
	oldVal, ok := os.LookupEnv("PUSHER_UTIL_TEST_VAR")
	os.Setenv("PUSHER_UTIL_TEST_VAR", "value_from_env")
	defer func() {
		if ok {
			os.Setenv("PUSHER_UTIL_TEST_VAR", oldVal)
		} else {
			os.Unsetenv("PUSHER_UTIL_TEST_VAR")
		}
	}()
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
	oldVal, ok := os.LookupEnv("PUSHER_UTIL_TEST_VAR")
	os.Setenv("PUSHER_UTIL_TEST_VAR", "bad_value_from_env")
	defer func() {
		if ok {
			os.Setenv("PUSHER_UTIL_TEST_VAR", oldVal)
		} else {
			os.Unsetenv("PUSHER_UTIL_TEST_VAR")
		}
	}()
	flagSet.Parse([]string{""})
	err := flagx.ArgsFromEnv(flagSet)
	if err == nil {
		t.Error("Should have had an error")
	} else {
		log.Printf("After an invalid Set() (err=%q), the flag is %d\n", err, *flagVal)
	}
}
