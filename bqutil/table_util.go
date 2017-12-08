//  Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package bqutil includes generally useful abstractions for simplifying
// interactions with bigquery.
// Production utilities should go here, but test facilities should go
// in a separate bqtest package.
package bqutil

import (
	"errors"
	"reflect"
	"strings"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// TableExt provides extensions to the bigquery Dataset and Table
// objects to streamline common actions.
// It encapsulates the Client and Dataset to simplify methods.
// TODO(gfr) Should this be called DatasetUtil ?
// TODO(gfr) Rename this to TableExt (extensions)
type TableExt struct {
	BqClient *bigquery.Client
	Dataset  *bigquery.Dataset
}

// NewTableExt creates a TableExt for a project.
// httpClient is used to inject mocks for the bigquery client.
// if httpClient is nil, a suitable default client is used.
// Additional bigquery ClientOptions may be optionally passed as final
//   clientOpts argument.  This is useful for testing credentials.
func NewTableExt(project, dataset string, clientOpts ...option.ClientOption) (TableExt, error) {
	ctx := context.Background()
	var bqClient *bigquery.Client
	var err error
	bqClient, err = bigquery.NewClient(ctx, project, clientOpts...)

	if err != nil {
		return TableExt{}, err
	}

	return TableExt{bqClient, bqClient.Dataset(dataset)}, nil
}

// ResultQuery constructs a query with common QueryConfig settings for
// writing results to a table.
// Generally, may need to change WriteDisposition.
func (util *TableExt) ResultQuery(query string, dryRun bool) *bigquery.Query {
	q := util.BqClient.Query(query)
	q.QueryConfig.DryRun = dryRun
	if strings.HasPrefix(query, "#legacySQL") {
		q.QueryConfig.UseLegacySQL = true
	}
	// Default for unqualified table names in the query.
	q.QueryConfig.DefaultProjectID = util.Dataset.ProjectID
	q.QueryConfig.DefaultDatasetID = util.Dataset.DatasetID
	return q
}

///////////////////////////////////////////////////////////////////
// Code to execute a single query and parse single row result.
///////////////////////////////////////////////////////////////////

// QueryAndParse executes a query that should return a single row, with
// all struct fields that match query columns filled in.
// The caller must pass in the *address* of an appropriate struct.
// TODO - extend this to also handle multirow results, by passing
// slice of structs.
func (util *TableExt) QueryAndParse(q string, structPtr interface{}) error {
	// TODO verify that structPtr is a ptr to struct.
	typeInfo := reflect.ValueOf(structPtr)

	if typeInfo.Type().Kind() != reflect.Ptr {
		return errors.New("Argument should be ptr to struct")
	}
	if reflect.Indirect(typeInfo).Kind() != reflect.Struct {
		return errors.New("Argument should be ptr to struct")
	}

	query := util.ResultQuery(q, false)
	it, err := query.Read(context.Background())
	if err != nil {
		return err
	}

	// We expect a single result row, so proceed accordingly.
	err = it.Next(structPtr)
	if err != nil {
		return err
	}
	var row map[string]bigquery.Value
	// If there are more rows, then something is wrong.
	err = it.Next(&row)
	if err != iterator.Done {
		return errors.New("multiple row data")
	}
	return nil
}
