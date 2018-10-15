package bqfake

// TODO: Implement context expiration checking.

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
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

	if os.Getenv("VERBOSE_CLIENT") != "" {
		log.Println("Called the underlying transport.")
		log.Println(req.URL)
		debug.PrintStack()
	}
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
// Most actions on objects derived from Client should be handled in methods on fakes.  Any actions
// that pass through to calls on the client will go to the DryRunClient.  The calls to the client
// can by accessing the underlying DryRunClient, or calls can be logged by setting the VERBOSE_CLIENT
// environment variable.
func NewClient(ctx context.Context, project string, opts ...option.ClientOption) (*Client, error) {
	dryRun, _ := DryRunClient()
	opts = append(opts, option.WithHTTPClient(dryRun))
	c, err := bigquery.NewClient(ctx, project, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{bqiface.AdaptClient(c), ctx}, nil
}

// Dataset creates a Dataset with an underlying DryRunClient.  Methods should generally be
// handled by overrides at some level before hitting the client.
func (client Client) Dataset(ds string) bqiface.Dataset {
	return Dataset{Dataset: client.Client.Dataset(ds), tables: make(map[string]*Table)}
}

func (client Client) Query(query string) bqiface.Query {
	return Query{client.Client.Query(query)}
}
