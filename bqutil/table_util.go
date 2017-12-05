package bqutil

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

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
func NewTableUtil(project, dataset string, httpClient *http.Client) (TableUtil, error) {
	ctx := context.Background()
	var bqClient *bigquery.Client
	var err error
	if httpClient != nil {
		opt := option.WithHTTPClient(httpClient)
		bqClient, err = bigquery.NewClient(ctx, project, opt)
	} else {
		// Creates a client.
		bqClient, err = bigquery.NewClient(ctx, project)
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

// TableInfo contains the critical stats for a specific table
// or partition.
type TableInfo struct {
	Name             string
	IsPartitioned    bool
	NumBytes         int64
	NumRows          uint64
	CreationTime     time.Time
	LastModifiedTime time.Time
}

// GetInfoMatching finDataset all tables matching table filter.
// and collects the basic stats about each of them.
// Returns slice ordered by decreasing age.
func (util *TableUtil) GetInfoMatching(dataset, filter string) []TableInfo {
	result := make([]TableInfo, 0)
	ctx := context.Background()
	ti := util.Dataset.Tables(ctx)
	var t *bigquery.Table
	var err error
	for t, err = ti.Next(); err == nil; t, err = ti.Next() {
		// TODO should this be starts with?  Or a regex?
		if strings.Contains(t.TableID, filter) {
			meta, err := t.Metadata(ctx)
			if err != nil {
				log.Println(err)
			} else {
				if meta.Type != bigquery.RegularTable {
					continue
				}
				ts := TableInfo{
					Name:             t.TableID,
					IsPartitioned:    meta.TimePartitioning != nil,
					NumBytes:         meta.NumBytes,
					NumRows:          meta.NumRows,
					CreationTime:     meta.CreationTime,
					LastModifiedTime: meta.LastModifiedTime,
				}
				log.Println(t.TableID, " : ", meta.Name, " : ", meta.LastModifiedTime)
				result = append(result, ts)
			}
		}
	}
	if err != nil {
		log.Println(err)
	}
	sort.Slice(result[:], func(i, j int) bool {
		return result[i].LastModifiedTime.Before(result[j].LastModifiedTime)
	})
	return result
}

// DestinationQuery constructs a query with common Config settings for
// writing results to a table.
// Generally, may need to change WriteDisposition.
func (util *TableUtil) DestinationQuery(query string, dest *bigquery.Table) *bigquery.Query {
	q := util.bqClient.Query(query)
	if dest != nil {
		q.QueryConfig.Dst = dest
	} else {
		q.QueryConfig.DryRun = true
	}
	q.QueryConfig.AllowLargeResults = true
	// Default for unqualified table names in the query.
	q.QueryConfig.DefaultProjectID = util.Dataset.ProjectID
	q.QueryConfig.DefaultDatasetID = util.Dataset.DatasetID
	q.QueryConfig.DisableFlattenedResults = true
	return q
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

///////////////////////////////////////////////////////////////////
// Specific queries.
///////////////////////////////////////////////////////////////////

// TODO - really should take the one that was parsed last, instead
// of random
var dedupTemplate = "" +
	"#standardSQL\n" +
	"# Delete all duplicate rows based on test_id\n" +
	"SELECT * except (row_number)\n" +
	"FROM (\n" +
	"  SELECT *, ROW_NUMBER() OVER (PARTITION BY test_id) row_number\n" +
	"  FROM `%s`)\n" +
	"WHERE row_number = 1\n"

// Dedup executes a query that dedups and writes to an appropriate
// partition.
func (util *TableUtil) Dedup(src string, overwrite bool, DatasettProject, DatasettDataset, DatasettTable string) {
	queryString := fmt.Sprintf(dedupTemplate, src)
	Datasett := util.bqClient.DatasetInProject(DatasettProject, DatasettDataset).Table(DatasettTable)
	q := util.DestinationQuery(queryString, Datasett)
	if overwrite {
		q.QueryConfig.WriteDisposition = bigquery.WriteTruncate
	}
	job, err := q.Run(context.Background())
	if err != nil {
		// TODO add metric.
		log.Println(err)
	}
	log.Println(job.ID())
	log.Println(job.LastStatus())
	status, err := job.Wait(context.Background())
	if err != nil {
		log.Println(err)
	} else {
		log.Println(status)
		if status.Done() {
			log.Println("Done")
			log.Printf("%+v\n", *status.Statistics)
			log.Printf("%+v\n", status.Statistics.Details)
		}
	}
}

// TODO - really should take the one that was parsed last, instead
// of random
var dedupInPlace = "" +
	"# Delete all duplicate rows based on test_id\n" +
	"DELETE\n" +
	"  `%s` copy\n" +
	"WHERE\n" +
	"  CONCAT(copy.test_id, CAST(copy.parse_time AS string)) IN (\n" +
	"  SELECT\n" +
	"    CONCAT(test_id, CAST(parse_time AS string))\n" +
	"  FROM (\n" +
	"    SELECT\n" +
	"      test_id,\n" +
	"      parse_time,\n" +
	"      ROW_NUMBER() OVER (PARTITION BY test_id) row_number\n" +
	"    FROM\n" +
	"      `%s`)\n" +
	"  WHERE\n" +
	"    row_number > 1 )"

// DedupInPlace executes a query that dedups a table.
// TODO interpret and return status.
func (util *TableUtil) DedupInPlace(src string) {
	queryString := fmt.Sprintf(dedupInPlace, src, src)
	q := util.ResultQuery(queryString, false)
	job, err := q.Run(context.Background())
	if err != nil {
		// TODO add metric.
		log.Println(err)
	}
	log.Println(job.ID())
	log.Println(job.LastStatus())
	status, err := job.Wait(context.Background())
	if err != nil {
		log.Println(err)
	} else {
		log.Println(status)
		if status.Done() {
			log.Println("Done")
			log.Printf("%+v\n", *status.Statistics)
			log.Printf("%+v\n", status.Statistics.Details)
		}
	}
}
