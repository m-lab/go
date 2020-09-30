package gcsfake_test

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"google.golang.org/api/iterator"

	"github.com/m-lab/go/cloudtest/gcsfake"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func assertStifaceClient() { func(c stiface.Client) {}(&gcsfake.GCSClient{}) }

func countAll(t *testing.T, it stiface.ObjectIterator) (total int, normal int, prefix int) {
	for o, err := it.Next(); err != iterator.Done; o, err = it.Next() {
		if err != nil {
			t.Fatal(err, "when attempting it.Next()")
			continue
		}

		total++

		if o.Prefix != "" {
			prefix++
			log.Println("Skipping", o.Prefix)
			continue
		}
		if o.Updated.Before(time.Now().Add(-time.Minute)) {
			continue
		}
		normal++
	}
	return
}

func TestGCSClient(t *testing.T) {
	// Use a fake queue client.
	ctx := context.Background()

	fc := gcsfake.GCSClient{}
	fc.AddTestBucket("foobar",
		&gcsfake.BucketHandle{
			ObjAttrs: []*storage.ObjectAttrs{
				{Name: "ndt/2019/01/01/obj1", Updated: time.Now()},
				{Name: "ndt/2019/01/01/obj2", Updated: time.Now()},
				{Name: "ndt/2019/01/01/obj3"},                             // Will be filtered out by the "since" filter.
				{Name: "ndt/2019/01/01/subdir/obj4", Updated: time.Now()}, // filtered because of subdir.
				{Name: "ndt/2019/01/01/subdir/obj5", Updated: time.Now()},
				{Name: "obj6", Updated: time.Now()}, // Will be filtered by prefix.
			}})

	bucket := fc.Bucket("foobar")
	if bucket == nil {
		t.Fatal("Bucket is nil")
	}
	// Check that the bucket is valid, by fetching it's attributes.
	// Bypass check if we are running travis tests.
	_, err := bucket.Attrs(ctx)
	if err != nil {
		t.Error(err)
	}

	type test struct {
		prefix string
		n      int
		p      int
	}
	tests := []test{
		{"ndt/2019/01/01", 2, 1},     // Prefix that is a directory, but without the final /
		{"ndt/2019/01/01/", 2, 1},    // Should work both with and without slash.
		{"ndt/2019/01/01/obj", 2, 0}, // Should work with a prefix that isn't a directory.
	}

	for _, tt := range tests {
		qry := storage.Query{
			Delimiter: "/",
			Prefix:    tt.prefix,
		}
		it := bucket.Objects(ctx, &qry)

		_, n, p := countAll(t, it)
		if n != tt.n {
			t.Error("Expected", tt.n, "items, got", n)
		}
		if p != tt.p {
			t.Error("Expected", tt.p, "prefix, got", p)
		}
	}

	fc.Close()
}
