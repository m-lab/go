// Package bqfake provides tools to construct fake bigquery datasets, tables, query responses, etc.
// DEPRECATED - please use cloudtest/bqfake instead!
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
	"github.com/googleapis/google-cloud-go-testing/bigquery/bqiface"
	"google.golang.org/api/iterator"
)

var (
	// ErrTypeAssertionFailed may be returned by the row iterator when the target type is incorrect.
	ErrTypeAssertionFailed = errors.New("type assertion failed")
)

// Table implements part of the bqiface.Table interface required for basic testing
// Other parts of the interface should be implemented as needed.
type Table struct {
	bqiface.Table
	ds   Dataset
	name string
	// NOTE: TableType is used to indicate if this is initialized
	metadata *bigquery.TableMetadata
}

// ProjectID implements the bqiface method.
func (tbl Table) ProjectID() string {
	return tbl.ds.ProjectID()
}

// DatasetID implements the bqiface method.
func (tbl Table) DatasetID() string {
	return tbl.ds.DatasetID()
}

// TableID implements the bqiface method.
func (tbl Table) TableID() string {
	return tbl.name
}

// FullyQualifiedName implements the bqiface method.
func (tbl Table) FullyQualifiedName() string {
	return tbl.ProjectID() + ":" + tbl.DatasetID() + "." + tbl.name
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
// DEPRECATED - use cloudtest/bqfake
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

// Dataset implements part of the bqiface.Dataset interface.
type Dataset struct {
	bqiface.Dataset
	tables map[string]*Table
}

// Table implements the bqiface method.
func (ds Dataset) Table(name string) bqiface.Table {
	t, ok := ds.tables[name]
	if !ok {
		t = &Table{ds: ds, name: name, metadata: &bigquery.TableMetadata{}}
		// TODO is this better? t = &Table{ds: ds.Dataset, name: name, metadata: &pm}
		ds.tables[name] = t
	}
	return t
}

// ClientConfig contains configuration for injecting result and error values.
type ClientConfig[Row any] struct {
	QueryConfig[Row]
}

// QueryConfig contains configuration for injecting query results and error values.
type QueryConfig[Row any] struct {
	ReadErr error
	RowIteratorConfig[Row]
}

// RowIteratorConfig contains configuration for injecting row iteration results and error values.
type RowIteratorConfig[Row any] struct {
	IterErr error
	Rows    []Row
}

// Query implements parts of bqiface.Query to allow some very basic
// unit tests.
type Query[Row any] struct {
	bqiface.Query
	config QueryConfig[Row]
}

// SetQueryConfig is used to set the ReadErr or RowIteratorConfig.
func (q Query[Row]) SetQueryConfig(bqiface.QueryConfig) {
	log.Println("SetQueryConfig not implemented")
}

func (q Query[Row]) Run(context.Context) (bqiface.Job, error) {
	log.Println("Run not implemented")
	return Job{}, nil
}

func (q Query[Row]) Read(context.Context) (bqiface.RowIterator, error) {
	if q.config.ReadErr != nil {
		return nil, q.config.ReadErr
	}
	return &RowIterator[Row]{config: q.config.RowIteratorConfig}, nil
}

// Job implements parts of bqiface.Job to allow some very basic
// unit tests.
type Job struct {
	bqiface.Job
}

func (j Job) Wait(context.Context) (*bigquery.JobStatus, error) {
	log.Println("Wait not implemented")
	return nil, nil
}

type RowIterator[Row any] struct {
	bqiface.RowIterator
	config RowIteratorConfig[Row]
	index  int
}

func (r *RowIterator[Row]) TotalRows() uint64 {
	return uint64(len(r.config.Rows))
}

func (r *RowIterator[Row]) Next(dst interface{}) error {
	// Check config for an error.
	if r.config.IterErr != nil {
		return r.config.IterErr
	}
	// Allow an empty config to return Done.
	if r.index >= len(r.config.Rows) {
		return iterator.Done
	}
	v, ok := dst.(*Row)
	if !ok {
		return ErrTypeAssertionFailed
	}
	*v = r.config.Rows[r.index]
	r.index++
	return nil
}
