package flagx

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// KeyValue parses "key=value" pairs from a given argument. The KeyValue flag is
// designed to be used for repeatable arguments. Each use of the flag will add
// another key value pair (or overwrite a previous one if the same key is used).
type KeyValue struct {
	pairs map[string]kvsource
}

type kvsource struct {
	value string
	fname string
}

// Set parses key=value argument. Multiple pairs may be separated with a comma,
// i.e. "key1=value1,key2=value2". When the first character of the value is
// prefixed by with "@", i.e. "key1=@file1", Set reads the file content for the
// key value.
func (kv *KeyValue) Set(kvs string) error {
	pairs := strings.Split(kvs, ",")
	for _, pair := range pairs {
		fields := strings.SplitN(pair, "=", 2)
		if len(fields) != 2 {
			return fmt.Errorf("bad key/value: %q split on '=' into %d pieces (should have been 2)", pair, len(fields))
		}
		if kv.pairs == nil {
			kv.pairs = make(map[string]kvsource)
		}
		if len(fields[1]) > 0 && fields[1][0] == '@' {
			fname := fields[1][1:]
			b, err := ioutil.ReadFile(fname)
			if err != nil {
				return err
			}
			kv.pairs[fields[0]] = kvsource{
				value: string(b),
				fname: fname,
			}
		} else {
			kv.pairs[fields[0]] = kvsource{value: fields[1]}
		}
	}
	return nil
}

// String serializes parsed arguments into a form similiar to how they were
// added from the command line. Order is not preserved.
func (kv *KeyValue) String() string {
	fields := []string{}
	for k, s := range kv.pairs {
		if s.fname == "" {
			fields = append(fields, fmt.Sprintf("%s=%s", k, s.value))
		} else {
			fields = append(fields, fmt.Sprintf("%s=@%s", k, s.fname))
		}
	}
	return strings.Join(fields, ",")
}

// Get returns all of the KeyValue pairs as a map[string]string. The returned
// value is a copy.
func (kv *KeyValue) Get() map[string]string {
	h := make(map[string]string)
	for k, s := range kv.pairs {
		h[k] = s.value
	}
	return h
}
