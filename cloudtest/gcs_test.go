package cloudtest_test

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/storage/stiface"
	"github.com/m-lab/go/cloudtest"
	"google.golang.org/api/iterator"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

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

	fc := cloudtest.GCSClient{}
	fc.AddTestBucket("foobar",
		cloudtest.BucketHandle{
			ObjAttrs: []*storage.ObjectAttrs{
				&storage.ObjectAttrs{Name: "ndt/2019/01/01/obj1", Updated: time.Now()},
				&storage.ObjectAttrs{Name: "ndt/2019/01/01/obj2", Updated: time.Now()},
				&storage.ObjectAttrs{Name: "ndt/2019/01/01/obj3"},
				&storage.ObjectAttrs{Name: "ndt/2019/01/01/subdir/obj4", Updated: time.Now()},
				&storage.ObjectAttrs{Name: "ndt/2019/01/01/subdir/obj5", Updated: time.Now()},
				&storage.ObjectAttrs{Name: "obj6", Updated: time.Now()},
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

	qry := storage.Query{
		Delimiter: "/",
		Prefix:    "ndt/2019/01/01/",
	}
	it := bucket.Objects(ctx, &qry)

	_, n, p := countAll(t, it)
	if n != 2 {
		t.Error("Expected 2 items, got", n)
	}
	if p != 1 {
		t.Error("Expected 1 prefix, got", p)
	}

	qry = storage.Query{
		Delimiter: "/",
		Prefix:    "ndt/2019/01/01",
	}
	it = bucket.Objects(ctx, &qry)

	_, n, p = countAll(t, it)
	if n != 2 {
		t.Error("Expected 2 items, got", n)
	}
	if p != 1 {
		t.Error("Expected 1 prefix, got", p)
	}

	qry = storage.Query{
		Delimiter: "/",
		Prefix:    "ndt/2019/01/01/obj",
	}
	it = bucket.Objects(ctx, &qry)

	_, n, p = countAll(t, it)
	if n != 2 {
		t.Error("Expected 2 items, got", n)
	}
	if p != 0 {
		t.Error("Expected 0 prefix, got", p)
	}

	fc.Close()
}
