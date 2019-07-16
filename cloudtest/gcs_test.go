package cloudtest_test

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/m-lab/go/cloudtest"
	"google.golang.org/api/iterator"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
				&storage.ObjectAttrs{Name: "obj4", Updated: time.Now()},
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
		Prefix:    "ndt/2019/01/01",
	}
	it := bucket.Objects(ctx, &qry)

	count := 0
	for o, err := it.Next(); err != iterator.Done; o, err = it.Next() {
		if err != nil {
			// TODO - should this retry?
			// log the underlying error, with added context
			t.Error(err, "when attempting it.Next()")
			continue
		}

		if o.Prefix != "" {
			log.Println("Skipping", o.Prefix)
			continue
		}
		if o.Updated.Before(time.Now().Add(-time.Minute)) {
			continue
		}
		log.Println(o.Name)
		count++
	}
	if count != 2 {
		t.Error("Expected 2 items", count)
	}
	fc.Close()
	t.Fail()
}
