package bqfake_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/m-lab/etl-gardener/cloud/bq"
	bqfake "github.com/m-lab/go/bq-fake"
	"github.com/m-lab/go/dataset"
)

func TestCachedMeta(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "mlab-testing")
	if err != nil {
		panic(err)
	}
	ds := c.Dataset("etl")

	{
		meta := bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
		meta.TimePartitioning = &bigquery.TimePartitioning{Expiration: 0 * time.Second}
		ds.(bqfake.Dataset).AddTable("DedupTest", &meta)
	}

	tbl := ds.Table("DedupTest")
	meta, err := tbl.Metadata(ctx)
	if err != nil {
		t.Error(err)
	} else if meta == nil {
		t.Error("Meta should not be nil")
	}

	dsExt := dataset.Dataset{ds, *c}
	at := bq.NewAnnotatedTable(tbl, &dsExt)
	// Fetch cache detail - which hits backend
	meta, err = at.CachedMeta(ctx)
	if err != nil {
		t.Error(err)
	} else if meta == nil {
		t.Error("Meta should not be nil")
	}
	// Fetch again, exercising the cached code path.
	meta, err = at.CachedMeta(ctx)
	if err != nil {
		t.Error(err)
	} else if meta == nil {
		t.Error("Meta should not be nil")
	}
}
