package bqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
)

// PrettyPrint generates a formatted json representation of a Schema.
// It simplifies the schema by removing zero valued fields, and compacting
// each field record onto a single line.
// Intended for diagnostics and debugging.  Not suitable for production use.
func PrettyPrint(schema bigquery.Schema, simplify bool) (string, error) {
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(jsonBytes), "\n")
	before := ""
	output := &bytes.Buffer{}

	for _, line := range lines {
		// Remove Required from all fields.
		trim := strings.Trim(strings.TrimSpace(line), ",") // remove leading space, trailing comma
		switch trim {
		case `"Schema": null`:
			fallthrough
		case `"Repeated": false`:
			fallthrough
		case `"Required": false`:
			if !simplify {
				fmt.Fprint(output, before, trim)
				before = ", "
			}
		case `"Required": true`:
			fmt.Fprint(output, before, trim)
			before = ", "
		case `"Schema": [`:
			fallthrough
		case `[`:
			fmt.Fprintf(output, "%s%s\n", before, trim)
			before = ""
		case `{`:
			fmt.Fprint(output, line)
			before = ""
		case `}`:
			fmt.Fprintln(output, strings.TrimSpace(line))
		case `]`:
			fmt.Fprint(output, line)
			before = ""
		default:
			fmt.Fprint(output, before, trim)
			before = ", "
		}
	}
	fmt.Fprintln(output)
	return output.String(), nil
}

// Customize recursively traverses a schema, substituting any fields that have a matching
// name in the provided map.
func Customize(schema bigquery.Schema, subs map[string]bigquery.FieldSchema) bigquery.Schema {
	// We have to copy the schema, to avoid corrupting the bigquery fieldCache.
	out := make(bigquery.Schema, len(schema))
	for i := range schema {
		out[i] = &bigquery.FieldSchema{}
		*out[i] = *schema[i]
		fs := out[i]
		s, ok := subs[fs.Name]
		if ok {
			*fs = s

		} else {
			if fs.Type == bigquery.RecordFieldType {
				fs.Schema = Customize(fs.Schema, subs)
			}
		}

	}
	return out
}

// RemoveRequired recursively traverses a schema, setting Required to false in all fields
// that are not fundamentally required by BigQuery
func RemoveRequired(schema bigquery.Schema) bigquery.Schema {
	// We have to copy the schema, to avoid corrupting the bigquery fieldCache.
	out := make(bigquery.Schema, len(schema))
	for i := range schema {
		out[i] = &bigquery.FieldSchema{}
		*out[i] = *schema[i]
		fs := out[i]
		switch fs.Type {
		case bigquery.RecordFieldType:
			fs.Required = false
			fs.Schema = RemoveRequired(fs.Schema)

		// These field types seem to be always required.
		case bigquery.TimeFieldType:
		case bigquery.TimestampFieldType:
		case bigquery.DateFieldType:
		case bigquery.DateTimeFieldType:

		default:
			fs.Required = false
		}
	}

	return out
}
