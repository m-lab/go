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

func assertDataset(ds Dataset) {
	func(cc bqiface.Dataset) {}(ds)
}
