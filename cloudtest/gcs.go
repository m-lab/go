package cloudtest

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/storage/stiface"
	"google.golang.org/api/iterator"
)

// GCSClient provides a fake storage client that can be customized with arbitrary fake bucket contents.
type GCSClient struct {
	stiface.Client
	buckets map[string]BucketHandle
}

// AddTestBucket adds a fake bucket for testing.
func (c *GCSClient) AddTestBucket(name string, bh BucketHandle) {
	if c.buckets == nil {
		c.buckets = make(map[string]BucketHandle, 5)
	}
	c.buckets[name] = bh
}

// Close implements stiface.Client.Close
func (c GCSClient) Close() error {
	return nil
}

// Bucket implements stiface.Client.Bucket
func (c GCSClient) Bucket(name string) stiface.BucketHandle {
	return c.buckets[name]
}

// BucketHandle provides a fake BucketHandle implementation for testing.
type BucketHandle struct {
	stiface.BucketHandle
	ObjAttrs []*storage.ObjectAttrs // Objects that will be returned by iterator
}

// Attrs implements trivial stiface.BucketHandle.Attrs
func (bh BucketHandle) Attrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return &storage.BucketAttrs{}, nil
}

// Objects implements stiface.BucketHandle.Objects
func (bh BucketHandle) Objects(ctx context.Context, q *storage.Query) stiface.ObjectIterator {
	// TODO - should check if ctx has expired?
	obj := make([]*storage.ObjectAttrs, 0, len(bh.ObjAttrs))
	dir := ""
	for i := range bh.ObjAttrs {
		if !strings.HasPrefix(bh.ObjAttrs[i].Name, q.Prefix) {
			continue
		}
		if q.Delimiter != "" {
			suffix := strings.Trim(bh.ObjAttrs[i].Name[len(q.Prefix):], q.Delimiter)
			parts := strings.Split(suffix, q.Delimiter)
			if len(parts) > 1 {
				if dir != parts[0] {
					dir = parts[0]
					obj = append(obj, &storage.ObjectAttrs{Prefix: strings.Trim(q.Prefix, q.Delimiter) + q.Delimiter + parts[0]})
				}
				continue
			}
		}
		obj = append(obj, bh.ObjAttrs[i])
	}
	n := 0
	return objIt{next: &n, objects: obj}
}

// objIt provides a fake stiface.ObjectIterator
type objIt struct {
	stiface.ObjectIterator
	objects []*storage.ObjectAttrs
	next    *int
}

// Next implements stiface.ObjectIterator.Next
func (it objIt) Next() (*storage.ObjectAttrs, error) {
	if *it.next >= len(it.objects) {
		return nil, iterator.Done
	}
	*it.next++
	return it.objects[*it.next-1], nil
}
