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
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/go-test/deep"
	"github.com/m-lab/etl/bq"
)

func TestGetTableStats(t *testing.T) {
	client, _ := LoggingCloudClient() // Use this for creating the ResponseBody.
	//client := getTableStatsClient()
	util, err := bq.NewTableUtil("mlab-sandbox", "validation", client)
	if err != nil {
		t.Fatal(err)
	}

	stats := util.GetTableStats("dedup")

	// This creates the metadata response we expect.
	var want bigquery.TableMetadata
	json.Unmarshal([]byte(wantString), &want)

	if diff := deep.Equal(stats, want); diff != nil {
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
	util, err := bq.NewTableUtil("mlab-sandbox", "validation", client)
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
		where partition_id = "%s" `, "tableutil_tests", "20171016")
	x, err := util.QueryAndParse(queryString, PartitionInfo{})
	info := x.(PartitionInfo)
	if err != nil {
		t.Error()
	}
	if info.PartitionID != "20171016" {
		t.Error("Incorrect PartitionID")
	}
	log.Printf("%+v\n", info)
}
