// Package dsfake implements a fake dsiface.Client
// If you make changes to existing code, please test whether it breaks
// existing clients, e.g. in etl-gardener.
package dsfake

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"reflect"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/datastore/dsiface"
)

// NOTE: This is over-restrictive, but fine for current purposes.
func validateDatastoreEntity(e interface{}) error {
	v := reflect.ValueOf(e)
	if v.Kind() != reflect.Ptr {
		return datastore.ErrInvalidEntityType
	}
	// NOTE: This is over-restrictive, but fine for current purposes.
	if reflect.Indirect(v).Kind() != reflect.Struct {
		return datastore.ErrInvalidEntityType
	}
	return nil
}

// ErrNotImplemented is returned if a dsiface function is unimplemented.
var ErrNotImplemented = errors.New("Not implemented")

// Client implements a crude datastore test client.  It is somewhat
// simplistic and incomplete.  It works only for basic Put, Get, and Delete,
// but may not always work correctly.
type Client struct {
	dsiface.Client // For unimplemented methods
	lock           sync.Mutex
	objects        map[datastore.Key]reflect.Value
}

// NewClient returns a fake client that satisfies dsiface.Client.
func NewClient() *Client {
	if flag.Lookup("test.v") == nil {
		log.Fatal("DSFakeClient should only be used in tests")
	}
	return &Client{objects: make(map[datastore.Key]reflect.Value, 10)}
}

// Close implements dsiface.Client.Close
func (c *Client) Close() error { return nil }

// Count implements dsiface.Client.Count
func (c *Client) Count(ctx context.Context, q *datastore.Query) (n int, err error) {
	return 0, ErrNotImplemented
}

// Delete implements dsiface.Client.Delete
func (c *Client) Delete(ctx context.Context, key *datastore.Key) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.objects[*key]
	if !ok {
		return datastore.ErrNoSuchEntity
	}
	delete(c.objects, *key)
	return nil
}

// Get implements dsiface.Client.Get
func (c *Client) Get(ctx context.Context, key *datastore.Key, dst interface{}) (err error) {
	err = validateDatastoreEntity(dst)
	if err != nil {
		return err
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	v := reflect.ValueOf(dst)
	o, ok := c.objects[*key]
	if !ok {
		return datastore.ErrNoSuchEntity
	}
	reflect.Indirect(v).Set(o)
	return nil
}

// Put mplements dsiface.Client.Put
func (c *Client) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	err := validateDatastoreEntity(src)
	if err != nil {
		return nil, err
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	v := reflect.ValueOf(src)
	c.objects[*key] = reflect.Indirect(v)
	return key, nil
}

// DumpKeys lists all keys saved in the fake client.
func (c *Client) DumpKeys() []datastore.Key {
	c.lock.Lock()
	defer c.lock.Unlock()
	keys := make([]datastore.Key, len(c.objects))
	i := 0
	for k, v := range c.objects {
		keys[i] = k
		i++
		log.Output(2, fmt.Sprint(k, v))
	}

	return keys
}
