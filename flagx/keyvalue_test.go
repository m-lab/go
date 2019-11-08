package flagx_test

import (
	"flag"
	"log"
	"reflect"
	"testing"

	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/rtx"
)

func TestKeyValue(t *testing.T) {
	kv := &flagx.KeyValue{}
	rtx.Must(kv.Set("MLAB.thing1=thing2"), "Could not set thing1 to thing2")
	rtx.Must(kv.Set("MLAB.a=b"), "Could not set a to b")
	err := kv.Set("MLAB.c:d")
	if err == nil {
		t.Error("Should have had an error inserting the keypair MLAB.c:d")
	}
	if !reflect.DeepEqual(kv.Get(), map[string]string{
		"MLAB.thing1": "thing2",
		"MLAB.a":      "b",
	}) {
		t.Errorf("%v is wrong", kv.String())
	}

	if kv.String() == "" {
		t.Error("KeyValue.String() isn't expected to be useful, but it should not be empty")
	}
}

// If this compiles successfully then flagx.KeyValue conforms to the flag.Value
// interface.
func AssertKeyValueIsFlagValue(kv *flagx.KeyValue) {
	func(f flag.Value) {}(kv)
}

func Example() {
	metadata := flagx.KeyValue{}
	flag.Var(&metadata, "metadata", "Key-value pairs to be added to the metadata (flag may be repeated)")
	// Commandline flags should look like: -metadata key1=val1 -metadata key2=val2
	flag.Parse()

	log.Println(metadata.Get())
}
