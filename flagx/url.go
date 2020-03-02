package flagx

import (
	"net/url"

	"github.com/m-lab/go/rtx"
)

// URL is a flag type for parsing URL strings and handling errors during flag
// parsing. Use MustNewURL() to specify a default value.
type URL struct {
	*url.URL
}

// MustNewURL creates a new flagx.URL initialized with the given value. Failure
// to parse is fatal. For example:
//   f := flagx.MustNewURL("http://example.com")
func MustNewURL(s string) URL {
	u := URL{}
	rtx.Must(u.Set(s), "Failed to parse and set given URL %q", s)
	return u
}

// Get returns the inner *url.URL type.
func (u *URL) Get() *url.URL {
	return u.URL
}

// Set parses a URL string and stores the result.
func (u *URL) Set(s string) error {
	var err error
	(*u).URL, err = url.Parse(s)
	if err != nil {
		return err
	}
	return nil
}

// String formats the underlying URL as a string.
func (u *URL) String() string {
	if u.URL == nil {
		return ""
	}
	return u.URL.String()
}
