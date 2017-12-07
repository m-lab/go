// +build integration

package bqutil_test

// This file contains integration tests, which should be run
// infrequently, with appropriate credentials.  These tests depend
// on the state of our bigquery tables, so they may start failing
// if the tables are changed.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/go-test/deep"
	"github.com/m-lab/go/bqutil"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//var wantStringTest1 = `{"Name":"","Description":"","Schema":[{"Name":"test_id","Description":"","Repeated":false,"Required":false,"Type":"STRING","Schema":null}],"ViewQuery":"","UseLegacySQL":false,"UseStandardSQL":false,"TimePartitioning":{"Expiration":0},"ExpirationTime":"0001-01-01T00:00:00Z","Labels":null,"ExternalDataConfig":null,"FullID":"mlab-testing:go.TestGetTableStats","Type":"TABLE","CreationTime":"2017-12-06T12:19:16.218-05:00","LastModifiedTime":"2017-12-06T12:19:16.218-05:00","NumBytes":7,"NumRows":1,"StreamingBuffer":null,"ETag":"\"cX5UmbB_R-S07ii743IKGH9YCYM/MTUxMjU4MDc1NjIxOA\""}`

var wantTableMetadata = `{"Name":"","Description":"","Schema":[{"Name":"test_id","Description":"","Repeated":false,"Required":false,"Type":"STRING","Schema":null}],"ViewQuery":"","UseLegacySQL":false,"UseStandardSQL":false,"TimePartitioning":{"Expiration":0},"ExpirationTime":"0001-01-01T00:00:00Z","Labels":null,"ExternalDataConfig":null,"FullID":"mlab-testing:go.TestGetTableStats","Type":"TABLE","CreationTime":"2017-12-06T12:19:16.218-05:00","LastModifiedTime":"2017-12-06T12:19:16.218-05:00","NumBytes":7,"NumRows":1,"StreamingBuffer":null,"ETag":"\"cX5UmbB_R-S07ii743IKGH9YCYM/MTUxMjU4MDc1NjIxOA\""}`

// TestGetTableStats does a live test against a sandbox test table.
func TestGetTableStats(t *testing.T) {
	client, _ := LoggingCloudClient() // Use this for creating the ResponseBody.

	opts := []option.ClientOption{option.WithHTTPClient(client)}
	if os.Getenv("TRAVIS") != "" {
		authOpt := option.WithCredentialsFile("../travis-testing.key")
		opts = append(opts, authOpt)
	}
	util, err := bqutil.NewTableUtil("mlab-testing", "go", opts...)
	if err != nil {
		t.Fatal(err)
	}

	table := util.Dataset.Table("TestGetTableStats")
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

	if diff := deep.Equal(stats, want); diff != nil {
		actual, _ := json.Marshal(stats)
		log.Printf("Actual json:\n%%s\n", string(actual))
		t.Error(diff)
	}
}

// PartitionInfo provides basic information about a partition.
type PartitionInfo struct {
	PartitionID  string    `qfield:"partition_id"`
	CreationTime time.Time `qfield:"created"`
	LastModified time.Time `qfield:"last_modified"`
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
	util, err := bqutil.NewTableUtil("mlab-testing", "go", opts...)
	if err != nil {
		log.Fatal(err)
	}

	// This uses legacy, because PARTITION_SUMMARY is not supported in standard.
	queryString := fmt.Sprintf(
		`#legacySQL
		SELECT
		  partition_id,
		  msec_to_timestamp(creation_time) AS created,
		  msec_to_timestamp(last_modified_time) AS last_modified
		FROM
		  [%s$__PARTITIONS_SUMMARY__]
		where partition_id = "%s" `, "TestQueryAndParse", "20170101")
	x, err := util.QueryAndParse(queryString, PartitionInfo{})
	info := x.(PartitionInfo)
	if err != nil {
		t.Error()
	}
	if info.PartitionID != "20170101" {
		t.Error("Incorrect PartitionID")
	}
	log.Printf("%+v\n", info)
}
