package dsfake_test

import (
	"log"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/m-lab/go/cloudtest/dsfake"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func must(t *testing.T, err error) {
	if err != nil {
		log.Output(2, err.Error())
		t.Fatal(err)
	}
}

type Object struct {
	Value string
}

func TestDSFake(t *testing.T) {
	client := dsfake.NewClient()

	k1 := datastore.NameKey("TestDSFake", "o1", nil)
	k1.Namespace = "dsfake"
	k2 := datastore.NameKey("TestDSFake", "o2", nil)
	k2.Namespace = "dsfake"

	o1 := Object{"o1"}
	_, err := client.Put(nil, k1, &o1)
	must(t, err)

	var o1a Object
	// This should fail because it o1a isn't a pointer
	err = client.Get(nil, k1, o1a)
	if err != datastore.ErrInvalidEntityType {
		t.Error("Should detect non-pointer")
	}

	must(t, client.Get(nil, k1, &o1a))
	if o1a.Value != o1.Value {
		t.Fatal("Failed put/get", o1a, o1)
	}

	// A second object should not interfere with the first.
	o2 := Object{"o2"}
	_, err = client.Put(nil, k2, &o2)
	must(t, err)

	// Check that Get still fetches the correct o1 value
	var o1b Object
	must(t, client.Get(nil, k1, &o1b))
	if o1b.Value != o1.Value {
		t.Fatal("Apparent object collision", o1b, o1)
	}

	client.DumpKeys()
	o2.Value = "local-o2"
	// Check that changing original object doesn't change the stored value.
	var o2a Object
	must(t, client.Get(nil, k2, &o2a))
	if o2a.Value != "o2" {
		t.Error("Changing local modifies persisted value", o2a.Value, "!=", "o2")
	}

	// test DumpKeys()
	keys := client.DumpKeys()
	if len(keys) != 2 {
		t.Fatal("Should be 2 keys")
	}
	if keys[0] != *k1 && keys[1] != *k1 {
		t.Error("Missing key", k1)
	}
	if keys[0] != *k2 && keys[1] != *k2 {
		t.Error("Missing key", k2)
	}

	// Test Delete()
	must(t, client.Delete(nil, k1))
	err = client.Get(nil, k1, &o1b)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal("delete failed")
	}

}
