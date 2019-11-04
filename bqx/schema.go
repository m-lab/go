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
	"github.com/m-lab/go/rtx"
	"gopkg.in/yaml.v2"
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
		trim := strings.Trim(strings.TrimSpace(line), ",")
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

// CustomizeAppend recursively traverses a schema, appending the
// bigquery.FieldSchema to existing fields matching a name in the provided map.
func CustomizeAppend(schema bigquery.Schema, additions map[string]*bigquery.FieldSchema) bigquery.Schema {
	// We have to copy the schema, to avoid corrupting the bigquery fieldCache.
	custom := make(bigquery.Schema, len(schema))
	for i := range schema {
		custom[i] = &bigquery.FieldSchema{}
		*custom[i] = *schema[i]
		fs := custom[i]
		s, ok := additions[fs.Name]
		if ok {
			fs.Schema = append(fs.Schema, s)

		} else {
			if fs.Type == bigquery.RecordFieldType {
				fs.Schema = CustomizeAppend(fs.Schema, additions)
			}
		}

	}
	return custom
}

// Customize recursively traverses a schema, substituting any fields that have
// a matching name in the provided map.
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

// RemoveRequired recursively traverses a schema, setting Required to false in
// all fields that are not fundamentally required by BigQuery.
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

// These errors are self-explanatory.
var (
	ErrInvalidProjectName = errors.New("Invalid project name")
	ErrInvalidDatasetName = errors.New("Invalid dataset name")
	ErrInvalidTableName   = errors.New("Invalid table name")
	ErrInvalidFQTable     = errors.New("Invalid fully qualified table name")
)

var (
	projectRegex = regexp.MustCompile("^[a-z0-9-]+$")
	datasetRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")
	tableRegex   = regexp.MustCompile("^[a-zA-Z0-9_]+$")
)

// PDT contains a bigquery project, dataset, and table name.
type PDT struct {
	Project string
	Dataset string
	Table   string
}

// ParsePDT parses and validates a fully qualified bigquery table name of the
// form project.dataset.table.  None of the elements needs to exist, but all
// must conform to the corresponding naming restrictions.
func ParsePDT(fq string) (PDT, error) {
	parts := strings.Split(fq, ".")
	if len(parts) != 3 {
		return PDT{}, ErrInvalidFQTable
	}
	if !projectRegex.MatchString(parts[0]) {
		return PDT{}, ErrInvalidProjectName
	}
	if !datasetRegex.MatchString(parts[1]) {
		return PDT{}, ErrInvalidDatasetName
	}
	if !tableRegex.MatchString(parts[2]) {
		return PDT{}, ErrInvalidTableName
	}
	return PDT{parts[0], parts[1], parts[2]}, nil
}

// UpdateTable will update an existing table.  Returns error if the table
// doesn't already exist, or if the schema changes are incompatible.
func (pdt PDT) UpdateTable(ctx context.Context, client *bigquery.Client, schema bigquery.Schema) error {
	// See if dataset exists, or create it.
	ds := client.Dataset(pdt.Dataset)
	_, err := ds.Metadata(ctx)
	if err != nil {
		// TODO if we see errors showing up here.
		// TODO possibly retry if this is a transient error.
		// apiErr, ok := err.(*googleapi.Error)
		log.Println(err) // So we can discover these and add explicit handling.
		return err
	}
	t := ds.Table(pdt.Table)

	meta, err := t.Metadata(ctx)
	if err != nil {
		return err
	}

	// If table already exists, attempt to update the schema.
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

// CreateTable will create a new table, or fail if the table already exists.
// It will also set appropriate time-partitioning field and clustering fields
// if non-nil arguments are provided.  Returns error if the dataset does not
// already exist, or if other errors are encountered.
func (pdt PDT) CreateTable(ctx context.Context, client *bigquery.Client, schema bigquery.Schema, description string,
	partitioning *bigquery.TimePartitioning, clustering *bigquery.Clustering) error {
	ds := client.Dataset(pdt.Dataset)

	if _, err := ds.Metadata(ctx); err != nil {
		// TODO if we see errors showing up here.
		// TODO possibly retry if this is a transient error.
		// apiErr, ok := err.(*googleapi.Error)
		log.Println(err) // So we can discover these and add explicit handling.
		return err
	}

	t := ds.Table(pdt.Table)

	meta := &bigquery.TableMetadata{
		Schema:           schema,
		TimePartitioning: partitioning,
		Clustering:       clustering,
		Description:      description,
	}

	err := t.Create(ctx, meta)

	if err != nil {
		// TODO if we see errors showing up here.
		// TODO possibly retry if this is a transient error.
		// apiErr, ok := err.(*googleapi.Error)
		log.Println(err) // So we can discover these and add explicit handling.
		return err
	}

	return nil
}

// SchemaDoc contains bigquery.Schema field Descriptions as read from an auxiliary source, such as YAML.
type SchemaDoc map[string]map[string]string

// NewSchemaDoc reads the given file and attempts to parse it as a SchemaDoc. Errors are fatal.
func NewSchemaDoc(docs []byte) SchemaDoc {
	sd := SchemaDoc{}
	err := yaml.Unmarshal(docs, &sd)
	rtx.Must(err, "Failed to unmarshal schema doc")
	return sd
}

// UpdateSchemaDescription walks each field in the given schema and assigns the
// Description field in place using values found in the given SchemaDoc.
func UpdateSchemaDescription(schema bigquery.Schema, docs SchemaDoc) error {
	WalkSchema(
		schema, func(fields []*bigquery.FieldSchema) error {
			var ok bool
			var d map[string]string
			// Starting with the longest prefix, stop looking for descriptions on first match.
			prefix := []string{}
			for i := range fields {
				prefix = append(prefix, fields[i].Name)
			}
			for start := 0; start < len(prefix) && !ok; start++ {
				path := strings.Join(prefix[start:], ".")
				d, ok = docs[path]
			}
			if !ok {
				// This is not an error, the field simply doesn't have extra description.
				return nil
			}
			field := fields[len(fields)-1]
			if field.Description != "" {
				log.Printf("WARNING: Overwriting existing description for %q: %q",
					field.Name, field.Description)
			}
			field.Description = d["Description"]
			return nil
		},
	)
	return nil
}

// WalkSchema visits every FieldSchema object in the given schema by calling the visit function.
// The prefix is a path of field names from the top level to the current Field.
func WalkSchema(schema bigquery.Schema, visit func(fields []*bigquery.FieldSchema) error) error {
	return walkSchema([]*bigquery.FieldSchema{}, schema, visit)
}

func walkSchema(
	prefix []*bigquery.FieldSchema, schema bigquery.Schema,
	visit func(fields []*bigquery.FieldSchema) error) error {
	fields := ([]*bigquery.FieldSchema)(schema)
	for _, field := range fields {
		path := append(prefix, field)
		err := visit(path)
		if err != nil {
			return err
		}
		if field.Type == bigquery.RecordFieldType {
			err := walkSchema(path, field.Schema, visit)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
