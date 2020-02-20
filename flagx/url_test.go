package flagx_test

import (
	"flag"
	"net/url"
	"testing"

	"github.com/go-test/deep"
	"github.com/m-lab/go/flagx"
)

func TestURL(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    *url.URL
		wantErr bool
	}{
		{
			name: "success-gs",
			s:    "gs://bucket/path/object.tar.gz",
			want: &url.URL{
				Scheme: "gs",
				Host:   "bucket",
				Path:   "/path/object.tar.gz",
			},
		},
		{
			name: "success-https",
			s:    "https://bucket:1234/path/object.tar.gz?this=that",
			want: &url.URL{
				Scheme:   "https",
				Host:     "bucket:1234",
				Path:     "/path/object.tar.gz",
				RawQuery: "this=that",
			},
		},
		{
			name:    "error-bad-url-format",
			s:       "://this-is-not-a-url",
			wantErr: true,
		},
		{
			name: "error-empty-url",
			s:    "",
			want: &url.URL{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to set the URL.
			u := flagx.URL{}
			if u.String() != "" {
				t.Errorf("URL.String() empty URL flag returned a value %q", u.String())
			}
			if err := u.Set(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("URL.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Skip remaining checks if we expect an error.
			if tt.wantErr {
				return
			}
			// Does the original URL match the reformatted URL?
			if u.String() != tt.s {
				t.Errorf("URL.String() got = %q, want %q", u.String(), tt.s)
			}
			// Does the parsed URL match the expected URL?
			if diff := deep.Equal(u.Get(), tt.want); diff != nil {
				t.Errorf("URL.Set()\ngot %#v\nwant %#v\ndifferences: %v", u.URL, tt.want, diff)
			}
			// Try to initialize a flag using the original test value.
			s := flagx.MustNewURL(tt.s)
			if diff := deep.Equal(u, s); diff != nil {
				t.Errorf("URL.Set and MustNewURL returned different values: %#v", diff)
			}
		})
	}
}

// Verify that flagx.URL implements the flag.Value interface.
func assertFlagValueURL(b flagx.URL) {
	func(in flag.Value) {}(&b)
}
