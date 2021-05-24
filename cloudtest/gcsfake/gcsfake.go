//  Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cloudtest provides utilities for testing, e.g. cloud
// service tests using mock http Transport, fake storage client, etc.
package gcsfake

import (
	"bytes"
	"context"
	"errors"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"google.golang.org/api/iterator"
)

// GCSClient provides a fake storage client that can be customized with arbitrary fake bucket contents.
type GCSClient struct {
	stiface.Client
	buckets map[string]*BucketHandle
}

// AddTestBucket adds a fake bucket for testing.
func (c *GCSClient) AddTestBucket(name string, bh *BucketHandle) {
	if c.buckets == nil {
		c.buckets = make(map[string]*BucketHandle, 5)
	}
	c.buckets[name] = bh
}

// Close implements stiface.Client.Close
func (c *GCSClient) Close() error {
	return nil
}

// Bucket implements stiface.Client.Bucket
func (c *GCSClient) Bucket(name string) stiface.BucketHandle {
	return c.buckets[name]
}

// BucketHandle provides a fake BucketHandle implementation for testing.
type BucketHandle struct {
	stiface.BucketHandle
	ObjAttrs       []*storage.ObjectAttrs // Objects that will be returned by iterator
	Objs           map[string]*ObjectHandle
	WritesMustFail bool
	ClosesMustFail bool
}

// NewBucketHandle creates a new empty BucketHandle.
func NewBucketHandle() *BucketHandle {
	return &BucketHandle{
		ObjAttrs: make([]*storage.ObjectAttrs, 0),
		Objs:     make(map[string]*ObjectHandle),
	}
}

// Attrs implements trivial stiface.BucketHandle.Attrs
func (bh *BucketHandle) Attrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return &storage.BucketAttrs{}, nil
}

// Object returns an ObjectHandle for the specified object name if it exists
// in this bucket, or a new ObjectHandle otherwise.
func (bh *BucketHandle) Object(name string) stiface.ObjectHandle {
	if o, ok := bh.Objs[name]; ok {
		return o
	}
	return &ObjectHandle{
		Name:           name,
		Bucket:         bh,
		Data:           new(bytes.Buffer),
		WritesMustFail: bh.WritesMustFail,
		ClosesMustFail: bh.ClosesMustFail,
	}
}

// Objects implements stiface.BucketHandle.Objects
func (bh *BucketHandle) Objects(ctx context.Context, q *storage.Query) stiface.ObjectIterator {
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
	return &objIt{ctx: ctx, next: &n, objects: obj}
}

// objIt provides a fake stiface.ObjectIterator
type objIt struct {
	ctx context.Context
	stiface.ObjectIterator
	objects []*storage.ObjectAttrs
	next    *int
}

// Next implements stiface.ObjectIterator.Next
func (it *objIt) Next() (*storage.ObjectAttrs, error) {
	if it.ctx.Err() != nil {
		return nil, it.ctx.Err()
	}
	if *it.next >= len(it.objects) {
		return nil, iterator.Done
	}
	*it.next++
	return it.objects[*it.next-1], nil
}

// ObjectHandle implements stiface.ObjectHandle
type ObjectHandle struct {
	stiface.ObjectHandle
	Name           string
	Bucket         *BucketHandle
	Data           *bytes.Buffer
	WritesMustFail bool
	ClosesMustFail bool
}

// NewReader returns a fakeReader for this ObjectHandle.
func (o *ObjectHandle) NewReader(context.Context) (stiface.Reader, error) {
	return &fakeReader{
		buf: o.Data,
	}, nil
}

// NewWriter returns a fakeWrite for this ObjectHandle.
func (o *ObjectHandle) NewWriter(context.Context) stiface.Writer {
	return &fakeWriter{
		object:        o,
		buf:           o.Data,
		mustFail:      o.WritesMustFail,
		closeMustFail: o.ClosesMustFail,
	}
}

type fakeWriter struct {
	stiface.Writer
	object        *ObjectHandle
	buf           *bytes.Buffer
	mustFail      bool
	closeMustFail bool
}

// Write writes data to the fake bucket. The object is created if it does not
// exist already.
func (w *fakeWriter) Write(p []byte) (int, error) {
	if w.mustFail {
		return 0, errors.New("write failed")
	}
	w.object.Bucket.Objs[w.object.Name] = w.object
	return w.buf.Write(p)
}
func (w *fakeWriter) Close() error {
	if w.closeMustFail {
		return errors.New("close failed")
	}
	return nil
}

type fakeReader struct {
	stiface.Reader
	buf *bytes.Buffer
}

func (r *fakeReader) Read(p []byte) (int, error) {
	return r.buf.Read(p)
}
