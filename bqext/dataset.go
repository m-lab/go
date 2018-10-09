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

// Package bqext includes generally useful abstractions for simplifying
// interactions with bigquery.
// Production extensions should go here, but test facilities should go
// in a separate bqtest package.
package bqext

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Dataset provides extensions to the bigquery Dataset and Dataset
// objects to streamline common actions.
// It encapsulates the Client and Dataset to simplify methods.
// DEPRECATED - use bqiface version in go/dataset
type Dataset struct {
	*bigquery.Dataset // Exposes Dataset API directly.
	BqClient          *bigquery.Client
}

// NewDataset creates a Dataset for a project.
// httpClient is used to inject mocks for the bigquery client.
// if httpClient is nil, a suitable default client is used.
// Additional bigquery ClientOptions may be optionally passed as final
//   clientOpts argument.  This is useful for testing credentials.
// DEPRECATED - use bqiface version in go/dataset
func NewDataset(project, dataset string, clientOpts ...option.ClientOption) (Dataset, error) {
	ctx := context.Background()
	var bqClient *bigquery.Client
	var err error
	bqClient, err = bigquery.NewClient(ctx, project, clientOpts...)

	if err != nil {
		return Dataset{}, err
	}

	return Dataset{bqClient.Dataset(dataset), bqClient}, nil
}

// ResultQuery constructs a query with common QueryConfig settings for
// writing results to a table.
// Generally, may need to change WriteDisposition.
// DEPRECATED - use bqiface version in go/dataset
func (dsExt *Dataset) ResultQuery(query string, dryRun bool) *bigquery.Query {
	q := dsExt.BqClient.Query(query)
	q.QueryConfig.DryRun = dryRun
	if strings.HasPrefix(query, "#legacySQL") {
		q.QueryConfig.UseLegacySQL = true
	}
	// Default for unqualified table names in the query.
	q.QueryConfig.DefaultProjectID = dsExt.ProjectID
	q.QueryConfig.DefaultDatasetID = dsExt.DatasetID
	return q
}

///////////////////////////////////////////////////////////////////
// Code to execute a single query and parse single row result.
///////////////////////////////////////////////////////////////////

// QueryAndParse executes a query that should return a single row, with
// all struct fields that match query columns filled in.
// The caller must pass in the *address* of an appropriate struct.
// TODO - extend this to also handle multirow results, by passing
// slice of structs.
// DEPRECATED - use bqiface version in go/dataset
func (dsExt *Dataset) QueryAndParse(q string, structPtr interface{}) error {
	typeInfo := reflect.ValueOf(structPtr)

	if typeInfo.Type().Kind() != reflect.Ptr {
		return errors.New("Argument should be ptr to struct")
	}
	if reflect.Indirect(typeInfo).Kind() != reflect.Struct {
		return errors.New("Argument should be ptr to struct")
	}

	query := dsExt.ResultQuery(q, false)
	it, err := query.Read(context.Background())
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
// DEPRECATED - use bqiface version in go/dataset
type PartitionInfo struct {
	PartitionID  string
	CreationTime time.Time
	LastModified time.Time
}

// GetPartitionInfo provides basic information about a partition.
// DEPRECATED - use bqiface version in go/dataset
func (dsExt Dataset) GetPartitionInfo(table string, partition string) (PartitionInfo, error) {
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

	err := dsExt.QueryAndParse(queryString, &pi)
	if err != nil {
		log.Println(err, ":", queryString)
		return PartitionInfo{}, err
	}
	return pi, nil
}

// DestinationQuery constructs a query with common Config settings for
// writing results to a table.
// If dest is nil, then this will create a DryRun query.
// TODO - should disposition be an opts... field instead?
// DEPRECATED - use bqiface version in go/dataset
func (dsExt *Dataset) DestQuery(query string, dest *bigquery.Table, disposition bigquery.TableWriteDisposition) *bigquery.Query {
	q := dsExt.BqClient.Query(query)
	if dest != nil {
		q.QueryConfig.Dst = dest
	} else {
		q.QueryConfig.DryRun = true
	}
	q.QueryConfig.WriteDisposition = disposition
	q.QueryConfig.AllowLargeResults = true
	// Default for unqualified table names in the query.
	q.QueryConfig.DefaultProjectID = dsExt.ProjectID
	q.QueryConfig.DefaultDatasetID = dsExt.DatasetID
	q.QueryConfig.DisableFlattenedResults = true
	return q
}

// ExecDestQuery executes a destination or dryrun query, and returns status or error.
func (dsExt *Dataset) ExecDestQuery(q *bigquery.Query) (*bigquery.JobStatus, error) {
	if q.QueryConfig.Dst == nil && q.QueryConfig.DryRun == false {
		return nil, errors.New("query must be a destination or dry run")
	}
	job, err := q.Run(context.Background())
	if err != nil {
		return nil, err
	}
	log.Println("JobID:", job.ID())
	status, err := job.Wait(context.Background())
	if err != nil {
		return status, err
	}
	return status, nil
}

///////////////////////////////////////////////////////////////////
// Specific queries.
///////////////////////////////////////////////////////////////////

// TODO Obsolete - delete.
var dedupTemplate = `
	#standardSQL
	# Delete all duplicate rows, partitioning by specified field
	SELECT * except (row_number)
	FROM (
	  SELECT *, ROW_NUMBER() OVER (PARTITION BY %s) row_number
	  FROM ` + "`%s`" + `)
	WHERE row_number = 1`

// This template expects to be executed on a table containing a single day's data, such
// as measurement-lab:batch.ndt_20170601.
//
// Some tests are collected as both uncompressed and compressed files. In some historical
// archives (June 2017), files for a single test appear in different tar files, which
// results in duplicate rows.
// This query strips the gz, finds duplicates, and chooses the best row -  prefering gzipped
// files, and prefering later parse_time.
var dedupTemplate2 = `
	#standardSQL
	# Delete all duplicate rows based on test_id, preferring gz over non-gz, later parse_time
	SELECT * except (row_number, gz, stripped_id)
    from (
		select *, ROW_NUMBER() OVER (PARTITION BY stripped_id order by gz DESC, parse_time DESC) row_number
        FROM (
	        SELECT *, regexp_replace(test_id, ".gz$", "") as stripped_id, regexp_extract(test_id, ".*(.gz)$") as gz
	        FROM ` + "`%s`" + `
        )
    )
	WHERE row_number = 1`

// Dedup_Alpha executes a query that dedups and writes to destination partition.
// This function is alpha status.  The interface may change without notice
// or major version number change.
//
// `src` is relative to the project:dataset of dsExt.
// `dedupOn` names the field to be used for dedupping.  (Currently ignored)
// `destTable` specifies the table to write to, typically created with
//   dsExt.BqClient.DatasetInProject(...).Table(...)
//
// NOTE: If destination table is partitioned, destTable MUST include the partition
// suffix to avoid accidentally overwriting the entire table.
// DEPRECATED - use bqiface version in go/dataset
func (dsExt *Dataset) Dedup_Alpha(src string, dedupOn string, destTable *bigquery.Table) (*bigquery.JobStatus, error) {
	if !strings.Contains(destTable.TableID, "$") {
		meta, err := destTable.Metadata(context.Background())
		if err == nil && meta.TimePartitioning != nil {
			log.Println(err)
			return nil, errors.New("Destination table must specify partition")
		}
	}

	//log.Printf("Removing dups (of %s) and writing to %s.%s\n", dedupOn, destTable.DatasetID, destTable.TableID)
	// queryString := fmt.Sprintf(dedupTemplate, dedupOn, src)
	log.Println("Warning - ignoring dedup field:", dedupOn)
	log.Printf("Removing dups (of test_id) and writing to %s.%s\n", destTable.DatasetID, destTable.TableID)
	queryString := fmt.Sprintf(dedupTemplate2, src)
	query := dsExt.DestQuery(queryString, destTable, bigquery.WriteTruncate)
	return dsExt.ExecDestQuery(query)
}
