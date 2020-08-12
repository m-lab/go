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

// +build integration

package bqx_test

// This file contains integration tests, which should be run
// infrequently, with appropriate credentials.  These tests depend
// on the state of our bigquery tables, so they may start failing
// if the tables are changed.

// TODO (issue #8) tests that use bq tables should create them from scratch.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-test/deep"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"cloud.google.com/go/bigquery"

	"github.com/m-lab/go/cloud/bqx"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Looks like the Location field defaults to US.  Might be different if run in different locale.
const wantTableMetadata2 = `{"Location":"US","Schema":[{"Name":"test_id","Type":"STRING"}],"TimePartitioning":{},"FullID":"mlab-testing:go.TestGetTableStats","Type":"TABLE","CreationTime":"2017-12-06T12:19:16.218-05:00","LastModifiedTime":"2017-12-06T12:19:16.218-05:00","NumBytes":7,"NumLongTermBytes":7,"NumRows":1}`

func clientOpts() []option.ClientOption {
	opts := []option.ClientOption{}
	if os.Getenv("TRAVIS") != "" {
		authOpt := option.WithCredentialsFile("../travis-testing.key")
		opts = append(opts, authOpt)
	}
	return opts
}

// TestGetTableStats does a live test against a sandbox test table.
func TestGetTableStats(t *testing.T) {
	client, _ := LoggingCloudClient() // Use this for creating the ResponseBody.

	opts := clientOpts()
	opts = append(opts, option.WithHTTPClient(client))

	tExt, err := bqx.NewDataset("mlab-testing", "go", opts...)
	if err != nil {
		t.Fatal(err)
	}

	table := tExt.Table("TestGetTableStats")
	ctx := context.Background()
	stats, err := table.Metadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// This creates the metadata response we expect.
	var want bigquery.TableMetadata
	err = json.Unmarshal([]byte(wantTableMetadata2), &want)
	if err != nil {
		actual, _ := json.Marshal(stats)
		log.Printf("Actual json:\n%s\n", string(actual))
		t.Fatal(err)
	}

	stats.ETag = "" // Ignore this field in comparison.
	if diff := deep.Equal(*stats, want); diff != nil {
		actual, _ := json.Marshal(stats)
		t.Logf("Actual json:\n%s\n", string(actual))
		t.Error(diff)
	}
}

// partitionInfo provides basic information about a partition.
// Note that a similar struct is defined in dataset.go, but this
// one is used for testing the QueryAndParse method.
type partitionInfo struct {
	PartitionID  string
	CreationTime time.Time
	LastModified time.Time
}

func TestQueryAndParse(t *testing.T) {
	// This logs all the requests and responses, for debugging purposes.
	// Turns out this test causes three http requests to the backend.
	client, _ := LoggingCloudClient() // Use this for creating the ResponseBody.
	opts := clientOpts()
	opts = append(opts, option.WithHTTPClient(client))

	tExt, err := bqx.NewDataset("mlab-testing", "go", opts...)
	if err != nil {
		t.Fatal(err)
	}

	// This uses legacy, because PARTITION_SUMMARY is not supported in standard.
	queryString := fmt.Sprintf(
		`#legacySQL
		SELECT
		  partition_id as PartitionID,
		  msec_to_timestamp(creation_time) AS created,
		  msec_to_timestamp(last_modified_time) AS last_modified
		FROM
		  [%s$__PARTITIONS_SUMMARY__]
		where partition_id = "%s" `, "TestQueryAndParse", "20170101")
	pi := partitionInfo{}

	// Should be simple struct...
	err = tExt.QueryAndParse(queryString, []partitionInfo{})
	if err == nil {
		t.Error("Should produce error on slice input")
	}
	// Non-pointer...
	err = tExt.QueryAndParse(queryString, pi)
	if err == nil {
		t.Error("Should produce error on slice input")
	}

	// Correct behavior.
	err = tExt.QueryAndParse(queryString, &pi)
	if err != nil {
		t.Fatal(err)
	}
	if pi.PartitionID != "20170101" {
		t.Error("Incorrect PartitionID")
	}
}

// This test is always failing.
func TestPartitionInfo(t *testing.T) {
	client, _ := LoggingCloudClient()
	opts := clientOpts()
	opts = append(opts, option.WithHTTPClient(client))
	util, err := bqx.NewDataset("mlab-testing", "go", opts...)
	if err != nil {
		t.Fatal(err)
	}

	info, err := util.GetPartitionInfo("TestDedupDest", "19990101")
	if info.PartitionID != "19990101" {
		t.Error("Incorrect PartitionID", info)
	}
}
