package bqext

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
func Customize(schema bigquery.Schema, subs map[string]bigquery.FieldSchema) {
	for i := range schema {
		fs := schema[i]
		s, ok := subs[fs.Name]
		if ok {
			*fs = s
		} else {
			if fs.Type == bigquery.RecordFieldType {
				Customize(fs.Schema, subs)
			}
		}
	}
}

// RemoveRequired recursively traverses a schema, setting Required to false in all fields
// that are not fundamentally required by BigQuery
func RemoveRequired(schema bigquery.Schema) {
	for i := range schema {
		fs := schema[i]
		switch fs.Type {
		case bigquery.RecordFieldType:
			fs.Required = false
			RemoveRequired(fs.Schema)

		case bigquery.TimeFieldType:
		case bigquery.TimestampFieldType:
		case bigquery.DateFieldType:
		case bigquery.DateTimeFieldType:

		default:
			fs.Required = false
		}
	}
}
