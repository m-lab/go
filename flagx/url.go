package flagx

import (
	"net/url"

	"github.com/m-lab/go/rtx"
)

// URL is a flag type that parses the given URL string, handling errors during
// flag parsing. See MustNewURL() to set a default value.
type URL struct {
	*url.URL
}

// MustNewURL creates a new URL flag initalized with the given value. For example:
//   f := flagx.MustNewURL("http://example.com")
func MustNewURL(s string) URL {
	u := URL{}
	rtx.Must(u.Set(s), "Failed to parse and set given URL %q", s)
	return u
}

// Get retrieves the inner URL.
func (u URL) Get() *url.URL {
	return u.URL
}

// Set accepts a URL, parses it, and stores the result.
func (u *URL) Set(s string) error {
	var err error
	(*u).URL, err = url.Parse(s)
	if err != nil {
		return err
	}
	return nil
}

// String formats the underlying URL as a string.
func (u URL) String() string {
	return u.URL.String()
}
