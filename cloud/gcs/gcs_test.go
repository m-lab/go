package gcs_test

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/storage"

	"github.com/m-lab/go/cloud/gcs"
	"github.com/m-lab/go/rtx"

	"github.com/m-lab/go/cloudtest/gcsfake"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestHasFiles(t *testing.T) {
	fc := gcsfake.GCSClient{}
	fc.AddTestBucket("foobar",
		gcsfake.BucketHandle{
			ObjAttrs: []*storage.ObjectAttrs{
				{Name: "ndt/2019/01/01/obj1", Size: 101, Updated: time.Now()},
				{Name: "ndt/2019/01/01/obj2", Size: 2020, Updated: time.Now()},
			}})

	bh, err := gcs.GetBucket(context.Background(), fc, "foobar")
	rtx.Must(err, "GetBucket")
	if ok, _ := bh.HasFiles(context.Background(), "ndt/2019"); ok {
		t.Error("Should be false")
	}
	if ok, _ := bh.HasFiles(context.Background(), "ndt/2019/01/01"); !ok {
		t.Error("Should be true")
	}
}

func TestGetFilesSince(t *testing.T) {
	fc := gcsfake.GCSClient{}
	fc.AddTestBucket("foobar",
		gcsfake.BucketHandle{
			ObjAttrs: []*storage.ObjectAttrs{
				{Name: "ndt/2019/01/01/obj1", Size: 101, Updated: time.Now()},
				{Name: "ndt/2019/01/01/obj2", Size: 2020, Updated: time.Now()},
				{Name: "ndt/2019/01/01/obj3"},
				{Name: "ndt/2019/01/01/subdir/obj4", Updated: time.Now()},
				{Name: "ndt/2019/01/01/subdir/obj5", Updated: time.Now()},
				{Name: "obj6", Updated: time.Now()},
			}})

	bh, err := gcs.GetBucket(context.Background(), fc, "foobar")
	rtx.Must(err, "GetBucket")
	files, bytes, err := bh.GetFilesSince(context.Background(), "ndt/2019/01/01/", nil, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Error("Expected 2 files, got", len(files))
	}
	if bytes != 2121 {
		t.Error("Expected total 2121 bytes, got", bytes)
	}
}

func TestGetFilesSince_Context(t *testing.T) {
	fc := gcsfake.GCSClient{}
	fc.AddTestBucket("foobar",
		gcsfake.BucketHandle{
			ObjAttrs: []*storage.ObjectAttrs{
				{Name: "ndt/2019/01/01/obj1", Size: 101, Updated: time.Now()},
			}})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	bh, err := gcs.GetBucket(context.Background(), fc, "foobar")
	rtx.Must(err, "GetBucket")
	files, _, err := bh.GetFilesSince(ctx, "ndt/2019/01/01/", nil, time.Now().Add(-time.Minute))

	if err != context.Canceled {
		t.Error("Should return context.Canceled", err)
	}
	if files != nil {
		t.Error("Should return nil files", files)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 0)
	defer cancel()
	time.Sleep(time.Millisecond)

	files, _, err = bh.GetFilesSince(ctx, "ndt/2019/01/01/", nil, time.Now().Add(-time.Minute))

	if err != context.DeadlineExceeded {
		t.Error("Should return context.Canceled", err)
	}
	if files != nil {
		t.Error("Should return nil files", files)
	}
}
