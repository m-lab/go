package bqx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
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

var (
	ErrInvalidProject = errors.New("Invalid project name")
	ErrInvalidDataset = errors.New("Invalid dataset name")
	ErrInvalidTable   = errors.New("Invalid table name")
	ErrInvalidFQTable = errors.New("Invalid fully qualified table name")

	projectRegex = regexp.MustCompile("[a-z0-9-]+")
	datasetRegex = regexp.MustCompile("[a-zA-Z0-9_]+")
	tableRegex   = regexp.MustCompile("[a-zA-Z0-9_]+")
)

type pdt struct {
	Project string
	Dataset string
	Table   string
}

func parsePDT(fq string) (*pdt, error) {
	parts := strings.Split(fq, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidFQTable
	}
	if !projectRegex.MatchString(parts[0]) {
		return nil, ErrInvalidProject
	}
	if !datasetRegex.MatchString(parts[1]) {
		return nil, ErrInvalidDataset
	}
	if !tableRegex.MatchString(parts[2]) {
		return nil, ErrInvalidTable
	}
	return &pdt{parts[0], parts[1], parts[2]}, nil
}

// TODO maybe include flag to enable dataset creation?
func CreateOrUpdateTable(ctx context.Context, table string,
	schema bigquery.Schema, partitioning *bigquery.TimePartitioning, clustering *bigquery.Clustering) error {
	pdt, err := parsePDT(table)
	if err != nil {
		return err
	}
	client, err := bigquery.NewClient(ctx, pdt.Project)
	if err != nil {
		return err
	}

	ds := client.Dataset(pdt.Dataset)

	ds.Create(ctx, nil)

	t := ds.Table(pdt.Table)

	meta, err := t.Metadata(ctx)
	if err != nil {
		// Table probably doesn't exist
		log.Println(err)

		meta = &bigquery.TableMetadata{Schema: schema}
		meta.TimePartitioning = partitioning
		meta.Clustering = clustering

		err = t.Create(ctx, meta)
		return err
	}

	changes := bigquery.TableMetadataToUpdate{
		Schema: schema,
	}

	md, err := t.Update(ctx, changes, meta.ETag)
	if err != nil {
		return err
	}
	log.Printf("%+v\n", md)
	return nil
}
