package osx_test

import (
	"os"
	"testing"

	"github.com/m-lab/go/osx"
)

func TestMustSetenv(t *testing.T) {
	key := "TEST_SETENV_THIS_VARIABLE_DOES_NOT_CURRENTLY_EXIST"
	if _, present := os.LookupEnv(key); present {
		t.Error("key must not be present when the test is run")
	}
	revert1 := osx.MustSetenv(key, "value1")
	val, present := os.LookupEnv(key)
	if !present {
		t.Error("key must be present after MustSetenv")
	}
	if val != "value1" {
		t.Errorf("variable %q has the wrong value %q (should be \"value1\")", key, val)
	}
	revert2 := osx.MustSetenv(key, "value2")
	val, present = os.LookupEnv(key)
	if !present {
		t.Error("key must be present after MustSetenv")
	}
	if val != "value2" {
		t.Errorf("variable %q has the wrong value %q (should be \"value2\")", key, val)
	}
	revert2()
	val, present = os.LookupEnv(key)
	if !present {
		t.Error("key must be present after MustSetenv")
	}
	if val != "value1" {
		t.Errorf("variable %q has the wrong value %q (should be \"value1\")", key, val)
	}
	revert1()
	if _, present := os.LookupEnv(key); present {
		t.Error("key must not be present after the final revert1() call")
	}
}

func ExampleMustSetenv() {
	revert := osx.MustSetenv("PATH", "/temp/bin")
	defer revert()
}
