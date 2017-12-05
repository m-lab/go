package bqutil_test

// Pretty ugly implementation.  Will need to improve this before using
// the strategy more widely.  Possibly should use one of the go-vcr tools.

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

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

// This is captured using LoggingClient.
var ResponseBody = "{\n \"kind\": \"bigquery#table\",\n \"etag\": \"\\\"cX5UmbB_R-S07ii743IKGH9YCYM/MTQ5OTQ0MTc2NTEwOA\\\"\",\n \"id\": \"mlab-sandbox:validation.dedup\",\n \"selfLink\": \"https://www.googleapis.com/bigquery/v2/projects/mlab-sandbox/datasets/validation/tables/dedup\",\n \"tableReference\": {\n  \"projectId\": \"mlab-sandbox\",\n  \"datasetId\": \"validation\",\n  \"tableId\": \"dedup\"\n },\n \"schema\": {\n  \"fields\": [\n   {\n    \"name\": \"day\",\n    \"type\": \"TIMESTAMP\",\n    \"mode\": \"NULLABLE\"\n   },\n   {\n    \"name\": \"total\",\n    \"type\": \"INTEGER\",\n    \"mode\": \"NULLABLE\"\n   },\n   {\n    \"name\": \"dist\",\n    \"type\": \"INTEGER\",\n    \"mode\": \"NULLABLE\"\n   }\n  ]\n },\n \"numBytes\": \"888\",\n \"numLongTermBytes\": \"888\",\n \"numRows\": \"37\",\n \"creationTime\": \"1499291057539\",\n \"lastModifiedTime\": \"1499441765108\",\n \"type\": \"TABLE\",\n \"location\": \"US\"\n}\n"

// This is the expected TableMetadata, json encoded.
var wantString = `{"Name":"","Description":"","Schema":[{"Name":"day","Description":"","Repeated":false,"Required":false,"Type":"TIMESTAMP","Schema":null},{"Name":"total","Description":"","Repeated":false,"Required":false,"Type":"INTEGER","Schema":null},{"Name":"dist","Description":"","Repeated":false,"Required":false,"Type":"INTEGER","Schema":null}],"ViewQuery":"","UseLegacySQL":false,"UseStandardSQL":false,"TimePartitioning":null,"ExpirationTime":"0001-01-01T00:00:00Z","Labels":null,"ExternalDataConfig":null,"FullID":"mlab-sandbox:validation.dedup","Type":"TABLE","CreationTime":"2017-07-05T17:44:17.539-04:00","LastModifiedTime":"2017-07-07T11:36:05.108-04:00","NumBytes":888,"NumRows":37,"StreamingBuffer":null,"ETag":"\"cX5UmbB_R-S07ii743IKGH9YCYM/MTQ5OTQ0MTc2NTEwOA\""}`

// Client that returns canned response from metadata request.
func getTableStatsClient() *http.Client {
	c := make(chan *http.Response, 10)
	client := cloudtest.NewChannelClient(c)

	resp := &http.Response{}
	resp.StatusCode = http.StatusOK
	resp.Status = "OK"
	resp.Body = nopCloser{bytes.NewReader([]byte(ResponseBody))}
	c <- resp

	return client
}

func LoggingCloudClient() (*http.Client, error) {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/bigquery")
	if err != nil {
		return nil, err
	}
	return cloudtest.LoggingClient(client)
}

func TestGetTableStatsMock(t *testing.T) {
	// client, _ := LoggingCloudClient() // Use this for creating the ResponseBody.
	client := getTableStatsClient()
	util, err := bqutil.NewTableUtil("mlab-sandbox", "validation", client)
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
