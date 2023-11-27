package flagx

import (
	"regexp"
	"strings"
)

// KeyValueEscaped parses "key=value" pairs from a given argument. The KeyValue flag is
// designed to be used for repeatable arguments (separated by an unescaped ',').
// Escaped characters in the key/value will remain escaped (i.e., key\/=value\,
// will be interpreted as key=`key\/` and value=`value\,“).
// Each use of the flag will add another key value pair (or overwrite a previous one
// if the same key is used).
type KeyValueEscaped struct {
	KeyValue
}

// Set parses key=value argument. Multiple pairs may be separated with an unescaped comma,
// i.e. "key1=value1,key2=value2".
// When the first character of the value is prefixed by with "@", i.e. "key1=@file1",
// Set reads the file content for the key value.
func (kve *KeyValueEscaped) Set(kvs string) error {
	reg := regexp.MustCompile(".*?[^\\\\](,|$)")
	pairs := reg.FindAllString(kvs, -1)
	for i := 0; i < len(pairs)-1; i++ {
		pairs[i] = strings.TrimSuffix(pairs[i], ",")
	}
	return kve.formPairs(pairs)
}

// String serializes parsed arguments into a form similiar to how they were
// added from the command line. Order is not preserved.
func (kve *KeyValueEscaped) String() string {
	return kve.KeyValue.String()
}

// Get returns all of the KeyValueEscaped pairs as a map[string]string. The returned
// value is a copy.
func (kve *KeyValueEscaped) Get() map[string]string {
	return kve.KeyValue.Get()
}
