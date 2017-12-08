package bqutil_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"

	"google.golang.org/api/option"

	"cloud.google.com/go/bigquery"
	"github.com/go-test/deep"
	"github.com/m-lab/go/bqutil"
	"github.com/m-lab/go/cloudtest"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
)

type nopCloser struct {
	io.Reader
}

func (nc nopCloser) Close() error { return nil }

// LoggingCloudClient is used to get the ResponseBody used in the test client below.
func LoggingCloudClient() (*http.Client, error) {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/bigquery")
	if err != nil {
		return nil, err
	}
	return cloudtest.LoggingClient(client)
}

// This was captured using LoggingClient.
var injectedResponseBody = `
{ "kind": "bigquery#table", "etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/MTUxMjU4MDc1NjIxOA\"",
	"id": "mlab-testing:go.TestGetTableStats",
	"selfLink": "https://www.googleapis.com/bigquery/v2/projects/mlab-testing/datasets/go/tables/TestGetTableStats",
	"tableReference": { "projectId": "mlab-testing", "datasetId": "go", "tableId": "TestGetTableStats" },
	"schema": { "fields": [
	  { "name": "test_id", "type": "STRING", "mode": "NULLABLE" } ] },
	"timePartitioning": { "type": "DAY" }, "numBytes": "7", "numLongTermBytes": "0", "numRows": "1",
	"creationTime": "1512580756218", "lastModifiedTime": "1512580756218", "type": "TABLE", "location": "US" }`

// This is the expected TableMetadata, json encoded
var wantTableMetadata2 = `{"Name":"","Description":"","Schema":[{"Name":"test_id","Description":"","Repeated":false,"Required":false,"Type":"STRING","Schema":null}],"ViewQuery":"","UseLegacySQL":false,"UseStandardSQL":false,"TimePartitioning":{"Expiration":0},"ExpirationTime":"0001-01-01T00:00:00Z","Labels":null,"ExternalDataConfig":null,"FullID":"mlab-testing:go.TestGetTableStats","Type":"TABLE","CreationTime":"2017-12-06T12:19:16.218-05:00","LastModifiedTime":"2017-12-06T12:19:16.218-05:00","NumBytes":7,"NumRows":1,"StreamingBuffer":null,"ETag":"\"cX5UmbB_R-S07ii743IKGH9YCYM/MTUxMjU4MDc1NjIxOA\""}`

// Client that returns canned response from metadata request.
// Pretty ugly implementation.  Will need to improve this before using
// the strategy more widely.  Possibly should use one of the go-vcr tools.
func getTableStatsClient() *http.Client {
	c := make(chan *http.Response, 10)
	client := cloudtest.NewChannelClient(c)

	resp := &http.Response{}
	resp.StatusCode = http.StatusOK
	resp.Status = "OK"
	resp.Body = nopCloser{bytes.NewReader([]byte(injectedResponseBody))}
	c <- resp

	return client
}

func TestGetTableStatsMock(t *testing.T) {
	//client, _ := LoggingCloudClient() // Use this for creating the ResponseBody.
	//opts := []option.ClientOption{option.WithHTTPClient(client)}
	opts := []option.ClientOption{option.WithHTTPClient(getTableStatsClient())}
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
	json.Unmarshal([]byte(wantTableMetadata2), &want)

	if diff := deep.Equal(*stats, want); diff != nil {
		t.Error(diff)
	}
}
