package bqutil

import (
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// TableUtil provides utility functions on tables in a dataset.
// It encapsulates the client and Dataset to simplify methoDataset.
// TODO(gfr) Should this be called DatasetUtil ?
type TableUtil struct {
	bqClient *bigquery.Client
	Dataset  *bigquery.Dataset
}

// NewTableUtil creates a TableUtil for a project.
// httpClient is used to inject mocks for the bigquery client.
// if nil, a suitable default client is used.
// TODO - if travis, need to use a service account, from the
func NewTableUtil(project, dataset string, httpClient *http.Client, clientOpts ...option.ClientOption) (TableUtil, error) {
	ctx := context.Background()
	var bqClient *bigquery.Client
	var err error
	if httpClient != nil {
		opt := option.WithHTTPClient(httpClient)
		bqClient, err = bigquery.NewClient(ctx, project, append(clientOpts, opt)...)
	} else {
		// Creates a client.
		bqClient, err = bigquery.NewClient(ctx, project, clientOpts...)
	}

	if err != nil {
		return TableUtil{}, err
	}

	return TableUtil{bqClient, bqClient.Dataset(dataset)}, nil
}

// GetTableStats fetches the Metadata for a table.
// TODO(gfr) Is this worth having, or is it non-idiomatic?
func (util *TableUtil) GetTableStats(table string) bigquery.TableMetadata {
	t := util.Dataset.Table(table)

	ctx := context.Background()
	meta, err := t.Metadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return *meta
}

// ResultQuery constructs a query with common Config settings for
// writing results to a table.
// Generally, may need to change WriteDisposition.
func (util *TableUtil) ResultQuery(query string, dryRun bool) *bigquery.Query {
	q := util.bqClient.Query(query)
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

// ParseModel will parse a bigquery row map into a new item matching
// model.  Type of model must be struct annotated with qfield tags.
func ParseModel(row map[string]bigquery.Value, model interface{}) (interface{}, error) {
	typeOfModel := reflect.ValueOf(model).Type()

	ptr := reflect.New(typeOfModel).Interface()
	s := reflect.ValueOf(ptr).Elem()

	for i := 0; i < typeOfModel.NumField(); i++ {
		field := s.Field(i)
		tag := typeOfModel.Field(i).Tag.Get("qfield")
		v, ok := row[tag]
		if ok {
			field.Set(reflect.ValueOf(v)) // Will break if types don't match.
		}
	}
	// This is the result we want!!!
	return s.Interface(), nil
}

// QueryAndParse executes a query that should return a single row, with
// column labels matching the qfields tags in the provided model struct.
func (util *TableUtil) QueryAndParse(q string, model interface{}) (interface{}, error) {
	query := util.ResultQuery(q, false)
	it, err := query.Read(context.Background())
	if err != nil {
		log.Print(err)
		return model, err
	}

	// We expect a single result row, so proceed accordingly.
	var row map[string]bigquery.Value
	err = it.Next(&row)
	if err != nil {
		log.Println(err)
		return model, err
	}
	x, err := ParseModel(row, model)
	if err != nil {
		log.Println(err)
		return model, err
	}
	// If there are more rows, then something is wrong.
	err = it.Next(&row)
	if err != iterator.Done {
		return model, errors.New("multiple row data")
	}
	return x, nil
}
