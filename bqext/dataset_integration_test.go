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

package bqext_test

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

	"cloud.google.com/go/bigquery"
	"github.com/go-test/deep"
	"github.com/m-lab/go/bqext"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// TestGetTableStats does a live test against a sandbox test table.
func TestGetTableStats(t *testing.T) {
	client, _ := LoggingCloudClient() // Use this for creating the ResponseBody.

	opts := []option.ClientOption{option.WithHTTPClient(client)}
	if os.Getenv("TRAVIS") != "" {
		authOpt := option.WithCredentialsFile("../travis-testing.key")
		opts = append(opts, authOpt)
	}
	tExt, err := bqext.NewDataset("mlab-testing", "go", opts...)
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
	err = json.Unmarshal([]byte(wantTableMetadata), &want)
	if err != nil {
		actual, _ := json.Marshal(stats)
		log.Printf("Actual json:\n%s\n", string(actual))
		t.Fatal(err)
	}

	stats.ETag = "" // Ignore this field in comparison.
	if diff := deep.Equal(*stats, want); diff != nil {
		actual, _ := json.Marshal(stats)
		log.Printf("Actual json:\n%s\n", string(actual))
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
	opts := []option.ClientOption{option.WithHTTPClient(client)}
	if os.Getenv("TRAVIS") != "" {
		authOpt := option.WithCredentialsFile("../travis-testing.key")
		opts = append(opts, authOpt)
	}
	tExt, err := bqext.NewDataset("mlab-testing", "go", opts...)
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

func clientOpts() []option.ClientOption {
	opts := []option.ClientOption{}
	if os.Getenv("TRAVIS") != "" {
		authOpt := option.WithCredentialsFile("../travis-testing.key")
		opts = append(opts, authOpt)
	}
	return opts
}

// TODO - should build the test tables from scratch.  See https://github.com/m-lab/go/issues/8

func TestDedup_Alpha(t *testing.T) {
	start := time.Now() // Later, we will compare partition time to this.

	tExt, err := bqext.NewDataset("mlab-testing", "etl", clientOpts()...)
	if err != nil {
		t.Fatal(err)
	}

	// First check that source table has expected number of rows.
	// TestDedupSrc should have 6 rows, of which 4 should be unique.
	type QR struct {
		NumRows int64
	}
	result := QR{}
	err = tExt.QueryAndParse("select count(test_id) as NumRows from `TestDedupSrc_19990101`", &result)
	if result.NumRows != 6 {
		t.Fatal("Source table has wrong number rows: ", result.NumRows)
	}

	destTable := tExt.BqClient.DatasetInProject("mlab-testing", "etl").Table("TestDedupDest$19990101")
	_, err = tExt.Dedup_Alpha("TestDedupSrc_19990101", "test_id", destTable)
	if err != nil {
		t.Fatal(err)
	}

	pi, err := tExt.GetPartitionInfo("TestDedupDest", "19990101")
	if err != nil {
		t.Fatal(err)
	}
	if pi.CreationTime.Before(start) {
		t.Error("Partition not overwritten??? ", pi.CreationTime)
	}

	err = tExt.QueryAndParse("select count(test_id) as NumRows from `TestDedupDest` where _PARTITIONTIME = timestamp("+`"1999-01-01 00:00:00"`+")", &result)
	if err != nil {
		t.Fatal(err)
	}
	if result.NumRows != 4 {
		t.Error("Destination has wrong number of rows: ", result.NumRows)
	}
}

func TestPartitionInfo(t *testing.T) {
	util, err := bqext.NewDataset("mlab-testing", "etl", clientOpts()...)
	if err != nil {
		t.Fatal(err)
	}

	info, err := util.GetPartitionInfo("TestDedupDest", "19990101")
	if info.PartitionID != "19990101" {
		t.Error("Incorrect PartitionID", info)
	}
}
