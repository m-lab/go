package frombigquery

import (
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/bigquery"
)

//---------------------------------------------------------------------------------------
// Stuff from schema.go
//---------------------------------------------------------------------------------------

var (
	errNoStruct             = errors.New("bigquery: can only infer schema from struct or pointer to struct")
	errUnsupportedFieldType = errors.New("bigquery: unsupported type of field in struct")
	errInvalidFieldName     = errors.New("bigquery: invalid name of field in struct")
)

var typeOfByteSlice = reflect.TypeOf([]byte{})

var schemaCache Cache

type cacheVal struct {
	schema bigquery.Schema
	err    error
}

func inferSchemaReflectCached(t reflect.Type) (bigquery.Schema, error) {
	cv := schemaCache.Get(t, func() interface{} {
		s, err := inferSchemaReflect(t)
		return cacheVal{s, err}
	}).(cacheVal)
	return cv.schema, cv.err
}

func inferSchemaReflect(t reflect.Type) (bigquery.Schema, error) {
	rec, err := hasRecursiveType(t, nil)
	if err != nil {
		return nil, err
	}
	if rec {
		return nil, fmt.Errorf("bigquery: schema inference for recursive type %s", t)
	}
	return inferStruct(t)
}

func inferStruct(t reflect.Type) (bigquery.Schema, error) {
	switch t.Kind() {
	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return nil, errNoStruct
		}
		t = t.Elem()
		fallthrough

	case reflect.Struct:
		return inferFields(t)
	default:
		return nil, errNoStruct
	}
}

// inferFieldSchema infers the FieldSchema for a Go type
func inferFieldSchema(rt reflect.Type) (*bigquery.FieldSchema, error) {
	switch rt {
	case typeOfByteSlice:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.BytesFieldType}, nil
	case typeOfGoTime:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.TimestampFieldType}, nil
	case typeOfDate:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.DateFieldType}, nil
	case typeOfTime:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.TimeFieldType}, nil
	case typeOfDateTime:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.DateTimeFieldType}, nil
	}
	if isSupportedIntType(rt) {
		return &bigquery.FieldSchema{Required: true, Type: bigquery.IntegerFieldType}, nil
	}
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		et := rt.Elem()
		if et != typeOfByteSlice && (et.Kind() == reflect.Slice || et.Kind() == reflect.Array) {
			// Multi dimensional slices/arrays are not supported by BigQuery
			return nil, errUnsupportedFieldType
		}

		f, err := inferFieldSchema(et)
		if err != nil {
			return nil, err
		}
		f.Repeated = true
		f.Required = false
		return f, nil
	case reflect.Struct, reflect.Ptr:
		nested, err := inferStruct(rt)
		if err != nil {
			return nil, err
		}
		return &bigquery.FieldSchema{Required: true, Type: bigquery.RecordFieldType, Schema: nested}, nil
	case reflect.String:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.StringFieldType}, nil
	case reflect.Bool:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.BooleanFieldType}, nil
	case reflect.Float32, reflect.Float64:
		return &bigquery.FieldSchema{Required: true, Type: bigquery.FloatFieldType}, nil
	default:
		return nil, errUnsupportedFieldType
	}
}

// inferFields extracts all exported field types from struct type.
func inferFields(rt reflect.Type) (bigquery.Schema, error) {
	var s bigquery.Schema
	fields, err := fieldCache.Fields(rt)
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		f, err := inferFieldSchema(field.Type)
		if err != nil {
			return nil, err
		}
		f.Name = field.Name
		s = append(s, f)
	}
	return s, nil
}

// isSupportedIntType reports whether t can be properly represented by the
// BigQuery INTEGER/INT64 type.
func isSupportedIntType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return true
	default:
		return false
	}
}

// typeList is a linked list of reflect.Types.
type typeList struct {
	t    reflect.Type
	next *typeList
}

func (l *typeList) has(t reflect.Type) bool {
	for l != nil {
		if l.t == t {
			return true
		}
		l = l.next
	}
	return false
}

// hasRecursiveType reports whether t or any type inside t refers to itself, directly or indirectly,
// via exported fields. (Schema inference ignores unexported fields.)
func hasRecursiveType(t reflect.Type, seen *typeList) (bool, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false, nil
	}
	if seen.has(t) {
		return true, nil
	}
	fields, err := fieldCache.Fields(t)
	if err != nil {
		return false, err
	}
	seen = &typeList{t, seen}
	// Because seen is a linked list, additions to it from one field's
	// recursive call will not affect the value for subsequent fields' calls.
	for _, field := range fields {
		ok, err := hasRecursiveType(field.Type, seen)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
