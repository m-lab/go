package frombigquery

//========================================================================================
// This file contains code pulled from bigquery golang libraries, to emulate the library
// behavior, without hitting the backend.  It also allows examination of the rows that
// are ultimately sent to the service.
//========================================================================================
import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"

	"github.com/GoogleCloudPlatform/google-cloud-go-testing/bigquery/bqiface"

	"cloud.google.com/go/bigquery"
	bqv2 "google.golang.org/api/bigquery/v2"
)

//---------------------------------------------------------------------------------------
// Stuff from uploader.go
//---------------------------------------------------------------------------------------

// This is an fake for Uploader, for use in debugging, and tests.
// See bigquery.Uploader for field info.
type FakeUploader struct {
	bqiface.Uploader

	t                   *bigquery.Table
	SkipInvalidRows     bool
	IgnoreUnknownValues bool
	TableTemplateSuffix string

	Rows    []*InsertionRow // Most recently inserted rows, for testing/debugging.
	Request *bqv2.TableDataInsertAllRequest
	// Set this with SetErr to return an error.  Error is cleared on each call.
	Err       error
	CallCount int // Number of times Put is called.
}

func (u *FakeUploader) SetErr(err error) {
	u.Err = err
}

func NewFakeUploader() *FakeUploader {
	return new(FakeUploader)
}

// Put uploads one or more rows to the BigQuery service.
//
// If src is ValueSaver, then its Save method is called to produce a row for uploading.
//
// If src is a struct or pointer to a struct, then a schema is inferred from it
// and used to create a StructSaver. The InsertID of the StructSaver will be
// empty.
//
// If src is a slice of ValueSavers, structs, or struct pointers, then each
// element of the slice is treated as above, and multiple rows are uploaded.
//
// Put returns a PutMultiError if one or more rows failed to be uploaded.
// The PutMultiError contains a RowInsertionError for each failed row.
//
// Put will retry on temporary errors (see
// https://cloud.google.com/bigquery/troubleshooting-errors). This can result
// in duplicate rows if you do not use insert IDs. Also, if the error persists,
// the call will run indefinitely. Pass a context with a timeout to prevent
// hanging calls.
func (u *FakeUploader) Put(ctx context.Context, src interface{}) error {
	u.CallCount++
	if u.Err != nil {
		t := u.Err
		u.Err = nil
		return t
	}

	savers, err := valueSavers(src)
	if err != nil {
		log.Printf("Put: %v\n", err)
		log.Printf("src: %v\n", src)
		//debug.PrintStack()
		return err
	}
	return u.putMulti(ctx, savers)
}

func valueSavers(src interface{}) ([]bigquery.ValueSaver, error) {
	saver, ok, err := toValueSaver(src)
	if err != nil {
		return nil, err
	}
	if ok {
		return []bigquery.ValueSaver{saver}, nil
	}
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%T is not a ValueSaver, struct, struct pointer, or slice", src)

	}
	var savers []bigquery.ValueSaver
	for i := 0; i < srcVal.Len(); i++ {
		s := srcVal.Index(i).Interface()
		saver, ok, err := toValueSaver(s)
		log.Println(saver)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("src[%d] has type %T, which is not a ValueSaver, struct or struct pointer", i, s)
		}
		savers = append(savers, saver)
	}
	return savers, nil
}

// Make a ValueSaver from x, which must implement ValueSaver already
// or be a struct or pointer to struct.
func toValueSaver(x interface{}) (bigquery.ValueSaver, bool, error) {
	if saver, ok := x.(bigquery.ValueSaver); ok {
		return saver, ok, nil
	}
	v := reflect.ValueOf(x)
	// Support Put with []interface{}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false, nil
	}
	schema, err := inferSchemaReflect(v.Type())
	if err != nil {
		return nil, false, err
	}
	return &bigquery.StructSaver{Struct: x, Schema: schema}, true, nil
}

func (u *FakeUploader) putMulti(ctx context.Context, src []bigquery.ValueSaver) error {
	var rows []*InsertionRow
	for _, saver := range src {
		row, insertID, err := saver.Save()
		if err != nil {
			log.Printf("%v\n", err)
			debug.PrintStack()
			return err
		}
		rows = append(rows, &InsertionRow{InsertID: insertID, Row: row})
	}

	u.Rows = append(u.Rows, rows...)

	// Substitute for service call.
	var err error
	u.Request, err = insertRows(rows)
	return err
}

// An InsertionRow represents a row of data to be inserted into a table.
type InsertionRow struct {
	// If InsertID is non-empty, BigQuery will use it to de-duplicate insertions of
	// this row on a best-effort basis.
	InsertID string
	// The data to be inserted, represented as a map from field name to Value.
	Row map[string]bigquery.Value
}

//---------------------------------------------------------------------------------------
// Stuff from service.go
//---------------------------------------------------------------------------------------
func insertRows(rows []*InsertionRow) (*bqv2.TableDataInsertAllRequest, error) {
	req := &bqv2.TableDataInsertAllRequest{}
	for _, row := range rows {
		m := make(map[string]bqv2.JsonValue)
		for k, v := range row.Row {
			m[k] = bqv2.JsonValue(v)
		}
		req.Rows = append(req.Rows, &bqv2.TableDataInsertAllRequestRows{
			InsertId: row.InsertID,
			Json:     m,
		})
	}
	// Truncated here, because the remainder hits the backend.
	return req, nil
}
