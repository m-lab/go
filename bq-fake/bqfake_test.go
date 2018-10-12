package bqfake_test

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/google-cloud-go-testing/bigquery/bqiface"
	bqfake "github.com/m-lab/go/bq-fake"
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
	defer r.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
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

func TestDataset(t *testing.T) {
	ds := bqfake.Dataset{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	ds.Table("Foobar")
}

func TestTable(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "mlab-testing")
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
	if fqn != "mlab-testing:etl.DedupTest" {
		t.Error("Got", fqn)
	}

	if tbl.TableID() != "DedupTest" {
		t.Error("Expected TableID() = DedupTest, got", tbl.TableID())
	}
}

func TestUninitializedTable(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "mlab-testing")
	if err != nil {
		panic(err)
	}
	ds := c.Dataset("etl")

	tbl := ds.Table("DedupTest")
	meta, err := tbl.Metadata(ctx)
	if err == nil {
		t.Error("Should return an error")
	} else if err.Error() != "Error 404: Not found: Table mlab-testing:etl.DedupTest, notFound" {
		t.Error("Wrong error", err)
	}
	if meta != nil {
		t.Error("Should have nil metadata")
	}

	log.Println(err)
}

func TestTableMetadata(t *testing.T) {
	ctx := context.Background()
	c, err := bqfake.NewClient(ctx, "mlab-testing")
	if err != nil {
		panic(err)
	}
	ds := c.Dataset("etl")

	// TODO - also test table before it exists
	// Test whether changes to table can be seen in existing table objects.

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
	c, err := bqfake.NewClient(ctx, "mlab-testing")
	if err != nil {
		panic(err)
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
