package flagx

import (
	"fmt"
	"strings"
)

// KeyValue is a way of setting the elements of a map[string]string individually
// as key-value pairs on the command-line. It is designed to be used as a
// repeated argument, where each invocation of the command-line argument will
// add another key value pair.
//
// One use for this library could be to specify metadata headers.
type KeyValue struct {
	pairs map[string]string
}

// Set breaks the string apart at the first equals sign and puts the result
// key-value pair into the map.
func (kv *KeyValue) Set(kvstring string) error {
	pair := strings.SplitN(kvstring, "=", 2)
	if len(pair) != 2 {
		return fmt.Errorf("Bad kay value pair %q split on '=' into %d pieces (should have been 2)", kvstring, len(pair))
	}
	if kv.pairs == nil {
		kv.pairs = make(map[string]string)
	}
	kv.pairs[pair[0]] = pair[1]
	return nil
}

// String converts the headers into a text representation. It's not expected to
// provide useful output here, but it is a function that is required in order to
// implement flag.Value.
func (kv *KeyValue) String() string {
	return fmt.Sprintf("%+v", kv.pairs)
}

// Get returns a copy of the KeyValue object as a map[string]string. This new
// map may be modified as desired without worrying about modifying the
// command-line arguments.
func (kv *KeyValue) Get() map[string]string {
	h := make(map[string]string)
	for k, v := range kv.pairs {
		h[k] = v
	}
	return h
}
