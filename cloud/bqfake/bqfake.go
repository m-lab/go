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
	"google.golang.org/api/iterator"
)

// Table wraps a bigquery.Table, overriding parts of the bqiface.Table interface required for basic testing
// Other parts of the interface should be overridden as needed.
type Table struct {
	bqiface.Table
	ds Dataset
	// NOTE: TableType is used to indicate if this is initialized
	metadata *bigquery.TableMetadata
}

// Metadata implements the bqiface method.
func (tbl Table) Metadata(ctx context.Context) (*bigquery.TableMetadata, error) {
	if tbl.metadata == nil {
		return nil, errors.New("Table object incorrectly initialized")
	}
	if tbl.metadata.Type == "" {
		log.Printf("Metadata %p %v\n", tbl.metadata, tbl.metadata)
		msg := fmt.Sprintf("Error 404: Not found: Table %s, notFound", tbl.FullyQualifiedName())
		return nil, errors.New(msg)
	}
	log.Printf("Metadata %p %v\n", tbl.metadata, tbl.metadata)
	return tbl.metadata, nil
}

// Create implements the bqiface method.
func (tbl Table) Create(ctx context.Context, meta *bigquery.TableMetadata) error {
	log.Println("Create", meta)
	if tbl.metadata == nil {
		return errors.New("Table object incorrectly initialized")
	}
	if tbl.metadata.Type != "" {
		return errors.New("TODO: should return a table exists error")
	}
	*tbl.metadata = *meta
	if tbl.metadata.Type == "" {
		tbl.metadata.Type = "TABLE"
	}
	log.Printf("Metadata %p %v\n", tbl.metadata, tbl.metadata)
	return nil
}

// Dataset wraps a concrete bigquery.Dataset, overriding parts of the bqiface.Dataset
// interface to allow some basic unit tests.
type Dataset struct {
	bqiface.Dataset
	tables map[string]*Table
}

// Table implements the bqiface method.
func (ds Dataset) Table(name string) bqiface.Table {
	t, ok := ds.tables[name]
	if !ok {
		t = &Table{ds: ds, metadata: &bigquery.TableMetadata{}}
		t.Table = ds.Dataset.Table(name)
		// TODO is this better? t = &Table{ds: ds.Dataset, name: name, metadata: &pm}
		ds.tables[name] = t
	}
	return t
}

// Query wraps a concrete bigquery.Query, overriding parts of bqiface.Query to allow
// some very basic unit tests.
type Query struct {
	bqiface.Query
	//JobIDConfig() *bigquery.JobIDConfig
}

func (q Query) Read(context.Context) (bqiface.RowIterator, error) {
	log.Println("Read not implemented")
	return RowIterator{}, nil
}

func (q Query) Run(context.Context) (bqiface.Job, error) {
	log.Println("Run not implemented")
	return Job{}, nil
}

// Job implements parts of bqiface.Job to allow some very basic
// unit tests.
type Job struct {
	bqiface.Job

	//ID() string
	//Location() string
	//Config() (bigquery.JobConfig, error)
	//Status(context.Context) (*bigquery.JobStatus, error)
	//LastStatus() *bigquery.JobStatus
	//Cancel(context.Context) error
	//Read(context.Context) (RowIterator, error)
}

func (j Job) Wait(context.Context) (*bigquery.JobStatus, error) {
	log.Println("Wait not implemented")
	return nil, nil
}

type RowIterator struct {
	bqiface.RowIterator
	//SetStartIndex(uint64)
	//Schema() bigquery.Schema
	//TotalRows() uint64
	//Next(interface{}) error
	//PageInfo() *iterator.PageInfo
}

func (r RowIterator) Next(interface{}) error {
	log.Println("Next not implemented")
	return iterator.Done
}
