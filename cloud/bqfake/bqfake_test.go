package bqfake_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/bigquery/bqiface"
	"github.com/m-lab/go/cloud/bqfake"
	"google.golang.org/api/iterator"
)

func init() {
	// Always prepend the filename and line number.
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// This fails to compile if Dataset does not satisfy the interface.
func assertDataset(ds bqfake.Dataset) {
	func(cc bqiface.Dataset) {}(ds)
}

// This fails to compile if Client does not satisfy the interface.
func assertClient(c bqfake.Client) {
	func(cc bqiface.Client) {}(c)
}

func TestDryRunClient(t *testing.T) {
	c, ct := bqfake.DryRunClient()

	r, err := c.Get("http://foobar")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		t.Fatal("wrong status code", r.StatusCode)
	}
	if ct.Count() != 1 {
		t.Error("Didn't see the Get")
	}
	if len(ct.Requests()) != 1 {
		t.Error("Wrong number of requests")
	}
}

func TestNewClientErr(t *testing.T) {
	ctx := context.Background()
	// These options are incompatible with one another and generate an error from bigquery.NewClient.
	opts := []option.ClientOption{option.WithAPIKey("asdf"), option.WithoutAuthentication()}
	c, err := bqfake.NewClient(ctx, "fakeProject", opts...)
	if err == nil {
		c.Close()
		t.Fatal("Should return constructing client error")
	}
}

func TestBadDataset(t *testing.T) {
	ds := bqfake.Dataset{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	ds.Table("Foobar")
}

func TestUninitializedTable(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}
	ds := c.Dataset("fakeDataset")

	tbl := ds.Table("DedupTest")
	meta, err := tbl.Metadata(ctx)
	if err == nil {
		t.Error("should return an error")
	} else if err.Error() != "Error 404: Not found: Table fakeProject:fakeDataset.DedupTest, notFound" {
		t.Error("wrong error", err)
	}
	if meta != nil {
		t.Error("meta should be nil")
	}

	// Improperly initialized Table
	tbl = bqfake.Table{}
	meta, err = tbl.Metadata(ctx)
	if err == nil {
		t.Error("Should return an initialization error")
	} else if err.Error() != "Table object incorrectly initialized" {
		t.Error("Incorrect error:", err)
	}

	meta = &bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
	err = tbl.Create(ctx, meta)
	if err == nil {
		t.Error("Should return an initialization error")
	} else if err.Error() != "Table object incorrectly initialized" {
		t.Error("Incorrect error:", err)
	}
}

func TestTable(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}
	ds := c.Dataset("etl")

	{
		meta := bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
		meta.TimePartitioning = &bigquery.TimePartitioning{Expiration: 0 * time.Second}
		tbl := ds.Table("DedupTest")
		err := tbl.Create(ctx, &meta)
		if err != nil {
			t.Fatal(err)
		}
	}

	tbl := ds.Table("DedupTest")
	fqn := tbl.FullyQualifiedName()
	if fqn != "fakeProject:etl.DedupTest" {
		t.Error("Got", fqn)
	}

	if tbl.TableID() != "DedupTest" {
		t.Error("Expected TableID() = DedupTest, got", tbl.TableID())
	}
}

func createTable(ctx context.Context, ds bqiface.Dataset, name string) error {
	meta := bigquery.TableMetadata{CreationTime: time.Now(), LastModifiedTime: time.Now(), NumBytes: 168, NumRows: 8}
	meta.TimePartitioning = &bigquery.TimePartitioning{Expiration: 0 * time.Second}
	tbl := ds.Table("DedupTest")
	return tbl.Create(ctx, &meta)
}

func TestTableMetadata(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}
	ds := c.Dataset("etl")

	// TODO: Test whether changes to table can be seen in existing table objects.
	err = createTable(ctx, ds, "DedupTest")
	if err != nil {
		t.Fatal(err)
	}

	tbl := ds.Table("DedupTest")
	log.Println(tbl)
	meta, err := tbl.Metadata(ctx)
	if err != nil {
		t.Error(err)
	} else if meta == nil {
		t.Error("Meta should not be nil")
	}

	// Try to create the table again - should get an error
	err = tbl.Create(ctx, &bigquery.TableMetadata{})
	if err == nil {
		t.Fatal("Should throw error")
	} else {
		log.Println(err)
	}
}

func TestQuery(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "fakeProject")
	if err != nil {
		t.Fatal(err)
	}

	q := c.Query("foobar")
	q.SetQueryConfig(bqiface.QueryConfig{})
	j, err := q.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}
	it, err := q.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = it.Next(struct{}{})
	if err != iterator.Done {
		t.Fatal("Expected iterator.Done, got", err)
	}
}

func TestNewQueryReadClient(t *testing.T) {
	tests := []struct {
		name    string
		config  bqfake.QueryConfig
		wantErr bool
	}{
		{
			name: "success",
			config: bqfake.QueryConfig{
				RowIteratorConfig: bqfake.RowIteratorConfig{
					Rows: []map[string]bigquery.Value{
						{"okay": 1.234},
					},
				},
			},
		},
		{
			name: "read-error",
			config: bqfake.QueryConfig{
				ReadErr: fmt.Errorf("Fake read error"),
			},
			wantErr: true,
		},
		{
			name: "iter-error",
			config: bqfake.QueryConfig{
				RowIteratorConfig: bqfake.RowIteratorConfig{
					IterErr: fmt.Errorf("Fake iter error"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			c := bqfake.NewQueryReadClient(tt.config)
			q := c.Query("SELECT 'fake-query-string'")
			it, err := q.Read(context.Background())
			if err != nil && !tt.wantErr {
				t.Errorf("Query().Read() error = %v", err)
			}
			if it == nil {
				return
			}
			var row map[string]bigquery.Value
			i := 0
			for err = it.Next(&row); err == nil; err = it.Next(&row) {
				if len(tt.config.RowIteratorConfig.Rows) > 0 &&
					!reflect.DeepEqual(row, tt.config.RowIteratorConfig.Rows[i]) {
					t.Errorf("UpdateSchemaDescription() schema mismatch; got %#v, want %#v",
						row, tt.config.RowIteratorConfig.Rows[i])
				}
				i++
			}
			if err == iterator.Done {
				return
			}
			// err != nil.
			if err != tt.config.RowIteratorConfig.IterErr {
				t.Errorf("Next() error; got %v, want %v", err, tt.config.RowIteratorConfig.IterErr)
				return
			}
		})
	}
}
