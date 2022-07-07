package siteinfotest

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// StringProvider implements an HTTPProvider but the response's content is
// a fixed string.
type StringProvider struct {
	Response string
}

func (prov StringProvider) Get(string) (*http.Response, error) {
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(prov.Response)),
		StatusCode: http.StatusOK,
	}, nil
}

// FileReaderProvider implements an HTTPProvider but the response's content
// comes from a configurable file.
type FileReaderProvider struct {
	Path           string
	MustFailToRead bool
}

func (prov FileReaderProvider) Get(string) (*http.Response, error) {
	// Note: it's the caller's responsibility to call Body.Close().
	f, err := os.Open(prov.Path)
	if err != nil {
		return nil, err
	}
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
		Body:       &FailingReadCloser{},
		StatusCode: http.StatusOK,
	}, nil
}

// FailingReadCloser is ReadCloser that fails.
type FailingReadCloser struct{}

func (FailingReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func (FailingReadCloser) Close() error {
	return nil
}
