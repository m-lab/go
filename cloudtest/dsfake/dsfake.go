// Package dsfake implements a fake dsiface.Client
// If you make changes to existing code, please test whether it breaks
// existing clients, e.g. in etl-gardener.
package dsfake

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"reflect"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/googleapis/google-cloud-go-testing/datastore/dsiface"
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
	objects        map[datastore.Key][]byte
}

// NewClient returns a fake client that satisfies dsiface.Client.
func NewClient() *Client {
	if flag.Lookup("test.v") == nil {
		log.Fatal("DSFakeClient should only be used in tests")
	}
	return &Client{objects: make(map[datastore.Key][]byte, 10)}
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
	o, ok := c.objects[*key]
	if !ok {
		return datastore.ErrNoSuchEntity
	}
	return json.Unmarshal(o, dst)
}

// Put mplements dsiface.Client.Put
func (c *Client) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	err := validateDatastoreEntity(src)
	if err != nil {
		return nil, err
	}
	js, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.objects[*key] = js
	return key, nil
}

// GetKeys lists all keys saved in the fake client.
func (c *Client) GetKeys() []datastore.Key {
	c.lock.Lock()
	defer c.lock.Unlock()
	keys := make([]datastore.Key, len(c.objects))
	i := 0
	for k := range c.objects {
		keys[i] = k
		i++
	}

	return keys
}
