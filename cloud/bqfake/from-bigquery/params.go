// Copyright 2016 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frombigquery

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"time"

	"cloud.google.com/go/civil"
	"github.com/m-lab/go/internal/fields"

	bq "google.golang.org/api/bigquery/v2"
)

//---------------------------------------------------------------------------------------
// Stuff from params.go
//---------------------------------------------------------------------------------------
var (
	// See https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types#timestamp-type.
	timestampFormat = "2006-01-02 15:04:05.999999-07:00"

	// See https://cloud.google.com/bigquery/docs/reference/rest/v2/tables#schema.fields.name
	validFieldName = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]{0,127}$")
)

const nullableTagOption = "nullable"

func bqTagParser(t reflect.StructTag) (name string, keep bool, other interface{}, err error) {
	name, keep, opts, err := fields.ParseStandardTag("bigquery", t)
	if err != nil {
		return "", false, nil, err
	}
	if name != "" && !validFieldName.MatchString(name) {
		return "", false, nil, errInvalidFieldName
	}
	for _, opt := range opts {
		if opt != nullableTagOption {
			return "", false, nil, fmt.Errorf(
				"bigquery: invalid tag option %q. The only valid option is %q",
				opt, nullableTagOption)
		}
	}
	return name, keep, opts, nil
}

var fieldCache = fields.NewCache(bqTagParser, nil, nil)

var (
	int64ParamType     = &bq.QueryParameterType{Type: "INT64"}
	float64ParamType   = &bq.QueryParameterType{Type: "FLOAT64"}
	boolParamType      = &bq.QueryParameterType{Type: "BOOL"}
	stringParamType    = &bq.QueryParameterType{Type: "STRING"}
	bytesParamType     = &bq.QueryParameterType{Type: "BYTES"}
	dateParamType      = &bq.QueryParameterType{Type: "DATE"}
	timeParamType      = &bq.QueryParameterType{Type: "TIME"}
	dateTimeParamType  = &bq.QueryParameterType{Type: "DATETIME"}
	timestampParamType = &bq.QueryParameterType{Type: "TIMESTAMP"}
	numericParamType   = &bq.QueryParameterType{Type: "NUMERIC"}
)

var (
	typeOfDate     = reflect.TypeOf(civil.Date{})
	typeOfTime     = reflect.TypeOf(civil.Time{})
	typeOfDateTime = reflect.TypeOf(civil.DateTime{})
	typeOfGoTime   = reflect.TypeOf(time.Time{})
	typeOfRat      = reflect.TypeOf(&big.Rat{})
)
