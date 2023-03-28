// Package bqfake provides tools to construct fake bigquery datasets, tables, query responses, etc.
// DEPRECATED - please use cloudtest/bqfake instead.
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

// Table implements part of the bqiface.Table interface required for basic testing
// Other parts of the interface should be implemented as needed.
type Table struct {
	bqiface.Table
	ds   Dataset
	name string
	// NOTE: TableType is used to indicate if this is initialized
	metadata *bigquery.TableMetadata
	loader   bqiface.Loader
	err      error
}

// TableOpts defines field options for Table.
type TableOpts struct {
	Dataset
	Name     string
	Metadata *bigquery.TableMetadata
	Loader   bqiface.Loader
	Error    error
}

// NewTable returns a new instance of Table.
func NewTable(opts TableOpts) *Table {
	return &Table{
		ds:       opts.Dataset,
		name:     opts.Name,
		metadata: opts.Metadata,
		loader:   opts.Loader,
		err:      opts.Error,
	}
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

// Update updates the table's `Schema`.
func (tbl Table) Update(ctx context.Context, md bigquery.TableMetadataToUpdate, etag string) (*bigquery.TableMetadata, error) {
	if md.Schema != nil {
		tbl.metadata.Schema = md.Schema
	}
	return tbl.metadata, nil
}

// LoaderFrom returns a bqiface.Loader.
func (tbl Table) LoaderFrom(src bigquery.LoadSource) bqiface.Loader {
	return tbl.loader
}

// Dataset implements part of the bqiface.Dataset interface.
type Dataset struct {
	bqiface.Dataset
	tables   map[string]*Table
	metadata *bqiface.DatasetMetadata
	err      error
}

// NewDataset returns a new instance of Dataset.
func NewDataset(t map[string]*Table, md *bqiface.DatasetMetadata, err error) *Dataset {
	return &Dataset{
		tables:   t,
		metadata: md,
		err:      err,
	}
}

// Metadata implements the bqiface method.
func (ds Dataset) Metadata(ctx context.Context) (*bqiface.DatasetMetadata, error) {
	if ds.metadata != nil {
		return ds.metadata, nil
	}
	return nil, errors.New("invalid dataset metadata")
}

// Create returns an error.
func (ds Dataset) Create(ctx context.Context, md *bqiface.DatasetMetadata) error {
	return ds.err
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
type ClientConfig struct {
	QueryConfig
}

// QueryConfig contains configuration for injecting query results and error values.
type QueryConfig struct {
	ReadErr error
	RowIteratorConfig
}

// RowIteratorConfig contains configuration for injecting row iteration results and error values.
type RowIteratorConfig struct {
	IterErr error
	Rows    []map[string]bigquery.Value
}

// Query implements parts of bqiface.Query to allow some very basic
// unit tests.
type Query struct {
	bqiface.Query
	config QueryConfig
}

// SetQueryConfig is used to set the ReadErr or RowIteratorConfig.
func (q Query) SetQueryConfig(bqiface.QueryConfig) {
	log.Println("SetQueryConfig not implemented")
}

func (q Query) Run(context.Context) (bqiface.Job, error) {
	log.Println("Run not implemented")
	return Job{}, nil
}

func (q Query) Read(context.Context) (bqiface.RowIterator, error) {
	if q.config.ReadErr != nil {
		return nil, q.config.ReadErr
	}
	return &RowIterator{config: q.config.RowIteratorConfig}, nil
}

// Loader implements parts of bqiface.Loader to allow for testing.
type Loader struct {
	bqiface.Loader
	job Job
	err error
}

// NewLoader returns a new instance of Loader.
func NewLoader(job Job, err error) *Loader {
	return &Loader{
		job: job,
		err: err,
	}
}

// Run returns a bqiface.Job and an error.
func (l Loader) Run(ctx context.Context) (bqiface.Job, error) {
	return l.job, l.err
}

// Job implements parts of bqiface.Job to allow some very basic
// unit tests.
type Job struct {
	bqiface.Job
	status *bigquery.JobStatus
	err    error
}

// NewJob returns a new instance of Job.
func NewJob(status *bigquery.JobStatus, err error) *Job {
	return &Job{
		status: status,
		err:    err,
	}
}

// Wait returns a *bigquery.JobStatus and an error.
func (j Job) Wait(context.Context) (*bigquery.JobStatus, error) {
	return j.status, j.err
}

type RowIterator struct {
	bqiface.RowIterator
	config RowIteratorConfig
	index  int
}

func (r *RowIterator) Next(dst interface{}) error {
	// Check config for an error.
	if r.config.IterErr != nil {
		return r.config.IterErr
	}
	// Allow an empty config to return Done.
	if r.index >= len(r.config.Rows) {
		return iterator.Done
	}
	v := dst.(*map[string]bigquery.Value)
	*v = r.config.Rows[r.index]
	r.index++
	return nil
}
