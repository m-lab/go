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
		return errors.New("TODO: should return a table exists error")
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

type Query struct {
	bqiface.Query
	//JobIDConfig() *bigquery.JobIDConfig
	//SetQueryConfig(QueryConfig)
	//Run(context.Context) (Job, error)
	//Read(context.Context) (RowIterator, error)
}

func (q Query) SetQueryConfig(bqiface.QueryConfig) {
	log.Println("SetQueryConfig not implemented")
}

func (q Query) Run(context.Context) (bqiface.Job, error) {
	log.Println("Run not implemented")
	return Job{}, nil
}

func (q Query) Read(context.Context) (bqiface.RowIterator, error) {
	log.Println("Read not implemented")
	return RowIterator{}, nil
}

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
