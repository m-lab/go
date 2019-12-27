package storagex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/m-lab/go/rtx"
	"google.golang.org/api/iterator"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

// fakeIter is the common interface for the implemented iter types below.
type fakeIter interface {
	itNext(it *storage.ObjectIterator) (*storage.ObjectAttrs, error)
}

// errIter allows injecting iteration errors.
type errIter struct {
	i int
}

func (e *errIter) itNext(it *storage.ObjectIterator) (*storage.ObjectAttrs, error) {
	if e.i > 0 {
		return nil, fmt.Errorf("Fake error")
	}
	e.i++
	return it.Next()
}

// iterDone allows injecting a single synthetic "directory" object.
type iterDone struct {
	i      int
	prefix string
}

func (d *iterDone) itNext(it *storage.ObjectIterator) (*storage.ObjectAttrs, error) {
	if d.i > 0 {
		return nil, iterator.Done
	}
	d.i++
	attr := &storage.ObjectAttrs{
		Name:   "",
		Prefix: d.prefix,
	}
	return attr, nil
}

func TestBucket_Walk(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	rtx.Must(err, "Failed to create client")
	visit := func(o *Object) error {
		t.Log(o.ObjectName())
		return nil
	}

	tests := []struct {
		name    string
		prefix  string
		iter    fakeIter
		wantErr bool
	}{
		{
			name:   "okay",
			prefix: "t1",
		},
		{
			name:    "okay-err",
			prefix:  "t1",
			iter:    &errIter{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		bucket := NewBucket(client.Bucket("m-lab-go-storagex-mlab-testing"))
		t.Run(tt.name, func(t *testing.T) {
			if tt.iter != nil {
				bucket.itNext = tt.iter.itNext
			}
			if err := bucket.Walk(ctx, tt.prefix, visit); (err != nil) != tt.wantErr {
				t.Errorf("walk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type errorWriter struct{}

func (f *errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("Fake write error")
}

func TestObject_Copy(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	rtx.Must(err, "Failed to create client")

	tests := []struct {
		name         string
		ObjectHandle *storage.ObjectHandle
		ctx          context.Context
		w            io.Writer
		wantW        string
		wantErr      bool
	}{
		{
			name:         "newreader-error",
			ctx:          context.Background(),
			ObjectHandle: &storage.ObjectHandle{},
			wantErr:      true,
		},
		{
			name:         "okay",
			ctx:          context.Background(),
			ObjectHandle: client.Bucket("m-lab-go-storagex-mlab-testing").Object("t1/okay.txt"),
			w:            &bytes.Buffer{},
			wantW:        "okay\n",
		},
		{
			name:         "bad-writer",
			ctx:          context.Background(),
			ObjectHandle: client.Bucket("m-lab-go-storagex-mlab-testing").Object("t1/okay.txt"),
			w:            &errorWriter{},
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Object{
				ObjectHandle: tt.ObjectHandle,
			}
			if err := o.Copy(tt.ctx, tt.w); (err != nil) != tt.wantErr {
				t.Errorf("Object.Copy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if buf, ok := tt.w.(*bytes.Buffer); ok {
				if gotW := buf.String(); gotW != tt.wantW {
					t.Errorf("Object.Copy() = %v, want %v", gotW, tt.wantW)
				}
			}
		})
	}
}

func TestObject_LocalName(t *testing.T) {
	// An empty client is sufficient, since we make no network operations.
	client := storage.Client{}

	tests := []struct {
		name         string
		ObjectHandle *storage.ObjectHandle
		prefix       string
		want         string
	}{
		{
			name:         "okay-remove-prefix",
			ObjectHandle: client.Bucket("m-lab-go-storagex-mlab-testing").Object("t1/okay.txt"),
			prefix:       "t1/",
			want:         "okay.txt",
		},
		{
			name:         "okay-return-basename",
			ObjectHandle: client.Bucket("m-lab-go-storagex-mlab-testing").Object("t1/okay.txt"),
			prefix:       "t1/okay.txt",
			want:         "okay.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Object{
				ObjectHandle: tt.ObjectHandle,
				prefix:       tt.prefix,
			}
			if got := o.LocalName(); got != tt.want {
				t.Errorf("Object.LocalName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBucket_Dirs(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	rtx.Must(err, "Failed to create client")
	tests := []struct {
		name    string
		b       *storage.BucketHandle
		iter    fakeIter
		want    []string
		wantErr bool
	}{
		{
			name: "success",
			b:    client.Bucket("m-lab-go-storagex-mlab-testing"),
			iter: &iterDone{prefix: "foo/"},
			want: []string{"foo/"},
		},
		{
			name:    "error-x",
			b:       client.Bucket("m-lab-go-storagex-mlab-testing"),
			iter:    &errIter{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bucket{
				BucketHandle: tt.b,
				itNext:       tt.iter.itNext,
			}
			ctx := context.Background()
			got, err := b.Dirs(ctx, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Bucket.Dirs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bucket.Dirs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleBucket_Walk() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	rtx.Must(err, "Failed to allocate storage.Client")

	bucket := NewBucket(client.Bucket("your-bucket-name"))
	bucket.Walk(ctx, "path/in/bucket", func(o *Object) error {
		fmt.Println(o.ObjectName(), o.LocalName())
		return nil
	})
}
