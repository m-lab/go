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
	"log"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/bigquery/bqiface"
)

// Table implements part of the bqiface.Table interface.
type Table struct {
	bqiface.Table
	ds       bqiface.Dataset
	name     string
	metadata **bigquery.TableMetadata
}

// ProjectID implements the bqiface method.
func (tbl Table) ProjectID() string {
	return tbl.ds.ProjectID()
}

// DatasetID implements the bqiface method.
func (tbl Table) DatasetID() string {
	return tbl.ds.DatasetID()
}

// FullyQualifiedName implements the bqiface method.
func (tbl Table) FullyQualifiedName() string {
	return tbl.ProjectID() + ":" + tbl.DatasetID() + "." + tbl.name
}

// Metadata implements the bqiface method.
func (tbl Table) Metadata(ctx context.Context) (*bigquery.TableMetadata, error) {
	if *tbl.metadata == nil {
		log.Printf("Metadata %p %p %v\n", tbl.metadata, *tbl.metadata, *tbl.metadata)
		msg := fmt.Sprintf("Error 404: Not found: Table %s, notFound", tbl.FullyQualifiedName())
		return nil, errors.New(msg)
	}
	log.Printf("Metadata %p %p %v\n", tbl.metadata, *tbl.metadata, *tbl.metadata)
	return *tbl.metadata, nil
}

// Create implements the bqiface method.
func (tbl Table) Create(ctx context.Context, meta *bigquery.TableMetadata) error {
	log.Println("Create", meta)
	if *tbl.metadata != nil {
		return errors.New("wrong error")
	}
	t := &bigquery.TableMetadata{}
	*tbl.metadata = t
	**tbl.metadata = *meta
	log.Printf("Metadata %p %p %v\n", tbl.metadata, *tbl.metadata, *tbl.metadata)
	return nil
}

// Dataset implements part of the bqiface.Dataset interface.
type Dataset struct {
	bqiface.Dataset
	tables map[string]*Table
}

// Table implements the bqiface method.
func (ds Dataset) Table(name string) bqiface.Table {
	if ds.tables == nil {
		panic("FakeDataset.tables not initialized")
	}
	t, ok := ds.tables[name]
	if !ok {
		var pm *bigquery.TableMetadata
		t = &Table{ds: ds, name: name, metadata: &pm}
		// TODO is this better? t = &Table{ds: ds.Dataset, name: name, metadata: &pm}
		ds.tables[name] = t
	}
	return t
}

// This fails to compile if Dataset does not satisfy the interface.
func assertDataset(ds Dataset) {
	func(cc bqiface.Dataset) {}(ds)
}
