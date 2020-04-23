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

package bqx_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/go-test/deep"

	"cloud.google.com/go/bigquery"

	"github.com/m-lab/go/cloud/bqx"
	"github.com/m-lab/go/cloudtest"
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

// These two are generated by the TestGetTableStats integration test.
const injectedResponseBody = `
{ "kind": "bigquery#table", "etag": "\"cX5UmbB_R-S07ii743IKGH9YCYM/MTUxMjU4MDc1NjIxOA\"",
	"id": "mlab-testing:go.TestGetTableStats",
	"selfLink": "https://www.googleapis.com/bigquery/v2/projects/mlab-testing/datasets/go/tables/TestGetTableStats",
	"tableReference": { "projectId": "mlab-testing", "datasetId": "go", "tableId": "TestGetTableStats" },
	"schema": { "fields": [
	  { "name": "test_id", "type": "STRING", "mode": "NULLABLE" } ] },
	"timePartitioning": { "type": "DAY" }, "numBytes": "7", "numLongTermBytes": "7", "numRows": "1",
	"creationTime": "1512580756218", "lastModifiedTime": "1512580756218", "type": "TABLE", "location": "" }`

// This is the expected TableMetadata, json encoded, with the ETag deleted.
const wantTableMetadata = `{"Schema":[{"Name":"test_id","Type":"STRING"}],"TimePartitioning":{},"FullID":"mlab-testing:go.TestGetTableStats","Type":"TABLE","CreationTime":"2017-12-06T12:19:16.218-05:00","LastModifiedTime":"2017-12-06T12:19:16.218-05:00","NumBytes":7,"NumLongTermBytes":7,"NumRows":1}`

// Client that returns canned response from metadata request.
// Pretty ugly implementation.  Will need to improve this before using
// the strategy more widely.  Possibly should use one of the go-vcr tools.
func getOKClient() *http.Client {
	c := make(chan *http.Response, 10)
	client := cloudtest.NewChannelClient(c)

	resp := &http.Response{}
	resp.StatusCode = http.StatusOK
	resp.Status = "OK"
	resp.Body = nopCloser{bytes.NewReader([]byte(injectedResponseBody))}
	c <- resp

	return client
}

// This tests GetTableStats, by using a captured response body
// and comparing against actual stats from a table in mlab-testing.
// That test runs as an integration test, and the logged response body
// can be found it that test's output.
func TestGetTableStatsMock(t *testing.T) {
	opts := []option.ClientOption{option.WithHTTPClient(getOKClient())}
	dsExt, err := bqx.NewDataset("mock", "mock", opts...)
	if err != nil {
		t.Fatal(err)
	}

	table := dsExt.Table("TestGetTableStats")
	ctx := context.Background()
	stats, err := table.Metadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// This creates the metadata response we expect.
	var want bigquery.TableMetadata
	json.Unmarshal([]byte(wantTableMetadata), &want)

	stats.ETag = "" // Ignore this field for comparison.
	// Location was introduced after bigquery v1.3.0, and may result in flaky tests
	// if included in the next release.
	// stats.Location = ""

	if diff := deep.Equal(*stats, want); diff != nil {
		t.Error(diff)
	}
}

// This test only check very basic stuff.  Intended mostly just to
// improve coverage metrics.
func TestResultQuery(t *testing.T) {
	// Create a dummy client.
	opts := []option.ClientOption{option.WithHTTPClient(getOKClient())}
	dsExt, err := bqx.NewDataset("mock", "mock", opts...)
	if err != nil {
		t.Fatal(err)
	}

	q := dsExt.ResultQuery("query string", true)
	qc := q.QueryConfig
	if !qc.DryRun {
		t.Error("DryRun should be set.")
	}

	q = dsExt.ResultQuery("query string", false)
	qc = q.QueryConfig
	if qc.DryRun {
		t.Error("DryRun should be false.")
	}
}

// This test only check very basic stuff.  Intended mostly just to
// improve coverage metrics.
func TestDestQuery(t *testing.T) {
	// Create a dummy client.
	opts := []option.ClientOption{option.WithHTTPClient(getOKClient())}
	dsExt, err := bqx.NewDataset("mock", "mock", opts...)
	if err != nil {
		t.Fatal(err)
	}

	q := dsExt.DestQuery("query string", nil, bigquery.WriteEmpty)
	qc := q.QueryConfig
	if qc.Dst != nil {
		t.Error("Destination should be nil.")
	}
	if !qc.DryRun {
		t.Error("DryRun should be set.")
	}

	q = dsExt.DestQuery("query string", dsExt.Table("foobar"), bigquery.WriteEmpty)
	qc = q.QueryConfig
	if qc.Dst.TableID != "foobar" {
		t.Error("Destination should be foobar.")
	}
	if qc.DryRun {
		t.Error("DryRun should be false.")
	}
}
