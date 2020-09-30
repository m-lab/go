package uploader

import (
	"context"
	"reflect"
	"testing"

	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"github.com/m-lab/go/cloudtest/gcsfake"
)

func TestNew(t *testing.T) {
	client := &gcsfake.GCSClient{}
	u := New(client, "bucket_name")
	if u == nil {
		t.Errorf("New() returned nil")
	}
}

func TestUploader_Upload(t *testing.T) {
	type fields struct {
		client stiface.Client
		bucket stiface.BucketHandle
	}
	type args struct {
		ctx     context.Context
		path    string
		content []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				bucket: &gcsfake.BucketHandle{
					Objs: make(map[string]*gcsfake.ObjectHandle, 0),
				},
			},
			args: args{
				ctx:     context.Background(),
				content: []byte("test"),
				path:    "this/is/a/test",
			},
		},
		{
			name: "write-fails",
			fields: fields{
				bucket: &gcsfake.BucketHandle{
					Objs:           make(map[string]*gcsfake.ObjectHandle, 0),
					WritesMustFail: true,
				},
			},
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
			u := &Uploader{
				client: tt.fields.client,
				bucket: tt.fields.bucket,
			}
			obj, err := u.Upload(tt.args.ctx, tt.args.path, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uploader.Upload() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && obj == nil {
				t.Errorf("Uploader.Upload() returned a nil object.")
			}
			// Check uploaded content.
			if err == nil {
				if bucket, ok := tt.fields.bucket.(*gcsfake.BucketHandle); ok {
					uploaded := bucket.Objs[tt.args.path]
					if !reflect.DeepEqual(uploaded.Data.Bytes(), tt.args.content) {
						t.Errorf("Uploader.Upload() didn't upload the expected data")
					}
				}
			}
		})
	}
}
