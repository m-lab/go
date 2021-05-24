package uploader

import (
	"bytes"
	"context"
	"io"

	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
)

// Uploader is a Google Cloud Storage uploader.
type Uploader struct {
	client stiface.Client
	bucket stiface.BucketHandle
}

// New returns a new Uploader using the specified Client.
func New(client stiface.Client, bucket string) *Uploader {
	return &Uploader{
		client: client,
		bucket: client.Bucket(bucket),
	}
}

// Upload uploads the provided buffer to the specified GCS path.
func (u *Uploader) Upload(ctx context.Context, path string, content []byte) (stiface.ObjectHandle, error) {
	obj := u.bucket.Object(path)
	w := obj.NewWriter(ctx)

	_, err := io.Copy(w, bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}

	// Avoid using defer w.Close() here as it would hide errors occurring
	// while closing the writer, such as permission errors.
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return obj, nil
}
