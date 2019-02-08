package uuid

import (
	"errors"
	"testing"
)

func TestFileAsProxyForBoottime(t *testing.T) {
	_, err := getPrefix("/this/file/does/not/exist")
	if err == nil {
		t.Error("Should have had an error on a non-existent file")
	}
}

func TestErrorDoesntCauseNullUUID(t *testing.T) {
	defer func(oldPrefix string, oldError error) {
		cachedPrefixString, cachedPrefixError = oldPrefix, oldError
	}(cachedPrefixString, cachedPrefixError)

	cachedPrefixString = ""
	cachedPrefixError = errors.New("An error for testing")

	id, err := FromCookie(0)
	if err == nil {
		t.Error("Should have had an error")
	}
	if id == "" {
		t.Error("An error should not cause an empty-string uuid to be returned")
	}
}

func TestBadHostnameDoesntCauseNullUUID(t *testing.T) {
	defer func(oldPrefix string, oldError error, oldOsHostname func() (string, error)) {
		cachedPrefixString, cachedPrefixError = oldPrefix, oldError
		osHostname = oldOsHostname
	}(cachedPrefixString, cachedPrefixError, osHostname)

	osHostname = func() (string, error) {
		return "", errors.New("An error for testing")
	}
	cachedPrefixString, cachedPrefixError = getPrefix("/proc")
	id, err := FromCookie(0)
	if err == nil {
		t.Error("Should have had an error")
	}
	if id == "" {
		t.Error("An error should not cause an empty-string uuid to be returned")
	}
}
