package siteinfotest

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

var (
	stringProviderResponse     = "lol"
	fileReaderProviderResponse = "test"
)

func TestStringProvider(t *testing.T) {
	provider := StringProvider{
		Response: stringProviderResponse,
	}
	resp, err := provider.Get("rofl")
	if err != nil {
		t.Errorf("did not expect an error, but got: %v", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("did not expect an error, but got: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected an http.StatusOk status, but got: %v", resp.StatusCode)
	}
	body := string(bodyBytes)
	if body != stringProviderResponse {
		t.Errorf("expected response body of '%s', but got: '%s'", stringProviderResponse, body)
	}
}

func TestFileReaderProvider(t *testing.T) {
	provider := FileReaderProvider{
		Path: "testdata/data.txt",
	}
	resp, err := provider.Get("rofl")
	if err != nil {
		t.Errorf("did not expect an error, but got: %v", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("did not expect an error, but got: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected an http.StatusOk status, but got: %v", resp.StatusCode)
	}
	body := strings.TrimSpace(string(bodyBytes))
	if body != fileReaderProviderResponse {
		t.Errorf("expected response body of '%s', but got: '%s'", fileReaderProviderResponse, body)
	}

	// Test error case where file does not exist.
	provider = FileReaderProvider{
		Path: "file/does/not/exist.txt",
	}
	_, err = provider.Get("rofl")
	if err == nil {
		t.Error("expected an error, but did not get one")
	}
}

func TestFailingProvider(t *testing.T) {
	provider := FailingProvider{}
	_, err := provider.Get("rofl")
	if err == nil {
		t.Error("expected an error, but did not get one")
	}
}

func TestFailingReadProvider(t *testing.T) {
	provider := FailingReadProvider{}
	resp, err := provider.Get("rofl")
	if err != nil {
		t.Errorf("did not expect an error, but got: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected an http.StatusOk status, but got: %v", resp.StatusCode)
	}
	_, err = io.ReadAll(resp.Body)
	if err == nil {
		t.Error("expected an error, but did not get one")
	}
}

func TestFailingReadCloser(t *testing.T) {
	readCloser := FailingReadCloser{}
	_, err := readCloser.Read([]byte("rofl"))
	if err == nil {
		t.Error("expected an error, but did not get one")
	}
	err = readCloser.Close()
	if err != nil {
		t.Errorf("did not expect an error, but got: %v", err)
	}
}
