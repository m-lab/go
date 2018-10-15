package bqfake

// TODO: Implement context expiration checking.

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/bigquery/bqiface"
	"google.golang.org/api/option"
)

// *******************************************************************
// DryRunClient, that just returns status ok and empty body
// *******************************************************************

// CountingTransport counts calls, and returns OK and empty body.
// `count` field should only be accessed using atomic.Foobar
type CountingTransport struct {
	count int32
	reqs  []*http.Request
}

// Count returns the client call count.
func (ct *CountingTransport) Count() int32 {
	return atomic.LoadInt32(&ct.count)
}

// Requests returns the entire req from the last request
func (ct *CountingTransport) Requests() []*http.Request {
	return ct.reqs
}

// RoundTrip implements the RoundTripper interface, logging the
// request, and the response body, (which may be json).
func (ct *CountingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(&ct.count, 1)

	// Create an empty response with StatusOK
	resp := &http.Response{}
	resp.StatusCode = http.StatusOK
	resp.Body = &nopCloser{strings.NewReader("")}

	// Save the request for testing.
	ct.reqs = append(ct.reqs, req)

	return resp, nil
}

type nopCloser struct {
	io.Reader
}

func (nc *nopCloser) Close() error { return nil }

// DryRunClient returns a client that just counts calls.
func DryRunClient() (*http.Client, *CountingTransport) {
	client := &http.Client{}
	tp := &CountingTransport{}
	client.Transport = tp
	return client, tp
}

// Client implements a fake client.
type Client struct {
	bqiface.Client
	ctx context.Context // Just for checking expiration/cancelation
}

// NewClient creates a new Client implementing bqiface.Client, with a dry run HTTPClient.
func NewClient(ctx context.Context, project string, opts ...option.ClientOption) (*Client, error) {
	dryRun, _ := DryRunClient()
	opts = append(opts, option.WithHTTPClient(dryRun))
	c, err := bigquery.NewClient(ctx, project, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{bqiface.AdaptClient(c), ctx}, nil
}

// Dataset creates a Dataset.
// TODO - understand how bqiface adapters/structs work, and make this return a Dataset
// that satisfies bqiface.Dataset interface?
func (client Client) Dataset(ds string) bqiface.Dataset {
	return Dataset{Dataset: client.Client.Dataset(ds), tables: make(map[string]*Table)}
}

func (client Client) Query(string) bqiface.Query {
	return Query{}
}
