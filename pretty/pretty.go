package pretty

import (
	"encoding/json"
	"fmt"
)

// Print formats v and writes to stdout.
func Print(v interface{}) (int, error) {
	return fmt.Println(Sprint(v))
}

// Sprint formats v and returns a string.
func Sprint(v interface{}) string {
	d, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err.Error()
	}
	return string(d)
}
