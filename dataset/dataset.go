//  Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package x_dataset extends bqiface.Dataset with useful abstractions for simplifying
// interactions with bigquery.
// This package is intended to replace bqext.
// It is currently untested and buggy!!
package dataset

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/googleapis/google-cloud-go-testing/bigquery/bqiface"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Dataset provides extensions to the bigquery Dataset and Dataset
// objects to streamline common actions.
// It encapsulates the Client and Dataset to simplify methods.
type Dataset struct {
	bqiface.Dataset // Exposes Dataset API directly.
	BqClient        bqiface.Client
}

// Errors returned by Dataset functions.
var (
	ErrNilBqClient = errors.New("nil BqClient")
	ErrNilQuery    = errors.New("BqClient.Query failed")
)

// NewDataset creates a Dataset for a project.
// httpClient is used to inject mocks for the bigquery client.
// if httpClient is nil, a suitable default client is used.
// Additional bigquery ClientOptions may be optionally passed as final
//   clientOpts argument.  This is useful for testing credentials.
// NOTE: Caller should close the BqClient when finished.
func NewDataset(ctx context.Context, project, dataset string, clientOpts ...option.ClientOption) (Dataset, error) {
	c, err := bigquery.NewClient(ctx, project, clientOpts...)
	if err != nil {
		return Dataset{}, err
	}
	bqClient := bqiface.AdaptClient(c)

	return Dataset{bqClient.Dataset(dataset), bqClient}, nil
}

func (dsExt *Dataset) queryConfig(query string, dryRun bool) bqiface.QueryConfig {
	qc := bqiface.QueryConfig{}
	qc.Q = query
	qc.DryRun = dryRun
	if strings.HasPrefix(query, "#legacySQL") {
		qc.UseLegacySQL = true
	}
	// Default for unqualified table names in the query.
	qc.DefaultProjectID = dsExt.ProjectID()
	qc.DefaultDatasetID = dsExt.DatasetID()
	return qc
}

// ResultQuery constructs a query with common QueryConfig settings for
// writing results to a table.
// Generally, may need to change WriteDisposition.
func (dsExt *Dataset) ResultQuery(query string, dryRun bool) (bqiface.Query, error) {
	if dsExt.BqClient == nil {
		return nil, ErrNilBqClient
	}
	q := dsExt.BqClient.Query(query)
	if q == nil {
		return nil, ErrNilQuery
	}
	qc := dsExt.queryConfig(query, dryRun)
	q.SetQueryConfig(qc)
	return q, nil
}

///////////////////////////////////////////////////////////////////
// Code to execute a single query and parse single row result.
///////////////////////////////////////////////////////////////////

// QueryAndParse executes a query that should return a single row, with
// all struct fields that match query columns filled in.
// The caller must pass in the *address* of an appropriate struct.
// TODO - extend this to also handle multirow results, by passing
// slice of structs.
func (dsExt *Dataset) QueryAndParse(ctx context.Context, q string, structPtr interface{}) error {
	typeInfo := reflect.ValueOf(structPtr)

	if typeInfo.Type().Kind() != reflect.Ptr {
		return errors.New("Argument should be ptr to struct")
	}
	if reflect.Indirect(typeInfo).Kind() != reflect.Struct {
		return errors.New("Argument should be ptr to struct")
	}

	query, err := dsExt.ResultQuery(q, false)
	if err != nil {
		return err
	}

	it, err := query.Read(ctx)
	if err != nil {
		return err
	}

	// We expect a single result row, so proceed accordingly.
	err = it.Next(structPtr)
	if err != nil {
		return err
	}
	var row map[string]bigquery.Value
	// If there are more rows, then something is wrong.
	err = it.Next(&row)
	if err != iterator.Done {
		return errors.New("multiple row data")
	}
	return nil
}

// PartitionInfo provides basic information about a partition.
type PartitionInfo struct {
	PartitionID  string
	CreationTime time.Time
	LastModified time.Time
}

// GetPartitionInfo provides basic information about a partition.
func (dsExt Dataset) GetPartitionInfo(ctx context.Context, table string, partition string) (PartitionInfo, error) {
	// This uses legacy, because PARTITION_SUMMARY is not supported in standard.
	queryString := fmt.Sprintf(
		`#legacySQL
		SELECT
		  partition_id as PartitionID,
		  msec_to_timestamp(creation_time) AS CreationTime,
		  msec_to_timestamp(last_modified_time) AS LastModified
		FROM
		  [%s$__PARTITIONS_SUMMARY__]
		where partition_id = "%s" `, table, partition)
	pi := PartitionInfo{}

	err := dsExt.QueryAndParse(ctx, queryString, &pi)
	if err != nil {
		log.Println(err, ":", queryString)
		return PartitionInfo{}, err
	}
	return pi, nil
}

// DestQuery constructs a query with common Config settings for
// writing results to a table.
// If dest is nil, then this will create a DryRun query.
// TODO - should disposition be an opts... field instead?
func (dsExt *Dataset) DestQuery(query string, dest bqiface.Table, disposition bigquery.TableWriteDisposition) bqiface.Query {
	qc := dsExt.queryConfig(query, dest == nil)
	qc.Dst = dest
	qc.WriteDisposition = disposition
	qc.AllowLargeResults = true
	// Default for unqualified table names in the query.
	qc.DisableFlattenedResults = true
	q := dsExt.BqClient.Query(query)
	q.SetQueryConfig(qc)
	return q
}
