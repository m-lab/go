package flagx

import (
	"fmt"
	"strings"
)

// KeyArrayValue maps a key to one or multiple values.
// It parses "key=value1,value2,..." pairs from a given argument
// and is designed to be used for repeatable arguments.
// Each use of the flag will either add a new key-[]value pair
// (or overwrite a previous one if the same key is used).
type KeyArrayValue struct {
	pairs map[string][]string
}

// Set parses key=value1,value2,... arguments. Only one key-[]value
// pair can be specified per call.
func (kav *KeyArrayValue) Set(pair string) error {
	p := strings.Split(pair, "=")
	if len(p) != 2 {
		return fmt.Errorf("bad input pair: %s split on '=' into %d pieces (should have been 2)",
			pair, len(p))
	}
	if kav.pairs == nil {
		kav.pairs = make(map[string][]string)
	}
	v := strings.Split(p[1], ",")
	kav.pairs[p[0]] = v
	return nil
}

// String returns the key-[]value pairs as a string.
func (kav *KeyArrayValue) String() string {
	var sb strings.Builder
	for k, v := range kav.pairs {
		joined := strings.Join(v, ",")
		sb.WriteString(fmt.Sprintf("%s:[%s],", k, joined))
	}
	s := sb.String()
	return strings.TrimSuffix(s, ",")
}

// Get returns all of the KeyArrayValue pairs as a map[string][]string.
// The returned value is a copy.
func (kav *KeyArrayValue) Get() map[string][]string {
	h := make(map[string][]string)
	for k, s := range kav.pairs {
		h[k] = s
	}
	return h
}
