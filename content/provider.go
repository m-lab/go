package content

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"github.com/m-lab/uuid-annotator/metrics"
)

// Errors that might be returned outside the package.
var (
	ErrUnsupportedURLScheme = errors.New("Unsupported URL scheme")
	ErrNoChange             = errors.New("Data is unchanged")
)

// Provider is the interface implemented by everything that can return raw files.
type Provider interface {
	// Get returns the raw file []byte read from the latest copy of the provider
	// URL. It may be called multiple times. Caching is left up to the individual
	// Provider implementation.
	Get(ctx context.Context) ([]byte, error)
}

// gcsProvider gets zip files from Google Cloud Storage.
type gcsProvider struct {
	bucket, filename string
	client           stiface.Client
	md5              []byte
}

func (g *gcsProvider) Get(ctx context.Context) ([]byte, error) {
	o := g.client.Bucket(g.bucket).Object(g.filename)
	oa, err := o.Attrs(ctx)
	if err != nil {
		return nil, err
	}
	if g.md5 != nil && bytes.Equal(g.md5, oa.MD5) {
		return nil, ErrNoChange
	}

	// Otherise, we know that either g.md5 == nil || g.md5 != oa.MD5.
	// Reload data only if the object changed or the data was never loaded in the first place.
	r, err := o.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if g.md5 != nil {
		metrics.GCSFilesLoaded.WithLabelValues(hex.EncodeToString(g.md5)).Set(0)
	}
	g.md5 = oa.MD5
	metrics.GCSFilesLoaded.WithLabelValues(hex.EncodeToString(g.md5)).Set(1)
	return data, nil
}

// fileProvider gets files from the local disk.
type fileProvider struct {
	filename string
	mtime    time.Time
}

func (f *fileProvider) Get(ctx context.Context) ([]byte, error) {
	s, err := os.Stat(f.filename)
	if err != nil {
		return nil, fmt.Errorf("Could not os.Stat(%q): %w", f.filename, err)
	}
	newtime := s.ModTime()
	if newtime == f.mtime {
		return nil, ErrNoChange
	}
	b, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return nil, err
	}
	f.mtime = newtime
	return b, nil
}

// httpsProvider gets files from public HTTPS URLs (i.e. no authentication).
type httpsProvider struct {
	u       url.URL
	timeout time.Duration
	client  *http.Client
}

func (h *httpsProvider) Get(ctx context.Context) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	r, err := http.NewRequestWithContext(reqCtx, http.MethodGet, h.u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// FromURL returns a new rawfile.Provider based on the passed-in URL. Supported
// URL schemes are currently: gs://bucket/filename, file:localpath, and
// https://. Whether the path contained in the URL is valid isn't known until
// the Get() method of the returned Provider is called. Unsupported URL schemes
// cause this to return ErrUnsupportedURLScheme.
//
// Users interested in having the daemon download the data directly from MaxMind
// using credentials should implement an alternate https case in the below
// handler. M-Lab doesn't need that case because we cache MaxMind's data to
// reduce load on their servers and to eliminate a runtime dependency on a third
// party service.
func FromURL(ctx context.Context, u *url.URL) (Provider, error) {
	switch u.Scheme {
	case "gs":
		client, err := storage.NewClient(ctx)
		filename := strings.TrimPrefix(u.Path, "/")
		if len(filename) == 0 {
			return nil, errors.New("Bad GS url, no filename detected")
		}
		return &gcsProvider{
			client:   stiface.AdaptClient(client),
			bucket:   u.Host,
			filename: filename,
		}, err
	case "file":
		if u.Path == "" {
			return &fileProvider{
				filename: u.Opaque,
			}, nil
		}
		return &fileProvider{
			filename: u.Path,
		}, nil

	case "https":
		return &httpsProvider{
			u:       *u,
			timeout: time.Minute,
			client:  http.DefaultClient,
		}, nil
	default:
		return nil, ErrUnsupportedURLScheme
	}
}
