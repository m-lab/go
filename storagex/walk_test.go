package storagex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/m-lab/go/rtx"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestBucket_Walk(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	rtx.Must(err, "Failed to create client")
	visit := func(o *Object) error {
		t.Log(o.ObjectName())
		return nil
	}

	i := 0
	itNextErr := func(it *storage.ObjectIterator) (*storage.ObjectAttrs, error) {
		if i > 0 {
			return nil, fmt.Errorf("Fake error")
		}
		i = i + 1
		return it.Next()
	}

	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{
			name:   "okay",
			prefix: "t1",
		},
		{
			name:    "okay-err",
			prefix:  "t1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		bucket := NewBucket(client.Bucket("soltesz-mlab-sandbox"))
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				itNext = itNextErr
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

func TestObjectImpl_Copy(t *testing.T) {
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
			ObjectHandle: client.Bucket("soltesz-mlab-sandbox").Object("t1/okay.txt"),
			w:            &bytes.Buffer{},
			wantW:        "okay\n",
		},
		{
			name:         "bad-writer",
			ctx:          context.Background(),
			ObjectHandle: client.Bucket("soltesz-mlab-sandbox").Object("t1/okay.txt"),
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

func TestObjectImpl_LocalName(t *testing.T) {
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
			ObjectHandle: client.Bucket("soltesz-mlab-sandbox").Object("t1/okay.txt"),
			prefix:       "t1/",
			want:         "okay.txt",
		},
		{
			name:         "okay-return-basename",
			ObjectHandle: client.Bucket("soltesz-mlab-sandbox").Object("t1/okay.txt"),
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
