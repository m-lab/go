package gcsfake

import (
	"bytes"
	"context"
	"log"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"github.com/m-lab/go/testingx"
	"google.golang.org/api/iterator"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func assertStifaceClient() { func(c stiface.Client) {}(&GCSClient{}) }

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

	fc := GCSClient{}
	fc.AddTestBucket("foobar",
		&BucketHandle{
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

func TestBucketHandle_Object(t *testing.T) {
	bh := &BucketHandle{}
	testObj := &ObjectHandle{
		Bucket: bh,
		Name:   "test/obj",
		Data:   new(bytes.Buffer),
	}
	bh.Objs = map[string]*ObjectHandle{
		"test/obj": testObj,
	}

	// Get an existing object.
	got := bh.Object("test/obj")
	if !reflect.DeepEqual(got, testObj) {
		t.Errorf("BucketHandle.Object() = %v, want %v", got, testObj)
	}

	// Get a new object.
	got = bh.Object("non/existing/obj")
	if fakeObj, ok := got.(*ObjectHandle); ok {
		if fakeObj.Name != "non/existing/obj" || fakeObj.Bucket != bh ||
			fakeObj.Data == nil || fakeObj.WritesMustFail {
			t.Errorf("Object() didn't return the expected ObjectHandle.")
		}
	} else {
		t.Errorf("Object() didn't return a fake ObjectHandle")
	}
}

func TestObjectHandle_NewReader(t *testing.T) {
	buf := new(bytes.Buffer)
	buf.WriteString("test")
	obj := &ObjectHandle{
		Data:           buf,
		WritesMustFail: true,
	}
	got, err := obj.NewReader(context.Background())
	testingx.Must(t, err, "NewReader() failed")
	if reader, ok := got.(*fakeReader); ok {
		if reader.buf != obj.Data {
			t.Errorf("NewReader() did not return the expected Reader")
		}
	} else {
		t.Errorf("NewReader() did not return a *fakeReader")
	}
}

func TestObjectHandle_NewWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	obj := &ObjectHandle{
		Data:           buf,
		WritesMustFail: true,
	}

	got := obj.NewWriter(context.Background())
	if fakeWriter, ok := got.(*fakeWriter); ok {
		if fakeWriter.object != obj || fakeWriter.buf != buf ||
			fakeWriter.mustFail != obj.WritesMustFail {
			t.Errorf("NewWriter() didn't return the expected Writer")
		}
	}
}

func Test_fakeWriter_Write(t *testing.T) {
	testStr := []byte("test")
	bh := &BucketHandle{
		Objs: make(map[string]*ObjectHandle, 0),
	}
	obj := &ObjectHandle{
		Bucket: bh,
	}
	w := &fakeWriter{
		object: obj,
		buf:    new(bytes.Buffer),
	}
	got, err := w.Write(testStr)
	if err != nil {
		t.Errorf("Write() returned an error: %v", err)
	}
	if got != len(testStr) || string(w.buf.Bytes()) != string(testStr) {
		t.Error("Write() didn't write the expected []byte")
	}

	w.mustFail = true
	got, err = w.Write(testStr)
	if err == nil {
		t.Errorf("Write(): expected err, got nil")
	}
}

func Test_fakeWriter_Close(t *testing.T) {
	w := &fakeWriter{}
	if err := w.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	w.closeMustFail = true
	if err := w.Close(); err == nil {
		t.Errorf("Close() did not return an error")
	}

}

func Test_fakeReader_Read(t *testing.T) {
	w := &fakeReader{
		buf: bytes.NewBuffer([]byte("test")),
	}
	got := make([]byte, 4)
	_, err := w.Read(got)
	testingx.Must(t, err, "Read() returned an error")
	if string(got) != "test" {
		t.Errorf("Read(): got %s, expected test", string(got))
	}
}
