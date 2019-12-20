// Package storagex extends cloud.google.com/go/storage.
package storagex

import (
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// Object extends the storage.ObjectHandle operations on GCS Objects. Objects are
// generated during a Bucket.Walk.
type Object struct {
	*storage.ObjectHandle
	prefix string
}

// LocalName returns a path suitable for creating a local file. The local name
// may include path components because it is derived from the original GCS Object
// name with the original Walk pathPrefix removed. If the pathPrefix equals the
// GCS Object name (such as when pathPrefix is a single object), then the Object
// base name is returned.
func (o *Object) LocalName() string {
	if o.ObjectName() == o.prefix {
		// For single-file downloads, remove everything but the basename.
		return path.Base(o.ObjectName())
	}
	// Remove the initial prefix from object name.
	return strings.TrimPrefix(o.ObjectName(), o.prefix)
}

// Copy writes the Object data to the given writer.
func (o *Object) Copy(ctx context.Context, w io.Writer) error {
	r, err := o.NewReader(ctx)
	if err != nil {
		log.Println("Failed to get reader for object:", err)
		return err
	}
	defer r.Close()
	if _, err := io.Copy(w, r); err != nil {
		log.Println("Failed to io.Copy object:", err)
		return err
	}
	return nil
}

// Bucket extends storage.BucketHandle operations.
type Bucket struct {
	*storage.BucketHandle
}

// NewBucket creates a new Bucket.
func NewBucket(b *storage.BucketHandle) *Bucket {
	return &Bucket{BucketHandle: b}
}

// Walk visits each GCS object under pathPrefix and calls visit with every object. The given
// pathPrefix may be a GCS object name, in which case Walk will visit only that object.
func (b *Bucket) Walk(ctx context.Context, pathPrefix string, visit func(o *Object) error) error {
	return walk(ctx, b, pathPrefix, pathPrefix, visit)
}

// WalkIf visits each GCS path object under pathRegex and calls visit with each.
func (b *Bucket) WalkIf(ctx context.Context, pathRegex string, visit func(o *Object) error) error {
	return walkIf(ctx, b, pathRegex, "", visit)
}

var itNext = func(it *storage.ObjectIterator) (*storage.ObjectAttrs, error) {
	return it.Next()
}

func walkIf(ctx context.Context, bucket *Bucket, pathRegex, rootPrefix string, visit func(o *Object) error) error {
	it := bucket.Objects(ctx, &storage.Query{Prefix: rootPrefix, Delimiter: "/"})
	for {
		attr, err := itNext(it)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			log.Println("failed to list bucket:", err)
			return err
		}
		fmt.Println("prefix", attr.Prefix)
		m, err := regexp.MatchString(pathRegex, attr.Prefix)
		if m {
			// We found an object.
			fmt.Println("no name", attr.Name, attr.Prefix)
			visit(&Object{ObjectHandle: bucket.Object(attr.Prefix), prefix: rootPrefix})
			err = walkIf(ctx, bucket, pathRegex, attr.Prefix, visit)
			if err != nil {
				return err
			}
		} // else if !strings.HasSuffix(attr.Name, "/") {
		// Pseudo-directory entries have no name.
		// fmt.Println("no suffix", attr.Name, attr.Prefix)
		// visit(&Object{ObjectHandle: bucket.Object(attr.Prefix), prefix: rootPrefix})
		// } else {
		// fmt.Println("else", attr.Name, attr.Prefix)
		// }
	}
}

// walk recursively iterates over every GCS Object in the given bucket whose
// names begin with prefix. Each Object is passed to `visit`.
func walk(ctx context.Context, bucket *Bucket, prefix, rootPrefix string, visit func(o *Object) error) error {
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix, Delimiter: "/"})
	for {
		attr, err := itNext(it)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			log.Println("failed to list bucket:", err)
			return err
		}
		if attr.Name == "" {
			// Pseudo-directory entries have no name.
			err = walk(ctx, bucket, attr.Prefix, rootPrefix, visit)
			if err != nil {
				return err
			}
		} else if !strings.HasSuffix(attr.Name, "/") {
			// We found an object.
			visit(&Object{ObjectHandle: bucket.Object(attr.Name), prefix: rootPrefix})
		}
	}
}

func ListDirs(ctx context.Context, bucket *Bucket, rootPrefix string) ([]string, error) {
	var ret []string
	it := bucket.Objects(ctx, &storage.Query{Prefix: rootPrefix, Delimiter: "/"})
	for {
		attr, err := itNext(it)
		if err == iterator.Done {
			return ret, nil
		}
		if err != nil {
			log.Println("failed to list bucket:", err)
			return nil, err
		}
		if strings.HasSuffix(attr.Prefix, "/") {
			ret = append(ret, attr.Prefix)
		}
	}
}
