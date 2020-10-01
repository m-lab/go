package uploader

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"github.com/m-lab/go/cloudtest/gcsfake"
	"github.com/m-lab/go/testingx"
)

func TestNew(t *testing.T) {
	client := &gcsfake.GCSClient{}
	u := New(client, "bucket_name")
	if u == nil {
		t.Errorf("New() returned nil")
	}
}

func TestUploader_Upload(t *testing.T) {
	// Initialize fake client with working and failing buckets.
	client := &gcsfake.GCSClient{}
	failingBucket := gcsfake.NewBucketHandle()
	failingBucket.WritesMustFail = true
	client.AddTestBucket("test_bucket", gcsfake.NewBucketHandle())
	client.AddTestBucket("failing_bucket", failingBucket)
	type args struct {
		ctx     context.Context
		path    string
		content []byte
	}
	tests := []struct {
		name    string
		client  stiface.Client
		bucket  string
		args    args
		wantErr bool
	}{
		{
			name:   "ok",
			client: client,
			bucket: "test_bucket",
			args: args{
				ctx:     context.Background(),
				content: []byte("test"),
				path:    "this/is/a/test",
			},
		},
		{
			name:   "write-fails",
			client: client,
			bucket: "failing_bucket",
			args: args{
				ctx:     context.Background(),
				content: []byte("failing write"),
				path:    "this/is/a/test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := New(tt.client, tt.bucket)
			obj, err := u.Upload(tt.args.ctx, tt.args.path, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uploader.Upload() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && obj == nil {
				t.Errorf("Uploader.Upload() returned a nil object.")
			}
			// Check uploaded content.
			if err == nil {
				b := tt.client.Bucket(tt.bucket)
				uploaded := b.Object(tt.args.path)

				reader, err := uploaded.NewReader(context.Background())
				testingx.Must(t, err, "cannot get a Reader for the uploaded file")
				content, err := ioutil.ReadAll(reader)
				testingx.Must(t, err, "cannot read the uploaded file's contents")

				if !reflect.DeepEqual(content, tt.args.content) {
					t.Errorf("Uploader.Upload() didn't upload the expected data")
				}
			}
		})
	}
}
