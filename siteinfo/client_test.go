package siteinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const switchesPath = "testdata/switches.json"

// string provider implements a HTTPProvider but the response's content is
// a fixed string.
type stringProvider struct {
	response string
}

func (prov stringProvider) Get(string) (*http.Response, error) {
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(prov.response)),
		StatusCode: http.StatusOK,
	}, nil
}

// fileReaderProvider implements a HTTPProvider but the response's content
// comes from a configurable file.
type fileReaderProvider struct {
	path           string
	mustFailToRead bool
}

func (prov fileReaderProvider) Get(string) (*http.Response, error) {
	// Note: it's the caller's responsibility to call Body.Close().
	f, _ := os.Open(prov.path)
	return &http.Response{
		Body:       ioutil.NopCloser(bufio.NewReader(f)),
		StatusCode: http.StatusOK,
	}, nil
}

// failingProvider always fails.
type failingProvider struct{}

func (prov failingProvider) Get(string) (*http.Response, error) {
	return nil, fmt.Errorf("error")
}

// failingReadProvider returns a Body whose Read() method always fails.
type failingReadProvider struct{}

func (prov failingReadProvider) Get(string) (*http.Response, error) {
	return &http.Response{
		Body:       &mockReadCloser{},
		StatusCode: http.StatusOK,
	}, nil
}

// mockReadCloser is ReadCloser that fails.
type mockReadCloser struct{}

func (mockReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func (mockReadCloser) Close() error {
	return nil
}

//
// Tests start here.
//

func TestNew(t *testing.T) {
	client := New("project", "v1", http.DefaultClient)
	if client == nil {
		t.Errorf("New() returned nil.")
	}
}

func TestClient_Switches(t *testing.T) {
	prov := &fileReaderProvider{
		path: "testdata/switches.json",
	}
	client := New("test", "v1", prov)

	// This should return the content of the test file.
	res, err := client.Switches()
	if err != nil {
		t.Errorf("Switches() returned err: %v", err)
	}

	if len(res) != 144 {
		t.Errorf("Switches(): wrong map len %d, expected %d", len(res), 144)
	}

	if _, ok := res["akl01"]; !ok {
		t.Errorf("Switches() didn't return the expected result.")
	}

	// Make the HTTP client fail.
	client.httpClient = &failingProvider{}
	res, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}

	// Make reading the response body fail.
	client.httpClient = &failingReadProvider{}
	res, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}

	// Make the JSON unmarshalling fail.
	client.httpClient = &stringProvider{"this will fail"}
	res, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}
}
