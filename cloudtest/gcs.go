package cloudtest

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/storage/stiface"
	"google.golang.org/api/iterator"
)

type GCSClient struct {
	stiface.Client
	buckets map[string]BucketHandle
}

func (c *GCSClient) AddTestBucket(name string, bh BucketHandle) {
	if c.buckets == nil {
		c.buckets = make(map[string]BucketHandle, 5)
	}
	c.buckets[name] = bh
}

func (c GCSClient) Close() error {
	return nil
}

func (c GCSClient) Bucket(name string) stiface.BucketHandle {
	return c.buckets[name]
}

type BucketHandle struct {
	stiface.BucketHandle
	ObjAttrs []*storage.ObjectAttrs // Objects that will be returned by iterator
}

func (bh BucketHandle) Attrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return &storage.BucketAttrs{}, nil
}

func (bh BucketHandle) Objects(ctx context.Context, q *storage.Query) stiface.ObjectIterator {
	// TODO - should check if ctx has expired?
	obj := make([]*storage.ObjectAttrs, 0, len(bh.ObjAttrs))
	for i := range bh.ObjAttrs {
		if strings.HasPrefix(bh.ObjAttrs[i].Name, q.Prefix) {
			obj = append(obj, bh.ObjAttrs[i])
		}
	}
	n := 0
	return objIt{next: &n, objects: obj}
}

type objIt struct {
	stiface.ObjectIterator
	objects []*storage.ObjectAttrs
	next    *int
}

func (it objIt) Next() (*storage.ObjectAttrs, error) {
	if *it.next >= len(it.objects) {
		return nil, iterator.Done
	}
	*it.next++
	return it.objects[*it.next-1], nil
}

func assertStifaceClient() { func(c stiface.Client) {}(&GCSClient{}) }
