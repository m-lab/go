package gcs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"google.golang.org/api/iterator"
)

// BucketHandle adds functionality to stiface.BucketHandle
type BucketHandle struct {
	stiface.BucketHandle
}

// GetFilesSince returns list of all normal file objects with prefix and mTime > after.
// prefix is the path not including gs://bucket-name/, including the final /
// Will retry iterator errors up to five total.
// returns (objects, byteCount, error)
// Performance:  This handles about 5000 objects/second, including objects rejected by the regex and time cutoff.
func (bh *BucketHandle) GetFilesSince(ctx context.Context, prefix string, filter *regexp.Regexp, after time.Time) ([]*storage.ObjectAttrs, int64, error) {
	qry := storage.Query{
		Delimiter: "/", // This prevents traversing subdirectories.
		Prefix:    prefix,
	}
	it := bh.Objects(ctx, &qry)
	if it == nil {
		log.Println("Nil object iterator for", bh)
		return nil, 0, fmt.Errorf("Object iterator is nil.  BucketHandle: %v Prefix: %s", bh, prefix)
	}

	files := make([]*storage.ObjectAttrs, 0, 1000)

	byteCount := int64(0)
	gcsErrCount := 0
	for o, err := it.Next(); err != iterator.Done; o, err = it.Next() {
		if err != nil {
			// These errors are not recoverable.
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil, 0, err
			}
			gcsErrCount++
			time.Sleep(time.Second) // Helps if there is a transient network issue.
			if gcsErrCount > 5 {
				log.Printf("Failed after %d files.\n", len(files))
				return files, byteCount, err
			}
			// log the underlying error, with added context
			log.Println(err, "when attempting it.Next()")
			continue
		}

		// Prefixes have empty Updated fields, so the first clause would
		// generally skip prefixes, but we add the second clause just in
		// case someone passes in an ancient *after* date.
		if !o.Updated.After(after) || len(o.Prefix) > 0 {
			continue
		}
		// Ignore files that don't match filter.
		if filter != nil && !filter.MatchString(o.Name) {
			continue
		}
		byteCount += o.Size
		files = append(files, o)
	}
	return files, byteCount, nil
}

// *******************************************************************
// Storage Bucket related stuff.
//  TODO move to another package?
// *******************************************************************

// GetBucket gets a storage bucket.
// TODO - this is currently duplicated in etl-gardener/state/state.go
func GetBucket(ctx context.Context, sClient stiface.Client, bucketName string) (*BucketHandle, error) {
	bucket := sClient.Bucket(bucketName)
	if bucket == nil {
		return nil, errors.New("Nil bucket")
	}
	// Check that the bucket is valid, by fetching it's attributes.
	// Bypass check if we are running travis tests.
	_, err := bucket.Attrs(ctx)
	if err != nil {
		return nil, err
	}
	return &BucketHandle{bucket}, nil
}
