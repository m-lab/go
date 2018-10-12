package bqfake_test

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/m-lab/etl-gardener/cloud/bq"
	bqfake "github.com/m-lab/go/bq-fake"
	"github.com/m-lab/go/dataset"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestCachedMeta(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "mlab-testing")
	if err != nil {
		panic(err)
	}
	ds := c.Dataset("etl")

	// TODO - also test table before it exists
	// Test whether changes to table can be seen in existing table objects.

	{
		meta := bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
		meta.TimePartitioning = &bigquery.TimePartitioning{Expiration: 0 * time.Second}
		tbl := ds.Table("DedupTest")
		err := tbl.Create(ctx, &meta)
		if err != nil {
			t.Fatal(err)
		}
	}

	tbl := ds.Table("DedupTest")
	log.Println(tbl)
	meta, err := tbl.Metadata(ctx)
	if err != nil {
		t.Error(err)
	} else if meta == nil {
		t.Error("Meta should not be nil")
	}

	dsExt := dataset.Dataset{Dataset: ds, BqClient: *c}
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
