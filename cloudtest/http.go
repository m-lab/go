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

// Package cloudtest provides utilities for testing, e.g. cloud
// service tests using mock http Transport, fake storage client, etc.
package cloudtest

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

/////////////////////////////////////////////////////////////////////
// LoggingTransport
/////////////////////////////////////////////////////////////////////

type loggingTransport struct {
	Transport http.RoundTripper
}

type nopCloser struct {
	io.Reader
}

func (nc *nopCloser) Close() error { return nil }

// Log the contents of a reader, returning a new reader with
// same content.
func loggingReader(r io.ReadCloser) io.ReadCloser {
	buf, _ := ioutil.ReadAll(r)
	r.Close()
	log.Printf("Response body:\n%+v\n", string(buf))
	return &nopCloser{bytes.NewReader(buf)}
}

// RoundTrip implements the RoundTripper interface, logging the
// request, and the response body, (which may be json).
func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Using %#v results in an escaped string we can use in code.
	log.Printf("Request:\n%#v\n", req)
	var resp *http.Response
	var err error
	// nil Transport is valid, so check for it.
	if t.Transport == nil {
		resp, err = http.DefaultTransport.RoundTrip(req)

	} else {
		resp, err = t.Transport.RoundTrip(req)
	}
	if err != nil {
		return nil, err
	}
	resp.Body = loggingReader(resp.Body)
	return resp, err
}

// NewLoggingClient returns an HTTP client that also logs all requests
// and responses.
func NewLoggingClient() *http.Client {
	client := &http.Client{}
	client.Transport = &loggingTransport{client.Transport}
	return client
}

// LoggingClient is an HTTP client that also logs all requests and
// responses.
func LoggingClient(client *http.Client) (*http.Client, error) {
	if client == nil {
		return nil, errors.New("client must not be nil")
	}
	if client == http.DefaultClient {
		return nil, errors.New("bad idea to add logging to default client")
	}

	client.Transport = &loggingTransport{client.Transport}
	return client, nil
}

/////////////////////////////////////////////////////////////////////
// NewChannelClient
// Provides a transport that gets http.Response from a channel.
/////////////////////////////////////////////////////////////////////

// channelTransport provides a RoundTripper that handles everything
// locally.
type channelTransport struct {
	//	Transport http.RoundTripper
	Responses <-chan *http.Response
}

// RoundTrip implements the RoundTripper interface, using a channel to
// provide http responses.  This will block if the channel is empty.
func (t channelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := <-t.Responses // may block
	resp.Request = req
	return resp, nil
}

// NewChannelClient is an HTTP client that ignores requests and returns
// responses provided by a channel.
func NewChannelClient(c <-chan *http.Response) *http.Client {
	client := &http.Client{}
	client.Transport = &channelTransport{c}

	return client
}
