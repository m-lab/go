package flagx

import (
	"fmt"
	"strings"
)

// KeyValueArray maps a key to one or multiple values.
// It parses "key=value1,value2,..." pairs from a given argument
// and is designed to be used for repeatable arguments.
// Each use of the flag will add a new key-[]value pair
// (or append to a previous one if the same key is used).
type KeyValueArray struct {
	pairs map[string][]string
}

// Set parses key=value1,value2,... arguments. Only one key-[]value
// pair can be specified per call.
func (kva *KeyValueArray) Set(pair string) error {
	p := strings.SplitN(pair, "=", 2)
	if len(p) != 2 {
		return fmt.Errorf("bad input pair: %s split on '=' into %d pieces (should have been 2)",
			pair, len(p))
	}
	if kva.pairs == nil {
		kva.pairs = make(map[string][]string)
	}
	v := strings.Split(p[1], ",")
	if _, ok := kva.pairs[p[0]]; !ok {
		kva.pairs[p[0]] = make([]string, 0)
	}
	kva.pairs[p[0]] = append(kva.pairs[p[0]], v...)
	return nil
}

// String returns the key-[]value pairs as a string.
func (kva *KeyValueArray) String() string {
	var sb strings.Builder
	for k, v := range kva.pairs {
		joined := strings.Join(v, ",")
		sb.WriteString(fmt.Sprintf("%s:[%s],", k, joined))
	}
	s := sb.String()
	return strings.TrimSuffix(s, ",")
}

// Get returns all of the KeyValueArray pairs as a map[string][]string.
// The returned value is a copy.
func (kva *KeyValueArray) Get() map[string][]string {
	h := make(map[string][]string)
	for k, s := range kva.pairs {
		c := make([]string, len(s))
		copy(c, s)
		h[k] = c
	}
	return h
}
