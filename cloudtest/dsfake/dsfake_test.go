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

	o := Object{"init"}
	key := datastore.NameKey("TestDSFake", "misc", nil)
	key.Namespace = "dsfake"

	_, err := client.Put(nil, key, &o)
	must(t, err)

	var r Object
	// This should fail because it r isn't a pointer
	err = client.Get(nil, key, r)
	if err != datastore.ErrInvalidEntityType {
		t.Error("Should detect non-pointer")
	}

	must(t, client.Get(nil, key, &r))
	if r.Value != o.Value {
		t.Fatal("Failed put/get")
	}

	must(t, client.Delete(nil, key))
	err = client.Get(nil, key, &r)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal("delete failed")
	}

}
