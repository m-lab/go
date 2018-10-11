// Package bqfake provides tools to construct fake bigquery datasets, tables, query responses, etc.
package bqfake

/* Outline:
The shadow components implement the corresponding bqiface interface, but instead of wrapping
bigquery objects, they implement fakes that can be constructed incrementally.

*/

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/bigquery/bqiface"
	"github.com/m-lab/etl-gardener/cloud"
	"github.com/m-lab/go/dataset"
	"google.golang.org/api/option"
)

// This defines a Dataset that returns a Table, that returns a canned Metadata.
type Table struct {
	bqiface.Table
	metadata *bigquery.TableMetadata
}

func (tbl Table) Metadata(ctx context.Context) (*bigquery.TableMetadata, error) {
	if tbl.metadata == nil {
		msg := fmt.Sprintf("Error 404: Not found: Table %s, notFound", tbl.FullyQualifiedName())
		return nil, errors.New(msg)
	}
	return tbl.metadata, nil
}

type Dataset struct {
	bqiface.Dataset
	tables map[string]*Table
}

func (ds Dataset) Table(name string) bqiface.Table {
	if ds.tables == nil {
		panic("FakeDataset.tables not initialized")
	}
	t, ok := ds.tables[name]
	if !ok {
		// Empty table that will panic on all calls.
		return &Table{}
	}
	return t
}

func (ds Dataset) AddTable(name string, meta *bigquery.TableMetadata) {
	ds.tables[name] = &Table{Table: ds.Dataset.Table(name), metadata: meta}
}

// creates a Dataset with a dry run client.
func NewDataset(ctx context.Context, project, ds string) dataset.Dataset {
	dryRun, _ := cloud.DryRunClient()
	c, err := bigquery.NewClient(ctx, project, option.WithHTTPClient(dryRun))
	if err != nil {
		panic(err)
	}

	bqClient := bqiface.AdaptClient(c)

	fake := Dataset{Dataset: bqClient.Dataset(ds), tables: make(map[string]*Table)}
	return dataset.Dataset{Dataset: &fake, BqClient: bqClient}
}

/*
func TestCachedMeta(t *testing.T) {
	ctx := context.Background()
	dsExt := newFakeDataset("mlab-testing", "etl")
	{
		meta := bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
		meta.TimePartitioning = &bigquery.TimePartitioning{Expiration: 0 * time.Second}
		dsExt.Dataset.(*FakeDataset).AddTable("DedupTest", &meta)
	}

*/
