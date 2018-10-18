package frombigquery

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
)

// parseCivilDateTime parses a date-time represented in a BigQuery SQL
// compatible format and returns a civil.DateTime.
func parseCivilDateTime(s string) (civil.DateTime, error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return civil.DateTime{}, fmt.Errorf("bigquery: bad DATETIME value %q", s)
	}
	return civil.ParseDateTime(parts[0] + "T" + parts[1])
}

// convertBasicType returns val as an interface with a concrete type specified by typ.
func convertBasicType(val string, typ bigquery.FieldType) (bigquery.Value, error) {
	switch typ {
	case bigquery.StringFieldType:
		return val, nil
	case bigquery.BytesFieldType:
		return base64.StdEncoding.DecodeString(val)
	case bigquery.IntegerFieldType:
		return strconv.ParseInt(val, 10, 64)
	case bigquery.FloatFieldType:
		return strconv.ParseFloat(val, 64)
	case bigquery.BooleanFieldType:
		return strconv.ParseBool(val)
	case bigquery.TimestampFieldType:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}
		secs := math.Trunc(f)
		nanos := (f - secs) * 1e9
		return bigquery.Value(time.Unix(int64(secs), int64(nanos)).UTC()), nil
	case bigquery.DateFieldType:
		return civil.ParseDate(val)
	case bigquery.TimeFieldType:
		return civil.ParseTime(val)
	case bigquery.DateTimeFieldType:
		return civil.ParseDateTime(val)
	case bigquery.NumericFieldType:
		r, ok := (&big.Rat{}).SetString(val)
		if !ok {
			return nil, fmt.Errorf("bigquery: invalid NUMERIC value %q", val)
		}
		return bigquery.Value(r), nil
	default:
		return nil, fmt.Errorf("unrecognized type: %s", typ)
	}
}
