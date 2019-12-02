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

	a := Object{"first"}
	key := datastore.NameKey("TestDSFake", "o1", nil)
	key.Namespace = "dsfake"
	key2 := datastore.NameKey("TestDSFake", "o2", nil)
	key2.Namespace = "dsfake"

	_, err := client.Put(nil, key, &a)
	must(t, err)

	var b Object
	// This should fail because it r isn't a pointer
	err = client.Get(nil, key, b)
	if err != datastore.ErrInvalidEntityType {
		t.Error("Should detect non-pointer")
	}

	must(t, client.Get(nil, key, &b))
	if b.Value != a.Value {
		t.Fatal("Failed put/get")
	}

	// A second object should not interfere with the first.
	// c := Object{"other"} // TODO - reuse b instead to test independence.
	b.Value = "other"
	_, err = client.Put(nil, key2, &b)
	must(t, err)
	client.DumpKeys()

	b.Value = ""
	must(t, client.Get(nil, key, &b))
	if b.Value != a.Value {
		t.Fatal("Apparent object collision", a, b)
	}

	// test DumpKeys
	keys := client.DumpKeys()
	if len(keys) != 2 {
		t.Fatal("Should be 2 keys")
	}
	if keys[0] != *key && keys[1] != *key {
		t.Error("Missing key", key)
	}
	if keys[0] != *key2 && keys[1] != *key2 {
		t.Error("Missing key", key2)
	}

	must(t, client.Delete(nil, key))
	err = client.Get(nil, key, &a)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal("delete failed")
	}

	var c Object
	// key2 object should still exist
	must(t, client.Get(nil, key2, &c))
	if c.Value != "other" {
		t.Error("Wrong value", c.Value)
	}

}
