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

// StringProvider implements a HTTPProvider but the response's content is
// a fixed string.
type StringProvider struct {
	response string
}

func (prov StringProvider) Get(string) (*http.Response, error) {
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(prov.response)),
		StatusCode: http.StatusOK,
	}, nil
}

// FileReaderProvider implements a HTTPProvider but the response's content
// comes from a configurable file.
type FileReaderProvider struct {
	path           string
	mustFailToRead bool
}

func (prov FileReaderProvider) Get(string) (*http.Response, error) {
	// Note: it's the caller's responsibility to call Body.Close().
	f, _ := os.Open(prov.path)
	return &http.Response{
		Body:       ioutil.NopCloser(bufio.NewReader(f)),
		StatusCode: http.StatusOK,
	}, nil
}

// FailingProvider always fails.
type FailingProvider struct{}

func (prov FailingProvider) Get(string) (*http.Response, error) {
	return nil, fmt.Errorf("error")
}

// FailingReadProvider returns a Body whose Read() method always fails.
type FailingReadProvider struct{}

func (prov FailingReadProvider) Get(string) (*http.Response, error) {
	return &http.Response{
		Body:       &MockReadCloser{},
		StatusCode: http.StatusOK,
	}, nil
}

// mockReadCloser is ReadCloser that fails.
type MockReadCloser struct{}

func (MockReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func (MockReadCloser) Close() error {
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
	prov := &FileReaderProvider{
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

	// Test working call
	client.httpClient = &StringProvider{
		response: `{
			"xyz03": {
				"auto_negotiation": "yes",
				"flow_control": "no",
				"ipv4_prefix": "10.0.0.0/26",
				"rstp": "yes",
				"switch_make": "juniper",
				"switch_model": "qfx5100",
				"uplink_port": "xe-0/0/45",
				"uplink_speed": "10g"
			}
		}`,
	}
	_, err = client.Switches()
	if err != nil {
		t.Error("Switches(): expected success, got error.")
	}

	// Make the HTTP client fail.
	client.httpClient = &FailingProvider{}
	_, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}

	// Make reading the response body fail.
	client.httpClient = &FailingReadProvider{}
	_, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}

	// Make the JSON unmarshalling fail.
	client.httpClient = &StringProvider{"this will fail"}
	_, err = client.Switches()
	if err == nil {
		t.Errorf("Switches(): expected err, got nil.")
	}
}

func TestClient_Projects(t *testing.T) {
	prov := &FileReaderProvider{
		path: "testdata/projects.json",
	}
	client := New("rofl", "v2", prov)

	// This should return the content of the test file.
	res, err := client.Projects()
	if err != nil {
		t.Errorf("Projects() returned err: %v", err)
	}

	if len(res) != 8 {
		t.Errorf("Projects(): wrong map len %d, expected %d", len(res), 8)
	}

	if _, ok := res["mlab3-lol01"]; !ok {
		t.Errorf("Projects() didn't return the expected result.")
	}

	// Test working HTTP client request
	client.httpClient = &StringProvider{
		response: `{"mlab1-xyz0t":"mlab-sandbox"}`,
	}
	_, err = client.Projects()
	if err != nil {
		t.Error("Projects(): expected success, got error.")
	}

	// Make the HTTP client fail.
	client.httpClient = &FailingProvider{}
	_, err = client.Projects()
	if err == nil {
		t.Errorf("Projects(): expected err, got nil.")
	}

	// Make reading the response body fail.
	client.httpClient = &FailingReadProvider{}
	_, err = client.Projects()
	if err == nil {
		t.Errorf("Projects(): expected err, got nil.")
	}

	// Make the JSON unmarshalling fail.
	client.httpClient = &StringProvider{"this will fail"}
	_, err = client.Projects()
	if err == nil {
		t.Errorf("Projects(): expected err, got nil.")
	}
}

func TestClient_Machines(t *testing.T) {
	prov := &FileReaderProvider{
		path: "testdata/machines.json",
	}
	client := New("testmachines", "v2", prov)

	// This should return the content of the test file.
	res, err := client.Machines()
	if err != nil {
		t.Errorf("Machines() returned err: %v", err)
	}

	if len(res) != 4 {
		t.Errorf("Machines(): wrong map len %d, expected %d", len(res), 8)
	}

	// Test working HTTP client request
	client.httpClient = &StringProvider{
		response: `[
          {
            "hostname": "mlab2-abc09.mlab-oti.measurement-lab.org",
            "ipv4": "192.168.5.150",
            "ipv6": "2004:42a8:144:6::150",
            "project": "mlab-oti"
          }
		]`,
	}

	_, err = client.Machines()
	if err != nil {
		t.Error("Machines(): expected success, got error.")
	}

	// Make the HTTP client fail.
	client.httpClient = &FailingProvider{}
	_, err = client.Machines()
	if err == nil {
		t.Errorf("Machines(): expected err, got nil.")
	}

	// Make reading the response body fail.
	client.httpClient = &FailingReadProvider{}
	_, err = client.Machines()
	if err == nil {
		t.Errorf("Machines(): expected err, got nil.")
	}

	// Make the JSON unmarshalling fail.
	client.httpClient = &StringProvider{"this will fail"}
	_, err = client.Machines()
	if err == nil {
		t.Errorf("Machines(): expected err, got nil.")
	}
}

func TestClient_SiteMachines(t *testing.T) {
	prov := &FileReaderProvider{
		path: "testdata/site-machines.json",
	}
	client := New("testsitemachines", "v2", prov)

	// This should return the content of the test file.
	res, err := client.SiteMachines()
	if err != nil {
		t.Errorf("SiteMachines() returned err: %v", err)
	}

	if len(res) != 3 {
		t.Errorf("SiteMachines(): wrong map len %d, expected %d", len(res), 3)
	}
	if len(res["abc01"]) != 4 {
		t.Errorf("SiteMachines(): wrong machine count for site abc01 %d, expected %d", len(res["abc01"]), 4)
	}
	if len(res["lol02"]) != 1 {
		t.Errorf("SiteMachines(): wrong machine count for site lol02 %d, expected %d", len(res["lol02"]), 1)
	}
	if len(res["xyz0t"]) != 2 {
		t.Errorf("SiteMachines(): wrong machine count for site xyz0t %d, expected %d", len(res["xyz0t"]), 2)
	}

	// Test working HTTP client request.
	client.httpClient = &StringProvider{
		response: `{
			"abc01": [
				"mlab1"
			]
		}`,
	}

	_, err = client.SiteMachines()
	if err != nil {
		t.Errorf("SiteMachines(): expected success, got error: %v", err)
	}

	// Make the HTTP client fail.
	client.httpClient = &FailingProvider{}
	_, err = client.SiteMachines()
	if err == nil {
		t.Error("SiteMachines(): expected err, got nil.")
	}

	// Make reading the response body fail.
	client.httpClient = &FailingReadProvider{}
	_, err = client.SiteMachines()
	if err == nil {
		t.Error("SiteMachines(): expected err, got nil.")
	}

	// Make the JSON unmarshalling fail.
	client.httpClient = &StringProvider{"this will fail"}
	_, err = client.SiteMachines()
	if err == nil {
		t.Error("SiteMachines(): expected err, got nil.")
	}
}
