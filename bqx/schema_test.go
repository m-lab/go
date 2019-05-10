package bqx_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/m-lab/go/bqx"
	"github.com/m-lab/go/rtx"

	"cloud.google.com/go/bigquery"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Embedded struct {
	EmbeddedA int32 // These will be required
	EmbeddedB int32
}

type inner struct {
	Integer   int32
	ByteSlice []byte   // byte slices become BigQuery BYTE type.
	ByteArray [24]byte // byte arrays are repeated integers.  Very inefficient.
	String    string
}

type outer struct {
	Embedded     // EmbeddedA and EmbeddedB will appear as top level fields.
	Inner        inner
	Timestamp    time.Time
	IntTimestamp int64 // `bigquery:"-"`
}

func expect(t *testing.T, sch bigquery.Schema, str string, count int) {
	if j1, err := json.Marshal(sch); err != nil || strings.Count(string(j1), str) != count {
		if err != nil {
			log.Fatal(err)
		}
		_, _, line, _ := runtime.Caller(1)
		pp, _ := bqx.PrettyPrint(sch, false)
		t.Errorf("line %d: %s got %d, wanted %d\n%s", line, str, strings.Count(string(j1), str), count, pp)
	}
}

func TestRemoveRequired(t *testing.T) {
	s, err := bigquery.InferSchema(outer{})
	rtx.Must(err, "")

	expect(t, s, `"Required":true`, 8)
	expect(t, s, `"Repeated":true`, 1) // From the ByteArray

	c := bqx.RemoveRequired(s)
	expect(t, c, `"Required":true`, 1)
}

func TestCustomize(t *testing.T) {
	s, err := bigquery.InferSchema(outer{})
	rtx.Must(err, "")

	subs := map[string]bigquery.FieldSchema{
		"ByteArray":    bigquery.FieldSchema{Name: "ByteArray", Description: "", Repeated: false, Required: true, Type: "INTEGER"},
		"IntTimestamp": bigquery.FieldSchema{Name: "IntTimestamp", Description: "", Repeated: false, Required: true, Type: "TIMESTAMP"},
	}
	c := bqx.Customize(s, subs) // Substitute integer for ByteSlice
	expect(t, c, `"Required":true`, 9)
	expect(t, c, `"Repeated":true`, 0) // because we replaced the ByteArray
	expect(t, c, `"BYTES"`, 1)
	expect(t, c, `"RECORD"`, 1)
}

func TestPrettyPrint(t *testing.T) {
	expected :=
		`[
  {"Name": "EmbeddedA", "Description": "", "Required": true, "Type": "INTEGER"},
  {"Name": "EmbeddedB", "Description": "", "Required": true, "Type": "INTEGER"},
  {"Name": "Inner", "Description": "", "Required": true, "Type": "RECORD", "Schema": [
      {"Name": "Integer", "Description": "", "Required": true, "Type": "INTEGER"},
      {"Name": "ByteSlice", "Description": "", "Required": true, "Type": "BYTES"},
      {"Name": "ByteArray", "Description": "", "Repeated": true, "Type": "INTEGER"},
      {"Name": "String", "Description": "", "Required": true, "Type": "STRING"}
    ]},
  {"Name": "Timestamp", "Description": "", "Required": true, "Type": "TIMESTAMP"},
  {"Name": "IntTimestamp", "Description": "", "Required": true, "Type": "INTEGER"}
]
`

	s, err := bigquery.InferSchema(outer{})
	rtx.Must(err, "")

	pp, err := bqx.PrettyPrint(s, true)
	rtx.Must(err, "")

	if pp != expected {
		t.Error("Pretty print lines don't match")
		ppLines := strings.Split(pp, "\n")
		expLines := strings.Split(expected, "\n")
		if len(ppLines) != len(expLines) {
			t.Error(len(ppLines), len(expLines))
		}
		for i := range ppLines {
			if ppLines[i] != expLines[i] {
				fmt.Printf("%d expected: %s, got: %s\n", i, expLines[i], ppLines[i])
			}
		}
	}
}

func deleteDatasetAndContents(ctx context.Context, table string) error {
	pdt, err := bqx.ParsePDTForTest(table)
	if err != nil {
		return err
	}
	client, err := bigquery.NewClient(ctx, pdt.Project)
	if err != nil {
		return err
	}

	ds := client.Dataset(pdt.Dataset)

	return ds.DeleteWithContents(ctx)
}

func randName(prefix string) string {
	return prefix + strconv.FormatInt(rand.Int63(), 36)
}

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that hits bigquery backend")
	}
	schema, err := bigquery.InferSchema(outer{})
	if err != nil {
		log.Fatal(err)
	}

	name := "mlab-testing." + randName("ds") + randName(".tbl")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Update non-existing table
	err = bqx.UpdateTable(ctx, name, schema)
	if err == nil {
		t.Error("Update non-existing table should have failed")
	}

	// Bad field
	err = bqx.CreateTable(ctx, name, schema, "description",
		&bigquery.TimePartitioning{Field: "NonExistentField"}, nil)
	if err == nil {
		t.Error("Should have failed", name)
		t.Fatal("")
	}

	// Create
	err = bqx.CreateTable(ctx, name, schema, "description",
		&bigquery.TimePartitioning{Field: "Timestamp"}, nil)
	if err != nil {
		t.Error(err)
	}

	// Update
	err = bqx.UpdateTable(ctx, name, schema)
	if err != nil {
		t.Error(err)
	}

	err = deleteDatasetAndContents(ctx, name)
	if err != nil {
		t.Error(err)
	}
}
